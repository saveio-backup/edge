package rest

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/utils/plot"
)

func GeneratePlotFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	system, ok := cmd["System"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	numericID, ok := cmd["NumericID"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	startNonce, ok := cmd["StartNonce"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	nonces, ok := cmd["Nonces"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	cfg := &plot.PlotConfig{
		Sys:        system,
		NumericID:  numericID,
		StartNonce: uint64(startNonce),
		Nonces:     uint64(nonces),
		Path:       path,
	}
	err := plot.Plot(cfg)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
	}

	m := make(map[string]interface{}, 0)
	m["NumericID"] = numericID
	m["StartNonce"] = startNonce
	m["Nonces"] = nonces
	m["Path"] = path
	m["PlotFileName"] = plot.GetPlotFileName(cfg)
	resp["Result"] = m
	return resp
}
