package dsp

import (
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/actor/client"
	"github.com/saveio/themis-go-sdk/usdt"
	"github.com/saveio/themis/common/log"
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
	log.Debugf("event notify all")
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
	if this.eventHub.lastNotifyHeight == 0 {
		this.eventHub.lastNotifyHeight = currentHeight
		log.Debugf("first set up %d", currentHeight)
	}
	if this.eventHub.lastNotifyHeight > currentHeight {
		return
	}
	if currentHeight-this.eventHub.lastNotifyHeight > 100 {
		this.eventHub.lastNotifyHeight = currentHeight
		client.EventNotifyInvolvedSmartContract()
		return
	}
	if this.Dsp == nil {
		return
	}
	log.Debugf("this.eventHub.lastNotifyHeight %d, current %d", this.eventHub.lastNotifyHeight, currentHeight)
	events, err := this.Dsp.GetSmartContractEventByEventIdAndHeights(usdt.USDT_CONTRACT_ADDRESS.ToBase58(),
		this.Dsp.WalletAddress(), 0, this.eventHub.lastNotifyHeight, currentHeight+1)
	this.eventHub.lastNotifyHeight = currentHeight
	if err != nil {
		return
	}
	if len(events) == 0 {
		return
	}
	log.Debugf("notifyNewSmartContractEvent from %d-%d", this.eventHub.lastNotifyHeight, currentHeight)
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
