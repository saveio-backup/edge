package rest

import (
	"reflect"
	"strconv"

	dspActorClient "github.com/saveio/dsp-go-sdk/actor/client"
	dUtils "github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/utils"
	"github.com/saveio/themis/common/log"
)

//Handle for Dsp
func RegisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["NodeAddr"].(string)
	if !ok || len(addr) == 0 {
		addr = dspActorClient.P2pGetPublicAddr()
	}
	log.Debugf("register node addr %s", addr)
	volumeStr, ok := cmd["Volume"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	volume, err := strconv.ParseUint(volumeStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	serviceTime, err := strconv.ParseUint(serviceTimeStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.RegisterNode(addr, volume, serviceTime)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func UnregisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	tx, err := dsp.DspService.UnregisterNode()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func NodeQuery(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	walletAddr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	fsNodeInfo, err := dsp.DspService.NodeQuery(walletAddr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Info"] = utils.ConvertStructToMap(reflect.ValueOf(fsNodeInfo).Elem())
	resp["Result"] = m
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
	if ok && len(volumeStr) > 0 {
		temp, err := strconv.ParseUint(volumeStr, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
		}
		volume = temp
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if ok && len(serviceTimeStr) > 0 {
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

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func NodeWithdrawProfit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	tx, err := dsp.DspService.NodeWithdrawProfit()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

//Handle for DNS
func RegisterUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	fileHash, ok := cmd["FileHash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	fileName, _ := cmd["FileName"].(string)
	blocksRoot, _ := cmd["BlocksRoot"].(string)
	fileOwner, _ := cmd["FileOwner"].(string)
	fileSize := utils.StringToUint64(cmd["FileSize"])
	totalCount := utils.StringToUint64(cmd["TotalCount"])

	tx, err := dsp.DspService.RegisterUrl(url, fileHash, fileName, blocksRoot, fileOwner, fileSize, totalCount)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func BindUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	link, ok := cmd["Link"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	tx, err := dsp.DspService.BindUrl(url, link)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func QueryLink(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	link, err := dsp.DspService.QueryLink(url)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Link"] = link
	resp["Result"] = m
	return resp
}

func DeleteUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	tx, err := dsp.DspService.DeleteFileUrl(url)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func UpdatePluginVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	urlTypeStr, ok := cmd["Type"].(string)
	urlType := uint64(0)
	if len(urlTypeStr) > 0 {
		urlTypeInt, err := strconv.ParseUint(urlTypeStr, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		urlType = uint64(urlTypeInt)
	}
	version, ok := cmd["Version"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	pt, _ := cmd["Platform"].(string)
	platformType := dsp.DspFileUrlPatformType(0)
	if len(pt) > 0 {
		platformTypeInt, err := strconv.ParseUint(pt, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		platformType = dsp.DspFileUrlPatformType(platformTypeInt)
	}
	fileHash, ok := cmd["FileHash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	_, err := dsp.DspService.GetFileInfo(fileHash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	img, ok := cmd["Img"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	title, ok := cmd["Title"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	changeLogCN, ok := cmd["ChangeLogCN"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	changeLogEN, ok := cmd["ChangeLogEN"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	changeLog := dUtils.ChangeLog{
		ZH: changeLogCN,
		EN: changeLogEN,
	}
	fileName, _ := cmd["FileName"].(string)
	blocksRoot, _ := cmd["BlocksRoot"].(string)
	fileOwner, _ := cmd["FileOwner"].(string)
	fileSize := utils.StringToUint64(cmd["FileSize"])
	totalCount := utils.StringToUint64(cmd["TotalCount"])

	tx, err := dsp.DspService.UpdatePluginVersion(url, fileHash, fileName, blocksRoot, fileOwner, version, img, title, changeLog, urlType, fileSize, totalCount, platformType)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func QueryPluginVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	pt, _ := cmd["Platform"].(string)
	platformType := dsp.DspFileUrlPatformType(0)
	if len(pt) > 0 {
		platformTypeInt, err := strconv.ParseUint(pt, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		platformType = dsp.DspFileUrlPatformType(platformTypeInt)
	}
	fileHash, ok := cmd["FileHash"].(string)
	if len(fileHash) > 0 {
		_, err := dsp.DspService.GetFileInfo(fileHash)
		if err != nil {
			return ResponsePackWithErrMsg(err.Code, err.Error.Error())
		}
	}
	log.Debugf("pt: %s, Platform %d", pt, platformType)
	pluginVersion, err := dsp.DspService.QueryPluginVersion(url, fileHash, int(platformType))
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Url"] = pluginVersion.Url
	m["Type"] = pluginVersion.Type
	m["Version"] = pluginVersion.Version
	m["FileHash"] = pluginVersion.FileHash
	m["Img"] = pluginVersion.Img
	m["Title"] = pluginVersion.Title
	m["ChangeLog"] = pluginVersion.ChangeLog
	m["Platform"] = pluginVersion.Platform
	resp["Result"] = m
	return resp
}

func RegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ip, ok := cmd["Ip"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	port, ok := cmd["Port"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	depositStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	deposit, err := strconv.ParseUint(depositStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.RegisterDns(ip, port, deposit)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func UnRegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	tx, err := dsp.DspService.UnRegisterDns()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func QuitDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	tx, err := dsp.DspService.QuitDns()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func AddPos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.AddPos(amount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func ReducePos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	tx, derr := dsp.DspService.ReducePos(amount)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
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
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
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
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, err := dsp.DspService.QueryHostInfo(addr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = ret
	return resp
}

func QueryPluginsInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, err := dsp.DspService.QueryPluginsInfo()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func QueryPublicIP(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, derr := dsp.DspService.GetExternalIP(addr)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}
