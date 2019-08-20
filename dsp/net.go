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
