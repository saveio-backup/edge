package dsp

import (
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/actor/client"
)

type EventHub struct {
	lastNotifyHeight    uint32
	lastDNSChannel      string
	isChannelSyncing    bool
	channelFilterHeight uint32
	completeTaskCount   int
	netstate            string
}

func NewEventHub() *EventHub {
	h := &EventHub{}
	return h
}

// notifyWhenStartup. notify events when start up
func (this *Endpoint) notifyWhenStartup() {
	if !config.WsEnabled() {
		return
	}
	client.EventNotifyAll()
}

func (this *Endpoint) notifyAccountLogout() {
	if !config.WsEnabled() {
		return
	}
	client.EventNotifyAccount()
}

func (this *Endpoint) notifyChannelProgress() {

	syncing, _ := this.IsChannelProcessBlocks()
	if this.eventHub.isChannelSyncing != syncing {
		this.eventHub.isChannelSyncing = syncing
		client.EventNotifyChannelSyncing()
	}

	progress, _ := this.GetFilterBlockProgress()
	if this.eventHub.channelFilterHeight != progress.Now {
		this.eventHub.channelFilterHeight = progress.Now
		client.EventNotifyChannelProgress()
	}
}

func (this *Endpoint) notifyNewSmartContractEvent() {
	if !config.WsEnabled() {
		return
	}

	currentHeight, _ := this.Dsp.GetCurrentBlockHeight()
	if this.eventHub.lastNotifyHeight >= currentHeight {
		return
	}

	event, err := this.GetAccountSmartContractEventByBlock(currentHeight)
	if err != nil {
		return
	}
	this.eventHub.lastNotifyHeight = currentHeight
	if event == nil {
		return
	}
	client.EventNotifyInvolvedSmartContract()
}

func (this *Endpoint) notifyIfSwitchChannel() {
	if !config.WsEnabled() {
		return
	}
	if !this.Dsp.HasDNS() {
		if len(this.eventHub.lastDNSChannel) == 0 {
			return
		}
		client.EventNotifySwitchChannel()
		return
	}
	if this.eventHub.lastDNSChannel == this.Dsp.CurrentDNSWallet() {
		return
	}
	this.eventHub.lastDNSChannel = this.Dsp.CurrentDNSWallet()
	client.EventNotifySwitchChannel()
}

func (this *Endpoint) notifyUploadingTransferList() {
	if !config.WsEnabled() {
		return
	}

	client.EventNotifyUploadTransferList()

	resp := this.GetTransferList(transferTypeComplete, 0, 0)
	completeCount := len(resp.Transfers)

	if completeCount == this.eventHub.completeTaskCount {
		return
	}
	this.eventHub.completeTaskCount = completeCount

	client.EventNotifyCompleteTransferList()
}

func (this *Endpoint) notifyDownloadingTransferList() {
	if !config.WsEnabled() {
		return
	}

	client.EventNotifyDownloadTransferList()
	resp := this.GetTransferList(transferTypeComplete, 0, 0)
	completeCount := len(resp.Transfers)

	if completeCount == this.eventHub.completeTaskCount {
		return
	}
	this.eventHub.completeTaskCount = completeCount

	client.EventNotifyCompleteTransferList()
}

func (this *Endpoint) notifyNewNetworkState() {
	if !config.WsEnabled() {
		return
	}
	state, _ := this.GetNetworkState()
	if this.eventHub.netstate == state.String() {
		return
	}
	this.eventHub.netstate = state.String()
	client.EventNotifyNetworkState()
	client.EventNotifyChannels()
}
