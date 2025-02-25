package rest

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common/log"
)

const TLS_PORT int = 443

type ApiServer interface {
	Start() error
	Stop()
}

func ResponsePack(errCode int64) map[string]interface{} {
	resp := map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    "",
		"Version": "1.0.0",
	}
	return resp
}

func ResponsePackWithErrMsg(errCode int64, errMsg string) map[string]interface{} {
	resp := map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    strings.ToUpper(errMsg),
		"Version": "1.0.0",
	}
	return resp
}

func GetDspService() *dsp.Endpoint {
	svr := dsp.DspService
	if svr != nil {
		return svr
	}
	return &dsp.Endpoint{}
}

// Handle for themis go sdk
// get node verison
func GetNodeVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	var version string
	if dsp.DspService != nil {
		var err *dsp.DspErr
		version, err = dsp.DspService.GetNodeVersion()
		if err != nil {
			return ResponsePackWithErrMsg(err.Code, err.Error.Error())
		}
	}
	resp["Result"] = version
	return resp
}

// get networkid
func GetNetworkId(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	resp["Result"] = dsp.DspService.GetNetworkId()
	return resp
}

//get block height
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	height, err := dsp.DspService.GetBlockHeight()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = height
	return resp
}

//get block hash by height
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	height, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	hash, derr := dsp.DspService.GetBlockHash(uint32(height))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	if len(hash) == 0 {
		return ResponsePack(dsp.CHAIN_UNKNOWN_BLOCK)
	}
	resp["Result"] = hash
	return resp
}

//get block by hash
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str := cmd["Hash"].(string)
	if len(str) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	raw, _ := cmd["Raw"].(string)
	block, err := dsp.DspService.GetBlockByHash(str, raw)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"], resp["Error"] = block, dsp.SUCCESS
	return resp
}

//get block height by transaction hash
func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok || len(str) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	height, err := dsp.DspService.GetBlockHeightByTxHash(str)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = height
	return resp
}

//get block transaction hashes by height
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	height, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	res, derr := dsp.DspService.GetBlockTxsByHeight(uint32(height))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = res
	return resp
}

//get block by height
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if len(param) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	raw, _ := cmd["Raw"].(string)
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	index := uint32(height)

	block, derr := dsp.DspService.GetBlockByHeight(index, raw)
	if err != nil || block == nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = block
	return resp
}

//get transaction by hash
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	raw, _ := cmd["Raw"].(string)
	tx, err := dsp.DspService.GetTransactionByHash(str, raw)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = tx
	return resp
}

//get smartcontract event by height
func GetSmartCodeEventTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if len(param) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	eInfos, derr := dsp.DspService.GetSmartCodeEventTxsByHeight(uint32(height))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = eInfos
	return resp
}

//get smartcontract event by transaction hash
func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {

	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	notify, err := dsp.DspService.GetSmartCodeEventByTxHash(str)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = notify
	return resp
}

//get smartcontract event by transaction hash
func GetSmartCodeEventByEventId(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("cmdcmd %v", cmd)
	resp := ResponsePack(dsp.SUCCESS)
	contract, ok := cmd["Contract"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	addr, _ := cmd["Addr"].(string)
	eventIdStr, _ := cmd["EventId"].(string)
	eventId := uint32(0)
	if len(eventIdStr) > 0 {
		eId, err := strconv.ParseUint(eventIdStr, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
		eventId = uint32(eId)
	}
	events, err := dsp.DspService.GetSmartContractEventByEventId(contract, addr, eventId)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = events
	return resp
}

//get contract state
func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	raw, _ := cmd["Raw"].(string)
	state, err := dsp.DspService.GetContractState(str, raw)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = state
	return resp
}

//get storage from contract
func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	key := cmd["Key"].(string)
	value, err := dsp.DspService.GetStorage(str, key)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = value
	return resp
}

//get balance of address
func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addrBase58, _ := cmd["Addr"].(string)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_DSP, dsp.ErrMaps[dsp.NO_DSP].Error())
	}
	balance, err := dsp.DspService.GetBalance(addrBase58)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = balance
	return resp
}

//get balance history of address
func GetBalanceHistory(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addrBase58, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	limitStr, ok := cmd["Limit"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	result, derr := dsp.DspService.GetBalanceHistory(addrBase58, limitStr)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = result
	return resp
}

//get merkle proof by transaction hash
func GetMerkleProof(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	proof, err := dsp.DspService.GetMerkleProof(str)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = proof
	return resp
}

//get avg gas price in block
//[TODO] need change themis hcom.GetGasPrice return gasprice and height as string
//[TODO] or just return gasprice
func GetGasPrice(cmd map[string]interface{}) map[string]interface{} {
	result, err := dsp.DspService.GetGasPrice()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp := ResponsePack(dsp.SUCCESS)
	resp["Result"] = result
	return resp
}

//get allowance
func GetAllowance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	fromAddrStr, ok := cmd["From"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	toAddrStr, ok := cmd["To"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	rsp, err := dsp.DspService.GetAllowance(asset, fromAddrStr, toAddrStr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = rsp
	return resp
}

//get memory pool transaction count
func GetMemPoolTxCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	count, err := dsp.DspService.GetMemPoolTxCount()
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = count
	return resp
}

//get memory poll transaction state
func GetMemPoolTxState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	entryInfo, err := dsp.DspService.GetMemPoolTxState(str)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}

	resp["Result"] = entryInfo
	return resp
}

func GetTxByHeightAndLimit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	txType, err := dsp.RequiredStrToUint64(cmd["Type"])
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	asset, _ := cmd["Asset"].(string)
	h, _ := cmd["Height"].(string)
	l, _ := cmd["Limit"].(string)
	skip, _ := cmd["SkipTxCountFromBlock"].(string)
	ignoreStr, _ := cmd["IgnoreOtherContract"].(string)
	ignoreOtherCont := len(ignoreStr) > 0 && strings.ToLower(ignoreStr) == "true"
	log.Debugf("GetTxByHeightAndLimit CMD %v", cmd)
	txs, derr := dsp.DspService.GetTxByHeightAndLimit(addr, asset, txType, h, l, skip, ignoreOtherCont)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = txs
	return resp
}

//asset transfer direct
func AssetTransferDirect(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	to, ok := cmd["To"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if strings.ToLower(asset) == "save" {
		asset = "usdt"
	}
	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	txHash, err := dsp.DspService.AssetTransferDirect(to, asset, amountStr)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = txHash
	return resp
}

//asset transfer direct
func BatchAssetTransferDirect(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	to, ok := cmd["To"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if strings.ToLower(asset) == "save" {
		asset = "usdt"
	}
	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	times, ok := cmd["Times"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	all, err := dsp.DspService.BatchAssetTransferDirect(to, asset, amountStr, int(times))
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = all
	return resp
}

func GetConfig(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	newCfg := dsp.DspService.GetConfigs()
	resp["Result"] = newCfg
	return resp
}

func SetConfig(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	newCfg, err := dsp.DspService.SetConfigs(cmd)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	resp["Result"] = newCfg
	return resp
}

func SwitchChain(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	chainId, ok := cmd["ChainId"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	config, _ := cmd["Config"].(string)
	err := dsp.DspService.SwitchChain(chainId, config)
	if err != nil {
		return ResponsePackWithErrMsg(err.Code, err.Error.Error())
	}
	return resp
}

func InvokeSmartContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	version, ok := cmd["Version"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	verBufs, err := hex.DecodeString(version)
	if err != nil || len(verBufs) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	contractAddr, ok := cmd["Contract"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	method, ok := cmd["Method"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params, _ := cmd["Params"].([]interface{})
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gasPrice, _ := cmd["GasPrice"].(float64)
	gasLimit, _ := cmd["GasLimit"].(float64)
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	tx, derr := dsp.DspService.InvokeNativeContract(verBufs[0], contractAddr, method, params, uint64(gasPrice), uint64(gasLimit))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}

func PreExecSmartContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	version, ok := cmd["Version"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	verBufs, err := hex.DecodeString(version)
	if err != nil || len(verBufs) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	contractAddr, ok := cmd["Contract"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	method, ok := cmd["Method"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params, _ := cmd["Params"].([]interface{})
	ret, derr := dsp.DspService.PreExecNativeContract(verBufs[0], contractAddr, method, params)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	m := make(map[string]interface{}, 0)
	m["Data"] = ret
	resp["Result"] = m
	return resp
}

func PreExecSmartContractToJSON(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	version, ok := cmd["Version"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	verBufs, err := hex.DecodeString(version)
	if err != nil || len(verBufs) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	contractAddr, ok := cmd["Contract"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	method, ok := cmd["Method"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	params, _ := cmd["Params"].([]interface{})
	ret, derr := dsp.DspService.PreExecNativeContract(verBufs[0], contractAddr, method, params)
	// fmt.Println("ret===", ret, derr)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	origData, _ := hex.DecodeString(ret.(string))
	// fmt.Println("origData", origData)
	var nm interface{}
	if err := json.Unmarshal(origData, &nm); err != nil {
		// fmt.Printf("err %v", err)
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	resp["Result"] = nm
	return resp
}

func GetFsContractSetting(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, derr := dsp.DspService.GetFsConfig()
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetNetworkState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, derr := dsp.DspService.GetNetworkState()
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetModuleState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	ret, derr := dsp.DspService.GetModuleState()
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}

func GetChainId(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	m := make(map[string]interface{}, 0)
	m["ChainId"] = dsp.DspService.GetChainId()
	resp["Result"] = m
	return resp
}

func ReconnectChannelPeers(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Peers"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	peers := make([]string, 0, len(v))
	for _, str := range v {
		peer, ok := str.(string)
		if !ok {
			continue
		}
		peers = append(peers, peer)
	}
	res := dsp.DspService.ReconnectChannelPeers(peers)
	m := make(map[string]interface{}, 0)
	m["Peers"] = res
	resp["Result"] = m
	return resp
}

func GetChainIdList(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	m := make(map[string]interface{}, 0)
	ids, derr := GetDspService().GetChainIdList()
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	m["Ids"] = ids
	currentId := ""
	if dsp.DspService != nil {
		currentId = GetDspService().GetChainId()
	}
	m["CurrentId"] = currentId
	resp["Result"] = m
	return resp
}

func RemoveDBDir(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	v, ok := cmd["Type"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	derr := dsp.DspService.RemoveDBDir(dsp.DBType(v))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	return resp
}

func RegisterHeader(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	header, ok := cmd["Header"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	desc, ok := cmd["Desc"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ttl, ok := cmd["Ttl"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	ret, derr := dsp.DspService.RegisterHeader(header, desc, uint64(ttl))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = ret
	return resp
}
