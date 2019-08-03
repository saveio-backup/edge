package utils

import (
	"strconv"
)

func UploadFile(path, desc string, WhiteList []string, encryptPassword, url string, share bool, duration, interval string, privilege uint64, copyNum string, storeType int64) ([]byte, error) {
	var err error
	var durationVal, intervalVal, copyNumVal interface{}
	if len(duration) > 0 {
		durationVal, err = strconv.ParseFloat(duration, 64)
		if err != nil {
			return nil, err
		}
	}
	if len(interval) > 0 {
		intervalVal, err = strconv.ParseFloat(interval, 64)
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
	ret, dErr := sendRpcRequest("uploadfile", []interface{}{path, desc, WhiteList, encryptPassword, url, share, durationVal, intervalVal, float64(privilege), copyNumVal, storeType})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func DownloadFile(fileHash, url, link, pwd string, maxPeerNum uint64, setFileName bool) ([]byte, error) {
	ret, dErr := sendRpcRequest("downloadfile", []interface{}{fileHash, url, link, pwd, float64(maxPeerNum), setFileName})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func DeleteFile(hash string) ([]byte, error) {
	ret, dErr := sendRpcRequest("deletefile", []interface{}{hash})
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
func CalculateUploadFee(path, duration, interval, times, copyNum string, WhiteList []string) ([]byte, error) {
	ret, dErr := sendRpcRequest("calculateuploadfee", []interface{}{path, duration, interval, times, copyNum, WhiteList})
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
func GetUserSpace(addr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getuserspace", []interface{}{addr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
func SetUserSpace(addr string, size, second map[string]interface{}) ([]byte, error) {
	ret, dErr := sendRpcRequest("setuserspace", []interface{}{addr, size, second})
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
