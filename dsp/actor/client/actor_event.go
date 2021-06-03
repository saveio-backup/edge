package client

import (
	"fmt"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/dsp-go-sdk/utils/async"
	"github.com/saveio/edge/common"
	sdkCom "github.com/saveio/themis-go-sdk/common"
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
	Events   []*sdkCom.SmartContactEvent
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

type NotifyNewTask struct {
	Type     int
	Id       string
	Response chan *NotifyResp
}

type NotifyNetworkState struct {
	Response chan *NotifyResp
}

type NotifyModuleState struct {
	Response chan *NotifyResp
}

type NotifyResp struct {
	Error error
}

func EventNotifyAll() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyAll{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannels() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyChannels{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyInvolvedSmartContract(events []*sdkCom.SmartContactEvent) error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyInvolvedSmartContract{
		Response: make(chan *NotifyResp, 1),
		Events:   events,
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyRevenue() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyRevenue{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyAccount() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyAccount{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifySwitchChannel() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifySwitchChannel{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannelSyncing() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyChannelSyncing{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyChannelProgress() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyChannelProgress{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyUploadTransferList() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyUploadTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyDownloadTransferList() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyDownloadTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyCompleteTransferList() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyCompleteTransferList{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyNetworkState() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyNetworkState{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyModuleState() error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyModuleState{
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}

func EventNotifyNewTask(taskType int, id string) error {
	if EventServerPid == nil {
		return fmt.Errorf("event server has not instance")
	}
	req := &NotifyNewTask{
		Type:     taskType,
		Id:       id,
		Response: make(chan *NotifyResp, 1),
	}
	f := async.TimeoutFunc(func() error {
		EventServerPid.Tell(req)
		resp := <-req.Response
		if resp != nil {
			return resp.Error
		}
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.EVENT_ACTOR_TIMEOUT)*time.Second)
}
