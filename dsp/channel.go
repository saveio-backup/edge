package dsp

import (
	"container/list"
	"strconv"
	"time"

	dsp "github.com/saveio/dsp-go-sdk"
	"github.com/saveio/dsp-go-sdk/common"
	chanCom "github.com/saveio/pylons/common"
	pylons_transfer "github.com/saveio/pylons/transfer"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

type FilterBlockProgress struct {
	Progress float32
	Start    uint32
	End      uint32
	Now      uint32
}

type ChannelInfo struct {
	ChannelId         uint32
	Balance           uint64
	BalanceFormat     string
	Address           string
	HostAddr          string
	TokenAddr         string
	Participant1State int
	ParticiPant2State int
	IsDNS             bool
	Connected         bool
}

type ChannelInfosResp struct {
	Balance       uint64
	BalanceFormat string
	Channels      []*ChannelInfo
}

var startChannelHeight uint32

func (this *Endpoint) SetFilterBlockRange() {
	currentBlockHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	startChannelHeight = this.Dsp.Channel.GetCurrentFilterBlockHeight()
	// endChannelHeight = currentBlockHeight
	log.Debugf("set filter block range %d %s %d %d", currentBlockHeight, err, startChannelHeight)
}

func (this *Endpoint) ResetChannelProgress() {
	log.Debugf("ResetChannelProgress")
	startChannelHeight = 0
	// endChannelHeight = 0
}

func (this *Endpoint) GetFilterBlockProgress() (*FilterBlockProgress, *DspErr) {
	progress := &FilterBlockProgress{}
	if this.Dsp == nil {
		return progress, nil
	}
	endChannelHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		log.Debugf("get channel err %s", err)
		return progress, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: ErrMaps[CHAIN_INTERNAL_ERROR]}
	}
	if endChannelHeight == 0 {
		return progress, nil
	}
	progress.Start = startChannelHeight
	progress.End = endChannelHeight
	now := this.Dsp.Channel.GetCurrentFilterBlockHeight()
	progress.Now = now
	log.Debugf("endChannelHeight %d, start %d", endChannelHeight, startChannelHeight)
	if endChannelHeight <= startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	rangeHeight := endChannelHeight - startChannelHeight
	if now >= rangeHeight+startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	p := float32(now-startChannelHeight) / float32(rangeHeight)
	progress.Progress = p
	log.Debugf("GetFilterBlockProgress start %d, now %d, end %d, progress %v", startChannelHeight, now, endChannelHeight, progress)
	return progress, nil
}

func (this *Endpoint) GetAllChannels() (interface{}, *DspErr) {
	log.Debugf("GetAllChannels")
	if this.Account == nil {
		return nil, &DspErr{Code: NO_ACCOUNT, Error: ErrMaps[NO_ACCOUNT]}
	}
	if this.Dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &ChannelInfosResp{}
	chResp, err := this.Dsp.Channel.AllChannels()
	if err != nil {
		log.Errorf("get all channels err %s", err)
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if chResp == nil {
		log.Debugf("dsp get all channel is nil")
		return resp, nil
	}
	resp.Balance = chResp.Balance
	resp.BalanceFormat = chResp.BalanceFormat
	resp.Channels = make([]*ChannelInfo, 0)
	if len(chResp.Channels) == 0 {
		log.Errorf("all channel is nil")
	}
	for _, ch := range chResp.Channels {
		hostAddr, _ := this.Dsp.GetExternalIP(ch.Address)
		chInfo, err := this.Dsp.Chain.Native.Channel.GetChannelInfo(uint64(ch.ChannelId), chainCom.ADDRESS_EMPTY, chainCom.ADDRESS_EMPTY)
		if err != nil {
			log.Errorf("get channel info err %s", err)
		}
		state1 := 1
		if chInfo != nil && chInfo.Participant1.IsCloser {
			state1 = 0
		}
		state2 := 1
		if chInfo != nil && chInfo.Participant2.IsCloser {
			state2 = 0
		}
		newCh := &ChannelInfo{
			ChannelId:         ch.ChannelId,
			Balance:           ch.Balance,
			BalanceFormat:     ch.BalanceFormat,
			Address:           ch.Address,
			HostAddr:          hostAddr,
			TokenAddr:         ch.TokenAddr,
			Participant1State: state1,
			ParticiPant2State: state2,
		}
		// TODO: feature to be discussed
		// address, err := chainCom.AddressFromBase58(ch.Address)
		// if err != nil {
		// 	resp.Channels = append(resp.Channels, newCh)
		// 	log.Errorf("parse from base58 err %s", err)
		// 	continue
		// }
		// info, err := this.Dsp.Chain.Native.Dns.GetDnsNodeByAddr(address)
		// if err != nil || info == nil {
		// 	resp.Channels = append(resp.Channels, newCh)
		// 	log.Errorf("err %s, info is nil %v", err, info)
		// 	continue
		// }
		// newCh.IsDNS = true
		if this.Dsp.DNS.DNSNode != nil && newCh.Address == this.Dsp.DNS.DNSNode.WalletAddr {
			newCh.Connected = true
		}
		resp.Channels = append(resp.Channels, newCh)
	}
	log.Debugf("GetAllChannels done: %s", resp)
	return resp, nil
}

//oniChannel api
func (this *Endpoint) OpenPaymentChannel(partnerAddr string, amount uint64) (chanCom.ChannelID, *DspErr) {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return 0, syncErr
	}
	if syncing {
		return 0, &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	log.Debugf("OpenPaymentChannel %s", partnerAddr)
	channelExist := this.Dsp.Channel.ChannelExist(partnerAddr)
	if !channelExist {
		return 0, &DspErr{Code: DSP_CHANNEL_EXIST, Error: ErrMaps[DSP_CHANNEL_EXIST]}
	}
	dnsUrl, err := this.Dsp.GetExternalIP(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_DNS_GET_EXTERNALIP_FAILED, Error: err}
	}
	err = this.Dsp.Channel.SetHostAddr(partnerAddr, dnsUrl)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	// err = this.Dsp.Channel.WaitForConnected(partnerAddr, time.Duration(common.WAIT_CHANNEL_CONNECT_TIMEOUT)*time.Second)
	// log.Debugf("openchannel wait for connected dns %s, partneraddr %s, err %s", dnsUrl, partnerAddr, err)
	// if err != nil {
	// 	log.Errorf("wait channel connected err %s %s", partnerAddr, err)
	// 	return 0, &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	// }
	id, err := this.Dsp.Channel.OpenChannel(partnerAddr, amount)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_OPEN_FAILED, Error: err}
	}

	url, ok := this.Dsp.DNS.OnlineDNS[partnerAddr]
	if ok {
		this.Dsp.DNS.DNSNode = &dsp.DNSNodeInfo{
			WalletAddr: partnerAddr,
			HostAddr:   url,
		}
	}
	return id, nil
}

func (this *Endpoint) ClosePaymentChannel(partnerAddr string) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	err := this.Dsp.Channel.ChannelClose(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_CLOSE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) DepositToChannel(partnerAddr string, realAmount uint64) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	bal, err := this.QuerySpecialChannelDeposit(partnerAddr)
	if err != nil {
		return err
	}
	err2 := this.Dsp.Channel.SetDeposit(partnerAddr, bal+realAmount)
	if err2 != nil {
		return &DspErr{Code: DSP_CHANNEL_DEPOSIT_FAILED, Error: err2}
	}
	return nil
}

func (this *Endpoint) MediaTransfer(paymentId int32, amount uint64, to string) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	url, err := this.Dsp.GetExternalIP(to)
	if err != nil {
		return &DspErr{Code: DSP_DNS_GET_EXTERNALIP_FAILED, Error: err}
	}
	err = this.Dsp.Channel.SetHostAddr(to, url)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	err = this.Dsp.Channel.WaitForConnected(to, time.Duration(common.WAIT_CHANNEL_CONNECT_TIMEOUT)*time.Second)
	if err != nil {
		log.Errorf("wait channel connected err %s %s", to, err)
		return &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	err = this.Dsp.Channel.MediaTransfer(paymentId, amount, to)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_MEDIATRANSFER_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with oniChannel
	return list.New()
}

func (this *Endpoint) QuerySpecialChannelDeposit(partnerAddr string) (uint64, *DspErr) {
	deposit, err := this.Dsp.Channel.GetTotalDepositBalance(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_DEPOSIT_FAILED, Error: err}
	}
	return deposit, nil
}

func (this *Endpoint) QuerySpecialChannelAvaliable(partnerAddr string) (uint64, *DspErr) {
	bal, err := this.Dsp.Channel.GetAvailableBalance(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_QUERY_AVA_BALANCE_FAILED, Error: err}
	}
	return bal, nil
}

// ChannelWithdraw. withdraw amount of asset from channel
func (this *Endpoint) ChannelWithdraw(partnerAddr string, realAmount uint64) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	bal, err := this.Dsp.Channel.GetAvailableBalance(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	if realAmount > bal {
		return &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	totalWithdraw, err := this.Dsp.Channel.GetTotalWithdraw(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_WITHDRAW_FAILED, Error: err}
	}
	totalbal := realAmount + totalWithdraw
	if totalbal-realAmount != totalWithdraw {
		return &DspErr{Code: DSP_CHANNEL_WITHDRAW_OVERFLOW, Error: err}
	}
	if totalbal == 0 {
		return &DspErr{Code: DSP_CHANNEL_WITHDRAW_WRONG_AMOUNT, Error: ErrMaps[DSP_CHANNEL_WITHDRAW_WRONG_AMOUNT]}
	}
	success, err := this.Dsp.Channel.Withdraw(partnerAddr, totalbal)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_WITHDRAW_FAILED, Error: err}
	}
	if !success {
		return &DspErr{Code: DSP_CHANNEL_WITHDRAW_FAILED, Error: ErrMaps[DSP_CHANNEL_WITHDRAW_FAILED]}
	}
	return nil
}

// ChannelCooperativeSettle. settle channel cooperatively
func (this *Endpoint) ChannelCooperativeSettle(partnerAddr string) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	err := this.Dsp.Channel.CooperativeSettle(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_CO_SETTLE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) QueryChannel(partnerAddrStr string) map[int]interface{} {
	getBalance := func(sender *pylons_transfer.NettingChannelEndState, receiver *pylons_transfer.NettingChannelEndState) chanCom.Balance {
		var senderTransferredAmount, receiverTransferredAmount chanCom.TokenAmount

		if sender.BalanceProof != nil {
			senderTransferredAmount = sender.BalanceProof.TransferredAmount
		}

		if receiver.BalanceProof != nil {
			receiverTransferredAmount = receiver.BalanceProof.TransferredAmount
		}

		result := sender.ContractBalance - senderTransferredAmount + receiverTransferredAmount
		return (chanCom.Balance)(result)
	}
	var matchPartner bool
	if len(partnerAddrStr) > 0 {
		//filter by partner if partner addr is valid
		_, err := chainCom.AddressFromBase58(partnerAddrStr)
		if err == nil {
			matchPartner = true
		}
	}

	channels := make(map[int]interface{})
	ChanList := this.GetChannelListByOwnerAddress("", "")
	for channel := ChanList.Front(); channel != nil; channel = channel.Next() {
		channelState := channel.Value.(*pylons_transfer.NettingChannelState)
		ourState := channelState.GetChannelEndState(0)
		partnerState := channelState.GetChannelEndState(1)
		partnerAddr := chainCom.Address(partnerState.GetAddress())
		if matchPartner && partnerAddrStr != partnerAddr.ToBase58() {
			continue
		}

		channels[int(channelState.GetIdentifier())] = map[string]interface{}{
			"channelID":                int(channelState.GetIdentifier()),
			"depositToContract":        ourState.GetContractBalance(),
			"ourBalance":               getBalance(ourState, partnerState),
			"partnerDepositToContract": partnerState.GetContractBalance(),
			"partnerBalance":           getBalance(partnerState, ourState),
			"partnerAddr":              partnerAddr.ToBase58(),
		}
	}

	return channels
}

func (this *Endpoint) QueryChannelByID(idstr, partnerAddr string) (interface{}, *DspErr) {
	id, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	channels := this.QueryChannel(partnerAddr)
	if channel, exist := channels[int(id)]; exist {
		return channel, nil
	} else {
		return make(map[string]interface{}), nil
	}
}
