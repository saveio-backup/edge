package server

import (
	"context"
	"errors"
	"fmt"
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
	"net"
)

type MessageHandler func(msgData interface{}, pid *actor.PID)

type P2PActor struct {
	channelNet          *network.Network
	dspNet              *network.Network
	tkActSvr            *tkActServer.TrackerActorServer
	props               *actor.Props
	msgHandlers         map[string]MessageHandler
	localPID            *actor.PID
	dnsHostAddrCallBack func(string) string
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

func (this *P2PActor) SetDNSHostAddrCallback(cb func(string) string) {
	this.dnsHostAddrCallBack = cb
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
	if this.localPID != nil {
		this.localPID.Stop()
		log.Debugf("stop local p2p pid success")
	}
	if this.dspNet != nil {
		this.dspNet.Stop()
		log.Debugf("stop dsp p2p pid success")
	}
	if this.channelNet != nil {
		this.channelNet.Stop()
		log.Debugf("stop channel p2p pid success")
	}
	if this.tkActSvr != nil {
		this.tkActSvr.Stop()
		log.Debugf("stop tracker p2p pid success")
	}
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
			hostAddr := this.dnsHostAddrCallBack(msg.Address)
			msg.Ret.Err = this.channelNet.Connect(hostAddr)
			msg.Ret.Done <- true
		}()
	case *chAct.SendReq:
		go func() {
			log.Debugf("msg.address:[%s]", msg.Address)
			msg.Ret.Err = this.channelNet.SendOnce(msg.Data, msg.Address)
			msg.Ret.Done <- true
		}()
	case *dspAct.ConnectReq:
		go func() {
			err := this.dspNet.Connect(msg.Address)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *dspAct.WaitForConnectedReq:
		go func() {
			err := this.dspNet.WaitForConnected(msg.Address, msg.Timeout)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *chAct.GetNodeNetworkStateReq:
		go func() {
			state, err := this.channelNet.GetConnStateByWallet(msg.Address)
			msg.Ret.State = int(state)
			msg.Ret.Err = err
			msg.Ret.Done <- true
		}()
	case *dspAct.SendReq:
		go func() {
			err := this.dspNet.Send(msg.Data, msg.SessionId, msg.MsgId, msg.Address, msg.SendTimeout)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *dspAct.BroadcastReq:
		go func() {
			var m map[string]error
			var err error
			if msg.Action != nil {
				m, err = this.dspNet.Broadcast(msg.Addresses, msg.Data, msg.SessionId, msg.MsgId, msg.Action)
			} else {
				m, err = this.dspNet.Broadcast(msg.Addresses, msg.Data, msg.SessionId, msg.MsgId)
			}
			msg.Response <- &dspAct.BroadcastResp{Result: m, Error: err}
		}()
	case *dspAct.PublicAddrReq:
		go func() {
			if this.dspNet == nil {
				msg.Response <- &dspAct.PublicAddrResp{Addr: "", Error: errors.New("dsp network is nil")}
				return
			}
			addr := this.dspNet.PublicAddr()
			msg.Response <- &dspAct.PublicAddrResp{Addr: addr}
		}()

	case *dspAct.SendAndWaitReplyReq:
		go func() {
			ret, err := this.dspNet.SendAndWaitReply(msg.Data, msg.SessionId, msg.MsgId, msg.Address, msg.SendTimeout)
			msg.Response <- &dspAct.RequestWithRetryResp{Data: ret, Error: err}
		}()
	case *dspAct.ClosePeerSessionReq:
		go func() {
			err := this.dspNet.ClosePeerSession(msg.Address, msg.SessionId)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *dspAct.PeerSessionSpeedReq:
		go func() {
			tx, rx, err := this.dspNet.GetPeerSessionSpeed(msg.Address, msg.SessionId)
			msg.Response <- &dspAct.P2PSpeedResp{Tx: tx, Rx: rx, Error: err}
		}()

	case *dspAct.ReconnectPeerReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				state, err := this.dspNet.GetConnStateByWallet(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2PResp{Error: nil}
					return
				}
				err = this.dspNet.HealthCheckPeer(msg.Address)
				msg.Response <- &dspAct.P2PResp{Error: err}
			case dspAct.P2PNetTypeChannel:
				state, err := this.channelNet.GetConnStateByWallet(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2PResp{Error: nil}
					return
				}
				err = this.channelNet.HealthCheckPeer(msg.Address)
				msg.Response <- &dspAct.P2PResp{Error: err}
			}
		}()
	case *dspAct.ConnectionExistReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				state, err := this.dspNet.GetConnStateByWallet(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2PBoolResp{Value: true, Error: nil}
					return
				}
				msg.Response <- &dspAct.P2PBoolResp{Value: false, Error: fmt.Errorf("get peer state failed, err %s", err)}
			case dspAct.P2PNetTypeChannel:
				state, err := this.channelNet.GetConnStateByWallet(msg.Address)
				if state == p2pNet.PEER_REACHABLE && err == nil {
					msg.Response <- &dspAct.P2PBoolResp{Value: true, Error: nil}
					return
				}
				msg.Response <- &dspAct.P2PBoolResp{Value: false, Error: fmt.Errorf("get peer state failed, err %s", err)}
			}
		}()
	case *dspAct.CompleteTorrentReq:
		go func() {
			walletAddr, err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2PResp{Error: err}
				return
			}
			log.Debugf("start announce request")
			annResp, err := this.tkActSvr.AnnounceRequestCompleteTorrent(&pm.CompleteTorrentReq{
				InfoHash: msg.Hash,
				Ip:       net.ParseIP(msg.IP),
				Port:     msg.Port,
			}, walletAddr)
			log.Debugf("CompleteTorrentReq announce response: %v, err %v\n", annResp, err)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *dspAct.TorrentPeersReq:
		go func() {
			walletAddr, err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2PStringsResp{Value: nil, Error: err}
				return
			}
			annResp, err := this.tkActSvr.AnnounceRequestTorrentPeers(&pm.GetTorrentPeersReq{
				InfoHash: msg.Hash,
				NumWant:  100,
			}, walletAddr)
			log.Debugf("TorrentPeersReq announce response: %v, err %v\n", annResp, err)
			if err != nil {
				msg.Response <- &dspAct.P2PStringsResp{Value: nil, Error: err}
			} else {
				msg.Response <- &dspAct.P2PStringsResp{Value: annResp.Peers, Error: nil}
			}
		}()
	case *dspAct.EndpointRegistryReq:
		go func() {
			walletAddr, err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				log.Errorf("connect err %v", err)
				msg.Response <- &dspAct.P2PResp{Error: err}
				return
			}
			log.Debugf("AnnounceRequestEndpointRegistry %s", walletAddr)
			annResp, err := this.tkActSvr.AnnounceRequestEndpointRegistry(&pm.EndpointRegistryReq{
				Wallet: msg.WalletAddr[:],
				Ip:     net.ParseIP(msg.IP),
				Port:   msg.Port,
			}, walletAddr)
			log.Debugf("EndpointRegistryReq announce response: %v, err %v\n", annResp, err)
			msg.Response <- &dspAct.P2PResp{Error: err}
		}()
	case *dspAct.GetEndpointReq:
		go func() {
			walletAddr, err := tkActClient.P2pConnect(msg.Address)
			if err != nil {
				msg.Response <- &dspAct.P2PStringResp{Value: "", Error: err}
				return
			}
			log.Debugf("AnnounceRequestGetEndpointAddr dns %v wallet %s", walletAddr, msg.WalletAddr.ToBase58())
			annResp, err := this.tkActSvr.AnnounceRequestGetEndpointAddr(&pm.QueryEndpointReq{
				Wallet: msg.WalletAddr[:],
			}, walletAddr)
			log.Debugf("GetEndpointReq announce response: %v, err %v\n", annResp, err)
			if err != nil {
				msg.Response <- &dspAct.P2PStringResp{Value: "", Error: err}
			} else {
				msg.Response <- &dspAct.P2PStringResp{Value: annResp.Peer, Error: nil}
			}
		}()
	case *dspAct.IsPeerNetQualityBadReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				bad := this.dspNet.IsPeerNetQualityBad(msg.Address)
				msg.Response <- &dspAct.P2PBoolResp{Value: bad}
			case dspAct.P2PNetTypeChannel:
				bad := this.channelNet.IsPeerNetQualityBad(msg.Address)
				msg.Response <- &dspAct.P2PBoolResp{Value: bad}
			}
		}()
	case *dspAct.AppendAddrToHealthCheckReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				this.dspNet.AppendAddrToHealthCheck(msg.Address)
			case dspAct.P2PNetTypeChannel:
				this.channelNet.AppendAddrToHealthCheck(msg.Address)
			}
			msg.Response <- &dspAct.P2PResp{Error: nil}
		}()
	case *dspAct.RemoveAddrFromHealthCheckReq:
		go func() {
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				this.dspNet.RemoveAddrFromHealthCheck(msg.Address)
			case dspAct.P2PNetTypeChannel:
				this.channelNet.RemoveAddrFromHealthCheck(msg.Address)
			}
			msg.Response <- &dspAct.P2PResp{Error: nil}
		}()
	case *dspAct.GetHostAddrReq:
		go func() {
			hostAddr := ""
			switch msg.NetType {
			case dspAct.P2PNetTypeDsp:
				pr := this.dspNet.GetPeerFromWalletAddr(msg.Address)
				if pr != nil {
					hostAddr = pr.GetHostAddr()
				}
			case dspAct.P2PNetTypeChannel:
				pr := this.channelNet.GetPeerFromWalletAddr(msg.Address)
				if pr != nil {
					hostAddr = pr.GetHostAddr()
				}
			}
			msg.Response <- &dspAct.P2PStringResp{Value: hostAddr, Error: nil}
		}()
	default:
		log.Error("[P2PActor] receive unknown message type! %T", msg)
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
