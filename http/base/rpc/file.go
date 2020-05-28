package rpc

import (
	"fmt"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

// file apis

func UploadFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Path", "Password", "Desc", "WhiteList", "EncryptPassword", "Url", "Share", "Duration", "Interval", "Privilege", "CopyNum", "StoreType"})
	v := rest.UploadFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DeleteFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Hash", "Password", "GasLimit"})
	v := rest.DeleteUploadFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DownloadFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Hash", "Url", "Link", "DecryptPassword", "MaxPeerNum", "SetFileName", "Password"})
	checkPasswordRet := rest.CheckPassword(params)
	errorCode, _ := checkPasswordRet["Error"].(int64)
	if errorCode != 0 {
		return responsePackError(dsp.ACCOUNT_PASSWORD_WRONG, "wrong password")
	}
	fmt.Println(params)
	v := rest.DownloadFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUploadFiles(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetUploadFiles(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetDownloadFiles(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetDownloadFiles(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetTransferList(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetTransferList(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func CalculateUploadFee(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Path", "Duration"})
	v := rest.CalculateUploadFee(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetDownloadFileInfo(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Url"})
	v := rest.GetDownloadFileInfo(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func EncryptFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Path", "Password"})
	v := rest.EncryptFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DecryptFile(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Path", "Password"})
	v := rest.DecryptFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetFileShareIncome(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Begin", "End", "Offset", "Limit"})
	v := rest.GetFileShareIncome(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetFileShareRevenue(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetFileShareRevenue(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func WhiteListOperate(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"FileHash", "Operation", "List"})
	v := rest.WhiteListOperate(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetFileWhiteList(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"FileHash"})
	v := rest.GetFileWhiteList(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUserSpace(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.GetUserSpace(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func SetUserSpace(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Addr", "Password", "Size", "Second"})
	v := rest.SetUserSpace(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUserSpaceRecords(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Addr", "Offset", "Limit"})
	v := rest.GetUserSpaceRecords(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
