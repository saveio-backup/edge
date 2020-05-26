package rpc

import (
	"fmt"

	"github.com/saveio/edge/http/rest"
)

func GetAllChannels(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetAllChannels(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func OpenChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner", "Password", "Amount"})
	v := rest.OpenChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func OpenToAllDNSChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Password", "Amount"})
	v := rest.OpenToAllDNSChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func CloseChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner", "Password"})
	v := rest.CloseChannel(params)
	fmt.Printf("rest close channel %v\n", v)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func CloseAllChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Password"})
	v := rest.CloseAllChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func DepositChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner", "Amount", "Password"})
	v := rest.DepositChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func WithdrawChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner", "Amount", "Password"})
	v := rest.WithdrawChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryChannelDeposit(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner"})
	v := rest.QueryChannelDeposit(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner"})
	v := rest.QueryChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func QueryChannelByID(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Id", "Partner"})
	v := rest.QueryChannelByID(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func TransferByChannel(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"To", "Amount", "PaymentId", "Password"})
	v := rest.TransferByChannel(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetChannelInitProgress(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetChannelInitProgress(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func ChannelCooperativeSettle(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{"Partner", "Password"})
	v := rest.ChannelCooperativeSettle(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
