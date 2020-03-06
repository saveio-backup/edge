package network

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
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
	dspMsg "github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/p2p/peer"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/pylons/network/transport/messages"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

type Network struct {
	P2p                  *network.Network                        // underlay network p2p instance
	builder              *network.Builder                        // network builder
	networkId            uint32                                  // network id
	intranetIP           string                                  // intranet ip
	proxyAddrs           []string                                // proxy address
	opCodes              map[opcode.Opcode]proto.Message         // opcodes
	pid                  *actor.PID                              // actor pid
	keys                 *crypto.KeyPair                         // crypto network keys
	kill                 chan struct{}                           // network stop signal
	handler              func(*network.ComponentContext, string) // network msg handler
	walletAddrFromPeerId func(string) string                     // peer key from id delegate
	peers                *sync.Map                               // peer clients
	lock                 *sync.RWMutex                           // lock for sync control
	peerForHealthCheck   *sync.Map                               // peerId for keep health check
	asyncRecvDisabled    bool                                    // disabled async receive msg
	connectLock          *sync.Map                               // connection lock for address
}

func NewP2P(opts ...NetworkOption) *Network {
	n := &Network{
		P2p: new(network.Network),
	}
	n.kill = make(chan struct{})
	n.peers = new(sync.Map)
	n.peerForHealthCheck = new(sync.Map)
	n.lock = new(sync.RWMutex)
	n.connectLock = new(sync.Map)
	// update by options
	n.walletAddrFromPeerId = func(id string) string {
		return id
	}
	for _, opt := range opts {
		opt.apply(n)
	}
	return n
}

func (this *Network) WalletAddrFromPeerId(peerId string) string {
	return this.walletAddrFromPeerId(peerId)
}

// GetProxyServer. get working proxy server
func (this *Network) GetProxyServer() *network.ProxyServer {
	if this.P2p == nil {
		return &network.ProxyServer{}
	}
	addr, peerId := this.P2p.GetWorkingProxyServer()
	return &network.ProxyServer{
		IP:     addr,
		PeerID: peerId,
	}
}

// IsProxyAddr. check if the address is a proxy address
func (this *Network) IsProxyAddr(addr string) bool {
	for _, a := range this.proxyAddrs {
		if a == addr {
			return true
		}
	}
	return false
}

// Protocol. get network protocol
func (this *Network) Protocol() string {
	return getProtocolFromAddr(this.PublicAddr())
}

// PublicAddr. get network public host address
func (this *Network) PublicAddr() string {
	return this.P2p.ID.Address
}

// GetPID returns p2p actor
func (this *Network) GetPID() *actor.PID {
	return this.pid
}

// Start. start network
func (this *Network) Start(protocol, addr, port string) error {
	builderOpt := []network.BuilderOption{
		network.WriteFlushLatency(1 * time.Millisecond),
		network.WriteTimeout(dspCom.NETWORK_STREAM_WRITE_TIMEOUT),
		// network.WriteBufferSize(common.MAX_WRITE_BUFFER_SIZE),
	}
	builder := network.NewBuilderWithOptions(builderOpt...)
	this.builder = builder

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
	for k, v := range this.opCodes {
		err := opcode.RegisterMessageType(k, v)
		if err != nil {
			log.Errorf("register messages failed %v, %s", k, err)
		} else {
			log.Debugf("register msg %v success", k)
		}
	}
	this.P2p.SetDialTimeout(time.Duration(common.NETWORK_DIAL_TIMEOUT) * time.Second)
	this.P2p.SetCompressFileSize(common.COMPRESS_DATA_SIZE)
	if len(this.proxyAddrs) > 0 {
		log.Debugf("set proxy mode %v", this.proxyAddrs)
		this.P2p.EnableProxyMode(true)
		this.P2p.SetProxyServer([]network.ProxyServer{
			network.ProxyServer{
				IP: this.proxyAddrs[0],
			},
		})
	}
	this.P2p.SetNetworkID(this.networkId)
	go this.P2p.Listen()
	this.P2p.BlockUntilListening()
	err = this.startProxy(builder)
	if err != nil {
		log.Errorf("start proxy failed, err: %s", err)
	}
	if len(this.P2p.ID.Address) == 6 {
		return errors.New("invalid address")
	}
	go this.healthCheckService()
	return nil
}

// Stop. stop network
func (this *Network) Stop() {
	close(this.kill)
	this.P2p.Close()
}

// GetPeerFromWalletAddr. get peer from wallet addr
func (this *Network) GetPeerFromWalletAddr(walletAddr string) *peer.Peer {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	p, ok := this.peers.Load(walletAddr)
	if !ok {
		return nil
	}
	pr, ok := p.(*peer.Peer)
	if pr == nil || !ok {
		return nil
	}
	return pr
}

func (this *Network) GetWalletFromHostAddr(hostAddr string) string {
	walletAddr := ""
	this.peers.Range(func(key, value interface{}) bool {
		pr, ok := value.(*peer.Peer)
		if pr == nil || !ok {
			return true
		}
		if pr.GetHostAddr() == hostAddr {
			walletAddr = key.(string)
			return false
		}
		return true
	})
	return walletAddr
}

// Connect. connect to peer with host address, store its peer id after success
func (this *Network) Connect(hostAddr string) error {
	log.Debugf("connect to host addr %s", hostAddr)
	if this == nil {
		return fmt.Errorf("network is nil")
	}
	connLockValue, _ := this.connectLock.LoadOrStore(hostAddr, new(sync.RWMutex))
	connLock := connLockValue.(*sync.RWMutex)
	connLock.Lock()
	defer connLock.Unlock()

	_, ok := this.peers.Load(hostAddr)
	if ok {
		return this.waitForConnectedByHost(hostAddr, time.Duration(15)*time.Second)
	}
	walletAddr := this.GetWalletFromHostAddr(hostAddr)
	var pr *peer.Peer
	if len(walletAddr) > 0 {
		if this.IsConnReachable(walletAddr) {
			log.Debugf("connection exist %s", hostAddr)
			return nil
		}
		p, ok := this.peers.Load(walletAddr)
		if ok {
			pr = p.(*peer.Peer)
		} else {
			pr = peer.New(hostAddr)
		}
	} else {
		pr = peer.New(hostAddr)
	}
	pr.SetState(peer.ConnectStateConnecting)
	this.peers.Store(hostAddr, pr)
	log.Debugf("bootstrap to %v....", hostAddr)
	peerIds := this.P2p.Bootstrap([]string{hostAddr})
	log.Debugf("bootstrap to %v, %v", hostAddr, peerIds)
	if len(peerIds) == 0 {
		return fmt.Errorf("peer id is emptry from bootstraping to %s", hostAddr)
	}
	peerId := peerIds[0]
	walletAddr = this.walletAddrFromPeerId(peerId)
	pr.SetPeerId(peerId)
	pr.SetState(peer.ConnectStateConnected)
	log.Debugf("client %p, peerId %v, peer %p", this.P2p.GetPeerClient(peerId), peerId, pr)
	pr.SetClient(this.P2p.GetPeerClient(peerId))
	this.peers.Delete(hostAddr)
	this.peers.Store(walletAddr, pr)
	return nil
}

// GetPeerStateByAddress. get peer state by peerId
func (this *Network) GetConnStateByWallet(walletAddr string) (network.PeerState, error) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return 0, fmt.Errorf("peer not found %s", walletAddr)
	}
	peerId := pr.GetPeerId()
	log.Debugf("get peer id of wallet %s, %s, pr %p", walletAddr, peerId, pr)
	s, err := this.P2p.GetRealConnState(peerId)
	if err != nil {
		return s, err
	}
	client := this.P2p.GetPeerClient(peerId)
	if client == nil {
		return s, fmt.Errorf("get peer %s client is nil", peerId)
	}
	return s, err
}

// IsConnReachable. check if peer state reachable
func (this *Network) IsConnReachable(walletAddr string) bool {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	if this.P2p == nil || len(walletAddr) == 0 {
		return false
	}
	state, err := this.GetConnStateByWallet(walletAddr)
	log.Debugf("get peer state by addr: %s %v %s", walletAddr, state, err)
	if err != nil {
		return false
	}
	if state != network.PEER_REACHABLE {
		return false
	}
	return true
}

// WaitForConnected. poll to wait for connected
func (this *Network) WaitForConnected(walletAddr string, timeout time.Duration) error {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return fmt.Errorf("peer %s not found", walletAddr)
	}
	peerId := pr.GetPeerId()
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		state, err := this.GetConnStateByWallet(walletAddr)
		client := this.P2p.GetPeerClient(peerId)
		if client != nil {
			log.Debugf("peer id %s", client.ClientID())
		}
		log.Debugf("GetPeerStateByAddress state addr:%s, :%d, err %v", peerId, state, err)
		if state == network.PEER_REACHABLE && client != nil && len(client.ClientID()) > 0 {
			return nil
		}
		<-time.After(interval)
	}
	return fmt.Errorf("wait for connecting %s timeout", peerId)
}

func (this *Network) HealthCheckPeer(walletAddr string) error {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	if !this.P2p.ProxyModeEnable() {
		if _, ok := this.peerForHealthCheck.Load(walletAddr); !ok {
			log.Debugf("ignore check health of peer %s, because proxy mode is disabled", walletAddr)
			return nil
		}
		log.Debugf("proxy mode is disabled, but %s exist in health check list", walletAddr)
	}
	if len(walletAddr) == 0 {
		log.Debugf("health check empty peer id")
		return nil
	}
	peerState, err := this.GetConnStateByWallet(walletAddr)
	log.Debugf("get peer state of %s, state %d, err %s", walletAddr, peerState, err)
	if peerState != network.PEER_REACHABLE {
		log.Debugf("health check peer: %s unreachable, err: %s ", walletAddr, err)
	}
	if err == nil && peerState == network.PEER_REACHABLE {
		return nil
	}
	time.Sleep(time.Duration(common.BACKOFF_INIT_DELAY) * time.Second)
	proxy := this.GetProxyServer()
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return fmt.Errorf("peer not found %s", walletAddr)
	}
	peerId := pr.GetPeerId()
	if len(proxy.IP) > 0 && peerId == proxy.PeerID {
		log.Debugf("reconnect proxy server ....")
		err = this.P2p.ReconnectProxyServer(proxy.IP, proxy.PeerID)
		if err != nil {
			log.Errorf("proxy reconnect failed, err %s", err)
			this.WaitForConnected(walletAddr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
			return err
		}
	} else {
		log.Debugf("reconnect %s", walletAddr)
		err = this.reconnect(walletAddr)
		if err != nil {
			this.WaitForConnected(walletAddr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
			return err
		}
	}
	err = this.WaitForConnected(walletAddr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second)
	if err != nil {
		log.Errorf("reconnect peer failed , err: %s", err)
		return err
	}
	log.Debugf("reconnect peer success: %s", walletAddr)
	return nil
}

// SendOnce send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) SendOnce(msg proto.Message, walletAddr string) error {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	if this.P2p.ProxyModeEnable() && len(this.GetProxyServer().PeerID) > 0 {
		if err := this.HealthCheckPeer(this.walletAddrFromPeerId(this.GetProxyServer().PeerID)); err != nil {
			return err
		}
	}
	if err := this.HealthCheckPeer(walletAddr); err != nil {
		return err
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return fmt.Errorf("peer not found %s", walletAddr)
	}
	peerId := pr.GetPeerId()
	state, _ := this.GetConnStateByWallet(walletAddr)
	if state != network.PEER_REACHABLE {
		return fmt.Errorf("can not send to inactive peer %s", walletAddr)
	}
	signed, err := this.P2p.PrepareMessage(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("prepare message failed, err %s", err)
	}
	err = this.P2p.Write(peerId, signed)
	log.Debugf("write msg done sender:%s, to %s, nonce: %d", signed.GetSender().Address, peerId, signed.GetMessageNonce())
	if err != nil {
		return fmt.Errorf("send message to %s failed ", walletAddr)
	}
	return nil
}

// Send send msg to peer asynchronous
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, sessionId, msgId, walletAddr string, sendTimeout time.Duration) error {

	if this.P2p.ProxyModeEnable() && len(this.GetProxyServer().PeerID) > 0 {
		log.Debugf("check proxy %v", this.GetProxyServer())
		if err := this.HealthCheckPeer(this.walletAddrFromPeerId(this.GetProxyServer().PeerID)); err != nil {
			return err
		}
	}
	log.Debugf("before send, health check it")
	if err := this.HealthCheckPeer(walletAddr); err != nil {
		return err
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return fmt.Errorf("peer not found %s", walletAddr)
	}
	state, _ := this.GetConnStateByWallet(walletAddr)
	if state != network.PEER_REACHABLE {
		if err := this.HealthCheckPeer(walletAddr); err != nil {
			return fmt.Errorf("can not send to inactive peer %s", walletAddr)
		}
	}
	this.stopKeepAlive()
	log.Debugf("send msg %s to %s", msgId, walletAddr)
	var err error
	if sendTimeout > 0 {
		err = pr.StreamSend(sessionId, msgId, msg, sendTimeout)
	} else {
		err = pr.Send(msgId, msg)
	}
	log.Debugf("send msg %s to %s done", msgId, walletAddr)
	this.restartKeepAlive()
	return err
}

// Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) SendAndWaitReply(msg proto.Message, sessionId, msgId, walletAddr string, sendTimeout time.Duration) (
	proto.Message, error) {
	if this == nil {
		return nil, errors.New("no network")
	}
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	if this.P2p.ProxyModeEnable() && len(this.GetProxyServer().PeerID) > 0 {
		if err := this.HealthCheckPeer(this.walletAddrFromPeerId(this.GetProxyServer().PeerID)); err != nil {
			return nil, err
		}
	}
	if err := this.HealthCheckPeer(walletAddr); err != nil {
		return nil, err
	}
	state, _ := this.GetConnStateByWallet(walletAddr)
	if state != network.PEER_REACHABLE {
		if err := this.HealthCheckPeer(walletAddr); err != nil {
			return nil, fmt.Errorf("can not send to inactive peer %s", walletAddr)
		}
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return nil, fmt.Errorf("peer %s not found", walletAddr)
	}
	this.stopKeepAlive()
	log.Debugf("send msg %s to %s", msgId, walletAddr)
	var err error
	var resp proto.Message
	if sendTimeout > 0 {
		resp, err = pr.StreamSendAndWaitReply(sessionId, msgId, msg, sendTimeout)
	} else {
		resp, err = pr.SendAndWaitReply(msgId, msg)
	}
	log.Debugf("send and wait reply done  %s", err)
	this.restartKeepAlive()
	return resp, err
}

func (this *Network) AppendAddrToHealthCheck(walletAddr string) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return
	}
	peerId := pr.GetPeerId()
	this.peerForHealthCheck.Store(walletAddr, peerId)
}

func (this *Network) RemoveAddrFromHealthCheck(walletAddr string) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	this.peerForHealthCheck.Delete(walletAddr)
}

// ClosePeerSession
func (this *Network) ClosePeerSession(walletAddr, sessionId string) error {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	p, ok := this.peers.Load(walletAddr)
	if !ok {
		return fmt.Errorf("peer %s not found", walletAddr)
	}
	pr := p.(*peer.Peer)
	return pr.CloseSession(sessionId)
}

// GetPeerSendSpeed. return send speed for peer
func (this *Network) GetPeerSessionSpeed(walletAddr, sessionId string) (uint64, uint64, error) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	p, ok := this.peers.Load(walletAddr)
	if !ok {
		return 0, 0, fmt.Errorf("peer %s not found", walletAddr)
	}
	pr := p.(*peer.Peer)
	return pr.GetSessionSpeed(sessionId)
}

// Broadcast. broadcast same msg to peers. Handle action if send msg success.
// If one msg is sent failed, return err. But the previous success msgs can not be recalled.
// callback(responseMsg, responseToAddr).
func (this *Network) Broadcast(addrs []string, msg proto.Message, sessionId, msgId string,
	callbacks ...func(proto.Message, string) bool) (map[string]error, error) {
	if this.P2p.ProxyModeEnable() && len(this.GetProxyServer().PeerID) > 0 {
		err := this.HealthCheckPeer(this.walletAddrFromPeerId(this.GetProxyServer().PeerID))
		if err != nil {
			return nil, err
		}
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
				walletAddr := this.GetWalletFromHostAddr(req.addr)
				log.Debugf("broadcast msg %s to host %s wallet %s", msgId, req.addr, walletAddr)
				if len(walletAddr) == 0 || !this.IsConnReachable(walletAddr) {
					log.Debugf("%s not exist, connecting...", req.addr)
					if err := this.Connect(req.addr); err != nil {
						log.Errorf("broadcast msg connect err %s", err)
						done <- &broadcastResp{
							addr: req.addr,
							err:  err,
						}
						continue
					}
					walletAddr = this.GetWalletFromHostAddr(req.addr)
				}
				var res proto.Message
				var err error
				if callbacks == nil || len(callbacks) == 0 {
					err = this.Send(msg, sessionId, msgId, walletAddr, 0)
				} else {
					res, err = this.SendAndWaitReply(msg, sessionId, msgId, walletAddr, 0)
				}
				if err != nil {
					log.Errorf("broadcast msg to %s err %s", walletAddr, err)
					done <- &broadcastResp{
						addr: walletAddr,
						err:  err,
					}
					continue
				} else {
					log.Debugf("receive reply msg from %s", walletAddr)
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
			m[result.addr] = result.err
			if atomic.LoadInt32(&stop) > 0 {
				return m, nil
			}
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

//P2P network msg receive. transfer to actor_channel
func (this *Network) Receive(ctx *network.ComponentContext, message proto.Message, hostAddr, peerId string) error {
	// TODO check message is nil
	walletAddr := this.walletAddrFromPeerId(peerId)
	if strings.Contains(walletAddr, this.Protocol()) || len(walletAddr) == 0 {
		log.Debugf("receive %T msg from peer %s", message, peerId)
		return nil
	}
	log.Debugf("peerId %s wallet %s, message %T", peerId, walletAddr, message)
	state, err := this.GetConnStateByWallet(walletAddr)
	switch message.(type) {
	case *messages.Processed:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.Delivered:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.SecretRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.RevealSecret:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.BalanceProof:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.DirectTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.LockedTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.RefundTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.LockExpired:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.WithdrawRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.Withdraw:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.CooperativeSettleRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.CooperativeSettle:
		act.OnBusinessMessage(message, walletAddr)
	case *dspMsg.Message:
		log.Debugf("Network.Receive %T Msg %s, state: %d, err: %s", message, hostAddr, state, err)
		msg := message.(*dspMsg.Message)
		p, ok := this.peers.Load(walletAddr)
		if !ok {
			log.Warnf("receive a msg, but peer not found %s", walletAddr)
			return nil
		}
		pr, ok := p.(*peer.Peer)
		if !ok {
			log.Warnf("convert p to peer failed %s", walletAddr)
			return nil
		}
		if pr.IsMsgReceived(msg.MsgId) {
			log.Warnf("receive a duplicated msg, ignore it %s", msg.MsgId)
			return nil
		}
		pr.AddReceivedMsg(msg.MsgId)
		if len(msg.Syn) > 0 {
			// reply to origin request msg, no need to enter handle router
			pr, ok := this.peers.Load(walletAddr)
			log.Debugf("receive reply msg from %s, sync id %s", walletAddr, msg.Syn)
			if !ok {
				log.Warnf("receive a unknown msg from %s", walletAddr)
				return nil
			}
			pr.(*peer.Peer).Receive(msg.Syn, msg)
			return nil
		}
		if this.handler != nil {
			this.handler(ctx, walletAddr)
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

func (this *Network) GetClientTime(walletAddr string) (uint64, error) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	p, ok := this.peers.Load(walletAddr)
	if !ok {
		return 0, fmt.Errorf("[network GetClientTime] client is nil: %s", walletAddr)
	}
	t := p.(*peer.Peer).ActiveTime()
	ms := t.UnixNano() / int64(time.Millisecond)
	return uint64(ms), nil
}

func (this *Network) IsPeerNetQualityBad(walletAddr string) bool {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	totalFailed := &peer.FailedCount{}
	peerCnt := 0
	var peerFailed *peer.FailedCount
	this.peers.Range(func(key, value interface{}) bool {
		peerCnt++
		peerWalletAddr, _ := key.(string)
		pr, _ := value.(*peer.Peer)
		cnt := pr.GetFailedCnt()
		totalFailed.Dial += cnt.Dial
		totalFailed.Send += cnt.Send
		totalFailed.Recv += cnt.Recv
		totalFailed.Disconnect += cnt.Disconnect
		if peerWalletAddr == walletAddr {
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

func (this *Network) reconnect(walletAddr string) error {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	p, ok := this.peers.Load(walletAddr)
	var pr *peer.Peer
	if !ok {
		return fmt.Errorf("peer not found %s", walletAddr)
	}
	pr = p.(*peer.Peer)
	if len(walletAddr) > 0 && pr.State() == peer.ConnectStateConnecting {
		if err := this.WaitForConnected(walletAddr, time.Duration(common.MAX_WAIT_FOR_CONNECTED_TIMEOUT)*time.Second); err != nil {
			pr.SetState(peer.ConnectStateFailed)
			return err
		}
		pr.SetState(peer.ConnectStateConnected)
		return nil
	}
	pr.SetState(peer.ConnectStateConnecting)
	if len(walletAddr) > 0 {
		state, err := this.GetConnStateByWallet(walletAddr)
		if state == network.PEER_REACHABLE && err == nil {
			pr.SetState(peer.ConnectStateConnected)
			return nil
		}
	}
	peerIds := this.P2p.Bootstrap([]string{pr.GetHostAddr()})
	if len(peerIds) == 0 {
		pr.SetState(peer.ConnectStateFailed)
		return fmt.Errorf("reconnect %s failed, no peer ids", walletAddr)
	}
	peerId := peerIds[0]
	if len(peerId) == 0 {
		return fmt.Errorf("reconnect %s failed, no peer id return", walletAddr)
	}
	pr.SetState(peer.ConnectStateConnected)
	pr.SetPeerId(peerId)
	this.peers.Store(walletAddr, pr)
	return nil
}

func (this *Network) addProxyComponents(builder *network.Builder) {
	if len(this.proxyAddrs) == 0 {
		return
	}
	log.Debugf("enable %t", this.P2p.ProxyModeEnable())
	hasAdd := make(map[string]struct{})
	for _, proxyAddr := range this.proxyAddrs {
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
			if this.P2p.ProxyModeEnable() && len(this.GetProxyServer().PeerID) > 0 {
				this.HealthCheckPeer(this.walletAddrFromPeerId(this.GetProxyServer().PeerID))
			}
			this.peerForHealthCheck.Range(func(key, value interface{}) bool {
				addr, _ := key.(string)
				if len(addr) == 0 {
					return true
				}
				if shouldLog {
					addrState, err := this.GetConnStateByWallet(addr)
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

func (this *Network) updatePeerTime(walletAddr string) {
	if !this.isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	pr := this.GetPeerFromWalletAddr(walletAddr)
	if pr == nil {
		return
	}
	peerId := pr.GetPeerId()
	client := this.P2p.GetPeerClient(peerId)
	if client == nil {
		return
	}
	client.Time = time.Now()
}

func (this *Network) stopKeepAlive() {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	// var ka *keepalive.Component
	// var ok bool
	// for _, info := range this.P2p.Components.GetInstallComponents() {
	// 	ka, ok = info.Component.(*keepalive.Component)
	// 	if ok {
	// 		break
	// 	}
	// }
	// if ka == nil {
	// 	return
	// }
	// // stop keepalive for temporary
	// ka.Cleanup(this.P2p)
	// deleted := this.P2p.Components.Delete(ka)
	// log.Debugf("stop keep alive %t", deleted)
}

func (this *Network) restartKeepAlive() {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	// var ka *keepalive.Component
	// var ok bool
	// for _, info := range this.P2p.Components.GetInstallComponents() {
	// 	ka, ok = info.Component.(*keepalive.Component)
	// 	if ok {
	// 		break
	// 	}
	// }
	// if ka != nil {
	// 	return
	// }
	// options := []keepalive.ComponentOption{
	// 	keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
	// 	keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
	// }
	// ka = keepalive.New(options...)
	// err := this.builder.AddComponent(ka)
	// if err != nil {
	// 	return
	// }
	// ka.Startup(this.P2p)
}

// startProxy. start proxy service
func (this *Network) startProxy(builder *network.Builder) error {
	var err error
	log.Debugf("NATProxyServerAddrs :%v", this.proxyAddrs)
	for _, proxyAddr := range this.proxyAddrs {
		if len(proxyAddr) == 0 {
			continue
		}
		log.Debugf("set proxy mode")
		this.P2p.EnableProxyMode(true)
		this.P2p.SetProxyServer([]network.ProxyServer{
			network.ProxyServer{
				IP: proxyAddr,
			},
		})
		protocol := getProtocolFromAddr(proxyAddr)
		log.Debugf("start proxy will blocking...%s %s, networkId: %d",
			protocol, proxyAddr, this.networkId)
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
			proxyHostAddr, proxyPeerId := this.P2p.GetWorkingProxyServer()
			pr := peer.New(proxyHostAddr)
			pr.SetPeerId(proxyPeerId)
			this.peers.Store(proxyPeerId, pr)
			log.Debugf("start proxy finish, publicAddr: %s, proxy peer id %s", this.P2p.ID.Address, proxyPeerId)
			return nil
		case <-time.After(time.Duration(common.START_PROXY_TIMEOUT) * time.Second):
			err = fmt.Errorf("proxy: %s timeout", proxyAddr)
			log.Debugf("start proxy err :%s", err)
			break
		}
	}
	return err
}

// WaitForConnected. poll to wait for connected
func (this *Network) waitForConnectedByHost(hostAddr string, timeout time.Duration) error {
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		_, ok := this.peers.Load(hostAddr)
		if ok {
			continue
		}
		walletAddr := this.GetWalletFromHostAddr(hostAddr)
		if len(walletAddr) > 0 && this.IsConnReachable(walletAddr) {
			log.Debugf("connection exist %s", hostAddr)
			return nil
		}
		<-time.After(interval)
	}
	return fmt.Errorf("wait for connecting %s timeout", hostAddr)
}

func (this *Network) isValidWalletAddr(walletAddr string) bool {
	if this.GetProxyServer().PeerID == walletAddr {
		return true
	}
	// proxy wallet
	_, err := chainCom.AddressFromBase58(walletAddr)
	return err == nil

}

func getProtocolFromAddr(addr string) string {
	idx := strings.Index(addr, "://")
	if idx == -1 {
		return "tcp"
	}
	return addr[:idx]
}
