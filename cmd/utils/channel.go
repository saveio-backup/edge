package utils

import "github.com/saveio/themis/common/log"

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

// func ClosePaymentChannel(partnerAddr string) error {
// 	ret, dErr := sendRpcRequest("", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }

func DepositToChannel(partnerAddr, realAmount, password string) (interface{}, error) {
	ret, dErr := sendRpcRequest("depositchannel", []interface{}{partnerAddr, realAmount, password})
	if dErr != nil {
		return nil, dErr.Error
	}
	log.Debugf("ret %v", ret)
	return ret, nil
}

// func MediaTransfer(paymentId int32, amount uint64, to string) error {
// 	ret, dErr := sendRpcRequest("transferbychannel", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }
// func GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {

// }
// func QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
// 	ret, dErr := sendRpcRequest("querychanneldeposit", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }
// func QuerySpecialChannelAvaliable(partnerAddr string) (uint64, error) {
// 	ret, dErr := sendRpcRequest("", []interface{}{})
// 	if dErr != nil {
// 		return nil, dErr.Error
// 	}

// }
// func ChannelWithdraw(partnerAddr string, realAmount uint64) error {
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
