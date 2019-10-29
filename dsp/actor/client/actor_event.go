package client

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
)

var EventServerPid *actor.PID

func SetEventPid(eventPid *actor.PID) {
	EventServerPid = eventPid
}

type NotifyChannels struct {
	Response chan *NotifyResp
}

type NotifyAll struct {
	Response chan *NotifyResp
}

type NotifyInvolvedSmartContract struct {
	Response chan *NotifyResp
}

type NotifyRevenue struct {
	Response chan *NotifyResp
}

type NotifyAccount struct {
	Response chan *NotifyResp
}

type NotifySwitchChannel struct {
	Response chan *NotifyResp
}

type NotifyChannelSyncing struct {
	Response chan *NotifyResp
}

type NotifyChannelProgress struct {
	Response chan *NotifyResp
}

type NotifyUploadTransferList struct {
	Response chan *NotifyResp
}

type NotifyDownloadTransferList struct {
	Response chan *NotifyResp
}

type NotifyCompleteTransferList struct {
	Response chan *NotifyResp
}

type NotifyNetworkState struct {
	Response chan *NotifyResp
}

type NotifyResp struct {
	Error error
}

func EventNotifyAll() error {
	req := &NotifyAll{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannels() error {
	req := &NotifyChannels{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyInvolvedSmartContract() error {
	req := &NotifyInvolvedSmartContract{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyRevenue() error {
	req := &NotifyRevenue{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyAccount() error {
	req := &NotifyAccount{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifySwitchChannel() error {
	req := &NotifySwitchChannel{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannelSyncing() error {
	req := &NotifyChannelSyncing{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannelProgress() error {
	req := &NotifyChannelProgress{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyUploadTransferList() error {
	req := &NotifyUploadTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyDownloadTransferList() error {
	req := &NotifyDownloadTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyCompleteTransferList() error {
	req := &NotifyCompleteTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyNetworkState() error {
	req := &NotifyNetworkState{
		Response: make(chan *NotifyResp, 1),
	}
	f := utils.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}
