package rest

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
)

func UploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	desc, ok := cmd["Desc"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}

	whitelist := make([]string, 0)
	wh, _ := cmd["WhiteList"].([]interface{})
	for _, w := range wh {
		if _, ok := w.(string); !ok {
			continue
		}
		whitelist = append(whitelist, w.(string))
	}
	pwd, _ := cmd["EncryptPassword"].(string)
	url, _ := cmd["Url"].(string)
	share, _ := cmd["Share"].(bool)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	opt, err := dsp.DspService.UploadFile(path, desc, cmd["Duration"], cmd["Interval"],
		cmd["Privilege"], cmd["CopyNum"], cmd["StoreType"], cmd["RealFileSize"], pwd, url, whitelist, share)
	if err != nil {
		log.Errorf("upload file failed, err %v", err)
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	uploadOption := make(map[string]interface{})
	uploadOption["FileName"] = string(opt.FileDesc)
	uploadOption["RealFileSize"] = opt.FileSize
	uploadOption["ProveInterval"] = opt.ProveInterval
	uploadOption["ExpiredHeight"] = opt.ExpiredHeight
	uploadOption["Privilege"] = opt.Privilege
	uploadOption["CopyNum"] = opt.CopyNum
	uploadOption["Encrypt"] = opt.Encrypt
	uploadOption["EncryptPassword"] = string(opt.EncryptPassword)
	uploadOption["Url"] = string(opt.DnsURL)
	uploadOption["WhiteList"] = whitelist
	uploadOption["Share"] = opt.Share
	uploadOption["StorageType"] = opt.StorageType
	resp["Result"] = uploadOption
	return resp
}

func DeleteUploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gl, _ := cmd["GasLimit"].(string)
	gasLimit := uint64(0)
	if len(gl) > 0 {
		var err error
		gasLimit, err = strconv.ParseUint(gl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	ret, err := dsp.DspService.DeleteUploadFile(fileHash, gasLimit)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func DeleteUploadFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Hash"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gl, _ := cmd["GasLimit"].(string)
	gasLimit := uint64(0)
	if len(gl) > 0 {
		var err error
		gasLimit, err = strconv.ParseUint(gl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	fileHashes := make([]string, 0, len(v))
	for _, str := range v {
		fileHash, ok := str.(string)
		if !ok {
			continue
		}
		fileHashes = append(fileHashes, fileHash)
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	ret, err := dsp.DspService.DeleteUploadFiles(fileHashes, gasLimit)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func CalculateDeleteFilesFee(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["FileHashes"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	fileHashes := make([]string, 0, len(v))
	for _, str := range v {
		fileHash, ok := str.(string)
		if !ok {
			continue
		}
		fileHashes = append(fileHashes, fileHash)
	}
	ret, err := dsp.DspService.CalculateDeleteFilesFee(fileHashes)
	if err != nil {
		resp := ResponsePackWithErrMsg(err.Code, err.Error.Error())
		resp["Result"] = ret
		return resp
	}
	resp["Result"] = ret
	return resp
}

func DownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, _ := cmd["Hash"].(string)
	url, _ := cmd["Url"].(string)
	link, _ := cmd["Link"].(string)
	decryptedPassword, _ := cmd["DecryptPassword"].(string)
	max, ok := cmd["MaxPeerNum"].(float64)
	if !ok {
		max = 1
	}
	setFileName, _ := cmd["SetFileName"].(bool)
	inOrder, _ := cmd["InOrder"].(bool)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	// password, ok := cmd["Password"].(string)
	// if !ok {
	// 	return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	// }
	// if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
	// 	return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	// }
	err := dsp.DspService.DownloadFile(fileHash, url, link, decryptedPassword, uint64(max), setFileName, inOrder)
	if err != nil {
		log.Errorf("download file failed, err %v", err)
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

func DeleteDownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	ret, err := dsp.DspService.DeleteDownloadFile(fileHash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetUploadFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ft, _ := cmd["Type"].(string)
	fileType := dsp.DspFileListType(0)
	if len(ft) > 0 {
		fileTypeInt, err := strconv.ParseUint(ft, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		fileType = dsp.DspFileListType(fileTypeInt)
	}
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}

	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	fl, _ := cmd["Filter"].(string)
	filter := uint64(0)
	if len(fl) > 0 {
		var err error
		filter, err = strconv.ParseUint(fl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	log.Debugf("cmd :%v, type %d, offset %d limit %d", cmd, fileType, offset, limit)
	files, err := dsp.DspService.GetUploadFiles(fileType, offset, limit, dsp.UploadFileFilterType(filter))
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = files
	return resp
}

func GetDownloadFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ft, _ := cmd["Type"].(string)
	fileType := dsp.DspFileListType(0)
	if len(ft) > 0 {
		fileTypeInt, err := strconv.ParseUint(ft, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		fileType = dsp.DspFileListType(fileTypeInt)
	}
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}

	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	log.Debugf("cmd :%v, type %d, offset %d limit %d", cmd, fileType, offset, limit)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	fileinfos, err := dsp.DspService.GetDownloadFiles(fileType, offset, limit)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = fileinfos
	return resp
}

func GetTransferList(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	tt, ok := cmd["Type"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	transferTypeInt, err := strconv.ParseUint(tt, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	transferType := dsp.TransferType(transferTypeInt)
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	list, derr := dsp.DspService.GetTransferList(transferType, uint32(offset), uint32(limit))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = list
	return resp
}

func CalculateUploadFee(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("CalculateUploadFee cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	p, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	path, err := hex.DecodeString(p)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	res, derr := dsp.DspService.CalculateUploadFee(string(path), cmd["Duration"], cmd["Interval"],
		cmd["Times"], cmd["CopyNum"], cmd["WhiteList"], cmd["StoreType"])
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = res
	return resp
}

func GetDownloadFileInfo(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetDownloadFileInfo cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok || len(url) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	realUrl, err := hex.DecodeString(url)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	info, derr := dsp.DspService.GetDownloadFileInfo(string(realUrl))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = info
	return resp
}

func EncryptFile(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("EncryptFile cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok || len(path) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok || len(password) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	err := dsp.DspService.EncryptFile(string(path), password)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

func DecryptFile(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("DecryptFile cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok || len(path) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok || len(password) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	fileName, ok := cmd["FileName"].(string)
	// if !ok || len(fileName) == 0 {
	// 	return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	// }
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	err := dsp.DspService.DecryptFile(path, fileName, password)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

func GetFileShareIncome(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetFileShareIncome cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	begStr, ok := cmd["Begin"].(string)
	if !ok || len(begStr) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	begin, err := strconv.ParseUint(begStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	endStr, ok := cmd["End"].(string)
	if !ok || len(endStr) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	end, err := strconv.ParseUint(endStr, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	offset, err := dsp.OptionStrToUint64(cmd["Offset"])
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	limit, err := dsp.OptionStrToUint64(cmd["Limit"])
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	ret, derr := dsp.DspService.GetFileShareIncome(begin, end, offset, limit)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetFileShareRevenue(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetDownloadFileInfo cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	ret := make(map[string]interface{}, 0)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	revenue, err := dsp.DspService.GetFileRevene()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	ret["Revenue"] = revenue
	ret["RevenueFormat"] = utils.FormatUsdt(revenue)
	resp["Result"] = ret
	return resp
}

func WhiteListOperate(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("WhiteListOperate cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, ok := cmd["FileHash"].(string)
	if !ok || len(fileHash) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	op, ok := cmd["Operation"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	list, ok := cmd["List"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	whitelist := make([]*dsp.WhiteListRule, 0, len(list))
	for _, item := range list {
		l, ok := item.(map[string]interface{})
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		value, ok := l["Addr"]
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		addr, ok := value.(string)
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		value, ok = l["StartHeight"]
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		startHeight, ok := value.(float64)
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		value, ok = l["ExpiredHeight"]
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		expiredHeight, ok := value.(float64)
		if !ok {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}

		whitelist = append(whitelist, &dsp.WhiteListRule{
			Addr:          addr,
			StartHeight:   uint64(startHeight),
			ExpiredHeight: uint64(expiredHeight),
		})
	}
	log.Debugf("fileHash %v, op %v, list %v", fileHash, op, whitelist)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	tx, err := dsp.DspService.WhiteListOperation(fileHash, uint64(op), whitelist)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = tx
	return resp
}

func GetFileWhiteList(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetFileWhiteList cmd:%v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, ok := cmd["FileHash"].(string)
	if !ok || len(fileHash) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	list, err := dsp.DspService.GetWhitelist(fileHash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = list
	return resp
}

func GetUserSpace(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, _ := cmd["Addr"].(string)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	userspace, err := dsp.DspService.GetUserSpace(addr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = userspace
	return resp
}

func SetUserSpace(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	size, sizeOp, second, secondOp := float64(0), float64(0), float64(0), float64(0)
	sizeMap, ok := cmd["Size"].(map[string]interface{})
	if ok {
		size, _ = sizeMap["Value"].(float64)
		sizeOp, _ = sizeMap["Type"].(float64)
	}
	secondMap, ok := cmd["Second"].(map[string]interface{})
	if ok {
		second, _ = secondMap["Value"].(float64)
		secondOp, _ = secondMap["Type"].(float64)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	tx, err := dsp.DspService.SetUserSpace(addr, uint64(size), uint64(sizeOp), uint64(second), uint64(secondOp))
	if err != nil {
		log.Errorf("add user space err %s", err)
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]string, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func GetUserSpaceCost(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	log.Debugf("cmd %v , type %T", cmd, cmd["Size"])
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	size, sizeOp, second, secondOp := float64(0), float64(0), float64(0), float64(0)
	sizeMap, ok := cmd["Size"].(map[string]interface{})
	if ok {
		size, _ = sizeMap["Value"].(float64)
		sizeOp, _ = sizeMap["Type"].(float64)
	}
	secondMap, ok := cmd["Second"].(map[string]interface{})
	if ok {
		second, _ = secondMap["Value"].(float64)
		secondOp, _ = secondMap["Type"].(float64)
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	cost, err := dsp.DspService.GetUserSpaceCost(addr, uint64(size), uint64(sizeOp), uint64(second), uint64(secondOp))
	if err != nil {
		log.Errorf("get user space cost err %s", err)
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = cost
	return resp
}

func GetUserSpaceRecords(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	offset, _ := dsp.OptionStrToUint64(cmd["Offset"])
	limit, _ := dsp.OptionStrToUint64(cmd["Limit"])
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	ret, err := dsp.DspService.GetUserspaceRecords(addr, offset, limit)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	records := make(map[string]interface{})
	records["Records"] = ret
	resp["Result"] = records
	return resp
}

func GetStorageNodesInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, err := dsp.DspService.GetStorageNodesInfo()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetUploadFileInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	hash, ok := cmd["FileHash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, err := dsp.DspService.GetFileInfo(hash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func PauseUploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.PauseUploadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func ResumeUploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.ResumeUploadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func RetryUploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.RetryUploadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func CancelUploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gl, _ := cmd["GasLimit"].(string)
	gasLimit := uint64(0)
	if len(gl) > 0 {
		var err error
		gasLimit, err = strconv.ParseUint(gl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}

	ret, derr := dsp.DspService.CancelUploadFile(ids, gasLimit)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func PauseDownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.PauseDownloadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func ResumeDownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, err := dsp.DspService.ResumeDownloadFile(ids)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func RetryDownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.RetryDownloadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func CancelDownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.CancelDownloadFile(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func DeleteTransferRecord(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Ids"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ids := make([]string, 0, len(v))
	for _, str := range v {
		id, ok := str.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	ret, derr := dsp.DspService.DeleteTransferRecord(ids)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetTransferDetail(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	id, ok := cmd["Id"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	realId, err := hex.DecodeString(id)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	tt, ok := cmd["Type"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	transferTypeInt, err := strconv.ParseUint(tt, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	transferType := dsp.TransferType(transferTypeInt)
	var transfer interface{}
	var dspErr *dsp.DspErr
	if strings.Contains(string(realId), "://") {
		transfer, dspErr = dsp.DspService.GetTransferDetailByUrl(transferType, string(realId))
	} else {
		transfer, dspErr = dsp.DspService.GetTransferDetail(transferType, string(realId))
	}
	if dspErr != nil {
		return ResponsePackWithErrMsg(dspErr.Code, dspErr.Error.Error())
	}
	resp["Result"] = transfer
	return resp
}

func GetUploadFileProveDetail(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	hash, ok := cmd["FileHash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, err := dsp.DspService.GetProveDetail(hash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetPeerCountOfHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	hash, ok := cmd["FileHash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, err := dsp.DspService.GetPeerCountOfHash(hash)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	m := make(map[string]interface{})
	m["Count"] = ret
	resp["Result"] = m
	return resp
}
