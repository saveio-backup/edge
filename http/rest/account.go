package rest

import (
	edgeCmd "github.com/saveio/edge/cmd"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common/log"
)

func GetCurrentAccount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	service := dsp.DspService
	if service == nil {
		service = &dsp.Endpoint{}
	}
	acc, err := service.GetCurrentAccount()
	if err != nil {
		if acc == nil {
			return ResponsePackWithErrMsg(err.Code, err.Error.Error())
		}
		resp = ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = acc
	return resp
}

func NewAccount(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("NewAccount cmd %v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService != nil && dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	label, ok := cmd["Label"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	optionType, ok := cmd["KeyType"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	optionCurve, ok := cmd["Curve"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	optionScheme, ok := cmd["Scheme"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
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
	if dsp.DspService != nil && dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	privKeyStr, ok := cmd["PrivateKey"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
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
	if dsp.DspService != nil && dsp.DspService.AccountExists() {
		return ResponsePack(dsp.ACCOUNT_EXIST)
	}
	walletStr, ok := cmd["Wallet"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	walletPath, _ := cmd["WalletPath"].(string)
	acc, err := dsp.DspService.ImportWithWalletData(walletStr, password, walletPath)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = acc
	return resp
}

func ExportWalletFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	endpoint := dsp.DspService
	if endpoint == nil {
		endpoint = &dsp.Endpoint{}
	}
	walletResp, err := endpoint.ExportWalletFile()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = walletResp
	return resp
}

func ExportWIFPrivateKey(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	endpoint := dsp.DspService
	if endpoint == nil {
		endpoint = &dsp.Endpoint{}
	}
	if endpoint.GetDspAccount() == nil {
		return ResponsePackWithErrMsg(dsp.ACCOUNT_NOT_LOGIN, dsp.ErrMaps[dsp.ACCOUNT_NOT_LOGIN].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	ret, err := endpoint.ExportWIFPrivateKey()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

// Logout. logout current account
func Logout(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("Logout")
	resp := ResponsePack(dsp.SUCCESS)
	err := dsp.DspService.Logout()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

// Login. login current account
func Login(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	log.Debugf("login account with %s", password)
	ret, derr := dsp.DspService.Login(password)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

// CheckPassword.
func CheckPassword(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	err := dsp.DspService.CheckPassword(password)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}
