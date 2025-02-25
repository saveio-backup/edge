package dsp

import (
	"container/list"
	"math"
	"strconv"
	"time"

	dspConsts "github.com/saveio/dsp-go-sdk/consts"
	"github.com/saveio/edge/common"
	chanCom "github.com/saveio/pylons/common"
	pylons_transfer "github.com/saveio/pylons/transfer"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/common/log"
)

type FilterBlockProgress struct {
	Progress float32
	Start    uint32
	End      uint32
	Now      uint32
}

type ChannelInfo struct {
	ChannelId            uint32
	Balance              uint64
	BalanceFormat        string
	Address              string
	HostAddr             string
	TokenAddr            string
	IsParticipant1Closer bool
	IsParticipant2Closer bool
	IsDNS                bool
	IsOnline             bool
	Selected             bool
	CreatedAt            uint64
}

type ChannelInfosResp struct {
	Balance       uint64
	BalanceFormat string
	Channels      []*ChannelInfo
}

var startChannelHeight uint32

func (this *Endpoint) SetFilterBlockRange() {
	dsp := this.getDsp()
	if dsp == nil {
		return
	}
	currentBlockHeight, err := dsp.GetCurrentBlockHeight()
	startChannelHeight = dsp.GetCurrentFilterBlockHeight()
	log.Debugf("set filter block range %d %s %d %d", currentBlockHeight, err, startChannelHeight)
}

func (this *Endpoint) ResetChannelProgress() {
	log.Debugf("ResetChannelProgress")
	startChannelHeight = 0
}

func (this *Endpoint) IsChannelProcessBlocks() (bool, *DspErr) {
	if this.IsEvmChain() {
		return false, nil
	}
	dsp := this.getDsp()
	if dsp == nil {
		return false, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if !dsp.HasChannelInstance() {
		return false, &DspErr{Code: NO_CHANNEL, Error: ErrMaps[NO_CHANNEL]}
	}
	if dsp.ChannelFirstSyncing() {
		return true, nil
	}
	if !dsp.Running() {
		return false, nil
	}
	filterBlockHeight := dsp.GetCurrentFilterBlockHeight()
	now, getHeightErr := dsp.GetCurrentBlockHeight()
	log.Debugf("IsChannelProcessBlocks filterBlockHeight: %d, now :%d", filterBlockHeight, now)
	if getHeightErr != nil {
		return false, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
	}
	if filterBlockHeight+common.MAX_SYNC_HEIGHT_OFFSET <= now {
		this.SetFilterBlockRange()
		return true, nil
	}
	return false, nil
}

func (this *Endpoint) GetFilterBlockProgress() (*FilterBlockProgress, *DspErr) {
	progress := &FilterBlockProgress{}

	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	endChannelHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		log.Debugf("get channel err %s", err)
		return progress, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: ErrMaps[CHAIN_INTERNAL_ERROR]}
	}
	if this.IsEvmChain() {
		return &FilterBlockProgress{
			Progress: 1,
			Start:    endChannelHeight,
			End:      endChannelHeight,
			Now:      endChannelHeight,
		}, nil
	}

	if endChannelHeight == 0 {
		return progress, nil
	}
	if !dsp.HasChannelInstance() {
		return nil, &DspErr{Code: NO_CHANNEL, Error: ErrMaps[NO_CHANNEL]}
	}
	progress.Start = startChannelHeight
	progress.End = endChannelHeight
	now := dsp.GetCurrentFilterBlockHeight()
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
	log.Debugf("GetFilterBlockProgress start %d, now %d, end %d, progress %v",
		startChannelHeight, now, endChannelHeight, progress)
	return progress, nil
}

func (this *Endpoint) GetAllChannels() (*ChannelInfosResp, *DspErr) {
	log.Debugf("GetAllChannels")
	if this.GetDspAccount() == nil {
		return nil, &DspErr{Code: NO_ACCOUNT, Error: ErrMaps[NO_ACCOUNT]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &ChannelInfosResp{}
	chResp, err := dsp.AllChannels()
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
		log.Debugf("all channel is nil")
	}
	for _, ch := range chResp.Channels {
		hostAddr, _ := dsp.GetExternalIP(ch.Address)
		chInfo, err := dsp.GetChannelInfoFromChain(uint64(ch.ChannelId),
			chainCom.ADDRESS_EMPTY, chainCom.ADDRESS_EMPTY)
		if err != nil {
			log.Errorf("get channel info err %s", err)
		}
		newCh := &ChannelInfo{
			ChannelId:     ch.ChannelId,
			Balance:       ch.Balance,
			BalanceFormat: ch.BalanceFormat,
			Address:       ch.Address,
			HostAddr:      hostAddr,
			TokenAddr:     ch.TokenAddr,
		}
		log.Debugf("get channel %v balance %v format %v", newCh.ChannelId, newCh.Balance, newCh.BalanceFormat)
		if chInfo != nil {
			newCh.IsParticipant1Closer = chInfo.Participant1.IsCloser
			newCh.IsParticipant2Closer = chInfo.Participant2.IsCloser
		}
		infoFromDB, _ := dsp.Channel.GetChannelInfoFromDB(ch.Address)
		if infoFromDB != nil {
			newCh.CreatedAt = infoFromDB.CreatedAt / dspConsts.MILLISECOND_PER_SECOND
			newCh.IsDNS = infoFromDB.IsDNS
		} else {
			dsp.Channel.AddChannelInfo(uint64(ch.ChannelId), ch.Address)
			address, _ := chainCom.AddressFromBase58(ch.Address)
			info, _ := dsp.GetDnsNodeByAddr(address)
			if info != nil {
				dsp.Channel.SetChannelIsDNS(ch.Address, true)
				newCh.IsDNS = true
			}
			newCh.CreatedAt = uint64(time.Now().Unix())
		}
		if dsp.IsDNS(newCh.Address) {
			newCh.Selected = true
		}
		newCh.IsOnline = this.channelNet.IsConnReachable(newCh.Address)
		resp.Channels = append(resp.Channels, newCh)
	}
	log.Debugf("GetAllChannels done: %v", resp)
	return resp, nil
}

func (this *Endpoint) GetDNSHostAddr(walletAddr string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	hostAddr, err := dsp.GetExternalIP(walletAddr)
	if err != nil {
		return "", &DspErr{Code: DSP_CHANNEL_GET_HOSTADDR_ERROR, Error: err}
	}
	return hostAddr, nil
}

func (this *Endpoint) CurrentPaymentChannel() (*ChannelInfo, *DspErr) {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return nil, syncErr
	}
	if syncing {
		return nil, &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	log.Debugf("CurrentPaymentChannel")
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if !dsp.HasDNS() {
		return nil, &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}

	curChannel, err := dsp.GetChannelInfo(dsp.CurrentDNSWallet())
	if err != nil || curChannel == nil {
		return nil, &DspErr{Code: DSP_CHANNEL_GET_ALL_FAILED, Error: ErrMaps[DSP_CHANNEL_GET_ALL_FAILED]}
	}

	chInfo, err := dsp.GetChannelInfoFromChain(uint64(curChannel.ChannelId), chainCom.ADDRESS_EMPTY, chainCom.ADDRESS_EMPTY)
	if err != nil {
		log.Errorf("get channel info err %s", err)
	}
	hostAddr, _ := dsp.GetExternalIP(curChannel.Address)
	resp := &ChannelInfo{
		ChannelId:     curChannel.ChannelId,
		Balance:       curChannel.Balance,
		BalanceFormat: curChannel.BalanceFormat,
		Address:       curChannel.Address,
		HostAddr:      hostAddr,
		TokenAddr:     curChannel.TokenAddr,
		IsDNS:         true,
		Selected:      true,
		IsOnline:      this.channelNet.IsConnReachable(curChannel.Address),
	}
	if chInfo != nil {
		resp.IsParticipant1Closer = chInfo.Participant1.IsCloser
		resp.IsParticipant2Closer = chInfo.Participant2.IsCloser
	}
	return resp, nil
}

func (this *Endpoint) SwitchPaymentChannel(partnerAddr string) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	log.Debugf("SwitchPaymentChannel %s", partnerAddr)
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	exist := dsp.ChannelExist(partnerAddr)
	if !exist {
		return &DspErr{Code: DSP_CHANNEL_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_NOT_EXIST]}
	}

	if !dsp.IsDnsOnline(partnerAddr) {
		return &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}

	dsp.UpdateDNS(partnerAddr, dsp.GetOnlineDNSHostAddr(partnerAddr), true)
	this.notifyIfSwitchChannel()
	return nil
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
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	channelExist := dsp.ChannelExist(partnerAddr)
	if channelExist {
		return 0, &DspErr{Code: DSP_CHANNEL_EXIST, Error: ErrMaps[DSP_CHANNEL_EXIST]}
	}
	balance, err := dsp.BalanceOf(this.getDspWalletAddr())
	if err != nil {
		return 0, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	log.Debugf("OpenPaymentChannel %s, balance %d amount %d",
		partnerAddr, balance, amount+uint64(0.02*math.Pow10(constants.USDT_DECIMALS)))
	if balance == 0 || amount > balance {
		return 0, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	if amount+uint64(0.02*math.Pow10(constants.USDT_DECIMALS)) > balance {
		return 0, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	partnerAddress, err := chainCom.AddressFromBase58(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
	}
	info, _ := dsp.GetDnsNodeByAddr(partnerAddress)
	if info == nil {
		return 0, &DspErr{Code: DSP_CHANNEL_OPEN_TO_NO_DNS, Error: ErrMaps[DSP_CHANNEL_OPEN_TO_NO_DNS]}
	}
	id, err := dsp.OpenChannel(partnerAddr, amount)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_OPEN_FAILED, Error: err}
	}
	hostAddr, _ := dsp.GetExternalIP(partnerAddr)
	dsp.UpdateDNS(partnerAddr, hostAddr, false)
	dsp.Channel.SetChannelIsDNS(partnerAddr, true)
	return id, nil
}

func (this *Endpoint) OpenToAllDNSChannel(amount uint64) *DspErr {
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	onlineDNS := dsp.GetAllOnlineDNS()
	log.Debugf("open to all online dns %v", onlineDNS)
	for walletAddr, _ := range onlineDNS {
		_, err := this.OpenPaymentChannel(walletAddr, amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Endpoint) CloseAllChannel() *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	channels, err := this.GetAllChannels()
	if err != nil {
		return err
	}
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	for _, ch := range channels.Channels {
		err := dsp.CloseChannel(ch.Address)
		log.Debugf("closing channel of %s", ch.Address)
		if err != nil {
			return &DspErr{Code: DSP_CHANNEL_CLOSE_FAILED, Error: err}
		}
	}
	return nil
}

func (this *Endpoint) ClosePaymentChannel(partnerAddr string) *DspErr {
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	_, err := chainCom.AddressFromBase58(partnerAddr)
	if err != nil {
		return &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
	}
	if dsp.GetTaskMgr() != nil && dsp.GetTaskMgr().HasRunningDownloadTask() {
		return &DspErr{Code: DSP_EXIST_ACTIVE_DOWNLOAD_TASK, Error: ErrMaps[DSP_EXIST_ACTIVE_DOWNLOAD_TASK]}
	}
	closeCurDNS := dsp.CurrentDNSWallet() == partnerAddr
	if err := dsp.CloseChannel(partnerAddr); err != nil {
		return &DspErr{Code: DSP_CHANNEL_CLOSE_FAILED, Error: err}
	}
	if !closeCurDNS {
		return nil
	}
	dsp.ResetDNSNode()
	channels, _ := this.GetAllChannels()
	for _, ch := range channels.Channels {
		if ch.Address == partnerAddr {
			continue
		}
		if derr := this.SwitchPaymentChannel(ch.Address); derr != nil {
			return nil
		}
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
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	err2 := dsp.SetDeposit(partnerAddr, bal+realAmount)
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
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if !dsp.HasDNS() {
		return &DspErr{Code: NO_DNS, Error: ErrMaps[NO_DNS]}
	}
	err := dsp.WaitForConnected(dsp.CurrentDNSWallet(),
		time.Duration(dspConsts.WAIT_CHANNEL_CONNECT_TIMEOUT)*time.Second)
	if err != nil {
		log.Errorf("wait channel connected err %s %s", to, err)
		return &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	err = dsp.MediaTransfer(paymentId, amount, dsp.CurrentDNSWallet(), to)
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
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	deposit, err := dsp.GetTotalDepositBalance(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_DEPOSIT_FAILED, Error: err}
	}
	return deposit, nil
}

func (this *Endpoint) QuerySpecialChannelAvaliable(partnerAddr string) (uint64, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.GetAvailableBalance(partnerAddr)
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
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.GetAvailableBalance(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_INTERNAL_ERROR, Error: err}
	}
	if realAmount > bal {
		return &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	totalWithdraw, err := dsp.GetTotalWithdraw(partnerAddr)
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
	success, err := dsp.Withdraw(partnerAddr, totalbal)
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
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	err := dsp.CooperativeSettle(partnerAddr)
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

func (this *Endpoint) QueryChannelByID(idStr, partnerAddr string) (interface{}, *DspErr) {
	id, err := strconv.ParseInt(idStr, 10, 32)
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
