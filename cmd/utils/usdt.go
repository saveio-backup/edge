package utils

const (
	ASSET_USDT = "usdt"
)

func GetBalance(addr string) ([]byte, error) {
	ret, ontErr := sendRpcRequest("getbalance", []interface{}{addr})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	return ret, nil
}
