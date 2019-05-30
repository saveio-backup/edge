package dsp

import (
	"container/list"
	"errors"

	chanCom "github.com/saveio/pylons/common"
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

func (this *Endpoint) GetFilterBlockProgress() (float32, error) {
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

//oniChannel api
func (this *Endpoint) OpenPaymentChannel(partnerAddr string) (chanCom.ChannelID, error) {
	return this.Dsp.Channel.OpenChannel(partnerAddr)
}

func (this *Endpoint) ClosePaymentChannel(regAddr, tokenAddr, partnerAddr string, retryTimeout float64) {
	//[TODO] call channel close function of dsp-go-sdk
	return
}

func (this *Endpoint) DepositToChannel(partnerAddr string, totalDeposit uint64) error {
	return this.Dsp.Channel.SetDeposit(partnerAddr, totalDeposit)
}

func (this *Endpoint) Transfer(paymentId int32, amount uint64, to string) error {
	return this.Dsp.Channel.MediaTransfer(paymentId, amount, to)
}

func (this *Endpoint) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with oniChannel
	return list.New()
}

func (this *Endpoint) QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetTotalDepositBalance(partnerAddr)
}

func (this *Endpoint) QuerySpecialChannelAvaliable(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetAvaliableBalance(partnerAddr)
}

// ChannelWithdraw. withdraw amount of asset from channel
func (this *Endpoint) ChannelWithdraw(partnerAddr string, amount uint64) error {
	totalWithdraw, err := this.Dsp.Channel.GetTotalWithdraw(partnerAddr)
	if err != nil {
		return err
	}
	bal := amount + totalWithdraw
	if bal-amount != totalWithdraw {
		return errors.New("withdraw overflow")
	}
	success, err := this.Dsp.Channel.Withdraw(partnerAddr, bal)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("withdraw failed")
	}
	return nil
}

// ChannelCooperativeSettle. settle channel cooperatively
func (this *Endpoint) ChannelCooperativeSettle(partnerAddr string) error {
	return this.Dsp.Channel.CooperativeSettle(partnerAddr)
}
