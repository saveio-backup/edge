package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

func GetCurrentAccount(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetCurrentAccount(params)
	acc, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(acc)
}

func NewAccount(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 5 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Password", "Label", "KeyType", "Curve", "Scheme"})
	v := rest.NewAccount(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func ImportWithPrivateKey(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"PrivateKey", "Label", "Password"})
	v := rest.ImportWithPrivateKey(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func ExportPrivateKey(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Password"})
	v := rest.ExportWIFPrivateKey(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func ImportWithWalletData(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Wallet", "Password"})
	v := rest.ImportWithWalletData(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func ExportWalletFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.ExportWalletFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

// Logout. logout current account
func Logout(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.Logout(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
