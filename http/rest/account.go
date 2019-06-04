package rest

import (
	edgeCmd "github.com/saveio/edge/cmd"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common/log"
)

func GetCurrentAccount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	acc, err := dsp.DspService.GetCurrentAccount()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = acc
	return resp
}

func NewAccount(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("NewAccount cmd %v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	label, ok := cmd["Label"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	optionType, ok := cmd["KeyType"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	optionCurve, ok := cmd["Curve"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	optionScheme, ok := cmd["Scheme"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	createOnly, _ := cmd["CreateOnly"].(bool)
	keyType := edgeCmd.GetKeyTypeCode(optionType)
	curve := edgeCmd.GetCurveCode(optionCurve)
	scheme := edgeCmd.GetSchemeCode(optionScheme)
	acc, err := dsp.DspService.NewAccount(label, keyType, curve, scheme, []byte(password), createOnly)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = acc
	return resp
}

func ImportWithPrivateKey(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	privKeyStr, ok := cmd["PrivateKey"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	label, ok := cmd["Label"].(string)
	acc, err := dsp.DspService.ImportWithPrivateKey(privKeyStr, label, password)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	// TODO: save acc2 to wallet.dat
	resp["Result"] = acc
	return resp
}

func ImportWithWalletData(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	walletStr, ok := cmd["Wallet"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	acc, err := dsp.DspService.ImportWithWalletData(walletStr, password)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = acc
	return resp
}

func ExportWalletFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	walletResp, err := dsp.DspService.ExportWalletFile()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = walletResp
	return resp
}

func ExportWIFPrivateKey(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, err := dsp.DspService.ExportWIFPrivateKey()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

// Logout. logout current account
func Logout(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	err := dsp.DspService.Logout()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}
