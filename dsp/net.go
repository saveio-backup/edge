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

func (this *Endpoint) GetNetworkState() (*NetworkStateResp, *DspErr) {
	state := &NetworkStateResp{
		ChainState:        networkStateReachable,
		DNSState:          networkStateReachable,
		DspProxyState:     networkStateReachable,
		ChannelProxyState: networkStateReachable,
	}
	_, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		state.ChainState = networkStateUnReachable
	}
	if this.Dsp.DNS == nil || this.Dsp.DNS.DNSNode == nil {
		state.DNSState = networkStateUnReachable
	} else {
		if !this.channelNet.IsConnectionExists(this.Dsp.DNS.DNSNode.HostAddr) {
			state.DNSState = networkStateUnReachable
		}
	}
	if !this.dspNet.IsConnectionExists(this.dspNet.GetProxyServer()) {
		state.DspProxyState = networkStateUnReachable
	}
	if !this.channelNet.IsConnectionExists(this.channelNet.GetProxyServer()) {
		state.ChannelProxyState = networkStateUnReachable
	}
	return state, nil
}
