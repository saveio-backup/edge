package dsp

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/carrier/network"
	p2pNet "github.com/saveio/carrier/network"
	"github.com/saveio/carrier/network/components/backoff"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/themis/common/log"
)

type Network struct {
	proxySvrAddr  string
	net           *p2pNet.Network
	key           *crypto.KeyPair
	handler       func(*network.ComponentContext)
	peerStateChan chan *keepalive.PeerStateEvent
	ActivePeers   *sync.Map
}

func NewNetwork() *Network {
	n := &Network{}
	n.ActivePeers = new(sync.Map)
	n.peerStateChan = make(chan *keepalive.PeerStateEvent, 16)
	return n
}

func (this *Network) SetHandler(handler func(*network.ComponentContext)) {
	this.handler = handler
}

func (this *Network) SetNetworkKey(key *crypto.KeyPair) {
	this.key = key
}

func (this *Network) SetProxyServer(proxySvr string) {
	this.proxySvrAddr = proxySvr
}

func (this *Network) ListenAddr() string {
	return this.net.ID.Address
}

// PublicAddr address
func (this *Network) PublicAddr() string {
	return this.net.ID.Address
}

func (this *Network) Protocol() string {
	idx := strings.Index(this.PublicAddr(), "://")
	if idx == -1 {
		return "tcp"
	}
	return this.PublicAddr()[:idx]
}

func (this *Network) Receive(ctx *network.ComponentContext) error {
	log.Debugf("dsp network receive msg")
	if this.handler != nil {
		this.handler(ctx)
	}
	return nil
}

func (this *Network) Start(addr string) error {
	if this.net != nil {
		return fmt.Errorf("already listening at %s", addr)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	opcode.RegisterMessageType(opcode.Opcode(common.MSG_OP_CODE), &pb.Message{})
	protocolIndex := strings.Index(addr, "://")
	if protocolIndex == -1 {
		return errors.New("invalid address")
	}
	protocol := addr[:protocolIndex]
	builder := network.NewBuilderWithOptions(network.WriteFlushLatency(1 * time.Millisecond))
	builder.SetAddress(addr)
	if this.key != nil {
		log.Debugf("dsp use account key")
		builder.SetKeys(this.key)
	} else {
		log.Debugf("dsp use radom key")
		builder.SetKeys(ed25519.RandomKeyPair())
	}

	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
		keepalive.WithPeerStateChan(this.peerStateChan),
	}
	builder.AddComponent(keepalive.New(options...))
	backoffOptions := []backoff.ComponentOption{
		backoff.WithMaxAttempts(100), //try again times;
	}
	builder.AddComponent(backoff.New(backoffOptions...))
	netComponent := new(NetComponent)
	netComponent.Net = this
	builder.AddComponent(netComponent)
	if len(this.proxySvrAddr) > 0 {
		if protocol == "udp" {
			builder.AddComponent(new(proxy.UDPProxyComponent))
		} else if protocol == "kcp" {
			builder.AddComponent(new(proxy.KCPProxyComponent))
		}
	}
	net, err := builder.Build()
	if err != nil {
		return err
	}
	this.net = net
	log.Debugf("proxy %v, list %v", this.proxySvrAddr, addr)
	if len(this.proxySvrAddr) > 0 {
		this.net.SetProxyServer(this.proxySvrAddr)
	}
	go this.net.Listen()
	go this.PeerStateChange(this.syncPeerState)
	this.net.BlockUntilListening()
	log.Debugf("++++ dsp will blocking %s", this.proxySvrAddr)
	if len(this.proxySvrAddr) > 0 {
		if protocol == "udp" {
			this.net.BlockUntilUDPProxyFinish()
		} else if protocol == "kcp" {
			this.net.BlockUntilKCPProxyFinish()
		}
	}
	if len(net.ID.Address) == 6 {
		return errors.New("invalid address")
	}
	log.Debugf("+++++++++ proxy:%v addr:%v", this.proxySvrAddr, net.ID.Address)
	return nil
}

func (this *Network) PeerStateChange(fn func(*keepalive.PeerStateEvent)) {
	log.Debugf("[DSPNetwork] PeerStateChange")
	ka, reg := this.net.Component(keepalive.ComponentID)
	if !reg {
		log.Error("keepalive component do not reg")
		return
	}
	peerStateChan := ka.(*keepalive.Component).GetPeerStateChan()
	stopCh := ka.(*keepalive.Component).GetStopChan()
	for {
		select {
		case event := <-peerStateChan:
			fn(event)

		case <-stopCh:
			return

		}
	}
}
func (this *Network) syncPeerState(state *keepalive.PeerStateEvent) {
	log.Debugf("[syncPeerState] addr: %s state: %v", state.Address, state.State)
	switch state.State {
	case keepalive.PEER_REACHABLE:
		log.Debugf("[syncPeerState] addr: %s state: NetworkReachable\n", state.Address)
		this.ActivePeers.LoadOrStore(state.Address, struct{}{})
	case keepalive.PEER_UNKNOWN:
		log.Debugf("[syncPeerState] addr: %s state: PEER_UNKNOWN\n", state.Address)
	case keepalive.PEER_READY:
		log.Debugf("[syncPeerState] addr: %s state: Ready\n", state.Address)
	case keepalive.PEER_UNREACHABLE:
		this.ActivePeers.Delete(state.Address)
		log.Debugf("[syncPeerState] addr: %s state: NetworkUnreachable\n", state.Address)
	}
}

func (this *Network) Stop() error {
	log.Debugf("[DSPNetwork] Stop")
	if this.net == nil {
		return errors.New("network is down")
	}
	this.net.Close()
	return nil
}

func (this *Network) Dial(addr string) error {
	log.Debugf("[DSPNetwork] Dial")
	if this.net == nil {
		return errors.New("network is nil")
	}
	_, err := this.net.Dial(addr)
	return err
}

func (this *Network) Disconnect(addr string) error {
	log.Debugf("[DSPNetwork] Disconnet")
	if this.net == nil {
		return errors.New("network is nil")
	}
	peer, err := this.net.Client(addr)
	if err != nil {
		return err
	}
	return peer.Close()
}

// IsPeerListenning. check the peer is listening or not.
func (this *Network) IsPeerListenning(addr string) bool {
	log.Debugf("[DSPNetwork] IsPeerList")
	if this.net == nil {
		return false
	}
	err := this.Dial(addr)
	if err != nil {
		return false
	}
	err = this.Disconnect(addr)
	if err != nil {
		return false
	}
	return true
}

func (this *Network) IsConnectionExists(addr string) bool {
	log.Debugf("[DSPNetwork] IsConnectionExists")
	if this.net == nil {
		return false
	}
	return this.net.ConnectionStateExists(addr)
}

func (this *Network) Connect(addr ...string) error {
	log.Debugf("DSP Connect 20 %s", addr[0])

	for _, a := range addr {
		if this.IsConnectionExists(a) {
			log.Debugf("connection exist %s", a)
			continue
		}
		log.Debugf("bootstrap to %v", a)
		this.net.Bootstrap(a)
	}
	for _, a := range addr {
		err := this.WaitForConnected(a, time.Duration(10)*time.Second)
		if err != nil {
			return err
		}
		// exist := this.net.ConnectionStateExists(a)
		// if !exist {
		// 	return errors.New("connection not exist")
		// }
	}
	return nil
}

func (this *Network) WaitForConnected(addr string, timeout time.Duration) error {
	log.Debugf("[DSPNetwork] WaitForConnected")
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		_, ok := this.ActivePeers.Load(addr)
		log.Debugf("active peer %s ok %t", addr, ok)
		if ok {
			return nil
		}
		<-time.After(interval)
	}
	return errors.New("wait for connected timeout")
}

// Send send msg to peer
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, peer interface{}) error {
	log.Debugf("[DSPNetwork] send")
	client, err := this.loadClient(peer)
	if err != nil {
		return err
	}
	log.Debugf("loop active peers")
	this.ActivePeers.Range(func(key, value interface{}) bool {
		log.Debugf("before send, active peers has %s", key, value)
		return false
	})
	log.Debugf("send msg to %v", peer)
	// ctx := network.WithSignMessage(context.Background(), true)
	return client.Tell(context.Background(), msg)
}

// Request. send msg to peer and wait for response synchronously
func (this *Network) Request(msg proto.Message, peer interface{}) (proto.Message, error) {
	log.Debugf("[DSPNetwork] Request")
	client, err := this.loadClient(peer)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
	defer cancel()
	// ctx2 := network.WithSignMessage(ctx, true)
	return client.Request(ctx, msg)
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, peer interface{}, retry int) (proto.Message, error) {
	log.Debugf("[DSPNetwork] Request with retry")
	client, err := this.loadClient(peer)
	if err != nil {
		return nil, err
	}
	var res proto.Message
	for i := 0; i < retry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
		defer cancel()
		// ctx2 := network.WithSignMessage(ctx, true)
		res, err = client.Request(ctx, msg)
		if err == nil || err.Error() != "context deadline exceeded" {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Broadcast. broadcast same msg to peers. Handle action if send msg success.
// If one msg is sent failed, return err. But the previous success msgs can not be recalled.
// action(responseMsg, responseToAddr).
func (this *Network) Broadcast(addrs []string, msg proto.Message, needReply bool, stop func() bool, action func(proto.Message, string)) error {
	log.Debugf("[DSPNetwork] broadcast")
	wg := sync.WaitGroup{}
	maxRoutines := common.MAX_GOROUTINES_IN_LOOP
	if len(addrs) <= common.MAX_GOROUTINES_IN_LOOP {
		maxRoutines = len(addrs)
	}
	count := 0
	errs := new(sync.Map)
	errCount := 0
	for _, addr := range addrs {
		wg.Add(1)
		go func(to string) {
			defer wg.Done()
			log.Debugf("broadcast check is exists to %s", to)
			if !this.IsConnectionExists(to) {
				log.Debugf("not exist, connecting %v", to)
				err := this.Connect(to)
				if err != nil {
					errCount++
					errs.Store(to, err)
					return
				}
			} else {
				log.Debugf("connection exists")
			}
			var res proto.Message
			var err error
			if !needReply {
				err = this.Send(msg, to)
			} else {
				res, err = this.Request(msg, to)
			}
			if err != nil {
				errCount++
				errs.Store(to, err)
				return
			}
			if action != nil {
				action(res, to)
			}
		}(addr)
		count++
		if count >= maxRoutines {
			wg.Wait()
			// reset, start new round
			count = 0
		}
		if stop != nil && stop() {
			break
		}
		if errCount > 0 {
			break
		}
	}
	// wait again if last round count < maxRoutines
	wg.Wait()
	if stop != nil && stop() {
		return nil
	}
	if errCount == 0 {
		return nil
	}
	errs.Range(func(to, err interface{}) bool {
		log.Errorf("broadcast msg to %v, err %v", to, err)
		return false
	})
	return errors.New("broadcast failed")
}

func (this *Network) loadClient(peer interface{}) (*network.PeerClient, error) {
	addr, ok := peer.(string)
	if ok {
		client, err := this.net.Client(addr)
		if err != nil {
			return nil, err
		}
		if client == nil {
			return nil, errors.New("client is nil")
		}
		return client, nil
	}
	client, ok := peer.(*network.PeerClient)
	if !ok || client == nil {
		return nil, errors.New("invalid peer type")
	}
	return client, nil
}
