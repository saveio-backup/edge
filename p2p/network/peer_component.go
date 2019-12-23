package network

import (
	"github.com/saveio/carrier/network"
	"github.com/saveio/edge/p2p/peer"
	"github.com/saveio/themis/common/log"
)

type PeerComponent struct {
	Net *Network
}

func (this *PeerComponent) Startup(net *network.Network) {
}

func (this *PeerComponent) Cleanup(net *network.Network) {
}

func (this *PeerComponent) Receive(ctx *network.ComponentContext) error {
	return nil
}

func (this *PeerComponent) PeerConnect(client *network.PeerClient) {
	if client == nil || len(client.Address) == 0 {
		log.Warnf("peer has connected, but client is nil", client)
		return
	}
	if this.Net.IsProxyAddr(client.Address) {
		return
	}
	p, ok := this.Net.peers.LoadOrStore(client.Address, peer.New(client.Address))
	pr, ok := p.(*peer.Peer)
	if !ok {
		log.Errorf("convert peer to peer.Peer failed")
		return
	}
	log.Infof("peer %s has connected", client.Address)
	pr.SetClient(client)
}

func (this *PeerComponent) PeerDisconnect(client *network.PeerClient) {

}
