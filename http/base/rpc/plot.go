package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

func GeneratePlotFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 5 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	params := convertSliceToMap(cmd, []string{"System", "NumericID", "StartNonce", "Nonces", "Path"})
	v := rest.GeneratePlotFile(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetAllPlotFiles(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	params := convertSliceToMap(cmd, []string{"Path"})
	v := rest.GetAllPlotFiles(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func AddPlotFile(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	params := convertSliceToMap(cmd, []string{"FileName", "CreateSector"})
	v := rest.AddPlotFileToMine(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func AddPlotFiles(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	params := convertSliceToMap(cmd, []string{"Directory", "CreateSector"})
	v := rest.AddPlotFolderToMine(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
