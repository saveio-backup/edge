package server

import (
	"context"
	"fmt"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	dspAct "github.com/saveio/dsp-go-sdk/actor/client"
	"github.com/saveio/edge/p2p/network"
	chAct "github.com/saveio/pylons/actor/client"
	pm "github.com/saveio/scan/p2p/actor/messages"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
	tkActServer "github.com/saveio/scan/p2p/actor/tracker/server"
	"github.com/saveio/themis/common/log"
)

type MessageHandler func(msgData interface{}, pid *actor.PID)

type P2PActor struct {
	channelNet  *network.Network
	dspNet      *network.Network
	tkActSvr    *tkActServer.TrackerActorServer
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

func (this *P2PActor) SetChannelNetwork(net *network.Network) {
	this.channelNet = net
}

func (this *P2PActor) SetDspNetwork(net *network.Network) {
	this.dspNet = net
}

func (this *P2PActor) SetTrackerNet(tk *tkActServer.TrackerActorServer) {
	this.tkActSvr = tk
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

func (this *P2PActor) Stop() error {
	this.localPID.Stop()
	this.dspNet.Stop()
	this.channelNet.Stop()
	return nil
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
	case *chAct.ConnectReq:
		go func() {
			msg.Ret.Err = this.channelNet.Connect(msg.Address)
			msg.Ret.Done <- true
		}()
	case *chAct.CloseReq:
		go func() {
			msg.Ret.Err = this.channelNet.Close(msg.Address)
			msg.Ret.Done <- true
		}()
	case *chAct.SendReq:
		go func() {
			msg.Ret.Err = this.channelNet.Send(msg.Data, msg.Address)
			msg.Ret.Done <- true
		}()
	case *dspAct.ConnectReq:
		go func() {
			err := this.dspNet.ConnectAndWait(msg.Address)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	// case *dspAct.ChannelWaitForConnectedReq:
	// 	go func() {
	// 		err := this.channelNet.WaitForConnected(msg.Address, msg.Timeout)
	// 		msg.Response <- &dspAct.P2pResp{Error: err}
	// 	}()
	case *dspAct.WaitForConnectedReq:
		go func() {
			err := this.dspNet.WaitForConnected(msg.Address, msg.Timeout)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	case *chAct.GetNodeNetworkStateReq:
		go func() {
			state, err := this.channelNet.GetPeerStateByAddress(msg.Address)
			msg.Ret.State = int(state)
			msg.Ret.Err = err
			msg.Ret.Done <- true
		}()
	case *dspAct.CloseReq:
		go func() {
			err := this.dspNet.Disconnect(msg.Address)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	case *dspAct.SendReq:
		go func() {
			err := this.dspNet.Send(msg.Data, msg.Address)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	case *dspAct.BroadcastReq:
		go func() {
			m, err := this.dspNet.Broadcast(msg.Addresses, msg.Data, msg.NeedReply, msg.Action)
			msg.Response <- &dspAct.BroadcastResp{Result: m, Error: err}
		}()
	case *dspAct.PublicAddrReq:
		go func() {
			addr := this.dspNet.PublicAddr()
			msg.Response <- &dspAct.PublicAddrResp{Addr: addr}
		}()
	case *dspAct.RequestWithRetryReq:
		go func() {
			ret, err := this.dspNet.RequestWithRetry(msg.Data, msg.Address, msg.Retry, msg.Timeout)
			msg.Response <- &dspAct.RequestWithRetryResp{Data: ret, Error: err}
		}()
	case *dspAct.ReconnectPeerReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2pNetTypeDsp:
				state, err := this.dspNet.GetPeerStateByAddress(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2pResp{Error: nil}
					return
				}
				err = this.dspNet.ReconnectPeer(msg.Address)
				msg.Response <- &dspAct.P2pResp{Error: err}
			case dspAct.P2pNetTypeChannel:
				state, err := this.channelNet.GetPeerStateByAddress(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2pResp{Error: nil}
					return
				}
				err = this.channelNet.ReconnectPeer(msg.Address)
				msg.Response <- &dspAct.P2pResp{Error: err}
			}
		}()
	case *dspAct.ConnectionExistReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2pNetTypeDsp:
				state, err := this.dspNet.GetPeerStateByAddress(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2pBoolResp{Value: true, Error: nil}
					return
				}
				msg.Response <- &dspAct.P2pBoolResp{Value: false, Error: fmt.Errorf("get peer state failed, err %s", err)}
			case dspAct.P2pNetTypeChannel:
				state, err := this.channelNet.GetPeerStateByAddress(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2pBoolResp{Value: true, Error: nil}
					return
				}
				msg.Response <- &dspAct.P2pBoolResp{Value: false, Error: fmt.Errorf("get peer state failed, err %s", err)}
			}
		}()
	case *dspAct.CompleteTorrentReq:
		go func() {
			err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2pResp{Error: err}
				return
			}
			log.Debugf("start announce request")
			annResp, err := this.tkActSvr.AnnounceRequestCompleteTorrent(&pm.CompleteTorrentReq{
				InfoHash: msg.Hash,
				Ip:       net.ParseIP(msg.IP),
				Port:     msg.Port,
			}, msg.Address)
			log.Debugf("CompleteTorrentReq announce response: %v, err %v\n", annResp, err)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	case *dspAct.TorrentPeersReq:
		go func() {
			err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2pStringSliceResp{Value: nil, Error: err}
				return
			}
			annResp, err := this.tkActSvr.AnnounceRequestTorrentPeers(&pm.GetTorrentPeersReq{
				InfoHash: msg.Hash,
				NumWant:  100,
			}, msg.Address)
			log.Debugf("TorrentPeersReq announce response: %v, err %v\n", annResp, err)
			if err != nil {
				msg.Response <- &dspAct.P2pStringSliceResp{Value: nil, Error: err}
			} else {
				msg.Response <- &dspAct.P2pStringSliceResp{Value: annResp.Peers, Error: nil}
			}
		}()
	case *dspAct.EndpointRegistryReq:
		go func() {
			err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2pResp{Error: err}
				return
			}
			annResp, err := this.tkActSvr.AnnounceRequestEndpointRegistry(&pm.EndpointRegistryReq{
				Wallet: msg.WalletAddr[:],
				Ip:     net.ParseIP(msg.IP),
				Port:   msg.Port,
			}, msg.Address)
			log.Debugf("EndpointRegistryReq announce response: %v, err %v\n", annResp, err)
			msg.Response <- &dspAct.P2pResp{Error: err}
		}()
	case *dspAct.GetEndpointReq:
		go func() {
			err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2pStringResp{Value: "", Error: err}
				return
			}
			annResp, err := this.tkActSvr.AnnounceRequestGetEndpointAddr(&pm.QueryEndpointReq{
				Wallet: msg.WalletAddr[:],
			}, msg.Address)
			log.Debugf("GetEndpointReq announce response: %v, err %v\n", annResp, err)
			if err != nil {
				msg.Response <- &dspAct.P2pStringResp{Value: "", Error: err}
			} else {
				msg.Response <- &dspAct.P2pStringResp{Value: annResp.Peer, Error: nil}
			}
		}()
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
