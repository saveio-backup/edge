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
		websocket.Server().PushToNewSubscriber()
	case *edgeCli.NotifyChannels:
		websocket.Server().PushAllChannels()
	case *edgeCli.NotifyInvolvedSmartContract:
		websocket.Server().PushInvolvedSmartContractEvent()
	case *edgeCli.NotifyRevenue:
		websocket.Server().PushRevenue()
	case *edgeCli.NotifyAccount:
		websocket.Server().PushCurrentAccount()
	case *edgeCli.NotifySwitchChannel:
		websocket.Server().PushCurrentChannel()
		websocket.Server().PushNetworkState()
		websocket.Server().PushAllChannels()
	case *edgeCli.NotifyChannelSyncing:
		websocket.Server().PushChannelSyncing()
	case *edgeCli.NotifyChannelProgress:
		websocket.Server().PushChannelInitProgress()
	case *edgeCli.NotifyUploadTransferList:
		websocket.Server().PushUploadingTransferList()
	case *edgeCli.NotifyDownloadTransferList:
		websocket.Server().PushDownloadingTransferList()
	case *edgeCli.NotifyCompleteTransferList:
		websocket.Server().PushCompleteTransferList()
	case *edgeCli.NotifyNetworkState:
		websocket.Server().PushNetworkState()
	default:
		log.Errorf("[P2PActor] receive unknown message type! %v", msg)
	}
}

func (this *EventActorServer) GetLocalPID() *actor.PID {
	return this.localPID
}
