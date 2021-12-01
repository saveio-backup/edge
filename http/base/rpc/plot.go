package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
	"github.com/saveio/themis/common/log"
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
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	params := convertSliceToMap(cmd, []string{"TaskId", "FileName", "CreateSector"})
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

func GetAllProvedPlotFile(cmd []interface{}) map[string]interface{} {
	v := rest.GetAllProvedPlotFile(nil)
	log.Infof("GetAllProvedPlotFile task v %v", v)
	ret, err := parseRestResult(v)
	log.Infof("parse Rest Result %v", ret, err)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetAllPocTasks(cmd []interface{}) map[string]interface{} {
	v := rest.GetAllPocTasks(nil)
	log.Infof("get all task v %v", v)
	ret, err := parseRestResult(v)
	log.Infof("parse Rest Result %v", ret, err)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
