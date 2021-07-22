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
