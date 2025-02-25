package network

import (
	"github.com/saveio/carrier/network"
	"github.com/saveio/themis/common/log"
)

type NetComponent struct {
	*network.Component
	Net *Network
}

func (this *NetComponent) Startup(net *network.Network) {
}

func (this *NetComponent) Receive(ctx *network.ComponentContext) error {
	msg := ctx.Message()
	client := ctx.Client()
	var addr, peerId string
	if client != nil {
		addr = client.Address
		peerId = client.ClientID()
	}
	log.Debugf("receive msg from addr %s, peer %s", addr, peerId)
	return this.Net.Receive(ctx, msg, addr, peerId)
}
