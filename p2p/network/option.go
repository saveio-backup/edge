package network

import (
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/network"
)

type NetworkOption interface {
	apply(n *Network)
}

type NetworkFunc func(n *Network)

func (f NetworkFunc) apply(n *Network) {
	f(n)
}

func WithIntranetIP(intra string) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.intranetIP = intra
	})
}

func WithProxyAddr(proxy string) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.proxyAddr = proxy
	})
}

func WithKeys(keys *crypto.KeyPair) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.keys = keys
	})
}

func WithMsgHandler(handler func(*network.ComponentContext)) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.handler = handler
	})
}

func WithAsyncRecvMsgDisabled(asyncRecvMsgDisabled bool) NetworkOption {
	return NetworkFunc(func(n *Network) {
		n.asyncRecvDisabled = asyncRecvMsgDisabled
	})
}
