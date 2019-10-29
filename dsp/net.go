package dsp

import (
	"fmt"
	"time"
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

func (state *NetworkStateResp) String() string {
	str := ""
	if state == nil {
		return str
	}
	if state.Chain != nil {
		str += fmt.Sprintf("Host: %s, State: %d, UpdatedAt: %d\n", state.Chain.HostAddr, state.Chain.State, state.Chain.UpdatedAt)
	}
	if state.DNS != nil {
		str += fmt.Sprintf("Host: %s, State: %d, UpdatedAt: %d\n", state.DNS.HostAddr, state.DNS.State, state.DNS.UpdatedAt)
	}
	if state.DspProxy != nil {
		str += fmt.Sprintf("Host: %s, State: %d, UpdatedAt: %d\n", state.DspProxy.HostAddr, state.DspProxy.State, state.DspProxy.UpdatedAt)
	}
	if state.ChannelProxy != nil {
		str += fmt.Sprintf("Host: %s, State: %d, UpdatedAt: %d\n", state.ChannelProxy.HostAddr, state.ChannelProxy.State, state.ChannelProxy.UpdatedAt)
	}
	return str
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
		if this.channelNet.IsStateReachable(this.Dsp.DNS.DNSNode.HostAddr) {
			state.DNS.State = networkStateReachable
		} else {
			state.DNS.State = networkStateUnReachable
		}
	}
	if this.dspNet != nil {
		state.DspProxy.HostAddr = this.dspNet.GetProxyServer()
		updatedAt, _ := this.dspNet.GetClientTime(state.DspProxy.HostAddr)
		state.DspProxy.UpdatedAt = updatedAt
		if this.dspNet.IsStateReachable(state.DspProxy.HostAddr) {
			state.DspProxy.State = networkStateReachable
		} else {
			state.DspProxy.State = networkStateUnReachable
		}
	}
	if this.channelNet != nil {
		state.ChannelProxy.HostAddr = this.channelNet.GetProxyServer()
		updatedAt, _ := this.channelNet.GetClientTime(state.ChannelProxy.HostAddr)
		state.ChannelProxy.UpdatedAt = updatedAt
		if this.channelNet.IsStateReachable(state.ChannelProxy.HostAddr) {
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
