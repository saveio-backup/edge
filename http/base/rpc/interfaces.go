package rpc

import (
	"errors"

	"github.com/saveio/edge/dsp"
)

var DspService *dsp.Endpoint

func convertSliceToMap(cmd []interface{}, key []string) map[string]interface{} {
	if len(cmd) != len(key) {
		return nil
	}
	m := make(map[string]interface{})
	for i, k := range key {
		m[k] = cmd[i]
	}
	return m
}

func parseRestResult(ret map[string]interface{}) (interface{}, *dsp.DspErr) {
	code, ok := ret["Error"].(int64)
	if ok && code == 0 {
		result, _ := ret["Result"]
		return result, nil
	}
	if !ok {
		return nil, &dsp.DspErr{Code: dsp.INTERNAL_ERROR, Error: dsp.ErrMaps[dsp.INTERNAL_ERROR]}
	}
	errMsg, ok := ret["Desc"].(string)
	if !ok {
		return nil, &dsp.DspErr{Code: code, Error: dsp.ErrMaps[dsp.INTERNAL_ERROR]}
	}
	return nil, &dsp.DspErr{Code: code, Error: errors.New(errMsg)}
}
