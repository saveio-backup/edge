/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package websocket privides websocket server handler
package websocket

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/saveio/edge/common/config"
	Err "github.com/saveio/edge/http/base/error"
	"github.com/saveio/edge/http/rest"

	"github.com/saveio/edge/http/websocket/session"
	"github.com/saveio/themis/common/log"
)

const (
	WS_TOPIC_EVENT = 1
)

const (
	WS_ACTION_SUBSCRIBE = "subscribe"
)

type handler func(map[string]interface{}) map[string]interface{}
type Handler struct {
	handler  handler
	pushFlag bool
}

//subscribe event for client
type subscribe struct {
}
type WsServer struct {
	sync.RWMutex
	Upgrader     websocket.Upgrader
	listener     net.Listener
	server       *http.Server
	SessionList  *session.SessionList // websocket sessionlist
	ActionMap    map[string]Handler   //handler functions
	TxHashMap    map[string]string    //key: txHash   value:sessionId
	SubscribeMap map[string]subscribe //key: sessionId   value:subscribeInfo
}

//init websocket server
func InitWsServer() *WsServer {
	ws := &WsServer{
		Upgrader:     websocket.Upgrader{},
		SessionList:  session.NewSessionList(),
		TxHashMap:    make(map[string]string),
		SubscribeMap: make(map[string]subscribe),
	}
	return ws
}

//start websocket server
func (self *WsServer) Start() error {
	wsPort := int(config.Parameters.BaseConfig.PortBase) + config.Parameters.BaseConfig.WsPortOffset
	if wsPort == 0 {
		log.Error("Not configure HttpWsPort port ")
		return nil
	}
	self.registryMethod()
	self.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	log.Infof("test tls")
	tlsFlag := false
	if tlsFlag || wsPort%1000 == rest.TLS_PORT {
		var err error
		self.listener, err = self.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		self.listener, err = net.Listen("tcp", ":"+strconv.Itoa(wsPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	var done = make(chan bool)
	go self.checkSessionsTimeout(done)
	log.Infof("Start websocket at %d", wsPort)
	self.server = &http.Server{Handler: http.HandlerFunc(self.webSocketHandler)}
	err := self.server.Serve(self.listener)

	done <- true
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}
	return nil

}

func (self *WsServer) getSmartContractEvent() {
}

//registry handler method
func (self *WsServer) registryMethod() {

	heartbeat := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		self.Lock()
		defer self.Unlock()

		sessionId, _ := cmd["SessionId"].(string)
		sub := self.SubscribeMap[sessionId]
		resp["Action"] = "heartbeat"
		resp["Result"] = sub
		return resp
	}
	subscribe := func(cmd map[string]interface{}) map[string]interface{} {
		log.Debugf("subsrice from %v", cmd)
		resp := rest.ResponsePack(Err.SUCCESS)
		self.Lock()
		defer self.Unlock()

		sessionId, _ := cmd["SessionId"].(string)
		sub := self.SubscribeMap[sessionId]
		self.SubscribeMap[sessionId] = sub

		resp["Action"] = WS_ACTION_SUBSCRIBE
		resp["Result"] = sub
		return resp
	}
	getsessioncount := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Action"] = "getsessioncount"
		resp["Result"] = self.SessionList.GetSessionCount()
		return resp
	}
	actionMap := map[string]Handler{
		"heartbeat":         {handler: heartbeat},
		WS_ACTION_SUBSCRIBE: {handler: subscribe},
		"getsessioncount":   {handler: getsessioncount},
	}
	self.ActionMap = actionMap
}

func (self *WsServer) Stop() {
	if self.server != nil {
		self.server.Shutdown(context.Background())
		log.Infof("Close websocket ")
	}
}
func (self *WsServer) Restart() {
	go func() {
		time.Sleep(time.Second)
		self.Stop()
		time.Sleep(time.Second)
		go self.Start()
	}()
}

//check sessions timeout,if expire close the session
func (self *WsServer) checkSessionsTimeout(done chan bool) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var closeList []*session.Session
			self.SessionList.ForEachSession(func(v *session.Session) {
				if v.SessionTimeoverCheck() {
					resp := rest.ResponsePack(Err.SESSION_EXPIRED)
					v.Send(marshalResp(resp))
					closeList = append(closeList, v)
				}
			})
			for _, s := range closeList {
				self.SessionList.CloseSession(s)
			}

		case <-done:
			return
		}
	}
}

func (self *WsServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := self.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("websocket Upgrader: ", err)
		return
	}
	defer wsConn.Close()
	nsSession, err := self.SessionList.NewSession(wsConn)
	if err != nil {
		log.Error("websocket NewSession:", err)
		return
	}

	defer func() {
		self.deleteSubscribe(nsSession.GetSessionId())
		self.SessionList.CloseSession(nsSession)
		if err := recover(); err != nil {
			log.Fatal("websocket recover:", err)
		}
	}()

	log.Debugf("new session %v", nsSession.GetSessionId())
	for {
		_, bysMsg, err := wsConn.ReadMessage()
		if err == nil {
			log.Debugf("Read msg %s", bysMsg)
			if self.OnDataHandle(nsSession, bysMsg, r) {
				nsSession.UpdateActiveTime()
			}
			continue
		}
		e, ok := err.(net.Error)
		if !ok || !e.Timeout() {
			log.Infof("websocket conn:", err)
			return
		}
	}
}
func (self *WsServer) IsValidMsg(reqMsg map[string]interface{}) bool {
	if _, ok := reqMsg["Hash"].(string); !ok && reqMsg["Hash"] != nil {
		return false
	}
	if _, ok := reqMsg["Addr"].(string); !ok && reqMsg["Addr"] != nil {
		return false
	}
	if _, ok := reqMsg["AssetId"].(string); !ok && reqMsg["AssetId"] != nil {
		return false
	}
	return true
}

func (self *WsServer) OnDataHandle(curSession *session.Session, bysMsg []byte, r *http.Request) bool {

	var req = make(map[string]interface{})

	if err := json.Unmarshal(bysMsg, &req); err != nil {
		resp := rest.ResponsePack(Err.ILLEGAL_DATAFORMAT)
		curSession.Send(marshalResp(resp))
		log.Infof("websocket OnDataHandle:", err)
		return false
	}
	actionName, ok := req["Action"].(string)
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		curSession.Send(marshalResp(resp))
		return false
	}
	action, ok := self.ActionMap[actionName]
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		curSession.Send(marshalResp(resp))
		return false
	}
	if !self.IsValidMsg(req) {
		resp := rest.ResponsePack(Err.INVALID_PARAMS)
		curSession.Send(marshalResp(resp))
		return true
	}
	if height, ok := req["Height"].(float64); ok {
		req["Height"] = strconv.FormatInt(int64(height), 10)
	}
	if raw, ok := req["Raw"].(float64); ok {
		req["Raw"] = strconv.FormatInt(int64(raw), 10)
	}
	req["SessionId"] = curSession.GetSessionId()
	resp := action.handler(req)
	resp["Action"] = actionName
	resp["Id"] = req["Id"]

	curSession.Send(marshalResp(resp))

	switch actionName {
	case WS_ACTION_SUBSCRIBE:
		self.PushToNewSubscriber()
	}

	return true
}

func (self *WsServer) deleteSubscribe(sessionId string) {
	self.Lock()
	defer self.Unlock()
	delete(self.SubscribeMap, sessionId)
}

func marshalResp(resp map[string]interface{}) []byte {
	resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Infof("Websocket marshal json error:", err)
		return nil
	}

	return data
}

func (self *WsServer) Broadcast(sub int, resp map[string]interface{}) {
	self.Lock()
	defer self.Unlock()
	data, err := json.Marshal(resp)
	if err != nil {
		log.Infof("Websocket marshal json error:", err)
		return
	}
	for sid, _ := range self.SubscribeMap {
		s := self.SessionList.GetSessionById(sid)
		if s == nil {
			continue
		}
		s.Send(data)
	}
}

func (self *WsServer) initTlsListen() (net.Listener, error) {

	certPath := config.Parameters.BaseConfig.WsCertPath
	keyPath := config.Parameters.BaseConfig.WsKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	wsPort := strconv.FormatInt(int64(config.Parameters.BaseConfig.PortBase)+int64(config.Parameters.BaseConfig.WsPortOffset), 10)
	log.Info("TLS listen port is ", wsPort)
	listener, err := tls.Listen("tcp", ":"+wsPort, tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
