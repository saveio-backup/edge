package rest

import (
	"fmt"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
)

func GeneratePlotFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	system, ok := cmd["System"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	numericID, _ := cmd["NumericID"].(string)
	if len(numericID) == 0 {
		acc, err := dsp.DspService.GetCurrentAccount()
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error.Error())
		}
		numericID = fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))
	}

	startNonce, _ := utils.ToUint64(cmd["StartNonce"])
	nonces, _ := utils.ToUint64(cmd["Nonces"])

	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	size, _ := utils.ToUint64(cmd["Size"])
	num, _ := utils.ToUint64(cmd["Num"])

	if size == 0 {
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

		m := make(map[string]interface{})
		m["NumericID"] = numericID
		m["StartNonce"] = startNonce
		m["Nonces"] = nonces
		m["Path"] = path
		m["PlotFileName"] = plot.GetPlotFileName(cfg)

	}
	var err error
	startNonce, err = plot.GetMinStartNonce(numericID, path)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	nonces = size / plot.DEFAULT_PLOT_SIZEKB

	ms := make([]interface{}, 0)
	for i := uint64(0); i < uint64(num); i++ {

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

		m := make(map[string]interface{})
		m["NumericID"] = numericID
		m["StartNonce"] = startNonce
		m["Nonces"] = nonces
		m["Path"] = path
		m["PlotFileName"] = plot.GetPlotFileName(cfg)

		ms = append(ms, m)

		startNonce += nonces
	}
	resp["Result"] = ms
	return resp

}
