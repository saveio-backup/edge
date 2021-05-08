package utils

import (
	"fmt"
	"strconv"

	"github.com/saveio/edge/common/config"
)

func UploadFile(path, password, desc string, WhiteList []string, encryptPassword, url string, share bool,
	duration, proveLevel string, privilege uint64, copyNum string, storeType int64, realFileSize uint64) ([]byte, error) {
	var err error
	var durationVal, proveLevelVal, copyNumVal interface{}
	var durationF float64
	if len(duration) > 0 {
		durationF, err = strconv.ParseFloat(duration, 64)
		if err != nil {
			return nil, err
		}
		durationVal = durationF * float64(config.Parameters.BaseConfig.BlockTime)
	}
	if len(proveLevel) > 0 {
		proveLevelVal, err = strconv.ParseFloat(proveLevel, 64)
		if err != nil {
			return nil, err
		}
	}
	if len(copyNum) > 0 {
		copyNumVal, err = strconv.ParseFloat(copyNum, 64)
		if err != nil {
			return nil, err
		}
	}
	ret, dErr := sendRpcRequest("uploadfile", []interface{}{path, password, desc, WhiteList, encryptPassword, url,
		share, durationVal, proveLevelVal, float64(privilege), copyNumVal, storeType, realFileSize})
	if dErr != nil {
		fmt.Printf("dErr %v\n", dErr)
		return nil, dErr.Error
	}
	return ret, nil
}

func DownloadFile(args ...interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("downloadFile", args)
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func DeleteFile(args ...interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("deletefile", args)
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetUploadFiles(fileType, offset, limit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getuploadfiles", []interface{}{fileType, offset, limit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetDownloadFiles(fileType, offset, limit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getdownloadfiles", []interface{}{fileType, offset, limit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetTransferList(transferType, offset, limit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("gettransferlist", []interface{}{transferType, offset, limit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func CalculateUploadFee(path, duration, proveLevel, times, copyNum string, WhiteList []string) ([]byte, error) {
	ret, dErr := sendRpcRequest("calculateuploadfee", []interface{}{path, duration, proveLevel, times, copyNum, WhiteList})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetDownloadFileInfo(url string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getdownloadfileinfo", []interface{}{url})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func EncryptFile(path, password string) ([]byte, error) {
	ret, dErr := sendRpcRequest("encryptfile", []interface{}{path, password})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func DecryptFile(path, password string) ([]byte, error) {
	ret, dErr := sendRpcRequest("decryptfile", []interface{}{path, password})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetFileShareIncome(begin, end, offset, limit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getfileshareincome", []interface{}{begin, end, offset, limit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetFileShareRevenue() ([]byte, error) {
	ret, dErr := sendRpcRequest("getfilesharerevenue", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func WhiteListOperate(fileHash string, operation uint64, list []map[string]interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("whitelistoperate", []interface{}{fileHash, operation, list})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetFileWhiteList(fileHash string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getfilewhitelist", []interface{}{fileHash})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetFileProveDetail(fileHash string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getprovedetail", []interface{}{fileHash})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func GetUserSpace(addr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getuserspace", []interface{}{addr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func SetUserSpace(addr, password string, size, second map[string]interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("setuserspace", []interface{}{addr, password, size, second})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetUserSpaceCost(addr string, size, second map[string]interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("getuserspacecost", []interface{}{addr, size, second})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetUserSpaceRecords(addr, offset, limit string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getuserspacerecords", []interface{}{addr, offset, limit})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
