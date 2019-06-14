package utils

import (
	"encoding/json"
	"fmt"
)

func GetFilterBlockProgress() ([]byte, error) {
	ret, dErr := sendRpcRequest("getchannelinitprogress", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetAllChannels() ([]byte, error) {
	ret, dErr := sendRpcRequest("getallchannels", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func OpenPaymentChannel(partnerAddr string) (map[string]string, error) {
	ret, dErr := sendRpcRequest("openchannel", []interface{}{partnerAddr})
	if dErr != nil {
		return nil, dErr.Error
	}
	data := map[string]string{"Id": string(ret)}
	return data, nil
}

func ClosePaymentChannel(partnerAddr string) error {
	_, dErr := sendRpcRequest("closechannel", []interface{}{partnerAddr})
	if dErr != nil {
		return dErr.Error
	}
	return nil
}

func DepositToChannel(partnerAddr, realAmount, password string) (interface{}, error) {
	ret, dErr := sendRpcRequest("depositchannel", []interface{}{partnerAddr, realAmount, password})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func WithdrawChannel(partnerAddr, realAmount, password string) (interface{}, error) {
	ret, dErr := sendRpcRequest("withdrawchannel", []interface{}{partnerAddr, realAmount, password})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func MediaTransfer(paymentId, amount, to string) error {
	ret, dErr := sendRpcRequest("transferbychannel", []interface{}{to, amount, paymentId})
	fmt.Printf("ret:%v, dErr %v", ret, dErr)
	if dErr != nil {
		return dErr.Error
	}
	return nil
}
func QuerySpecialChannelDeposit(partnerAddr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("querychanneldeposit", []interface{}{partnerAddr})
	if dErr != nil {
		return nil, dErr.Error
	}
	data := make(map[string]string, 0)
	data["Amount"] = string(ret)
	bufs, _ := json.Marshal(data)
	return bufs, nil
}

// func QuerySpecialChannelAvaliable(partnerAddr string) (uint64, error) {
// 	ret, dErr := sendRpcRequest("", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }

// func ChannelCooperativeSettle(partnerAddr string) error {
// 	ret, dErr := sendRpcRequest("", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }
// func QueryChannel(partnerAddrStr string) map[int]interface{} {

// }
// func QueryChannelByID(idstr, partnerAddr string) (interface{}, error) {
// 	ret, dErr := sendRpcRequest("", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }
