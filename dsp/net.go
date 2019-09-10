package dsp

import (
	"time"

	"github.com/saveio/themis/common/log"
)

type networkState int

const (
	networkStateUnReachable networkState = iota
	networkStateReachable
)

type PeerState struct {
	HostAddr  string
	State     networkState
	UpdatedAt uint64
}

type NetworkStateResp struct {
	Chain        *PeerState
	DNS          *PeerState
	DspProxy     *PeerState
	ChannelProxy *PeerState
}

type ReconnectResp struct {
	HostAddr string
	Code     uint64
	Error    string
}

func (this *Endpoint) GetNetworkState() (*NetworkStateResp, *DspErr) {
	state := &NetworkStateResp{
		Chain:        &PeerState{},
		DNS:          &PeerState{},
		DspProxy:     &PeerState{},
		ChannelProxy: &PeerState{},
	}
	if this == nil || this.Dsp == nil {
		return state, nil
	}
	_, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err == nil {
		now := uint64(time.Now().Unix())
		state.Chain.State = networkStateReachable
		state.Chain.UpdatedAt = now
	}
	if this.Dsp.DNS != nil && this.Dsp.DNS.DNSNode != nil {
		state.DNS.HostAddr = this.Dsp.DNS.DNSNode.HostAddr
		updatedAt, _ := this.channelNet.GetClientTime(state.DNS.HostAddr)
		state.DNS.UpdatedAt = updatedAt
		if this.channelNet.IsConnectionExists(this.Dsp.DNS.DNSNode.HostAddr) {
			state.DNS.State = networkStateReachable
		} else {
			state.DNS.State = networkStateUnReachable
		}
	}
	if this.dspNet != nil {
		state.DspProxy.HostAddr = this.dspNet.GetProxyServer()
		updatedAt, _ := this.dspNet.GetClientTime(state.DspProxy.HostAddr)
		state.DspProxy.UpdatedAt = updatedAt
		connected, err := this.dspNet.IsProxyConnectionExists()
		if err != nil {
			log.Error("dsp proxy connection exist err %s", err)
		}
		if connected {
			state.DspProxy.State = networkStateReachable
		} else {
			state.DspProxy.State = networkStateUnReachable
		}
	}
	if this.channelNet != nil {
		state.ChannelProxy.HostAddr = this.channelNet.GetProxyServer()
		updatedAt, _ := this.channelNet.GetClientTime(state.ChannelProxy.HostAddr)
		state.ChannelProxy.UpdatedAt = updatedAt
		connected, err := this.channelNet.IsProxyConnectionExists()
		if err != nil {
			log.Error("channel proxy connection exist err %s", err)
		}
		if connected {
			state.ChannelProxy.State = networkStateReachable
		} else {
			state.ChannelProxy.State = networkStateUnReachable
		}
	}
	return state, nil
}

func (this *Endpoint) ReconnectChannelPeers(peers []string) []*ReconnectResp {
	resp := make([]*ReconnectResp, 0, len(peers))
	for _, p := range peers {
		if p == this.channelNet.GetProxyServer() {
			continue
		}
		res := &ReconnectResp{
			HostAddr: p,
		}
		err := this.channelNet.ReconnectPeer(p)
		if err != nil {
			res.Code = NET_RECONNECT_PEER_FAILED
			res.Error = err.Error()
		}
		resp = append(resp, res)
	}
	return resp
}
