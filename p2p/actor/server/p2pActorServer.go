package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	dspact "github.com/saveio/dsp-go-sdk/actor/client"
	"github.com/saveio/edge/p2p/networks/channel"
	"github.com/saveio/edge/p2p/networks/dsp"
	chact "github.com/saveio/pylons/actor/client"
	"github.com/saveio/themis/common/log"
)

type MessageHandler func(msgData interface{}, pid *actor.PID)

type P2PActor struct {
	channelNet  *channel.Network
	dspNet      *dsp.Network
	props       *actor.Props
	msgHandlers map[string]MessageHandler
	localPID    *actor.PID
}

func NewP2PActor() (*P2PActor, error) {
	var err error
	p2pActor := &P2PActor{
		msgHandlers: make(map[string]MessageHandler),
	}
	p2pActor.localPID, err = p2pActor.Start()
	if err != nil {
		return nil, err
	}
	return p2pActor, nil
}

func (this *P2PActor) SetChannelNetwork(net *channel.Network) {
	this.channelNet = net
}

func (this *P2PActor) SetDspNetwork(net *dsp.Network) {
	this.dspNet = net
}

func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "net_server")
	if err != nil {
		return nil, fmt.Errorf("[P2PActor] start error:%v", err)
	}
	this.localPID = localPid
	return localPid, err
}

func (this *P2PActor) Receive(ctx actor.Context) {
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
	case *chact.ConnectReq:
		err := this.channelNet.Connect(msg.Address)
		ctx.Sender().Request(&chact.P2pResp{Error: err}, ctx.Self())
	case *chact.CloseReq:
		err := this.channelNet.Close(msg.Address)
		ctx.Sender().Request(&chact.P2pResp{Error: err}, ctx.Self())
	case *chact.SendReq:
		err := this.channelNet.Send(msg.Data, msg.Address)
		ctx.Sender().Request(&chact.P2pResp{Error: err}, ctx.Self())
	case *dspact.ConnectReq:
		err := this.dspNet.Connect(msg.Address)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
	case *dspact.CloseReq:
		err := this.dspNet.Disconnect(msg.Address)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
	case *dspact.SendReq:
		err := this.dspNet.Send(msg.Data, msg.Address)
		log.Debugf("[p2pActor] send msg to %s, err %s", msg.Address, err)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
	case *dspact.BroadcastReq:
		err := this.dspNet.Broadcast(msg.Addresses, msg.Data, msg.NeedReply, msg.Stop, msg.Action)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
	case *dspact.PeerListeningReq:
		ret := this.dspNet.IsPeerListenning(msg.Address)
		var err error
		if !ret {
			err = errors.New("peer is not listening")
		}
		log.Debugf("is peer listening %s, ret %t, err %s", msg.Address, ret, err)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
	case *dspact.PublicAddrReq:
		addr := this.dspNet.PublicAddr()
		log.Debugf("receive PublicIPReq msg: %v", addr)
		ctx.Sender().Request(&dspact.PublicAddrResp{Addr: addr}, ctx.Self())
	case *dspact.RequestWithRetryReq:
		ret, err := this.dspNet.RequestWithRetry(msg.Data, msg.Address, msg.Retry)
		ctx.Sender().Request(&dspact.RequestWithRetryResp{Data: ret, Error: err}, ctx.Self())
	default:
		log.Error("[P2PActor] receive unknown message type!")
	}
}

func (this *P2PActor) Broadcast(message proto.Message) {
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.channelNet.P2p.Broadcast(ctx, message)
}

func (this *P2PActor) RegMsgHandler(msgName string, handler MessageHandler) {
	this.msgHandlers[msgName] = handler
}

func (this *P2PActor) UnRegMsgHandler(msgName string, handler MessageHandler) {
	delete(this.msgHandlers, msgName)
}

func (this *P2PActor) GetLocalPID() *actor.PID {
	return this.localPID
}
