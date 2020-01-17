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
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/keepalive/proxyKeepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	dspCom "github.com/saveio/dsp-go-sdk/common"
	dspNetCom "github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	dspMsg "github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/p2p/peer"
	"github.com/saveio/pylons/actor/msg_opcode"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/pylons/network/transport/messages"
	"github.com/saveio/themis/common/log"
)

var once sync.Once

type Network struct {
	P2p                *network.Network                // underlay network p2p instance
	builder            *network.Builder                // network builder
	intranetIP         string                          // intranet ip
	proxyAddr          string                          // proxy address
	pid                *actor.PID                      // actor pid
	keys               *crypto.KeyPair                 // crypto network keys
	kill               chan struct{}                   // network stop signal
	handler            func(*network.ComponentContext) // network msg handler
	peers              *sync.Map                       // peer clients
	lock               *sync.RWMutex                   // lock for sync control
	addrForHealthCheck *sync.Map                       // address for keep health check
	asyncRecvDisabled  bool                            // disabled async receive msg
}

func NewP2P() *Network {
	n := &Network{
		P2p: new(network.Network),
	}

	n.kill = make(chan struct{})
	n.peers = new(sync.Map)
	n.addrForHealthCheck = new(sync.Map)
	n.lock = new(sync.RWMutex)
	return n
}

func (this *Network) GetProxyServer() string {
	return this.proxyAddr
}

// IsProxyAddr. check if the address is a proxy address
func (this *Network) IsProxyAddr(addr string) bool {
	if len(this.proxyAddr) > 0 && addr == this.proxyAddr {
		return true
	}
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 &&
		strings.Contains(config.Parameters.BaseConfig.NATProxyServerAddrs, addr) {
		return true
	}
	return false
}

func (this *Network) Protocol() string {
	return getProtocolFromAddr(this.PublicAddr())
}

func (this *Network) PublicAddr() string {
	return this.P2p.ID.Address
}

// SetPID sets p2p actor
func (this *Network) SetPID(pid *actor.PID) {
	this.pid = pid
}

// GetPID returns p2p actor
func (this *Network) GetPID() *actor.PID {
	return this.pid
}

func (this *Network) Start(protocol, addr, port string, opts ...NetworkOption) error {
	builderOpt := []network.BuilderOption{
		network.WriteFlushLatency(1 * time.Millisecond),
		network.WriteTimeout(dspCom.NETWORK_STREAM_WRITE_TIMEOUT),
	}
	builder := network.NewBuilderWithOptions(builderOpt...)
	this.builder = builder
	for _, opt := range opts {
		opt.apply(this)
	}
	// set keys
	if this.keys != nil {
		log.Debugf("Network use account key")
		builder.SetKeys(this.keys)
	} else {
		log.Debugf("Network use RandomKeyPair key")
		builder.SetKeys(ed25519.RandomKeyPair())
	}
	intranetHost := "127.0.0.1"
	if len(this.intranetIP) > 0 {
		intranetHost = this.intranetIP
	}

	log.Debugf("network start at %s, listen addr %s",
		fmt.Sprintf("%s://%s:%s", protocol, intranetHost, port), fmt.Sprintf("%s://%s:%s", protocol, addr, port))
	builder.SetListenAddr(fmt.Sprintf("%s://%s:%s", protocol, intranetHost, port))
	builder.SetAddress(fmt.Sprintf("%s://%s:%s", protocol, addr, port))

	// add msg receiver
	recvMsgComp := new(NetComponent)
	recvMsgComp.Net = this
	builder.AddComponent(recvMsgComp)

	// add peer component
	peerComp := new(PeerComponent)
	peerComp.Net = this
	builder.AddComponent(peerComp)

	// add peer keepalive
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
	}
	builder.AddComponent(keepalive.New(options...))

	// add proxy keepalive
	builder.AddComponent(proxyKeepalive.New())

	// add ack reply
	ackOption := []ackreply.ComponentOption{
		ackreply.WithAckCheckedInterval(time.Second * common.ACK_MSG_CHECK_INTERVAL),
		ackreply.WithAckMessageTimeout(time.Second * common.MAX_ACK_MSG_TIMEOUT),
	}
	builder.AddComponent(ackreply.New(ackOption...))

	this.addProxyComponents(builder)

	var err error
	this.P2p, err = builder.Build()
	if err != nil {
		log.Error("[P2pNetwork] Start builder.Build error: ", err.Error())
		return err
	}
	this.P2p.DisableCompress()
	if this.asyncRecvDisabled {
		this.P2p.DisableMsgGoroutine()
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
	go this.healthCheckService()
	return nil
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
		log.Debugf("start proxy will blocking...%s %s, networkId: %d",
			protocol, proxyAddr, config.Parameters.BaseConfig.NetworkId)
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
			this.addrForHealthCheck.Store(proxyAddr, struct{}{})
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
		return fmt.Errorf("network is nil")
	}
	peerState, _ := this.GetPeerStateByAddress(tAddr)
	if peerState == network.PEER_REACHABLE {
		return nil
	}
	p, ok := this.peers.Load(tAddr)
	if ok && p.(*peer.Peer).State() == peer.ConnectStateConnecting {
		log.Debugf("already try to connect %s", tAddr)
		return nil
	}
	var pr *peer.Peer
	if !ok {
		pr = peer.New(tAddr)
		pr.SetState(peer.ConnectStateConnecting)
		this.peers.Store(tAddr, pr)
	} else {
		pr = p.(*peer.Peer)
	}
	log.Debugf("connecting....+")
	this.P2p.Bootstrap(tAddr)
	log.Debugf("connecting....+ done")
	pr.SetState(peer.ConnectStateConnected)
	return nil
}

func (this *Network) ConnectAndWait(addr string) error {
	p, ok := this.peers.Load(addr)
	var pr *peer.Peer
	if ok {
		pr = p.(*peer.Peer)
		if pr.State() == peer.ConnectStateConnecting {
			// already try to connect, don't retry before we get a result
			log.Info("already try to connect")
			err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
			if err != nil {
				pr.SetState(peer.ConnectStateFailed)
			} else {
				pr.SetState(peer.ConnectStateConnected)
			}
			return err
		}
	}
	if !ok {
		pr = peer.New(addr)
		pr.SetState(peer.ConnectStateConnecting)
		this.peers.Store(addr, pr)
	}
	if this.IsConnectionExists(addr) {
		log.Debugf("connection exist %s", addr)
		pr.SetState(peer.ConnectStateConnected)
		return nil
	}
	log.Debugf("bootstrap to %v", addr)
	this.P2p.Bootstrap(addr)
	err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
	if err != nil {
		pr.SetState(peer.ConnectStateFailed)
	} else {
		pr.SetState(peer.ConnectStateConnected)
	}
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
		state, err := this.GetPeerStateByAddress(addr)
		log.Debugf("GetPeerStateByAddress state addr:%s, :%d, err %v", addr, state, err)
		if state == network.PEER_REACHABLE {
			return nil
		}
		<-time.After(interval)
	}
	return fmt.Errorf("wait for connecting %s timeout", addr)
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

func (this *Network) HealthCheckPeer(addr string) error {
	if !this.P2p.ProxyModeEnable() {
		if _, ok := this.addrForHealthCheck.Load(addr); !ok {
			log.Debugf("ignore check health of peer %s, because proxy mode is disabled", addr)
			return nil
		}
		log.Debugf("proxy mode is disabled, but %s exist in health check list", addr)
	}
	if len(addr) == 0 {
		return nil
	}
	peerState, err := this.GetPeerStateByAddress(addr)
	log.Debugf("get peer state of %s, state %d, err %s", addr, peerState, err)
	if peerState != network.PEER_REACHABLE {
		log.Debugf("health check peer: %s unreachable, err: %s ", addr, err)
	}
	if err == nil && peerState == network.PEER_REACHABLE {
		return nil
	}
	time.Sleep(time.Duration(common.BACKOFF_INIT_DELAY) * time.Second)
	if addr == this.proxyAddr {
		log.Debugf("reconnect proxy server ....")
		err = this.P2p.ReconnectProxyServer(this.proxyAddr)
		if err != nil {
			log.Errorf("proxy reconnect failed, err %s", err)
			return err
		}
	} else {
		err = this.reconnectPeer(addr)
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

// SendOnce send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) SendOnce(msg proto.Message, toAddr string) error {
	err := this.HealthCheckPeer(this.proxyAddr)
	if err != nil {
		return err
	}
	err = this.HealthCheckPeer(toAddr)
	if err != nil {
		return err
	}
	state, _ := this.GetPeerStateByAddress(toAddr)
	if state != network.PEER_REACHABLE {
		return fmt.Errorf("can not send to inactive peer %s", toAddr)
	}
	signed, err := this.P2p.PrepareMessage(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to sign message")
	}
	err = this.P2p.Write(toAddr, signed)
	log.Debugf("write msg done sender:%s, to %s, nonce: %d", signed.GetSender().Address, toAddr, signed.GetMessageNonce())
	if err != nil {
		return fmt.Errorf("failed to send message to %s", toAddr)
	}
	return nil
}

// Send send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, msgId, toAddr string, sendTimeout time.Duration) error {
	if err := this.HealthCheckPeer(this.proxyAddr); err != nil {
		return err
	}
	log.Debugf("before send, health check it")
	if err := this.HealthCheckPeer(toAddr); err != nil {
		return err
	}
	state, _ := this.GetPeerStateByAddress(toAddr)
	if state != network.PEER_REACHABLE {
		if err := this.HealthCheckPeer(toAddr); err != nil {
			return fmt.Errorf("can not send to inactive peer %s", toAddr)
		}
	}
	p, ok := this.peers.Load(toAddr)
	log.Debugf("load peer success %t from %v", ok, toAddr)
	if !ok {
		return fmt.Errorf("peer not exist %s", toAddr)
	}
	this.stopKeepAlive()
	pr := p.(*peer.Peer)
	log.Debugf("send msg %s to %s", msgId, toAddr)
	var err error
	if sendTimeout > 0 {
		err = pr.StreamSend(msgId, msg, sendTimeout)
	} else {
		err = pr.Send(msgId, msg)
	}
	log.Debugf("send msg %s to %s done", msgId, toAddr)
	this.restartKeepAlive()
	return err
}

// Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) SendAndWaitReply(msg proto.Message, msgId, toAddr string, sendTimeout time.Duration) (
	proto.Message, error) {
	if this == nil {
		return nil, errors.New("no network")
	}
	if err := this.HealthCheckPeer(this.proxyAddr); err != nil {
		return nil, err
	}
	if err := this.HealthCheckPeer(toAddr); err != nil {
		return nil, err
	}
	state, _ := this.GetPeerStateByAddress(toAddr)
	if state != network.PEER_REACHABLE {
		if err := this.HealthCheckPeer(toAddr); err != nil {
			return nil, fmt.Errorf("can not send to inactive peer %s", toAddr)
		}
	}
	p, ok := this.peers.Load(toAddr)
	if !ok {
		return nil, fmt.Errorf("peer not exist %s", toAddr)
	}
	this.stopKeepAlive()
	pr := p.(*peer.Peer)
	log.Debugf("send msg %s to %s", msgId, toAddr)
	var err error
	var resp proto.Message
	if sendTimeout > 0 {
		resp, err = pr.StreamSendAndWaitReply(msgId, msg, sendTimeout)
	} else {
		resp, err = pr.SendAndWaitReply(msgId, msg)
	}
	log.Debugf("send and wait reply done  %s", err)
	this.restartKeepAlive()
	return resp, err
}

func (this *Network) AppendAddrToHealthCheck(addr string) {
	this.addrForHealthCheck.Store(addr, struct{}{})
}

func (this *Network) RemoveAddrFromHealthCheck(addr string) {
	this.addrForHealthCheck.Delete(addr)
}

// [Deprecated] Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) Request(msg proto.Message, peer string) (proto.Message, error) {
	err := this.HealthCheckPeer(this.proxyAddr)
	if err != nil {
		return nil, err
	}
	err = this.HealthCheckPeer(peer)
	if err != nil {
		return nil, err
	}
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	// init context for timeout handling
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT)*time.Second)
	defer cancel()
	log.Debugf("send request msg to %s", peer)
	resp, err := client.Request(ctx, msg, time.Duration(dspNetCom.REQUEST_MSG_TIMEOUT)*time.Second)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// [Deprecated] RequestWithRetry. send msg to peer and wait for response synchronously
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
			err = this.HealthCheckPeer(this.proxyAddr)
			if err != nil {
				continue
			}
			// check receiver state
			err = this.HealthCheckPeer(peer)
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
				ctx, cancel := context.WithTimeout(context.Background(),
					time.Duration(dspCom.ACTOR_MAX_P2P_REQ_TIMEOUT)*time.Second)
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

	return response, err
}

// Broadcast. broadcast same msg to peers. Handle action if send msg success.
// If one msg is sent failed, return err. But the previous success msgs can not be recalled.
// callback(responseMsg, responseToAddr).
func (this *Network) Broadcast(addrs []string, msg proto.Message, msgId string,
	callbacks ...func(proto.Message, string) bool) (map[string]error, error) {
	err := this.HealthCheckPeer(this.proxyAddr)
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
				if callbacks == nil || len(callbacks) == 0 {
					err = this.Send(msg, msgId, req.addr, 0)
				} else {
					res, err = this.SendAndWaitReply(msg, msgId, req.addr, 0)
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
				var finished bool
				for _, callback := range callbacks {
					if callback == nil {
						continue
					}
					if finished {
						continue
					}
					finished = callback(res, req.addr)
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
		select {
		case result := <-done:
			if atomic.LoadInt32(&stop) > 0 {
				return m, nil
			}
			m[result.addr] = result.err
			if len(m) != len(addrs) {
				continue
			}
			return m, nil
		case <-time.After(time.Duration(120) * time.Second):
			log.Debugf("broadcast wait too long")
			return m, nil
		}
	}
}

func (this *Network) reconnectPeer(addr string) error {
	p, ok := this.peers.Load(addr)
	var pr *peer.Peer
	if !ok {
		log.Warnf("peer no exist %s", addr)
		pr = peer.New(addr)
		this.peers.Store(addr, pr)
	} else {
		pr = p.(*peer.Peer)
	}
	if pr.State() == peer.ConnectStateConnecting {
		err := this.WaitForConnected(addr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
		if err != nil {
			pr.SetState(peer.ConnectStateFailed)
		} else {
			pr.SetState(peer.ConnectStateConnected)
		}
		return err
	}
	pr.SetState(peer.ConnectStateConnecting)
	state, err := this.GetPeerStateByAddress(addr)
	if state == network.PEER_REACHABLE && err == nil {
		pr.SetState(peer.ConnectStateConnected)
		return nil
	}
	err = this.P2p.ReconnectPeer(addr)
	if err != nil {
		pr.SetState(peer.ConnectStateFailed)
	} else {
		pr.SetState(peer.ConnectStateConnected)
	}
	return err
}

//P2P network msg receive. transfer to actor_channel
func (this *Network) Receive(ctx *network.ComponentContext, message proto.Message, from string) error {
	// TODO check message is nil
	state, err := this.GetPeerStateByAddress(from)
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
		log.Debugf("Network.Receive %T Msg %s, state: %d, err: %s", message, from, state, err)
		msg := message.(*dspMsg.Message)
		p, ok := this.peers.Load(from)
		if !ok {
			log.Warnf("receive a msg, but peer not found %s", from)
			return nil
		}
		pr, ok := p.(*peer.Peer)
		if !ok {
			log.Warnf("convert p to peer failed %s", from)
			return nil
		}
		if pr.IsMsgReceived(msg.MsgId) {
			log.Warnf("receive a duplicated msg, ignore it %s", msg.MsgId)
			return nil
		}
		pr.AddReceivedMsg(msg.MsgId)
		if len(msg.Syn) > 0 {
			// reply to origin request msg, no need to enter handle router
			pr, ok := this.peers.Load(from)
			log.Debugf("receive reply msg from %s, sync id %s", from, msg.Syn)
			if !ok {
				log.Warnf("receive a unknown msg from %s", from)
				return nil
			}
			pr.(*peer.Peer).Receive(msg.Syn, msg)
			return nil
		}
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
	p, ok := this.peers.Load(addr)
	if !ok {
		return 0, fmt.Errorf("[network GetClientTime] client is nil: %s", addr)
	}
	t := p.(*peer.Peer).ActiveTime()
	return uint64(t.Unix()), nil
}

func (this *Network) addProxyComponents(builder *network.Builder) {
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

func (this *Network) healthCheckService() {
	ti := time.NewTicker(time.Duration(common.MAX_HEALTH_CHECK_INTERVAL) * time.Second)
	startedAt := time.Now().Unix()
	for {
		select {
		case <-ti.C:
			shouldLog := ((time.Now().Unix()-startedAt)%60 == 0)
			this.addrForHealthCheck.Range(func(key, value interface{}) bool {
				addr, _ := key.(string)
				if len(addr) == 0 {
					return true
				}
				if shouldLog {
					addrState, err := this.GetPeerStateByAddress(addr)
					if err != nil {
						log.Errorf("publicAddr: %s, addr %s state: %d, err: %s",
							this.PublicAddr(), addr, addrState, err)
					} else {
						log.Debugf("publicAddr: %s, addr %s state: %d", this.PublicAddr(), addr, addrState)
					}
				}
				this.HealthCheckPeer(addr)
				return true
			})
		case <-this.kill:
			log.Debugf("stop health check proxy service when receive kill")
			ti.Stop()
			return
		}
	}
}

func (this *Network) getKeepAliveComponent() *keepalive.Component {
	for _, info := range this.P2p.Components.GetInstallComponents() {
		keepaliveCom, ok := info.Component.(*keepalive.Component)
		if !ok {
			continue
		}
		return keepaliveCom
	}
	return nil
}

func (this *Network) updatePeerTime(addr string) {
	client := this.P2p.GetPeerClient(addr)
	if client == nil {
		return
	}
	client.Time = time.Now()
}

func (this *Network) IsPeerNetQualityBad(addr string) bool {
	totalFailed := &peer.FailedCount{}
	peerCnt := 0
	var peerFailed *peer.FailedCount
	this.peers.Range(func(key, value interface{}) bool {
		peerCnt++
		peerAddr, _ := key.(string)
		pr, _ := value.(*peer.Peer)
		cnt := pr.GetFailedCnt()
		totalFailed.Dial += cnt.Dial
		totalFailed.Send += cnt.Send
		totalFailed.Recv += cnt.Recv
		totalFailed.Disconnect += cnt.Disconnect
		if peerAddr == addr {
			peerFailed = cnt
		}
		return false
	})
	if peerFailed == nil {
		return false
	}
	log.Debugf("peer failed %v, totalFailed %v, peer cnt %v", peerFailed, totalFailed, peerCnt)
	if (peerFailed.Dial > 0 && peerFailed.Dial >= totalFailed.Dial/peerCnt) ||
		(peerFailed.Send > 0 && peerFailed.Send >= totalFailed.Send/peerCnt) ||
		(peerFailed.Disconnect > 0 && peerFailed.Disconnect >= totalFailed.Disconnect/peerCnt) ||
		(peerFailed.Recv > 0 && peerFailed.Recv >= totalFailed.Recv/peerCnt) {
		return true
	}
	return false
}

func (this *Network) stopKeepAlive() {
	this.lock.Lock()
	defer this.lock.Unlock()
	var ka *keepalive.Component
	var ok bool
	for _, info := range this.P2p.Components.GetInstallComponents() {
		ka, ok = info.Component.(*keepalive.Component)
		if ok {
			break
		}
	}
	if ka == nil {
		return
	}
	// stop keepalive for temporary
	ka.Cleanup(this.P2p)
	deleted := this.P2p.Components.Delete(ka)
	log.Debugf("stop keep alive %t", deleted)
}

func (this *Network) restartKeepAlive() {
	this.lock.Lock()
	defer this.lock.Unlock()
	var ka *keepalive.Component
	var ok bool
	for _, info := range this.P2p.Components.GetInstallComponents() {
		ka, ok = info.Component.(*keepalive.Component)
		if ok {
			break
		}
	}
	if ka != nil {
		return
	}
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
	}
	ka = keepalive.New(options...)
	err := this.builder.AddComponent(ka)
	if err != nil {
		return
	}
	ka.Startup(this.P2p)
}

func getProtocolFromAddr(addr string) string {
	idx := strings.Index(addr, "://")
	if idx == -1 {
		return "tcp"
	}
	return addr[:idx]
}
