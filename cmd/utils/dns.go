package utils

func RegisterUrl(url, link string) ([]byte, error) {
	ret, dErr := sendRpcRequest("registerurl", []interface{}{url, link})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func BindUrl(url, link string) ([]byte, error) {
	ret, dErr := sendRpcRequest("bindurl", []interface{}{url, link})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

type LinkResp struct {
	Link string
}

func QueryLink(url string) ([]byte, error) {
	ret, dErr := sendRpcRequest("querylink", []interface{}{url})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func RegisterDns(ip, port, deposit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("registerdns", []interface{}{ip, port, deposit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func UnregisterDns() ([]byte, error) {
	ret, dErr := sendRpcRequest("unregisterdns", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QuitDns() ([]byte, error) {
	ret, dErr := sendRpcRequest("quitdns", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func AddPos(amount string) ([]byte, error) {
	ret, dErr := sendRpcRequest("addpos", []interface{}{amount})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func ReducePos(amount string) ([]byte, error) {
	ret, dErr := sendRpcRequest("reducepos", []interface{}{amount})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QueryRegInfos() ([]byte, error) {
	ret, dErr := sendRpcRequest("queryreginfos", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QueryRegInfo(pubKey string) ([]byte, error) {
	ret, dErr := sendRpcRequest("queryreginfo", []interface{}{pubKey})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QueryHostInfos() ([]byte, error) {
	ret, dErr := sendRpcRequest("queryhostinfos", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QueryHostInfo(addr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("queryhostinfo", []interface{}{addr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func QueryPublicIP(addr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("querypublicip", []interface{}{addr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
