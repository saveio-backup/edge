package utils

import (
	"encoding/json"

	"github.com/saveio/edge/dsp"
)

func GetCurrentAccount() (*dsp.AccountResp, error) {
	ret, ontErr := sendRpcRequest("getcurrentaccount", []interface{}{})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	var data *dsp.AccountResp
	err := json.Unmarshal(ret, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func NewAccount(password, label, keyType, curve, scheme string) (*dsp.AccountResp, error) {
	ret, ontErr := sendRpcRequest("newaccount", []interface{}{password, label, keyType, curve, scheme})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	var data *dsp.AccountResp
	err := json.Unmarshal(ret, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ImportWithPrivateKey(privateKey, label, password string) (*dsp.AccountResp, error) {
	ret, ontErr := sendRpcRequest("importwithprivatekey", []interface{}{privateKey, label, password})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	var data *dsp.AccountResp
	err := json.Unmarshal(ret, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ImportWithWalletData(wallet, password string) (*dsp.AccountResp, error) {
	ret, ontErr := sendRpcRequest("importwithwalletdata", []interface{}{wallet, password})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	var data *dsp.AccountResp
	err := json.Unmarshal(ret, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ExportWalletFile() (string, error) {
	ret, ontErr := sendRpcRequest("exportwalletfile", []interface{}{})
	if ontErr != nil {
		return "", ontErr.Error
	}
	old := string(ret)
	return old, nil
	// new := strings.Replace(old, "\\", "", -1)
	// return string(new), nil
}

func ExportPrivateKey(password string) (string, error) {
	ret, ontErr := sendRpcRequest("exportprivatekey", []interface{}{password})
	if ontErr != nil {
		return "", ontErr.Error
	}
	return string(ret), nil
}

// Logout. logout current account
func Logout() error {
	_, ontErr := sendRpcRequest("logout", []interface{}{})
	if ontErr != nil {
		return ontErr.Error
	}
	return nil
}
