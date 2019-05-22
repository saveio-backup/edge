package dsp

import "github.com/saveio/carrier/network"

type NetComponent struct {
	*network.Component
	Net *Network
}

func (this *NetComponent) Startup(net *network.Network) {
}

func (this *NetComponent) Receive(ctx *network.ComponentContext) error {
	return this.Net.Receive(ctx)
}
