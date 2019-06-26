package utils

func QueryNode(walletAddr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("nodequery", []interface{}{walletAddr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func RegisterNode(nodeAddr, volume, serviceTime string) ([]byte, error) {
	ret, dErr := sendRpcRequest("registernode", []interface{}{nodeAddr, volume, serviceTime})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func UnregisterNode() ([]byte, error) {
	ret, dErr := sendRpcRequest("unregisternode", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func NodeWithdrawProfit() ([]byte, error) {
	ret, dErr := sendRpcRequest("nodewithdrawprofit", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func NodeUpdate(nodeAddr, volume, serviceTime string) ([]byte, error) {
	ret, dErr := sendRpcRequest("nodeupdate", []interface{}{nodeAddr, volume, serviceTime})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
