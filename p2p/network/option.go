package network

import (
	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/network"
	"github.com/saveio/carrier/types/opcode"
)

type NetworkOption interface {
	apply(n *Network)
}

type NetworkFunc func(n *Network)

func (f NetworkFunc) apply(n *Network) {
	f(n)
}

func WithOpcodes(opCodes map[opcode.Opcode]proto.Message) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.opCodes = opCodes
	})
}

func WitPid(pid *actor.PID) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.pid = pid
	})
}

func WithIntranetIP(intra string) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.intranetIP = intra
	})
}

func WithProxyAddrs(proxy []string) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.proxyAddrs = proxy
	})
}

func WithWalletAddrFromPeerId(walletAddrFromPeerId func(string) string) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.walletAddrFromPeerId = walletAddrFromPeerId
	})
}

func WithNetworkId(networkId uint32) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.networkId = networkId
	})
}

func WithKeys(keys *crypto.KeyPair) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.keys = keys
	})
}

func WithMsgHandler(handler func(*network.ComponentContext, string)) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.handler = handler
	})
}

func WithAsyncRecvMsgDisabled(asyncRecvMsgDisabled bool) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.asyncRecvDisabled = asyncRecvMsgDisabled
	})
}
