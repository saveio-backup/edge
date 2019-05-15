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
	"github.com/saveio/carrier/network/proxy"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/themis/common/log"
)

type Network struct {
	*p2pNet.Component
	proxySvrAddr string
	net          *p2pNet.Network
	key          *crypto.KeyPair
	handler      func(*network.ComponentContext)
}

func NewNetwork() *Network {
	return &Network{}
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
	// return this.listenAddr
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
	builder := network.NewBuilderWithOptions(network.WriteFlushLatency(1 * time.Millisecond))
	if this.key != nil {
		log.Debugf("dsp use account key")
		builder.SetKeys(this.key)
	} else {
		builder.SetKeys(ed25519.RandomKeyPair())
	}
	builder.SetAddress(addr)
	builder.AddComponent(this)
	if len(this.proxySvrAddr) > 0 {
		builder.AddComponent(new(proxy.ProxyComponent))
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
	this.net.BlockUntilListening()
	log.Debugf("++++ dsp will blocking")
	if len(this.proxySvrAddr) > 0 {
		this.net.BlockUntilProxyFinish()
	}
	log.Debugf("+++++++++ proxy:%v addr:%v", this.proxySvrAddr, net.ID.Address)
	return nil
}

func (this *Network) Halt() error {
	if this.net == nil {
		return errors.New("network is down")
	}
	this.net.Close()
	return nil
}

func (this *Network) Dial(addr string) error {
	if this.net == nil {
		return errors.New("network is nil")
	}
	_, err := this.net.Dial(addr)
	return err
}

func (this *Network) Disconnect(addr string) error {
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
	if this.net == nil {
		return false
	}
	return this.net.ConnectionStateExists(addr)
}

func (this *Network) Connect(addr ...string) error {
	for _, a := range addr {
		if this.IsConnectionExists(a) {
			continue
		}
		this.net.Bootstrap(a)
	}
	// this.net.Bootstrap(addr...)
	for _, a := range addr {
		exist := this.net.ConnectionStateExists(a)
		if !exist {
			return errors.New("connection not exist")
		}
	}
	return nil
}

// Send send msg to peer
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, peer interface{}) error {
	client, err := this.loadClient(peer)
	if err != nil {
		return err
	}
	return client.Tell(context.Background(), msg)
}

// Request. send msg to peer and wait for response synchronously
func (this *Network) Request(msg proto.Message, peer interface{}) (proto.Message, error) {
	client, err := this.loadClient(peer)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
	defer cancel()
	return client.Request(ctx, msg)
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, peer interface{}, retry int) (proto.Message, error) {
	client, err := this.loadClient(peer)
	if err != nil {
		return nil, err
	}
	var res proto.Message
	for i := 0; i < retry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
		defer cancel()
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
	wg := sync.WaitGroup{}
	maxRoutines := common.MAX_GOROUTINES_IN_LOOP
	if len(addrs) <= common.MAX_GOROUTINES_IN_LOOP {
		maxRoutines = len(addrs)
	}
	count := 0
	errs := make(map[string]error, 0)
	for _, addr := range addrs {
		wg.Add(1)
		go func(to string) {
			defer wg.Done()
			log.Debugf("broadcast check is exists to %s", to)
			if !this.IsConnectionExists(to) {
				err := this.Connect(to)
				if err != nil {
					errs[to] = err
					return
				}
			}
			var res proto.Message
			var err error
			if !needReply {
				err = this.Send(msg, to)
			} else {
				res, err = this.Request(msg, to)
			}
			if err != nil {
				errs[to] = err
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
		if len(errs) > 0 {
			break
		}
	}
	// wait again if last round count < maxRoutines
	wg.Wait()
	if stop != nil && stop() {
		return nil
	}
	if len(errs) == 0 {
		return nil
	}
	for to, err := range errs {
		log.Errorf("broadcast msg to %s, err %s", to, err)
	}
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
