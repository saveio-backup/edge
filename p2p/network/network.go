package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/carrier/network"
	"github.com/saveio/carrier/network/components/ackreply"
	"github.com/saveio/carrier/network/components/backoff"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	dspCom "github.com/saveio/dsp-go-sdk/common"
	dspNetCom "github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	dspMsg "github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/pylons/actor/msg_opcode"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/pylons/network/transport/messages"
	"github.com/saveio/themis/common/log"
)

var once sync.Once

type Network struct {
	P2p                   *network.Network
	peerAddrs             []string
	listenAddr            string
	proxyAddr             string
	pid                   *actor.PID
	protocol              string
	address               string
	mappingAddress        string
	Keys                  *crypto.KeyPair
	keepaliveInterval     time.Duration
	keepaliveTimeout      time.Duration
	kill                  chan struct{}
	addressForHealthCheck *sync.Map
	addressForConnecting  *sync.Map
	handler               func(*network.ComponentContext)
	peerFailedCount       *sync.Map
}

func NewP2P() *Network {
	n := &Network{
		P2p: new(network.Network),
	}
	n.addressForHealthCheck = new(sync.Map)
	n.addressForConnecting = new(sync.Map)
	n.kill = make(chan struct{})
	n.peerFailedCount = new(sync.Map)
	return n

}

func (this *Network) SetProxyServer(address string) {
	this.proxyAddr = address
}

func (this *Network) GetProxyServer() string {
	return this.proxyAddr
}

func (this *Network) SetNetworkKey(key *crypto.KeyPair) {
	this.Keys = key
}

func (this *Network) SetHandler(handler func(*network.ComponentContext)) {
	this.handler = handler
}

func (this *Network) Protocol() string {
	return getProtocolFromAddr(this.PublicAddr())
}

func (this *Network) ListenAddr() string {
	return this.listenAddr
}

func (this *Network) PublicAddr() string {
	return this.P2p.ID.Address
}

func (this *Network) GetPeersIfExist() error {
	this.P2p.EachPeer(func(client *network.PeerClient) bool {
		this.peerAddrs = append(this.peerAddrs, client.Address)
		return true
	})
	return nil
}

// SetPID sets p2p actor
func (this *Network) SetPID(pid *actor.PID) {
	this.pid = pid
}

// GetPID returns p2p actor
func (this *Network) GetPID() *actor.PID {
	return this.pid
}

func (this *Network) Start(address string) error {
	this.protocol = getProtocolFromAddr(address)
	log.Debugf("Network protocol %s", this.protocol)
	builderOpt := []network.BuilderOption{
		network.WriteFlushLatency(1 * time.Millisecond),
		network.WriteTimeout(dspCom.NETWORK_STREAM_WRITE_TIMEOUT),
	}
	builder := network.NewBuilderWithOptions(builderOpt...)

	// set keys
	if this.Keys != nil {
		log.Debugf("Network use account key")
		builder.SetKeys(this.Keys)
	} else {
		log.Debugf("Network use RandomKeyPair key")
		builder.SetKeys(ed25519.RandomKeyPair())
	}

	builder.SetAddress(address)
	// add msg receiver
	component := new(NetComponent)
	component.Net = this
	builder.AddComponent(component)

	// add keepalive
	if this.keepaliveInterval == 0 {
		this.keepaliveInterval = keepalive.DefaultKeepaliveInterval
	}
	if this.keepaliveTimeout == 0 {
		this.keepaliveTimeout = keepalive.DefaultKeepaliveTimeout
	}
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(this.keepaliveInterval),
		keepalive.WithKeepaliveTimeout(this.keepaliveTimeout),
	}
	log.Debugf("keepalive interval: %d, timeout: %d", this.keepaliveInterval, this.keepaliveTimeout)
	builder.AddComponent(keepalive.New(options...))

	// add backoff
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 {
		backoff_options := []backoff.ComponentOption{
			backoff.WithInitialDelay(common.BACKOFF_INIT_DELAY * time.Second),
			backoff.WithMaxAttempts(common.BACKOFF_MAX_ATTEMPTS),
			backoff.WithPriority(65535),
			backoff.WithMaxInterval(time.Duration(300) * time.Second),
		}
		log.Debugf("backoff opt %v", backoff_options)
		builder.AddComponent(backoff.New(backoff_options...))
	}

	// add ack reply
	ackOption := []ackreply.ComponentOption{
		ackreply.WithAckCheckedInterval(time.Second * 3),
		ackreply.WithAckMessageTimeout(time.Second * 10),
	}
	builder.AddComponent(ackreply.New(ackOption...))

	this.AddProxyComponents(builder)

	var err error
	this.P2p, err = builder.Build()
	if err != nil {
		log.Error("[P2pNetwork] Start builder.Build error: ", err.Error())
		return err
	}
	once.Do(func() {
		for k, v := range msg_opcode.OpCodes {
			err := opcode.RegisterMessageType(k, v)
			if err != nil {
				log.Errorf("register messages failed %v", k)
			}
		}
		opcode.RegisterMessageType(opcode.Opcode(dspNetCom.MSG_OP_CODE), &pb.Message{})
	})
	this.P2p.SetDialTimeout(time.Duration(common.NETWORK_DIAL_TIMEOUT) * time.Second)
	this.P2p.SetCompressFileSize(common.COMPRESS_DATA_SIZE)
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 {
		this.P2p.EnableProxyMode(true)
		this.P2p.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddrs)
	}

	this.P2p.SetNetworkID(config.Parameters.BaseConfig.NetworkId)
	go this.P2p.Listen()
	this.P2p.BlockUntilListening()
	err = this.StartProxy(builder)
	if err != nil {
		log.Errorf("start proxy failed, err: %s", err)
	}
	if len(this.P2p.ID.Address) == 6 {
		return errors.New("invalid address")
	}
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 {
		go this.healthCheckProxyService()
	}
	return nil
}

func (this *Network) AddProxyComponents(builder *network.Builder) {
	servers := strings.Split(config.Parameters.BaseConfig.NATProxyServerAddrs, ",")
	hasAdd := make(map[string]struct{})
	for _, proxyAddr := range servers {
		if len(proxyAddr) == 0 {
			continue
		}
		protocol := getProtocolFromAddr(proxyAddr)
		_, ok := hasAdd[protocol]
		if ok {
			continue
		}
		switch protocol {
		case "udp":
			builder.AddComponent(new(proxy.UDPProxyComponent))
		case "kcp":
			builder.AddComponent(new(proxy.KCPProxyComponent))
		case "quic":
			builder.AddComponent(new(proxy.QuicProxyComponent))
		case "tcp":
			builder.AddComponent(new(proxy.TcpProxyComponent))
		}
		hasAdd[protocol] = struct{}{}
	}
}

func (this *Network) StartProxy(builder *network.Builder) error {
	var err error
	log.Debugf("NATProxyServerAddrs :%v", config.Parameters.BaseConfig.NATProxyServerAddrs)
	servers := strings.Split(config.Parameters.BaseConfig.NATProxyServerAddrs, ",")
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 && len(servers) == 0 {
		return fmt.Errorf("invalid proxy config: %s", config.Parameters.BaseConfig.NATProxyServerAddrs)
	}
	for _, proxyAddr := range servers {
		if len(proxyAddr) == 0 {
			continue
		}
		this.P2p.EnableProxyMode(true)
		this.P2p.SetProxyServer(proxyAddr)
		protocol := getProtocolFromAddr(proxyAddr)
		log.Debugf("start proxy will blocking...%s %s, networkId: %d", protocol, proxyAddr, config.Parameters.BaseConfig.NetworkId)
		done := make(chan struct{}, 1)
		go func() {
			switch protocol {
			case "udp":
				this.P2p.BlockUntilUDPProxyFinish()
			case "kcp":
				this.P2p.BlockUntilKCPProxyFinish()
			case "quic":
				this.P2p.BlockUntilQuicProxyFinish()
			case "tcp":
				this.P2p.BlockUntilTcpProxyFinish()
			}
			done <- struct{}{}
		}()
		select {
		case <-done:
			this.proxyAddr = proxyAddr
			log.Debugf("start proxy finish, publicAddr: %s", this.P2p.ID.Address)
			return nil
		case <-time.After(time.Duration(common.START_PROXY_TIMEOUT) * time.Second):
			err = fmt.Errorf("proxy: %s timeout", proxyAddr)
			log.Debugf("start proxy err :%s", err)
			break
		}
	}
	return err
}

func (this *Network) Stop() {
	close(this.kill)
	this.P2p.Close()
}

func (this *Network) Connect(tAddr string) error {
	log.Debugf("Connect %s", tAddr)
	if this == nil {
		return fmt.Errorf("[Connect] this is nil")
	}
	peerState, _ := this.GetPeerStateByAddress(tAddr)
	if peerState == network.PEER_REACHABLE {
		return nil
	}
	if _, ok := this.addressForHealthCheck.Load(tAddr); ok {
		// already try to connect, don't retry before we get a result
		log.Info("already try to connect %s", tAddr)
		return nil
	}

	this.addressForHealthCheck.Store(tAddr, struct{}{})
	this.P2p.Bootstrap(tAddr)
	return nil
}

func (this *Network) ConnectAndWait(addr string) error {
	if _, ok := this.addressForConnecting.Load(addr); ok {
		// already try to connect, don't retry before we get a result
		log.Info("already try to connect")
		err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
		this.AddDialFailedCnt(addr)
		return err
	}
	this.addressForConnecting.Store(addr, struct{}{})
	if this.IsConnectionExists(addr) {
		log.Debugf("connection exist %s", addr)
		this.addressForConnecting.Delete(addr)
		return nil
	}
	log.Debugf("bootstrap to %v", addr)
	this.P2p.Bootstrap(addr)
	err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
	if err != nil {
		this.AddDialFailedCnt(addr)
	}
	this.addressForConnecting.Delete(addr)
	return err
}

func (this *Network) GetPeerStateByAddress(addr string) (network.PeerState, error) {
	s, err := this.P2p.GetRealConnState(addr)
	if err != nil {
		return s, err
	}
	client := this.P2p.GetPeerClient(addr)
	if client == nil {
		return s, fmt.Errorf("get peer client failed %s", addr)
	}
	return s, err
}

// IsConnectionReachable. check if peer state reachable
func (this *Network) IsConnectionReachable(addr string) bool {
	state, err := this.GetPeerStateByAddress(addr)
	log.Debugf("get peer state by addr: %s %v %s", addr, state, err)
	if err != nil {
		return false
	}
	if state != network.PEER_REACHABLE {
		return false
	}
	return true
}

func (this *Network) WaitForConnected(addr string, timeout time.Duration) error {
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		state, _ := this.GetPeerStateByAddress(addr)
		log.Debugf("GetPeerStateByAddress state addr:%s, :%d", addr, state)
		if state == network.PEER_REACHABLE {
			return nil
		}
		<-time.After(interval)
	}
	return errors.New("wait for connected timeout")
}

func (this *Network) Close(tAddr string) error {
	peer, err := this.P2p.Client(tAddr)
	if err != nil {
		log.Error("[P2P Close] close addr: %s error ", tAddr)
	} else {
		this.addressForHealthCheck.Delete(tAddr)
		peer.Close()
	}
	return nil
}

func (this *Network) Disconnect(addr string) error {
	if this.P2p == nil {
		return errors.New("network is nil")
	}
	peer := this.P2p.GetPeerClient(addr)
	if peer == nil {
		return errors.New("client is nil")
	}
	return peer.Close()
}

func (this *Network) IsConnectionExists(addr string) bool {
	if this.P2p == nil {
		return false
	}
	return this.P2p.ConnectionStateExists(addr)
}

// IsStateReachable. check if state is reachable
func (this *Network) IsStateReachable(addr string) bool {
	if this.P2p == nil {
		return false
	}
	state, err := this.P2p.GetRealConnState(addr)
	if state == network.PEER_REACHABLE && err == nil {
		return true
	}
	return false
}

func (this *Network) IsProxyConnectionExists() (bool, error) {
	if this.P2p == nil {
		return false, nil
	}
	return this.P2p.ProxyConnectionStateExists()
}

// Send send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, msgId, toAddr string) error {
	if err := this.healthCheckPeer(this.proxyAddr); err != nil {
		return err
	}
	if err := this.healthCheckPeer(toAddr); err != nil {
		return err
	}
	state, _ := this.GetPeerStateByAddress(toAddr)
	if state != network.PEER_REACHABLE {
		if err := this.healthCheckPeer(toAddr); err != nil {
			return fmt.Errorf("can not send to inactive peer %s", toAddr)
		}
	}
	if len(msgId) == 0 {
		msgId = utils.GenIdByTimestamp()
	}
	client := this.P2p.GetPeerClient(toAddr)
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	if err := client.AsyncSendAndWaitAck(context.Background(), msg, msgId); err != nil {
		this.AddSendFailedCnt(toAddr)
		return fmt.Errorf("failed to send message to %s", toAddr)
	}
	status := <-client.AckStatusNotify
	log.Debugf("receive ack status %v", status)
	if status.MessageID == msgId && status.Status == network.ACK_SUCCESS {
		return nil
	}
	return nil
}

// Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) Request(msg proto.Message, peer string) (proto.Message, error) {
	err := this.healthCheckPeer(this.proxyAddr)
	if err != nil {
		return nil, err
	}
	err = this.healthCheckPeer(peer)
	if err != nil {
		return nil, err
	}
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	// init context for timeout handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT)*time.Second)
	defer cancel()
	log.Debugf("send request msg to %s", peer)
	resp, err := client.Request(ctx, msg, time.Duration(dspNetCom.REQUEST_MSG_TIMEOUT)*time.Second)
	if err != nil {
		this.AddSendFailedCnt(peer)
		return nil, err
	}
	return resp, nil
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, peer string, retry, replyTimeout int) (proto.Message, error) {
	result := make(chan proto.Message, 1)
	done := false
	requestLock := &sync.Mutex{}
	client := this.P2p.GetPeerClient(peer)
	if client != nil {
		log.Debugf("disable backoff of peer %s", peer)
		client.DisableBackoff()
	}
	var response proto.Message
	var err error
	go func() {
		for i := 0; i < retry; i++ {
			// check proxy state
			err = this.healthCheckPeer(this.proxyAddr)
			if err != nil {
				continue
			}
			// check receiver state
			err = this.healthCheckPeer(peer)
			if err != nil {
				continue
			}
			log.Debugf("send request msg to %s with retry %d", peer, i)
			defer log.Debugf("send request msg to %s with retry %d done", peer, i)
			// get peer client to send msg
			client := this.P2p.GetPeerClient(peer)
			if client == nil {
				log.Errorf("get peer client is nil %s", peer)
				this.WaitForConnected(peer, time.Duration(replyTimeout)*time.Second)
				continue
			}
			go func(msg proto.Message) {
				// init context for timeout handling
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT)*time.Second)
				defer cancel()
				// send msg by request api and wait for the response
				res, reqErr := client.Request(ctx, msg, time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT)*time.Second)
				if reqErr != nil {
					return
				}
				requestLock.Lock()
				defer requestLock.Unlock()
				if done {
					return
				}
				done = true
				result <- res
			}(msg)
			<-time.After(time.Duration(dspNetCom.REQUEST_MSG_TIMEOUT) * time.Second)
			requestLock.Lock()
			if done {
				requestLock.Unlock()
				break
			}
			requestLock.Unlock()
		}
	}()
	select {
	case res := <-result:
		response = res
		err = nil
	case <-time.After(time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT) * time.Second):
		response = nil
		err = errors.New("retry all request and failed")
	}
	client = this.P2p.GetPeerClient(peer)
	if client != nil {
		log.Debugf("enable backoff of peer %s", peer)
		client.EnableBackoff()
	}
	if err != nil {
		this.AddSendFailedCnt(peer)
	}
	return response, err
}

// Broadcast. broadcast same msg to peers. Handle action if send msg success.
// If one msg is sent failed, return err. But the previous success msgs can not be recalled.
// callback(responseMsg, responseToAddr).
func (this *Network) Broadcast(addrs []string, msg proto.Message, msgId string, needReply bool, callback func(proto.Message, string) bool) (map[string]error, error) {
	err := this.healthCheckPeer(this.proxyAddr)
	if err != nil {
		return nil, err
	}
	maxRoutines := dspNetCom.MAX_GOROUTINES_IN_LOOP
	if len(addrs) <= dspNetCom.MAX_GOROUTINES_IN_LOOP {
		maxRoutines = len(addrs)
	}
	type broadcastReq struct {
		addr string
	}
	type broadcastResp struct {
		addr string
		err  error
	}
	dispatch := make(chan *broadcastReq, 0)
	done := make(chan *broadcastResp, 0)
	stop := int32(0)
	for i := 0; i < maxRoutines; i++ {
		go func() {
			for {
				req, ok := <-dispatch
				log.Debugf("receive request from %s, ok %t", req, ok)
				if !ok || req == nil {
					break
				}
				if !this.IsConnectionExists(req.addr) {
					log.Debugf("broadcast msg check %v not exist, connecting...", req.addr)
					err := this.ConnectAndWait(req.addr)
					if err != nil {
						done <- &broadcastResp{
							addr: req.addr,
							err:  err,
						}
						continue
					}
				}
				var res proto.Message
				var err error
				if !needReply {
					err = this.Send(msg, msgId, req.addr)
				} else {
					res, err = this.Request(msg, req.addr)
				}
				if err != nil {
					log.Errorf("broadcast msg to %s err %s", req.addr, err)
					done <- &broadcastResp{
						addr: req.addr,
						err:  err,
					}
					continue
				} else {
					log.Debugf("receive reply msg from %s", req.addr)
				}
				if callback != nil {
					finished := callback(res, req.addr)
					if finished {
						atomic.AddInt32(&stop, 1)
					}
				}
				done <- &broadcastResp{
					addr: req.addr,
					err:  nil,
				}
				log.Debugf("send broadcast to done done")
			}
		}()
	}
	go func() {
		for _, addr := range addrs {
			dispatch <- &broadcastReq{
				addr: addr,
			}
			if atomic.LoadInt32(&stop) > 0 {
				return
			}
		}
	}()

	m := make(map[string]error)
	for {
		result := <-done
		if atomic.LoadInt32(&stop) > 0 {
			return m, nil
		}
		m[result.addr] = result.err
		if len(m) != len(addrs) {
			continue
		}
		return m, nil
	}
}

func (this *Network) ReconnectPeer(addr string) error {
	if _, ok := this.addressForConnecting.Load(addr); ok {
		err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
		return err
	}
	this.addressForConnecting.Store(addr, struct{}{})
	state, err := this.GetPeerStateByAddress(addr)
	if state == network.PEER_REACHABLE && err == nil {
		this.addressForConnecting.Delete(addr)
		return nil
	}
	err = this.P2p.ReconnectPeer(addr)
	this.addressForConnecting.Delete(addr)
	return err
}

//P2P network msg receive. transfer to actor_channel
func (this *Network) Receive(ctx *network.ComponentContext, message proto.Message, from string) error {
	// TODO check message is nil
	state, err := this.GetPeerStateByAddress(from)
	log.Debugf("Network.Receive %T Msg %s, state: %d, err: %s", message, from, state, err)
	switch message.(type) {
	case *messages.Processed:
		act.OnBusinessMessage(message, from)
	case *messages.Delivered:
		act.OnBusinessMessage(message, from)
	case *messages.SecretRequest:
		act.OnBusinessMessage(message, from)
	case *messages.RevealSecret:
		act.OnBusinessMessage(message, from)
	case *messages.BalanceProof:
		act.OnBusinessMessage(message, from)
	case *messages.DirectTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.LockedTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.RefundTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.LockExpired:
		act.OnBusinessMessage(message, from)
	case *messages.WithdrawRequest:
		act.OnBusinessMessage(message, from)
	case *messages.Withdraw:
		act.OnBusinessMessage(message, from)
	case *messages.CooperativeSettleRequest:
		act.OnBusinessMessage(message, from)
	case *messages.CooperativeSettle:
		act.OnBusinessMessage(message, from)
	case *dspMsg.Message:
		if this.handler != nil {
			this.handler(ctx)
		}
	default:
		// if len(msg.String()) == 0 {
		// 	log.Warnf("Receive keepalive/keepresponse msg from %s", from)
		// 	return nil
		// }
		// log.Errorf("[P2pNetwork Receive] unknown message type:%s type %T", msg.String(), message)
	}

	return nil
}

func (this *Network) GetClientTime(addr string) (uint64, error) {
	client, err := this.P2p.Client(addr)
	if err != nil {
		return 0, err
	}
	if client == nil {
		return 0, fmt.Errorf("[network GetClientTime] client is nil: %s", addr)
	}
	return uint64(client.Time.Unix()), nil
}

func (this *Network) GetPeerReqNonce(addr string) uint64 {
	client := this.P2p.GetPeerClient(addr)
	if client == nil {
		return 0
	}
	return client.RequestNonce
}

func (this *Network) healthCheckProxyService() {
	ti := time.NewTicker(time.Duration(common.MAX_HEALTH_CHECK_INTERVAL) * time.Second)
	startedAt := time.Now().Unix()
	for {
		select {
		case <-ti.C:
			if (time.Now().Unix()-startedAt)%60 == 0 {
				proxyState, err := this.GetPeerStateByAddress(this.proxyAddr)
				if err != nil {
					log.Errorf("publicAddr: %s, proxy state: %d, err: %s", this.PublicAddr(), proxyState, err)
				} else {
					log.Debugf("publicAddr: %s, proxy state: %d", this.PublicAddr(), proxyState)
				}
			}
			this.healthCheckPeer(this.proxyAddr)
		case <-this.kill:
			log.Debugf("stop health check proxy service when receive kill")
			ti.Stop()
			return
		}
	}
}

func (this *Network) healthCheckPeer(addr string) error {
	if len(addr) == 0 || len(this.proxyAddr) == 0 {
		return nil
	}
	peerState, err := this.GetPeerStateByAddress(addr)
	if peerState != network.PEER_REACHABLE {
		log.Debugf("health check peer: %s unreachable, err: %s ", addr, err)
	}
	if err == nil && peerState == network.PEER_REACHABLE {
		return nil
	}
	this.AddDisconnectCnt(addr)
	time.Sleep(time.Duration(common.BACKOFF_INIT_DELAY) * time.Second)
	if addr == this.proxyAddr {
		log.Debugf("reconnect proxy server ....")
		err = this.P2p.ReconnectProxyServer(this.proxyAddr)
		if err != nil {
			log.Errorf("proxy reconnect failed, err %s", err)
			return err
		}
	} else {
		err = this.ReconnectPeer(addr)
		if err != nil {
			return err
		}
	}
	err = this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
	if err != nil {
		log.Errorf("reconnect peer failed , err: %s", err)
		return err
	}
	log.Debugf("reconnect peer success: %s", addr)
	return nil
}

func getProtocolFromAddr(addr string) string {
	idx := strings.Index(addr, "://")
	if idx == -1 {
		return "tcp"
	}
	return addr[:idx]
}
