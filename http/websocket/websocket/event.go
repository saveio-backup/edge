package websocket

import (
	"encoding/hex"
	"fmt"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
	sdkCom "github.com/saveio/themis-go-sdk/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

// PushToNewSubscriber. push events for new subscriber
func (self *WsServer) PushToNewSubscriber() {
	self.PushCurrentAccount()
	self.PushAllChannels()
	self.PushBalance()
	self.PushRevenue()
	self.PushNetworkState()
	self.PushModuleState()
	self.PushChannelSyncing()
	self.PushChannelInitProgress()
	self.PushCurrentChannel()
	self.PushCurrentUserSpace()
	self.PushUploadingTransferList()
	self.PushDownloadingTransferList()
	self.PushCompleteTransferList()
}

// PushInvolvedSmartContractEvent. push events which involved in smart contract
func (self *WsServer) PushInvolvedSmartContractEvent(events []*sdkCom.SmartContactEvent) {
	self.PushAllChannels()
	self.PushBalance()
	self.PushCurrentChannel()
	self.PushCurrentUserSpace()
	if len(events) == 0 {
		return
	}
	self.PushSmartContractEvents(events)
}

// PushEvents. push events
func (self *WsServer) PushSmartContractEvents(events []*sdkCom.SmartContactEvent) {
	var state interface{}
	for _, e := range events {
		for _, n := range e.Notify {
			if n.ContractAddress == utils.OntFSContractAddress.ToHexString() {
				state = n.States
				break
			}
		}
	}
	if state == nil {
		return
	}
	resp := rest.GetCurrentAccount(nil)
	resp["Action"] = "smartcontractevents"
	resp["Result"] = state
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushCurrentAccount. push current account when login in new account
func (self *WsServer) PushCurrentAccount() {
	resp := rest.GetCurrentAccount(nil)
	resp["Action"] = "getcurrentaccount"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushAllChannels. push all channel list when open/close channels,
// when bootstrap channels, or channel network state changed
func (self *WsServer) PushAllChannels() {
	resp := rest.GetAllChannels(nil)
	resp["Action"] = "getallchannels"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushBalance. push events when balance changed
func (self *WsServer) PushBalance() {
	resp := rest.GetBalance(nil)
	resp["Action"] = "getbalance"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushRevenue. push revenue when sharing files
func (self *WsServer) PushRevenue() {
	resp := rest.GetFileShareRevenue(nil)
	resp["Action"] = "getfilesharerevenue"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushNetworkState. push networkstate for some required host
func (self *WsServer) PushNetworkState() {
	resp := rest.GetNetworkState(nil)
	resp["Action"] = "networkstate"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

func (self *WsServer) PushModuleState() {
	resp := rest.GetModuleState(nil)
	resp["Action"] = "modulestate"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushChannelSyncing. push channel syncing state when block need to synced
func (self *WsServer) PushChannelSyncing() {
	resp := rest.IsChannelSyncing(nil)
	resp["Action"] = "channelsyncing"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

//  PushChannelInitProgress. push channel init progress
func (self *WsServer) PushChannelInitProgress() {
	resp := rest.GetChannelInitProgress(nil)
	resp["Action"] = "channelinitprogress"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushCurrentChannel. push current channel if selected or channel state changed
func (self *WsServer) PushCurrentChannel() {
	resp := rest.CurrentChannel(nil)
	resp["Action"] = "currentchannel"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushCurrentUserSpace. push current user space if space changed
func (self *WsServer) PushCurrentUserSpace() {
	resp := rest.GetUserSpace(nil)
	resp["Action"] = "getuserspace"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushUploadingTransferList. push uploading transfer list if there are uploading files
func (self *WsServer) PushUploadingTransferList() {
	cmd := map[string]interface{}{
		"Type":   "1",
		"Offset": "0",
		"Limit":  "0",
	}

	resp := rest.GetTransferList(cmd)
	resp["Action"] = "gettransferlist"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushDownloadingTransferList. push downloading transfer list if there are downloading files
func (self *WsServer) PushDownloadingTransferList() {
	cmd := map[string]interface{}{
		"Type":   "2",
		"Offset": "0",
		"Limit":  "0",
	}
	resp := rest.GetTransferList(cmd)
	resp["Action"] = "gettransferlist"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

// PushCompleteTransferList. push complete transfer list
func (self *WsServer) PushCompleteTransferList() {
	cmd := map[string]interface{}{
		"Type":   "0",
		"Offset": "0",
		"Limit":  "0",
	}
	resp := rest.GetTransferList(cmd)
	resp["Action"] = "gettransferlist"
	list, ok := resp["Result"].(*dsp.TransferlistResp)
	if ok {
		for _, t := range list.Transfers {
			log.Debugf("ws notify complete task id %s, file %s,created at %d, updated %d, result %t", t.Id, t.FileHash, t.CreatedAt, t.UpdatedAt, t.Result == nil)
		}
	}
	self.Broadcast(WS_TOPIC_EVENT, resp)
}

func (self *WsServer) NotifyNewTask(transferType int, id string) {
	cmd := map[string]interface{}{
		"Type": fmt.Sprintf("%v", transferType),
		"Id":   hex.EncodeToString([]byte(id)),
	}
	resp := rest.GetTransferDetail(cmd)
	resp["Action"] = "newtask"
	self.Broadcast(WS_TOPIC_EVENT, resp)
}
