package dsp

import (
	"fmt"
	"time"

	"github.com/saveio/themis/common/log"
)

type networkState int

const (
	networkStateUnReachable networkState = iota
	networkStateReachable
	networkStateDisable
)

type PeerState struct {
	HostAddr  string
	State     networkState
	UpdatedAt uint64
}

type NetworkStateResp struct {
	Chain    *PeerState
	DNS      *PeerState
	DspProxy *PeerState
}

func (state *NetworkStateResp) String() string {
	str := ""
	if state == nil {
		return str
	}
	if state.Chain != nil {
		str += fmt.Sprintf("Host: %s, State: %d\n", state.Chain.HostAddr, state.Chain.State)
	}
	if state.DNS != nil {
		str += fmt.Sprintf("Host: %s, State: %d\n", state.DNS.HostAddr, state.DNS.State)
	}
	if state.DspProxy != nil {
		str += fmt.Sprintf("Host: %s, State: %d\n", state.DspProxy.HostAddr, state.DspProxy.State)
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
		Chain:    &PeerState{},
		DNS:      &PeerState{},
		DspProxy: &PeerState{},
	}
	if this == nil {
		return state, nil
	}
	dsp := this.getDsp()
	if dsp == nil {
		return state, nil
	}
	_, err := dsp.GetCurrentBlockHeight()
	if err == nil {
		now := uint64(time.Now().Unix())
		state.Chain.State = networkStateReachable
		state.Chain.UpdatedAt = now
	}
	if !dsp.HasChannelInstance() {
		state.DNS.State = networkStateDisable
	}
	if dsp.HasDNS() && dsp.ChannelRunning() && dsp.Running() {
		state.DNS.HostAddr = dsp.CurrentDNSHostAddr()
		updatedAt, _ := this.channelNet.GetClientTime(dsp.CurrentDNSWallet())
		state.DNS.UpdatedAt = updatedAt
		if this.channelNet.IsConnReachable(dsp.CurrentDNSWallet()) {
			state.DNS.State = networkStateReachable
		} else {
			state.DNS.State = networkStateUnReachable
		}
	}
	if this.dspNet != nil {
		log.Debugf("this.dspNet.GetProxyServer().PeerID +++ %s", this.dspNet.GetProxyServer().PeerID)
	}
	if this.dspNet != nil && len(this.dspNet.GetProxyServer().PeerID) > 0 {
		state.DspProxy.HostAddr = this.dspNet.GetProxyServer().IP
		log.Debugf("peer id %s", this.dspNet.GetProxyServer().PeerID)
		updatedAt, _ := this.dspNet.GetClientTime(this.dspNet.WalletAddrFromPeerId(this.dspNet.GetProxyServer().PeerID))
		state.DspProxy.UpdatedAt = updatedAt
		if this.dspNet.IsConnReachable(this.dspNet.WalletAddrFromPeerId(this.dspNet.GetProxyServer().PeerID)) {
			state.DspProxy.State = networkStateReachable
		} else {
			state.DspProxy.State = networkStateUnReachable
		}
	}

	return state, nil
}

func (this *Endpoint) ReconnectChannelPeers(peers []string) []*ReconnectResp {
	resp := make([]*ReconnectResp, 0, len(peers))
	for _, p := range peers {
		res := &ReconnectResp{
			HostAddr: p,
		}
		err := this.channelNet.HealthCheckPeer(this.channelNet.GetWalletFromHostAddr(p))
		if err != nil {
			res.Code = NET_RECONNECT_PEER_FAILED
			res.Error = err.Error()
		}
		resp = append(resp, res)
	}
	return resp
}
