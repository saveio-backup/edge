package rpc

import (
	Err "github.com/saveio/themis/http/base/error"
)

func responseSuccess(result interface{}) map[string]interface{} {
	return responsePack(Err.SUCCESS, result)
}
func responsePackError(errcode int64, errMsg string) map[string]interface{} {
	resp := map[string]interface{}{
		"error": errcode,
		"desc":  errMsg,
	}
	return resp
}

func responsePack(errcode int64, result interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"error":  errcode,
		"desc":   Err.ErrMap[errcode],
		"result": result,
	}
	return resp
}
