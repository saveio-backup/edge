package rpc

import (
	"fmt"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

func RegisterNode(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"NodeAddr", "Volume", "ServiceTime"})
	v := rest.RegisterNode(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func UnregisterNode(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.UnregisterNode(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func NodeQuery(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.NodeQuery(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func NodeUpdate(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"NodeAddr", "Volume", "ServiceTime"})
	v := rest.NodeUpdate(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func NodeWithdrawProfit(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.NodeWithdrawProfit(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func RegisterUrl(cmd []interface{}) map[string]interface{} {
	fmt.Printf("+++++++cmd %v\n", cmd)
	params := convertSliceToMap(cmd, []string{"Url", "FileHash", "FileName", "BlocksRoot", "FileOwner", "FileSize", "TotalCount"})
	v := rest.RegisterUrl(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func BindUrl(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Url", "Link"})
	v := rest.BindUrl(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryLink(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Url"})
	v := rest.QueryLink(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func RegisterHeader(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Header", "Desc", "Ttl"})
	v := rest.RegisterHeader(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryRegInfos(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.QueryRegInfos(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryRegInfo(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Pubkey"})
	v := rest.QueryRegInfo(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryHostInfos(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.QueryHostInfos(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryHostInfo(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.QueryHostInfo(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
func QueryPublicIP(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.QueryPublicIP(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
