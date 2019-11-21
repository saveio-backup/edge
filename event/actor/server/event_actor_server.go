package server

import (
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	edgeCli "github.com/saveio/edge/dsp/actor/client"
	"github.com/saveio/edge/http/websocket"
	"github.com/saveio/themis/common/log"
)

type EventActorServer struct {
	props    *actor.Props
	localPID *actor.PID
}

func NewEventActorServer() (*EventActorServer, error) {
	var err error
	eventActor := &EventActorServer{}
	eventActor.localPID, err = eventActor.Start()
	if err != nil {
		return nil, err
	}
	return eventActor, nil
}

func (this *EventActorServer) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "event_server")
	if err != nil {
		return nil, fmt.Errorf("[P2PActor] start error:%v", err)
	}
	this.localPID = localPid
	return localPid, err
}

func (this *EventActorServer) Stop() error {
	if this.localPID != nil {
		this.localPID.Stop()
	}
	return nil
}

func (this *EventActorServer) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[P2PActor] actor restarting")
	case *actor.Stopping:
		log.Warn("[P2PActor] actor stopping")
	case *actor.Stopped:
		log.Warn("[P2PActor] actor stopped")
	case *actor.Started:
		log.Debug("[P2PActor] actor started")
	case *actor.Restart:
		log.Warn("[P2PActor] actor restart")
	case *edgeCli.NotifyAll:
		go func() {
			websocket.Server().PushToNewSubscriber()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyChannels:
		go func() {
			websocket.Server().PushAllChannels()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyInvolvedSmartContract:
		go func() {
			websocket.Server().PushInvolvedSmartContractEvent()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyRevenue:
		go func() {
			websocket.Server().PushRevenue()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyAccount:
		go func() {
			websocket.Server().PushCurrentAccount()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifySwitchChannel:
		go func() {
			websocket.Server().PushCurrentChannel()
			websocket.Server().PushNetworkState()
			websocket.Server().PushAllChannels()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyChannelSyncing:
		go func() {
			websocket.Server().PushChannelSyncing()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyChannelProgress:
		go func() {
			log.Debugf("push msg...")
			websocket.Server().PushChannelInitProgress()
			msg.Response <- &edgeCli.NotifyResp{}
			log.Debugf("push msg...done")
		}()
	case *edgeCli.NotifyUploadTransferList:
		go func() {
			websocket.Server().PushUploadingTransferList()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyDownloadTransferList:
		go func() {
			websocket.Server().PushDownloadingTransferList()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyCompleteTransferList:
		go func() {
			websocket.Server().PushCompleteTransferList()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	case *edgeCli.NotifyNetworkState:
		go func() {
			websocket.Server().PushNetworkState()
			msg.Response <- &edgeCli.NotifyResp{}
		}()
	default:
		log.Errorf("[P2PActor] receive unknown message type! %v", msg)
	}
}

func (this *EventActorServer) GetLocalPID() *actor.PID {
	return this.localPID
}
