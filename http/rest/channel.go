package rest

import (
	"math"
	"strconv"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common/constants"
	"github.com/saveio/themis/common/log"
)

func GetAllChannels(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	ret, err := dsp.DspService.GetAllChannels()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func IsChannelSyncing(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	ret, err := dsp.DspService.IsChannelProcessBlocks()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = map[string]interface{}{"Syncing": ret}
	return resp
}

//Handle for channel
func OpenChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	log.Debugf("partnerAddrstr :%T, cmd %v", cmd["Partner"], cmd)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amountStr, ok := cmd["Amount"].(string)
	realAmount := uint64(0)
	if ok {
		amount, err := strconv.ParseFloat(amountStr, 10)
		if err != nil {
			ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		realAmount = uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	_, derr := dsp.DspService.GetAccount(dsp.DspService.GetWalletFilePath(), password)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	id, err := dsp.DspService.OpenPaymentChannel(partnerAddrstr, realAmount)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = id
	return resp
}

//Handle for channel
func CloseChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	_, derr := dsp.DspService.GetAccount(dsp.DspService.GetWalletFilePath(), password)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	err := dsp.DspService.ClosePaymentChannel(partnerAddrstr)
	log.Debugf("close channel %s %s", partnerAddrstr, err)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

func DepositChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amount, err := strconv.ParseFloat(amountstr, 10)
	if err != nil || amount <= 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	_, derr := dsp.DspService.GetAccount(dsp.DspService.GetWalletFilePath(), password)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	derr = dsp.DspService.DepositToChannel(partnerAddrstr, realAmount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	return resp
}

func WithdrawChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amount, err := strconv.ParseFloat(amountstr, 10)
	if err != nil || amount <= 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	_, derr := dsp.DspService.GetAccount(dsp.DspService.GetWalletFilePath(), password)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	derr = dsp.DspService.ChannelWithdraw(partnerAddrstr, realAmount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	return resp
}

func QueryChannelDeposit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	deposit, err := dsp.DspService.QuerySpecialChannelDeposit(partnerAddrstr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = deposit
	return resp
}

func QueryChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	partnerAddrStr, _ := cmd["Partner"].(string)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	resp["Result"] = dsp.DspService.QueryChannel(partnerAddrStr)
	return resp
}

func QueryChannelByID(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	idstr, ok := cmd["Id"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	partnerAddrStr, _ := cmd["Partner"].(string)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	res, err := dsp.DspService.QueryChannelByID(idstr, partnerAddrStr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = res
	return resp
}

func TransferByChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	log.Debugf("cmd %v", cmd)

	toAddrstr, ok := cmd["To"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amount, err := strconv.ParseUint(amountstr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	idstr, ok := cmd["PaymentId"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	paymentId, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	derr := dsp.DspService.MediaTransfer(int32(paymentId), amount, toAddrstr)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	return resp
}

func GetChannelInitProgress(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	progress, err := dsp.DspService.GetFilterBlockProgress()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = progress
	return resp
}
