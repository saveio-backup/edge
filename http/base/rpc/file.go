package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

// file apis

func UploadFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 7 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Path", "Desc", "WhiteList", "EncryptPassword", "Url", "Share", "Duration"})
	v := rest.UploadFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DeleteFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Hash"})
	v := rest.DeleteFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DownloadFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 5 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Hash", "Url", "Link", "Password", "MaxPeerNum"})
	v := rest.DownloadFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUploadFiles(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetUploadFiles(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetDownloadFiles(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetDownloadFiles(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetTransferList(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Type", "Offset", "Limit"})
	v := rest.GetTransferList(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func CalculateUploadFee(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Path", "Duration"})
	v := rest.CalculateUploadFee(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetDownloadFileInfo(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Url"})
	v := rest.GetDownloadFileInfo(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func EncryptFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Path", "Password"})
	v := rest.EncryptFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DecryptFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Path", "Password"})
	v := rest.DecryptFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetFileShareIncome(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 4 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
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
	if len(cmd) < 3 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"FileHash", "Operation", "List"})
	v := rest.WhiteListOperate(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetFileWhiteList(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"FileHash"})
	v := rest.GetFileWhiteList(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUserSpace(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.GetUserSpace(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func SetUserSpace(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 4 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Size", "Addr", "Size", "Second"})
	v := rest.SetUserSpace(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetUserSpaceRecords(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePack(dsp.INVALID_PARAMS, "")
	}
	params := convertSliceToMap(cmd, []string{"Addr", "Offset", "Limit"})
	v := rest.GetUserSpaceRecords(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
