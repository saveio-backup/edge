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
	"github.com/saveio/carrier/network/components/backoff"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	dspmsg "github.com/saveio/dsp-go-sdk/network/message/pb"
	edgeCom "github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/pylons/network/transport/messages"

	"github.com/saveio/themis/common/log"
)

var once sync.Once

const (
	OpCodeProcessed opcode.Opcode = 1000 + iota
	OpCodeDelivered
	OpCodeSecretRequest
	OpCodeRevealSecret
	OpCodeSecretMsg
	OpCodeDirectTransfer
	OpCodeLockedTransfer
	OpCodeRefundTransfer
	OpCodeLockExpired
	OpCodeWithdrawRequest
	OpCodeWithdraw
	OpCodeCooperativeSettleRequest
	OpCodeCooperativeSettle
)

var opCodes = map[opcode.Opcode]proto.Message{
	OpCodeProcessed:                &messages.Processed{},
	OpCodeDelivered:                &messages.Delivered{},
	OpCodeSecretRequest:            &messages.SecretRequest{},
	OpCodeRevealSecret:             &messages.RevealSecret{},
	OpCodeSecretMsg:                &messages.Secret{},
	OpCodeDirectTransfer:           &messages.DirectTransfer{},
	OpCodeLockedTransfer:           &messages.LockedTransfer{},
	OpCodeRefundTransfer:           &messages.RefundTransfer{},
	OpCodeLockExpired:              &messages.LockExpired{},
	OpCodeWithdrawRequest:          &messages.WithdrawRequest{},
	OpCodeWithdraw:                 &messages.Withdraw{},
	OpCodeCooperativeSettleRequest: &messages.CooperativeSettleRequest{},
	OpCodeCooperativeSettle:        &messages.CooperativeSettle{},
}

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
}

func NewP2P() *Network {
	n := &Network{
		P2p: new(network.Network),
	}
	n.addressForHealthCheck = new(sync.Map)
	n.addressForConnecting = new(sync.Map)
	n.kill = make(chan struct{})
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
		network.WriteTimeout(edgeCom.NET_STREAM_WRITE_TIMEOUT),
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
		this.keepaliveTimeout = edgeCom.KEEPALIVE_TIMEOUT * time.Second
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
			backoff.WithInitialDelay(edgeCom.BACKOFF_INIT_DELAY * time.Second),
			backoff.WithMaxAttempts(edgeCom.BACKOFF_MAX_ATTEMPTS),
			backoff.WithPriority(65535),
			backoff.WithMaxInterval(time.Duration(300) * time.Second),
		}
		log.Debugf("backoff opt %v", backoff_options)
		builder.AddComponent(backoff.New(backoff_options...))
	}

	this.AddProxyComponents(builder)

	var err error
	this.P2p, err = builder.Build()
	if err != nil {
		log.Error("[P2pNetwork] Start builder.Build error: ", err.Error())
		return err
	}
	once.Do(func() {
		for k, v := range opCodes {
			err := opcode.RegisterMessageType(k, v)
			if err != nil {
				log.Errorf("register messages failed %v", k)
			}
		}
		opcode.RegisterMessageType(opcode.Opcode(common.MSG_OP_CODE), &pb.Message{})
	})
	this.P2p.SetCompressFileSize(edgeCom.COMPRESS_DATA_SIZE)
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
		case <-time.After(time.Duration(edgeCom.START_PROXY_TIMEOUT) * time.Second):
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
	if this == nil {
		return fmt.Errorf("[Connect] this is nil")
	}
	peerState, _ := this.GetPeerStateByAddress(tAddr)
	if peerState == network.PEER_REACHABLE {
		return nil
	}
	if _, ok := this.addressForHealthCheck.Load(tAddr); ok {
		// already try to connect, don't retry before we get a result
		log.Info("already try to connect")
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
		err := this.WaitForConnected(addr, time.Duration(edgeCom.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
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
	err := this.WaitForConnected(addr, time.Duration(edgeCom.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
	this.addressForConnecting.Delete(addr)
	return err
}

func (this *Network) GetPeerStateByAddress(addr string) (network.PeerState, error) {
	return this.P2p.GetRealConnState(addr)
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

func (this *Network) IsProxyConnectionExists() (bool, error) {
	if this.P2p == nil {
		return false, nil
	}
	return this.P2p.ProxyConnectionStateExists()
}

// Send send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, toAddr string) error {
	err := this.healthCheckPeer(this.proxyAddr)
	if err != nil {
		return err
	}
	err = this.healthCheckPeer(toAddr)
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
	log.Debugf("send request msg to %s", peer)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
	defer func() {
		log.Debugf("cancel send request msg to peer %s", peer)
		cancel()
	}()
	return client.Request(ctx, msg)
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, peer string, retry, timeout int) (proto.Message, error) {
	var err error
	var res proto.Message
	client := this.P2p.GetPeerClient(peer)
	if client != nil {
		log.Debugf("disable backoff of peer %s", peer)
		client.DisableBackoff()
	}
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
		// get peer client to send msg
		client := this.P2p.GetPeerClient(peer)
		if client == nil {
			log.Errorf("get peer client is nil %s", peer)
			this.WaitForConnected(peer, time.Duration(timeout/retry)*time.Second)
			continue
		}
		// init context for timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout/retry)*time.Second)
		defer cancel()
		// send msg by request api and wait for the response
		res, err = client.Request(ctx, msg)
		if err == nil {
			break
		}
	}
	client = this.P2p.GetPeerClient(peer)
	if client != nil {
		log.Debugf("enable backoff of peer %s", peer)
		client.EnableBackoff()
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Broadcast. broadcast same msg to peers. Handle action if send msg success.
// If one msg is sent failed, return err. But the previous success msgs can not be recalled.
// callback(responseMsg, responseToAddr).
func (this *Network) Broadcast(addrs []string, msg proto.Message, needReply bool, callback func(proto.Message, string) bool) (map[string]error, error) {
	err := this.healthCheckPeer(this.proxyAddr)
	if err != nil {
		return nil, err
	}
	maxRoutines := common.MAX_GOROUTINES_IN_LOOP
	if len(addrs) <= common.MAX_GOROUTINES_IN_LOOP {
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
					err = this.Send(msg, req.addr)
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
					log.Debugf("receive reply msg from %s, msg:%s", req.addr, msg.String())
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
		err := this.WaitForConnected(addr, time.Duration(edgeCom.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
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
	case *messages.Secret:
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
	case *dspmsg.Message:
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

func (this *Network) healthCheckProxyService() {
	ti := time.NewTicker(time.Duration(edgeCom.MAX_HEALTH_CHECK_INTERVAL) * time.Second)
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
			return
		}
	}
}

func (this *Network) healthCheckPeer(addr string) error {
	if len(addr) == 0 || len(this.proxyAddr) == 0 {
		return nil
	}
	peerState, err := this.GetPeerStateByAddress(addr)
	if err == nil && peerState == network.PEER_REACHABLE {
		return nil
	}
	log.Debugf("health check peer: %s unreachable, err: %s ", addr, err)
	time.Sleep(time.Duration(edgeCom.BACKOFF_INIT_DELAY) * time.Second)
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
	err = this.WaitForConnected(addr, time.Duration(edgeCom.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
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
