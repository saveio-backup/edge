package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

func CreateSector(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"SectorId", "ProveLevel", "Size"})
	v := rest.CreateSector(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DeleteSector(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"SectorId"})
	v := rest.DeleteSector(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetSectorInfo(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"SectorId"})
	v := rest.GetSectorInfo(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetSectorInfosForNode(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.GetSectorInfosForNode(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
