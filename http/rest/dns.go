package rest

import (
	"encoding/hex"

	"github.com/saveio/edge/dsp"
)

func GetAllDNS(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	allOnlineDns, err := dsp.DspService.GetAllOnlineDNS()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	list := make([]map[string]string, 0, len(allOnlineDns))
	for addr, host := range allOnlineDns {
		m := make(map[string]string)
		m["WalletAddr"] = addr
		m["HostAddr"] = host
		list = append(list, m)
	}
	resp["Result"] = list
	return resp
}

func GetHashFromUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_DSP, dsp.ErrMaps[dsp.NO_DSP].Error())
	}
	urlHex, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	url, err := hex.DecodeString(urlHex)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	hash, derr := dsp.DspService.GetFileHashFromUrl(string(url))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Hash"] = hash
	resp["Result"] = m
	return resp
}

func UpdateFileUrlLink(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_DSP, dsp.ErrMaps[dsp.NO_DSP].Error())
	}
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	hash, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	pwd, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	fileName, _ := cmd["FileName"].(string)
	fileSize, _ := cmd["FileSize"].(float64)
	totalCount, _ := cmd["TotalCount"].(float64)
	if checkErr := dsp.DspService.CheckPassword(pwd); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	tx, err := dsp.DspService.UpdateFileUrlLink(url, hash, fileName, uint64(fileSize), uint64(totalCount))
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}
