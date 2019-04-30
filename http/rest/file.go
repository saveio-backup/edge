package rest

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"strconv"

	"github.com/saveio/dsp-go-sdk/common"
	clicom "github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	berr "github.com/saveio/edge/http/base/error"
	http "github.com/saveio/edge/http/common"
	"github.com/saveio/edge/http/util"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/onifs"
)

func UploadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	desc, ok := cmd["Desc"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	currentAccount := DspService.Dsp.CurrentAccount()
	fssetting, err := DspService.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	duration, _ := cmd["Duration"].(float64)
	interval, ok := cmd["Interval"].(float64)
	if !ok || interval == 0 {
		interval = float64(fssetting.DefaultProvePeriod)
	}
	times, ok := cmd["Times"].(float64)
	if !ok || times == 0 {
		//TODO
		userspace, err := DspService.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		if userspace == nil {
			return ResponsePack(berr.DSP_USERSPACE_NO_ENOUGH)
		}
		currentHeight, err := DspService.Dsp.Chain.GetCurrentBlockHeight()
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		log.Debugf("userspace.ExpireHeight %d, current: %d", userspace.ExpireHeight, currentHeight)
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return ResponsePack(berr.DSP_USERSPACE_EXPIRED)
		}
		if duration > 0 && (uint64(currentHeight)+uint64(duration)) > userspace.ExpireHeight {
			return ResponsePack(berr.DSP_DURATION_EXCEED_EXPIRED)
		}
		if duration == 0 {
			duration = float64(userspace.ExpireHeight) - float64(currentHeight)
		}
		times = math.Ceil(duration / float64(interval))
		log.Debugf("userspace.ExpireHeight %d, current %d, duration :%v, times :%v", userspace.ExpireHeight, currentHeight, duration, times)
	}
	privilege, ok := cmd["Privilege"].(float64)
	if !ok {
		privilege = onifs.PUBLIC
	}
	copynum, ok := cmd["CopyNum"].(float64)
	if !ok {
		copynum = float64(fssetting.DefaultCopyNum)
	}
	password, _ := cmd["EncryptPassword"].(string)
	url, ok := cmd["Url"].(string)
	if !ok {
		// random
		b := make([]byte, clicom.DSP_URL_RAMDOM_NAME_LEN/2)
		_, err := rand.Read(b)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		url = DspService.Dsp.Chain.Native.Dns.GetCustomUrlHeader() + hex.EncodeToString(b)
	}
	find, err := DspService.Dsp.Chain.Native.Dns.QueryUrl(url, DspService.Dsp.CurrentAccount().Address)
	if find != nil || err == nil {
		return ResponsePack(berr.DSP_URL_EXISTS)
	}
	whitelist := make([]string, 0)
	wh, _ := cmd["WhiteList"].([]interface{})
	for _, w := range wh {
		if _, ok := w.(string); !ok {
			continue
		}
		whitelist = append(whitelist, w.(string))
	}
	share, _ := cmd["Share"].(bool)
	opt := &common.UploadOption{
		FileDesc:        desc,
		ProveInterval:   uint64(interval),
		ProveTimes:      uint32(times),
		Privilege:       uint32(privilege),
		CopyNum:         uint32(copynum),
		Encrypt:         len(password) > 0,
		EncryptPassword: password,
		RegisterDns:     len(url) > 0,
		BindDns:         len(url) > 0,
		DnsUrl:          url,
		WhiteList:       whitelist,
		Share:           share,
	}
	optBuf, _ := json.Marshal(opt)
	log.Debugf("path %s, UploadOption :%s\n", path, optBuf)
	go func() {
		baseName := filepath.Base(path)
		data := make([]byte, 1*1024*1024)
		_, err := rand.Read(data)
		if err != nil {
			log.Errorf("make rand data err %s", err)
			return
		}
		md5Ret := md5.Sum(data)
		path = config.Parameters.FsConfig.FsFileRoot + "/" + baseName
		log.Debugf("path:%s, md5Ret :%s", path, hex.EncodeToString(md5Ret[:]))
		ioutil.WriteFile(path, []byte(data), 0666)
		ret, err := DspService.Dsp.UploadFile(path, opt)
		// os.Remove(path)
		if err != nil {
			log.Errorf("upload failed err %s", err)
			return
		}
		log.Info(ret)
	}()
	return resp
}

func DeleteFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	fileHash, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fi, err := DspService.Dsp.Chain.Native.Fs.GetFileInfo(fileHash)
	if fi != nil && err == nil && fi.FileOwner.ToBase58() == DspService.Dsp.WalletAddress() {
		tx, err := DspService.Dsp.DeleteUploadedFile(fileHash)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		resp["Result"] = tx
		return resp
	}
	err = DspService.Dsp.DeleteDownloadedFile(fileHash)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	return resp
}

func DownloadFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	fileHash, _ := cmd["Hash"].(string)
	url, _ := cmd["Url"].(string)
	link, _ := cmd["Link"].(string)
	password, _ := cmd["Password"].(string)
	max, ok := cmd["MaxPeerNum"].(float64)
	if !ok {
		max = 1
	}
	if len(fileHash) > 0 {
		go func() {
			// TODO: get file name
			err := DspService.Dsp.DownloadFile(fileHash, "", common.ASSET_USDT, true, password, false, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return resp
	}
	if len(url) > 0 {
		go func() {
			err := DspService.Dsp.DownloadFileByUrl(url, common.ASSET_USDT, true, password, false, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return resp
	}
	if len(link) > 0 {
		go func() {
			err := DspService.Dsp.DownloadFileByLink(fileHash, common.ASSET_USDT, true, password, false, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return resp
	}
	return resp
}

func GetUploadFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	ft, _ := cmd["Type"].(string)
	fileType := http.DSP_FILE_LIST_TYPE(0)
	if len(ft) > 0 {
		fileTypeInt, err := strconv.ParseUint(ft, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		fileType = http.DSP_FILE_LIST_TYPE(fileTypeInt)
	}
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}

	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}
	log.Debugf("cmd :%v, type %d, offset %d limit %d", cmd, fileType, offset, limit)
	files, err := DspService.GetUploadFiles(fileType, offset, limit)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	resp["Result"] = files
	return resp
}

func GetDownloadFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	ft, _ := cmd["Type"].(string)
	fileType := http.DSP_FILE_LIST_TYPE(0)
	if len(ft) > 0 {
		fileTypeInt, err := strconv.ParseUint(ft, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		fileType = http.DSP_FILE_LIST_TYPE(fileTypeInt)
	}
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}

	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}
	log.Debugf("cmd :%v, type %d, offset %d limit %d", cmd, fileType, offset, limit)

	fileinfos, err := DspService.GetDownloadFiles(fileType, offset, limit)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	resp["Result"] = fileinfos
	return resp
}

func GetTransferList(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	tt, ok := cmd["Type"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	transferTypeInt, err := strconv.ParseUint(tt, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	transferType := dsp.TransferType(transferTypeInt)
	of, _ := cmd["Offset"].(string)
	offset := uint64(0)
	if len(of) > 0 {
		var err error
		offset, err = strconv.ParseUint(of, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}

	li, _ := cmd["Limit"].(string)
	limit := uint64(0)
	if len(li) > 0 {
		var err error
		limit, err = strconv.ParseUint(li, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
	}
	list := DspService.GetTransferList(transferType, offset, limit)
	resp["Result"] = list
	return resp
}

func CalculateUploadFee(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("CalculateUploadFee cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	p, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	path, err := hex.DecodeString(p)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	currentAccount := DspService.Dsp.CurrentAccount()
	fssetting, err := DspService.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	duration, err := util.OptionStrToFloat64(cmd["Duration"], 0)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	interval, err := util.OptionStrToFloat64(cmd["Interval"], float64(fssetting.DefaultProvePeriod))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	times, err := util.OptionStrToFloat64(cmd["Times"], 0)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if times == 0 {
		userspace, err := DspService.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		if userspace == nil {
			return ResponsePack(berr.DSP_USERSPACE_NO_ENOUGH)
		}
		currentHeight, err := DspService.Dsp.Chain.GetCurrentBlockHeight()
		if err != nil {
			return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
		}
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return ResponsePack(berr.DSP_USERSPACE_EXPIRED)
		}
		if duration > 0 && (uint64(currentHeight)+uint64(duration)) > userspace.ExpireHeight {
			return ResponsePack(berr.DSP_DURATION_EXCEED_EXPIRED)
		}
		if duration == 0 {
			duration = float64(userspace.ExpireHeight) - float64(currentHeight)
		}
		times = math.Ceil(duration / float64(interval))
		log.Debugf("userspace.ExpireHeight %d, current %d, duration :%v, times :%v", userspace.ExpireHeight, currentHeight, duration, times)
	}
	copynum, err := util.OptionStrToFloat64(cmd["CopyNum"], float64(fssetting.DefaultCopyNum))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	wh, err := util.OptionStrToFloat64(cmd["WhiteList"], 0)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	log.Debugf("interval %v, times: %v, copynum:%v, wh:%v, path: %v", interval, times, copynum, wh, path)
	// TEST:
	fee, err := DspService.CalculateUploadFee(string("./wallet.dat"), uint64(interval), uint32(times), uint32(copynum), uint64(wh))
	// fee, err := DspService.CalculateUploadFee(string(path), uint64(interval), uint32(times), uint32(copynum), uint64(wh))
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	feeFormat := utils.FormatUsdt(fee)
	type calculateResp struct {
		Fee       uint64
		FeeFormat string
	}

	resp["Result"] = &calculateResp{
		Fee:       fee,
		FeeFormat: feeFormat,
	}
	return resp
}

func GetDownloadFileInfo(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetDownloadFileInfo cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok || len(url) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	realUrl, err := hex.DecodeString(url)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	info := DspService.GetDownloadFileInfo(string(realUrl))
	if info == nil {
		return ResponsePackWithErrMsg(berr.DSP_QUERY_FILE_ERROR, "file not found")
	}
	resp["Result"] = info
	return resp
}

func EncryptFile(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("EncryptFile cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok || len(path) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	// realPath, err := hex.DecodeString(path)
	// if err != nil {
	// 	return ResponsePack(berr.INVALID_PARAMS)
	// }
	password, ok := cmd["Password"].(string)
	if !ok || len(password) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	err := DspService.EncryptFile(string(path), password)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_CRYPTO_ERROR, err.Error())
	}
	return resp
}

func DecryptFile(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("DecryptFile cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	path, ok := cmd["Path"].(string)
	if !ok || len(path) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	// realPath, err := hex.DecodeString(path)
	// if err != nil {
	// 	return ResponsePack(berr.INVALID_PARAMS)
	// }
	password, ok := cmd["Password"].(string)
	if !ok || len(password) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	err := DspService.DecryptFile(string(path), password)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_CRYPTO_ERROR, err.Error())
	}
	return resp
}

func GetFileShareIncome(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetDownloadFileInfo cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	begStr, ok := cmd["Begin"].(string)
	if !ok || len(begStr) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	begin, err := strconv.ParseUint(begStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	endStr, ok := cmd["End"].(string)
	if !ok || len(endStr) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	end, err := strconv.ParseUint(endStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	offset, err := util.OptionStrToUint64(cmd["Offset"])
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	limit, err := util.OptionStrToUint64(cmd["Limit"])
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	ret, err := DspService.GetFileShareIncome(begin, end, offset, limit)
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_SHARE_INCOME_ERROR, err.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetFileShareRevenue(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("GetDownloadFileInfo cmd:%v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	ret := make(map[string]interface{}, 0)
	revenue, err := DspService.GetFileRevene()
	if err != nil {
		return ResponsePackWithErrMsg(berr.DSP_FILE_ERROR, err.Error())
	}
	ret["Revenue"] = revenue
	ret["RevenueFormat"] = utils.FormatUsdt(revenue)
	resp["Result"] = ret
	return resp
}
