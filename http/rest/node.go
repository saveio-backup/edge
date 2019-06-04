package rest

import (
	"strconv"

	"github.com/saveio/edge/dsp"
)

//Handle for Dsp
func RegisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["NodeAddr"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	volumeStr, ok := cmd["Volume"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	volume, err := strconv.ParseUint(volumeStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	serviceTime, err := strconv.ParseUint(serviceTimeStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.RegisterNode(addr, volume, serviceTime)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func UnregisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	tx, err := dsp.DspService.UnregisterNode()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func NodeQuery(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	walletAddr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	fsNodeInfo, err := dsp.DspService.NodeQuery(walletAddr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = fsNodeInfo
	return resp
}

func NodeUpdate(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	var addr string
	var volume, serviceTime uint64

	_, ok := cmd["NodeAddr"].(string)
	if ok {
		addr = cmd["NodeAddr"].(string)
	}

	volumeStr, ok := cmd["Volume"].(string)
	if ok {
		temp, err := strconv.ParseUint(volumeStr, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
		}
		volume = temp
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if ok {
		temp, err := strconv.ParseUint(serviceTimeStr, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
		}
		serviceTime = temp
	}

	tx, err := dsp.DspService.NodeUpdate(addr, volume, serviceTime)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func NodeWithdrawProfit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	tx, err := dsp.DspService.NodeWithdrawProfit()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

//Handle for DNS
func RegisterUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	link, ok := cmd["Link"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	tx, err := dsp.DspService.RegisterUrl(url, link)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func BindUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	link, ok := cmd["Link"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	tx, err := dsp.DspService.BindUrl(url, link)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func QueryLink(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	link, err := dsp.DspService.QueryLink(url)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = link
	return resp
}

func RegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ip, ok := cmd["Ip"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	port, ok := cmd["Port"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}

	depositStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	deposit, err := strconv.ParseUint(depositStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.RegisterDns(ip, port, deposit)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = tx
	return resp
}

func UnRegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	tx, err := dsp.DspService.UnRegisterDns()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = tx
	return resp
}

func QuitDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	tx, err := dsp.DspService.QuitDns()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func AddPos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.AddPos(amount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func ReducePos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.ReducePos(amount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	resp["Result"] = tx
	return resp
}

func QueryRegInfos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, err := dsp.DspService.QueryRegInfos()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret.PeerPoolMap
	return resp
}

func QueryRegInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	pubkey, ok := cmd["Pubkey"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	ret, err := dsp.DspService.QueryRegInfo(pubkey)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func QueryHostInfos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, err := dsp.DspService.QueryHostInfos()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func QueryHostInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(dsp.INVALID_PARAMS)
	}
	ret, err := dsp.DspService.QueryHostInfo(addr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = ret
	return resp
}
