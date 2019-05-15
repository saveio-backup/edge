package rest

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	cutils "github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	berr "github.com/saveio/edge/http/base/error"
	hcomm "github.com/saveio/edge/http/common"
	"github.com/saveio/edge/http/util"
	"github.com/saveio/themis-go-sdk/usdt"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	chanCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/payload"
	bcomn "github.com/saveio/themis/http/base/common"
)

const TLS_PORT int = 443

var DspService *dsp.Endpoint

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

// Handle for themis go sdk
// get node verison
func GetNodeVersion(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	version, err := DspService.Chain.GetVersion()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = version
	return resp
}

// get networkid
func GetNetworkId(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = fmt.Sprintf("%d", config.Parameters.BaseConfig.NetworkId)
	return resp
}

//get block height
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	height, err := DspService.Chain.GetCurrentBlockHeight()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = height
	return resp
}

//get block hash by height
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	hash, err := DspService.Chain.GetBlockHash(uint32(height))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	if hash == common.UINT256_EMPTY {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}

	resp["Result"] = hash.ToHexString()
	return resp
}

//get block by hash
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	if len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}

	block, err := DspService.Chain.GetBlockByHash(str)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if block == nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if block.Header == nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}

	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)

		resp["Result"] = common.ToHexString(w.Bytes())
		resp["Error"] = berr.SUCCESS
		return resp
	}

	resp["Result"], resp["Error"] = bcomn.GetBlockInfo(block), berr.SUCCESS
	return resp
}

//get block height by transaction hash
func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok || len(str) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	height, err := DspService.Chain.GetBlockHeightByTxHash(str)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}

	resp["Result"] = height
	return resp
}

//get block transaction hashes by height
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok || len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	data, err := DspService.Chain.GetBlockTxHashesByHeight(uint32(height))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	res := hcomm.GetBlockTransactions(data)

	resp["Result"] = res
	return resp
}

//get block by height
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)

	block, err := DspService.Chain.GetBlockByHeight(index)
	if err != nil || block == nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
	} else {
		resp["Result"] = bcomn.GetBlockInfo(block)
	}
	return resp
}

//get transaction by hash
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	//[TODO] need support height later ？
	var height uint32

	tx, err := DspService.Chain.GetTransaction(str)
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}

	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		return resp
	}
	tran := bcomn.TransArryByteToHexString(tx)
	tran.Height = height
	resp["Result"] = tran
	return resp
}

//get smartcontract event by height
func GetSmartCodeEventTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param, ok := cmd["Height"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	eInfos, err := DspService.Chain.GetSmartContractEventByBlock(uint32(height))
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if eInfos == nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = eInfos
	return resp
}

//get smartcontract event by transaction hash
func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {

	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	notify, err := DspService.Chain.GetSmartContractEvent(str)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if notify == nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}

	resp["Result"] = notify
	return resp
}

//get contract state
func GetContractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	smartContract, err := DspService.Chain.GetSmartContract(str)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}

	var deployCode payload.DeployCode
	deployCode = payload.DeployCode(*smartContract)

	contract := &deployCode

	if contract == nil {
		return ResponsePack(berr.UNKNOWN_CONTRACT)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)

		contract.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		return resp
	}
	resp["Result"] = bcomn.TransPayloadToHex(contract)
	return resp
}

//get storage from contract
func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	value, err := DspService.Chain.GetStorage(str, item)

	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}

//get balance of address
func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addrBase58, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return ResponsePackWithErrMsg(berr.INVALID_PARAMS, err.Error())
	}
	balance, err := hcomm.GetBalance(address)
	if err != nil {
		return ResponsePackWithErrMsg(berr.INVALID_PARAMS, err.Error())
	}
	type balanceResp struct {
		Address       string
		Name          string
		Symbol        string
		Decimals      int
		Balance       uint64
		BalanceFormat string
	}
	usdt, err := strconv.ParseUint(balance.Usdt, 10, 64)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	bals := make([]*balanceResp, 0)
	bals = append(bals, &balanceResp{
		Address:       addrBase58,
		Name:          "Save Power",
		Symbol:        "SAVE",
		Decimals:      9,
		Balance:       usdt,
		BalanceFormat: utils.FormatUsdt(usdt),
	}, &balanceResp{
		Address:       addrBase58,
		Name:          "NEO",
		Symbol:        "NEO",
		Decimals:      1,
		Balance:       0,
		BalanceFormat: "0",
	}, &balanceResp{
		Address:       addrBase58,
		Name:          "Ontology",
		Symbol:        "ONT",
		Decimals:      1,
		Balance:       0,
		BalanceFormat: "0",
	})
	resp["Result"] = bals
	return resp
}

//get merkle proof by transaction hash
func GetMerkleProof(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	proof, err := DspService.Chain.GetMerkleProof(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = proof
	return resp
}

//get avg gas price in block
//[TODO] need change themis hcom.GetGasPrice return gasprice and height as string
//[TODO] or just return gasprice
func GetGasPrice(cmd map[string]interface{}) map[string]interface{} {
	result, err := DspService.Chain.GetGasPrice()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = result
	return resp
}

//get allowance
func GetAllowance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddrStr, ok := cmd["From"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddrStr, ok := cmd["To"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	fromAddr, err := bcomn.GetAddress(fromAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	toAddr, err := bcomn.GetAddress(toAddrStr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	rsp, err := bcomn.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = rsp
	return resp
}

//get memory pool transaction count
func GetMemPoolTxCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := DspService.Chain.GetMemPoolTxCount()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//get memory poll transaction state
func GetMemPoolTxState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	entryInfo, err := DspService.Chain.GetMemPoolTxState(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = entryInfo
	return resp
}

func GetTxByHeightAndLimit(cmd map[string]interface{}) map[string]interface{} {
	fmt.Printf("cmdsss:%v\n", cmd)
	resp := ResponsePack(berr.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	txType, err := util.RequiredStrToUint64(cmd["Type"])
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	asset, _ := cmd["Asset"].(string)
	if len(asset) == 0 {
		asset = "save"
	} else {
		asset = strings.ToLower(asset)
	}
	h, _ := cmd["Height"].(string)
	l, _ := cmd["Limit"].(string)

	height := uint32(0)
	if len(h) > 0 {
		height64, err := strconv.ParseUint(h, 10, 32)
		if err != nil {
			return ResponsePack(berr.INTERNAL_ERROR)
		}
		height = uint32(height64)
	}
	limit := uint32(0)
	if len(l) > 0 {
		limit64, err := strconv.ParseUint(l, 10, 32)
		if err != nil {
			return ResponsePack(berr.INTERNAL_ERROR)
		}
		limit = uint32(limit64)
	}
	currentHeight, err := DspService.Chain.GetCurrentBlockHeight()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if len(h) == 0 {
		height = currentHeight
	}
	if height > currentHeight {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	type txResp struct {
		Txid         string
		Type         uint
		From         string
		To           string
		Amount       uint64
		AmountFormat string
		FeeFormat    string
		Asset        string
		Timestamp    uint32
		BlockHeight  uint32
	}

	txs := make([]*txResp, 0)
	for i := int32(height); i >= 0; i-- {
		blk, err := DspService.Chain.GetBlockByHeight(uint32(i))
		if err != nil || blk == nil {
			continue
		}
		for _, t := range blk.Transactions {
			hash := t.Hash()
			event, err := DspService.Chain.GetSmartContractEvent(hash.ToHexString())
			if err != nil || event == nil {
				continue
			}
			for _, n := range event.Notify {
				states, ok := n.States.([]interface{})
				if !ok {
					continue
				}
				if len(states) != 4 || states[0] != "transfer" {
					continue
				}
				from := states[1].(string)
				to := states[2].(string)
				if asset == "save" && n.ContractAddress == usdt.USDT_CONTRACT_ADDRESS.ToHexString() {
					if txType == hcomm.TX_TYPE_ALL && (from != addr && to != addr) {
						continue
					}
					if txType == hcomm.TX_TYPE_SEND && from != addr {
						continue
					}
					if txType == hcomm.TX_TYPE_RECEIVE && to != addr {
						continue
					}
					amountFormat := utils.FormatUsdt(states[3].(uint64))
					sendType := hcomm.TX_TYPE_SEND
					if to == addr {
						sendType = hcomm.TX_TYPE_RECEIVE
					}
					tx := &txResp{
						Txid:         hash.ToHexString(),
						From:         from,
						To:           to,
						Type:         uint(sendType),
						Asset:        "save",
						Amount:       states[3].(uint64),
						AmountFormat: amountFormat,
						FeeFormat:    utils.FormatUsdt(10000000),
						BlockHeight:  uint32(i),
					}
					tx.Timestamp = blk.Header.Timestamp
					txs = append(txs, tx)
					if limit > 0 && uint32(len(txs)) >= limit {
						resp["Result"] = txs
						return resp
					}
					continue
				}
			}
		}
		if i == 0 {
			break
		}
	}
	resp["Result"] = txs
	return resp
}

//asset transfer direct
func AssetTransferDirect(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	wal, err := wallet.OpenWallet(config.WalletDatFilePath())
	if err != nil {
		return ResponsePack(berr.OPEN_WALLET_ERROR)
	}

	accDefault, err := wal.GetDefaultAccount([]byte(config.Parameters.BaseConfig.WalletPwd))
	if err != nil {
		return ResponsePack(berr.DEFAULT_ACCOUNT_NOT_FOUND)
	}

	to, ok := cmd["To"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	asset, ok := cmd["Asset"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if strings.ToLower(asset) == "save" {
		asset = "usdt"
	}

	var amount float64
	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	} else {
		temp, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return ResponsePack(berr.INVALID_PARAMS)
		}
		amount = temp
	}
	realAmount := uint64(amount * 1000000000)
	log.Debugf("amount :%v", realAmount)
	txHash, err := cutils.Transfer(chanCfg.DEFAULT_GAS_PRICE, chanCfg.DEFAULT_GAS_LIMIT, accDefault, asset, accDefault.Address.ToBase58(), to, realAmount)
	if err != nil {
		return ResponsePackWithErrMsg(berr.INTERNAL_ERROR, err.Error())
	}
	resp["Result"] = txHash
	return resp
}

//Handle for Dsp
func RegisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addr, ok := cmd["NodeAddr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	volumeStr, ok := cmd["Volume"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	volume, err := strconv.ParseUint(volumeStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	serviceTime, err := strconv.ParseUint(serviceTimeStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.RegisterNode(addr, volume, serviceTime)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

func UnregisterNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	tx, err := DspService.UnregisterNode()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

func NodeQuery(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	walletAddr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	fsNodeInfo, err := DspService.NodeQuery(walletAddr)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = fsNodeInfo
	return resp
}

func NodeUpdate(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	var addr string
	var volume, serviceTime uint64

	_, ok := cmd["NodeAddr"].(string)
	if ok {
		addr = cmd["NodeAddr"].(string)
	}

	volumeStr, ok := cmd["Volume"].(string)
	if ok {
		temp, err := strconv.ParseUint(volumeStr, 10, 64)
		if err != nil {
			return ResponsePack(berr.INVALID_PARAMS)
		}
		volume = temp
	}

	serviceTimeStr, ok := cmd["ServiceTime"].(string)
	if ok {
		temp, err := strconv.ParseUint(serviceTimeStr, 10, 64)
		if err != nil {
			return ResponsePack(berr.INVALID_PARAMS)
		}
		serviceTime = temp
	}

	tx, err := DspService.NodeUpdate(addr, volume, serviceTime)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

func NodeWithdrawProfit(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	tx, err := DspService.NodeWithdrawProfit()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

//Handle for DNS
func RegisterUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	link, ok := cmd["Link"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.Dsp.RegisterFileUrl(url, link)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

func BindUrl(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	link, ok := cmd["Link"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.Dsp.BindFileUrl(url, link)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = tx
	return resp
}

func QueryLink(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	url, ok := cmd["Url"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	link := DspService.Dsp.GetLinkFromUrl(url)

	resp["Result"] = link
	return resp
}

func RegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	ip, ok := cmd["Ip"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	port, ok := cmd["Port"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	depositStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	deposit, err := strconv.ParseUint(depositStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.Chain.Native.Dns.DNSNodeReg([]byte(ip), []byte(port), deposit)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = hex.EncodeToString(common.ToArrayReverse(tx.ToArray()))
	return resp
}

func UnRegisterDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	tx, err := DspService.Chain.Native.Dns.UnregisterDNSNode()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = hex.EncodeToString(common.ToArrayReverse(tx.ToArray()))
	return resp
}

func QuitDns(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	tx, err := DspService.Chain.Native.Dns.QuitNode()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = hex.EncodeToString(common.ToArrayReverse(tx.ToArray()))
	return resp
}

func AddPos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.Chain.Native.Dns.AddInitPos(amount)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = hex.EncodeToString(common.ToArrayReverse(tx.ToArray()))
	return resp
}

func ReducePos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	amountStr, ok := cmd["Amount"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	tx, err := DspService.Chain.Native.Dns.ReduceInitPos(amount)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = hex.EncodeToString(common.ToArrayReverse(tx.ToArray()))
	return resp
}

func QueryRegInfos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	ret, err := DspService.Chain.Native.Dns.GetPeerPoolMap()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = ret.PeerPoolMap
	return resp
}

func QueryRegInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	pubkey, ok := cmd["Pubkey"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	if pubkey == "self" {
		pubkey = ""
	}

	ret, err := DspService.Chain.Native.Dns.GetPeerPoolItem(pubkey)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = ret
	return resp
}

func QueryHostInfos(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	ret, err := DspService.Chain.Native.Dns.GetAllDnsNodes()
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = ret
	return resp
}

func QueryHostInfo(cmd map[string]interface{}) map[string]interface{} {
	var address common.Address

	resp := ResponsePack(berr.SUCCESS)

	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	if addr != "self" {
		tmpaddr, err := common.AddressFromBase58(addr)
		if err != nil {
			return ResponsePack(berr.INVALID_PARAMS)
		}
		address = tmpaddr
	}

	ret, err := DspService.Chain.Native.Dns.GetDnsNodeByAddr(address)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	resp["Result"] = ret
	return resp
}

func GetUserSpace(cmd map[string]interface{}) map[string]interface{} {
	fmt.Println("GetUserSpace")
	resp := ResponsePack(berr.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	type userspace struct {
		Used      uint64
		Remain    uint64
		ExpiredAt uint64
		Balance   uint64
	}
	space, err := DspService.Dsp.GetUserSpace(addr)
	if err != nil || space == nil {
		log.Errorf("get user space err %s, space %v", err, space)
		resp["Result"] = &userspace{
			Used:      0,
			Remain:    0,
			ExpiredAt: 0,
			Balance:   0,
		}
		return resp
	}
	currentHeight, err := DspService.Chain.GetCurrentBlockHeight()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	interval := uint64(0)
	if space.ExpireHeight > uint64(currentHeight) {
		interval = space.ExpireHeight - uint64(currentHeight)
	}
	fmt.Printf("space.ExpireHeight %d\n", space.ExpireHeight)
	resp["Result"] = &userspace{
		Used:      space.Used,
		Remain:    space.Remain,
		ExpiredAt: uint64(time.Now().Unix()) + interval,
		Balance:   space.Balance,
	}
	return resp
}

func AddUserSpace(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	size, ok := cmd["Size"].(float64)
	if !ok {
		size = 0
	}
	second, ok := cmd["Second"].(float64)
	if !ok {
		second = 0
	}
	tx, err := DspService.Dsp.AddUserSpace(addr, uint64(size), uint64(second))
	if err != nil {
		log.Errorf("add user space err %s", err)
		return ResponsePackWithErrMsg(berr.INTERNAL_ERROR, err.Error())
	}
	resp["Result"] = tx

	return resp
}

func RevokeUserSpace(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	size, ok := cmd["Size"].(float64)
	if !ok {
		size = 0
	}
	second, ok := cmd["Second"].(float64)
	if !ok {
		second = 0
	}
	tx, err := DspService.Dsp.RevokeUserSpace(addr, uint64(size), uint64(second))
	if err != nil {
		log.Errorf("add user space err %s", err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = tx

	return resp
}
