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

// Package websocket privides a function to start websocket server
package websocket

import (
	"github.com/saveio/edge/http/websocket/websocket"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/types"
)

var ws *websocket.WsServer

func Server() *websocket.WsServer {
	return ws
}

func StartServer() {
	go func() {
		ws = websocket.InitWsServer()
		ws.Start()
	}()
}

func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer()
		ws.Start()
		return
	}
	ws.Restart()
}

func pushSmartCodeEvent(v interface{}) {
	if ws == nil {
		return
	}
	rs, ok := v.(types.SmartCodeEvent)
	if !ok {
		log.Errorf("[PushSmartCodeEvent]", "SmartCodeEvent err")
		return
	}
	log.Debugf("rs %v", rs)
	panic("123")
	// go func() {
	// 	switch object := rs.Result.(type) {
	// 	case *event.LogEventArgs:
	// 	case *event.ExecuteNotify:
	// 	default:
	// 	}
	// }()
}
