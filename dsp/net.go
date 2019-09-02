package dsp

type networkState int

const (
	networkStateUnReachable networkState = iota
	networkStateReachable
)

type NetworkStateResp struct {
	ChainState        networkState
	DNSState          networkState
	DspProxyState     networkState
	ChannelProxyState networkState
}

type ReconnectResp struct {
	IP    string
	Code  uint64
	Error string
}

func (this *Endpoint) GetNetworkState() (*NetworkStateResp, *DspErr) {
	state := &NetworkStateResp{
		ChainState:        networkStateUnReachable,
		DNSState:          networkStateUnReachable,
		DspProxyState:     networkStateUnReachable,
		ChannelProxyState: networkStateUnReachable,
	}
	if this == nil || this.Dsp == nil {
		return state, nil
	}
	_, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err == nil {
		state.ChainState = networkStateReachable
	}
	if this.Dsp.DNS != nil && this.Dsp.DNS.DNSNode != nil && this.channelNet.IsConnectionExists(this.Dsp.DNS.DNSNode.HostAddr) {
		state.DNSState = networkStateReachable
	}
	if this.dspNet != nil && this.dspNet.IsConnectionExists(this.dspNet.GetProxyServer()) {
		state.DspProxyState = networkStateReachable
	}
	if this.channelNet != nil && this.channelNet.IsConnectionExists(this.channelNet.GetProxyServer()) {
		state.ChannelProxyState = networkStateReachable
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
			IP: p,
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
