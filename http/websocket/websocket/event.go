package websocket

import (
	"github.com/saveio/edge/http/rest"
)

// PushToNewSubscriber. push events for new subscriber
func (self *WsServer) PushToNewSubscriber() {
	self.PushCurrentAccount()
	self.PushAllChannels()
	self.PushBalance()
	self.PushRevenue()
	self.PushNetworkState()
	self.PushChannelSyncing()
	self.PushChannelInitProgress()
	self.PushCurrentChannel()
	self.PushCurrentUserSpace()
	self.PushUploadingTransferList()
	self.PushDownloadingTransferList()
	self.PushCompleteTransferList()
}

// PushInvolvedSmartContractEvent. push events which involved in smart contract
func (self *WsServer) PushInvolvedSmartContractEvent() {
	self.PushAllChannels()
	self.PushBalance()
	self.PushCurrentChannel()
	self.PushCurrentUserSpace()
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
	self.Broadcast(WS_TOPIC_EVENT, resp)
}
