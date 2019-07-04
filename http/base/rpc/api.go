package rpc

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

// api interfaces

// Handle for themis go sdk
// get node verison
func GetNodeVersion(cmd []interface{}) map[string]interface{} {

	params := convertSliceToMap(cmd, []string{})
	v := rest.GetNodeVersion(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)

}

// get networkid
func GetNetworkId(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetNetworkId(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)

}

//get block height
func GetBlockHeight(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetBlockHeight(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get block hash by height
func GetBlockHash(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Height"})
	v := rest.GetBlockHash(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get block by hash
func GetBlockByHash(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash", "Raw"})
	v := rest.GetBlockByHash(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get block height by transaction hash
func GetBlockHeightByTxHash(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash"})
	v := rest.GetBlockHeightByTxHash(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get block transaction hashes by height
func GetBlockTxsByHeight(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Height"})
	v := rest.GetBlockTxsByHeight(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get block by height
func GetBlockByHeight(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Height", "Raw"})
	v := rest.GetBlockByHeight(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get transaction by hash
func GetTransactionByHash(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash", "Raw"})
	v := rest.GetTransactionByHash(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get smartcontract event by height
func GetSmartCodeEventTxsByHeight(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Height"})
	v := rest.GetSmartCodeEventTxsByHeight(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get smartcontract event by transaction hash
func GetSmartCodeEventByTxHash(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash"})
	v := rest.GetSmartCodeEventByTxHash(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

//get contract state
func GetContractState(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash", "Raw"})
	v := rest.GetContractState(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetStorage(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 2 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash", "Key"})
	v := rest.GetStorage(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetBalance(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr"})
	v := rest.GetBalance(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetMerkleProof(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash"})
	v := rest.GetMerkleProof(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetGasPrice(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetGasPrice(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetAllowance(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Asset", "From", "To"})
	v := rest.GetAllowance(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetMemPoolTxCount(cmd []interface{}) map[string]interface{} {
	params := convertSliceToMap(cmd, []string{})
	v := rest.GetMemPoolTxCount(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetMemPoolTxState(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Hash"})
	v := rest.GetMemPoolTxState(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func GetTxByHeightAndLimit(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 6 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"Addr", "Type", "Asset", "Height", "Limit", "SkipTxCountFromBlock"})
	v := rest.GetTxByHeightAndLimit(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func AssetTransferDirect(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 3 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"To", "Asset", "Amount"})
	v := rest.AssetTransferDirect(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}

func SetConfig(cmd []interface{}) map[string]interface{} {
	if len(cmd) < 1 {
		return responsePackError(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params := convertSliceToMap(cmd, []string{"DownloadPath"})
	v := rest.SetConfig(params)
	ret, err := parseRestResult(v)
	if err != nil {
		return responsePackError(err.Code, err.Error.Error())
	}
	return responseSuccess(ret)
}
