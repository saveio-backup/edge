package rest

import (
	"math"
	"strconv"

	berr "github.com/saveio/edge/http/base/error"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/pylons/transfer"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/constants"
)

func GetAllChannels(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	acc := DspService.Chain.Native.Fs.DefAcc
	if acc == nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = DspService.Dsp.Channel.AllChannels()
	return resp
}

//Handle for channel
func OpenChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	id, err := DspService.OpenPaymentChannel(partnerAddrstr)
	if err != nil {
		resp["Result"] = 0
	} else {
		resp["Result"] = id
	}
	return resp
}

func DepositChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amount, err := strconv.ParseFloat(amountstr, 10)
	if err != nil || amount <= 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	bal, err := DspService.QuerySpecialChannelDeposit(partnerAddrstr)
	if err != nil {
		return ResponsePackWithErrMsg(berr.CHANNEL_DEPOSIT_ERROR, err.Error())
	}
	err = DspService.DepositToChannel(partnerAddrstr, bal+realAmount)
	if err != nil {
		return ResponsePackWithErrMsg(berr.CHANNEL_DEPOSIT_ERROR, err.Error())
	}
	return resp
}

func WithdrawChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amount, err := strconv.ParseFloat(amountstr, 10)
	if err != nil || amount <= 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	bal, err := DspService.QuerySpecialChannelDeposit(partnerAddrstr)
	if err != nil {
		return ResponsePackWithErrMsg(berr.CHANNEL_WITHDRAW_ERROR, err.Error())
	}
	if realAmount > bal {
		return ResponsePackWithErrMsg(berr.CHANNEL_WITHDRAW_ERROR, "insuffience balance")
	}
	err = DspService.ChannelWithdraw(partnerAddrstr, realAmount)
	if err != nil {
		return ResponsePackWithErrMsg(berr.CHANNEL_WITHDRAW_ERROR, err.Error())
	}
	return resp
}

func QueryChannelDeposit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	partnerAddrstr, ok := cmd["Partner"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	deposit, err := DspService.QuerySpecialChannelDeposit(partnerAddrstr)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	} else {
		resp["Result"] = deposit
	}

	return resp
}

func QueryChannel(cmd map[string]interface{}) map[string]interface{} {
	getBalance := func(sender *transfer.NettingChannelEndState, receiver *transfer.NettingChannelEndState) chanCom.Balance {
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

	resp := ResponsePack(berr.SUCCESS)

	partnerAddrStr, ok := cmd["Partner"].(string)
	if ok {
		//filter by partner if partner addr is valid
		_, err := chainCom.AddressFromBase58(partnerAddrStr)
		if err == nil {
			matchPartner = true
		}
	}

	channels := make(map[int]interface{})
	ChanList := DspService.GetChannelListByOwnerAddress("", "")
	for channel := ChanList.Front(); channel != nil; channel = channel.Next() {
		channelState := channel.Value.(*transfer.NettingChannelState)
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

	resp["Result"] = channels
	return resp
}

func QueryChannelByID(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	idstr, ok := cmd["Id"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	id, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	res := QueryChannel(cmd)
	result := res["Result"]
	channels := result.(map[int]interface{})
	if channel, exist := channels[int(id)]; exist {
		resp["Result"] = channel
	} else {
		resp["Result"] = make(map[string]interface{})
	}

	return resp
}

func TransferByChannel(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	toAddrstr, ok := cmd["To"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amountstr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amount, err := strconv.ParseUint(amountstr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	idstr, ok := cmd["PaymentId"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	paymentId, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	DspService.Transfer(int32(paymentId), amount, toAddrstr)
	return resp
}
