package rpc

import (
	"github.com/saveio/themis/common/log"
	berr "github.com/saveio/themis/http/base/error"
)

func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.SUCCESS, true)
}
