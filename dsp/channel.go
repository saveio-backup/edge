package dsp

import (
	"container/list"
	"strconv"

	chanCom "github.com/saveio/pylons/common"
	pylons_transfer "github.com/saveio/pylons/transfer"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

var startChannelHeight, endChannelHeight uint32

func (this *Endpoint) SetFilterBlockRange() {
	currentBlockHeight, _ := this.Dsp.Chain.GetCurrentBlockHeight()
	startChannelHeight = this.Dsp.Channel.GetCurrentFilterBlockHeight()
	endChannelHeight = currentBlockHeight
}

func (this *Endpoint) ResetChannelProgress() {
	startChannelHeight = 0
	endChannelHeight = 0
}

func (this *Endpoint) GetFilterBlockProgress() (float32, *DspErr) {
	if this.Dsp == nil {
		return 0, nil
	}
	if endChannelHeight == 0 {
		return 0.0, nil
	}
	now := this.Dsp.Channel.GetCurrentFilterBlockHeight()
	if endChannelHeight <= startChannelHeight {
		return 1.0, nil
	}
	rangeHeight := endChannelHeight - startChannelHeight
	if now >= rangeHeight+startChannelHeight {
		return 1.0, nil
	}
	progress := float32(now-startChannelHeight) / float32(rangeHeight)
	log.Debugf("GetFilterBlockProgress start %d, now %d, end %d, progress %f", startChannelHeight, now, endChannelHeight, progress)
	return progress, nil
}

func (this *Endpoint) GetAllChannels() (interface{}, *DspErr) {
	if this.Account == nil {
		return nil, &DspErr{Code: NO_ACCOUNT, Error: ErrMaps[NO_ACCOUNT]}
	}
	return this.Dsp.Channel.AllChannels(), nil
}

//oniChannel api
func (this *Endpoint) OpenPaymentChannel(partnerAddr string) (chanCom.ChannelID, *DspErr) {
	id, err := this.Dsp.Channel.OpenChannel(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_OPEN_FAILED, Error: err}
	}
	return id, nil
}

func (this *Endpoint) ClosePaymentChannel(partnerAddr string) *DspErr {
	err := this.Dsp.Channel.ChannelClose(partnerAddr)
	if err != nil {
		return &DspErr{Code: DSP_CHANNEL_CLOSE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) DepositToChannel(partnerAddr string, realAmount uint64) *DspErr {
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
	err := this.Dsp.Channel.MediaTransfer(paymentId, amount, to)
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
	bal, err := this.Dsp.Channel.GetAvaliableBalance(partnerAddr)
	if err != nil {
		return 0, &DspErr{Code: DSP_CHANNEL_QUERY_AVA_BALANCE_FAILED, Error: err}
	}
	return bal, nil
}

// ChannelWithdraw. withdraw amount of asset from channel
func (this *Endpoint) ChannelWithdraw(partnerAddr string, realAmount uint64) *DspErr {
	bal, derr := this.QuerySpecialChannelDeposit(partnerAddr)
	if derr != nil {
		return derr
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
