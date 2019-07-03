package dsp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/saveio/edge/common/config"
	hcomm "github.com/saveio/edge/http/common"
	edgeUtils "github.com/saveio/edge/utils"
	"github.com/saveio/themis-go-sdk/usdt"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/payload"
	bcomn "github.com/saveio/themis/http/base/common"
	"github.com/saveio/themis/smartcontract/service/native/film"
	cusdt "github.com/saveio/themis/smartcontract/service/native/usdt"
)

const (
	TxTypeAll     = 0
	TxTypeSend    = 1
	TxTypeReceive = 2
)

var GitCommit string

func (this *Endpoint) GetNodeVersion() (string, *DspErr) {
	version, err := this.Dsp.Chain.GetVersion()
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	max := 6
	if len(GitCommit) < max {
		max = len(GitCommit)
	}
	return version + GitCommit[:max], nil
}

// get networkid
func (this *Endpoint) GetNetworkId() string {
	return fmt.Sprintf("%d", config.Parameters.BaseConfig.NetworkId)
}

//get block height
func (this *Endpoint) GetBlockHeight() (uint32, *DspErr) {
	height, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return 0, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return height, nil
}

//get block hash by height
func (this *Endpoint) GetBlockHash(height uint32) (string, *DspErr) {
	hash, err := this.Dsp.Chain.GetBlockHash(uint32(height))
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if hash == common.UINT256_EMPTY {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: ErrMaps[CHAIN_INTERNAL_ERROR]}
	}
	return hash.ToHexString(), nil
}

//get block by hash
func (this *Endpoint) GetBlockByHash(hash, raw string) (interface{}, *DspErr) {
	if len(hash) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	var getTxBytes = false
	if raw == "1" {
		getTxBytes = true
	}

	block, err := this.Dsp.Chain.GetBlockByHash(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if block == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_BLOCK, Error: ErrMaps[CHAIN_UNKNOWN_BLOCK]}
	}
	if block.Header == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_BLOCK, Error: ErrMaps[CHAIN_UNKNOWN_BLOCK]}
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return common.ToHexString(w.Bytes()), nil
	}
	return bcomn.GetBlockInfo(block), nil
}

//get block height by transaction hash
func (this *Endpoint) GetBlockHeightByTxHash(str string) (uint32, *DspErr) {
	if len(str) == 0 {
		return 0, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	height, err := this.Dsp.Chain.GetBlockHeightByTxHash(str)
	if err != nil {
		return height, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return height, nil
}

//get block transaction hashes by height
func (this *Endpoint) GetBlockTxsByHeight(height uint32) (interface{}, *DspErr) {
	data, err := this.Dsp.Chain.GetBlockTxHashesByHeight(uint32(height))
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	res := hcomm.GetBlockTransactions(data)
	return res, nil
}

//get block by height
func (this *Endpoint) GetBlockByHeight(height uint32, raw string) (interface{}, *DspErr) {
	var getTxBytes = false
	if raw == "1" {
		getTxBytes = true
	}
	block, err := this.Dsp.Chain.GetBlockByHeight(height)
	if err != nil || block == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_BLOCK, Error: ErrMaps[CHAIN_UNKNOWN_BLOCK]}
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return common.ToHexString(w.Bytes()), nil
	} else {
		return bcomn.GetBlockInfo(block), nil
	}
}

//get transaction by hash
func (this *Endpoint) GetTransactionByHash(hash, raw string) (interface{}, *DspErr) {
	tx, err := this.Dsp.Chain.GetTransaction(hash)
	if tx == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_TX, Error: ErrMaps[CHAIN_UNKNOWN_TX]}
	}
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		return common.ToHexString(w.Bytes()), nil
	}
	tran := bcomn.TransArryByteToHexString(tx)
	//[TODO] need support height later ？
	var height uint32
	tran.Height = height
	return tran, nil
}

//get smartcontract event by height
func (this *Endpoint) GetSmartCodeEventTxsByHeight(height uint32) (interface{}, *DspErr) {
	eInfos, err := this.Dsp.Chain.GetSmartContractEventByBlock(uint32(height))
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if eInfos == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_SMARTCONTRACT_EVENT, Error: ErrMaps[CHAIN_UNKNOWN_SMARTCONTRACT_EVENT]}
	}
	return eInfos, nil
}

//get smartcontract event by transaction hash
func (this *Endpoint) GetSmartCodeEventByTxHash(hash string) (interface{}, *DspErr) {
	notify, err := this.Dsp.Chain.GetSmartContractEvent(hash)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if notify == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_SMARTCONTRACT_EVENT, Error: ErrMaps[CHAIN_UNKNOWN_SMARTCONTRACT_EVENT]}
	}
	return notify, nil
}

//get contract state
func (this *Endpoint) GetContractState(hash, raw string) (interface{}, *DspErr) {
	smartContract, err := this.Dsp.Chain.GetSmartContract(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	var deployCode payload.DeployCode
	deployCode = payload.DeployCode(*smartContract)
	contract := &deployCode
	if contract == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_SMARTCONTRACT, Error: ErrMaps[CHAIN_UNKNOWN_SMARTCONTRACT]}
	}
	if raw == "1" {
		w := bytes.NewBuffer(nil)
		contract.Serialize(w)
		return common.ToHexString(w.Bytes()), nil
	}
	return bcomn.TransPayloadToHex(contract), nil
}

//get storage from contract
func (this *Endpoint) GetStorage(hash, key string) (string, *DspErr) {
	item, err := common.HexToBytes(key)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	value, err := this.Dsp.Chain.GetStorage(hash, item)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return common.ToHexString(value), nil
}

type BalanceResp struct {
	Address       string
	Name          string
	Symbol        string
	Decimals      int
	Balance       uint64
	BalanceFormat string
}

//get balance of address
func (this *Endpoint) GetBalance(address string) ([]*BalanceResp, *DspErr) {
	addr, err := common.AddressFromBase58(address)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	balance, err := hcomm.GetBalance(addr)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	usdt, err := strconv.ParseUint(balance.Usdt, 10, 64)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	bals := make([]*BalanceResp, 0)
	bals = append(bals, &BalanceResp{
		Address:       address,
		Name:          "Save Power",
		Symbol:        "SAVE",
		Decimals:      9,
		Balance:       usdt,
		BalanceFormat: utils.FormatUsdt(usdt),
	}, &BalanceResp{
		Address:       address,
		Name:          "NEO",
		Symbol:        "NEO",
		Decimals:      1,
		Balance:       0,
		BalanceFormat: "0",
	}, &BalanceResp{
		Address:       address,
		Name:          "Ontology",
		Symbol:        "ONT",
		Decimals:      1,
		Balance:       0,
		BalanceFormat: "0",
	})
	return bals, nil
}

//get merkle proof by transaction hash
func (this *Endpoint) GetMerkleProof(hash string) (interface{}, *DspErr) {
	proof, err := this.Dsp.Chain.GetMerkleProof(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return proof, nil
}

//get avg gas price in block
//[TODO] need change themis hcom.GetGasPrice return gasprice and height as string
//[TODO] or just return gasprice
func (this *Endpoint) GetGasPrice() (uint64, *DspErr) {
	price, err := this.Dsp.Chain.GetGasPrice()
	if err != nil {
		return 0, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return price, nil
}

//get allowance
func (this *Endpoint) GetAllowance(asset, from, to string) (string, *DspErr) {
	fromAddr, err := bcomn.GetAddress(from)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	toAddr, err := bcomn.GetAddress(to)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	tx, err := bcomn.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return tx, nil
}

//get memory pool transaction count
func (this *Endpoint) GetMemPoolTxCount() (interface{}, *DspErr) {
	cnt, err := this.Dsp.Chain.GetMemPoolTxCount()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return cnt, nil
}

//get memory poll transaction state
func (this *Endpoint) GetMemPoolTxState(hash string) (interface{}, *DspErr) {
	state, err := this.Dsp.Chain.GetMemPoolTxState(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return state, nil
}

type TxResp struct {
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

func (this *Endpoint) GetTxByHeightAndLimit(addr, asset string, txType uint64, heightStr, limitStr, skipTxCntStr string) ([]*TxResp, *DspErr) {
	log.Debugf("GetTxByHeightAndLimit %v %v %v %v %v %v", addr, asset, txType, heightStr, limitStr, skipTxCntStr)
	if len(asset) == 0 {
		asset = "save"
	} else {
		asset = strings.ToLower(asset)
	}
	h := heightStr
	l := limitStr
	height := uint32(0)
	if len(h) > 0 {
		height64, err := strconv.ParseUint(h, 10, 32)
		if err != nil {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
		}
		height = uint32(height64)
	}
	limit := uint32(0)
	if len(l) > 0 {
		limit64, err := strconv.ParseUint(l, 10, 32)
		if err != nil {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
		}
		limit = uint32(limit64)
	}

	currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	if len(h) == 0 {
		height = currentHeight
	}
	if height > currentHeight {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	skipTxCnt := uint64(0)
	if len(heightStr) > 0 && len(skipTxCntStr) > 0 {
		var err error
		skipTxCnt, err = strconv.ParseUint(skipTxCntStr, 10, 32)
		if err != nil {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
		}
	}

	txs := make([]*TxResp, 0)
	events, err := this.Dsp.Chain.GetSmartContractEventByEventId(usdt.USDT_CONTRACT_ADDRESS.ToBase58(), addr, cusdt.EVENT_USDT_STATE_CHANGE)
	log.Debugf("events-len %d, addr %s-%s-%d", len(events), usdt.USDT_CONTRACT_ADDRESS.ToBase58(), addr, cusdt.EVENT_USDT_STATE_CHANGE)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	// TODO: fixed this
	tempMap := make(map[string]struct{}, 0)
	hasSkip := uint64(0)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		log.Debugf("#d event %v", i, event.Notify)
		blockHeight, err := this.Dsp.Chain.GetBlockHeightByTxHash(event.TxHash)
		if err != nil {
			continue
		}
		if blockHeight > height {
			continue
		}
		for _, n := range event.Notify {
			log.Debugf("States %v, addrs %v", n.States, n.Addresses)
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
				if txType == TxTypeAll && (from != addr && to != addr) {
					continue
				}
				if txType == TxTypeSend && from != addr {
					continue
				}
				if txType == TxTypeReceive && to != addr {
					continue
				}
				tempKey := fmt.Sprintf("%s-%v-%v", event.TxHash, states[1], states[2])
				if _, ok := tempMap[tempKey]; ok {
					continue
				}
				tempMap[tempKey] = struct{}{}
				amountFormat := utils.FormatUsdt(states[3].(uint64))
				sendType := TxTypeSend
				if to == addr {
					sendType = TxTypeReceive
				}
				tx := &TxResp{
					Txid:         event.TxHash,
					From:         from,
					To:           to,
					Type:         uint(sendType),
					Asset:        "save",
					Amount:       states[3].(uint64),
					AmountFormat: amountFormat,
					FeeFormat:    utils.FormatUsdt(10000000),
					BlockHeight:  uint32(blockHeight),
				}
				blk, err := this.Dsp.Chain.GetBlockByHeight(blockHeight)
				if err != nil {
					continue
				}
				if skipTxCnt > 0 && skipTxCnt > hasSkip {
					hasSkip++
					continue
				}
				tx.Timestamp = blk.Header.Timestamp
				txs = append(txs, tx)
				if limit > 0 && uint32(len(txs)) >= limit {
					return txs, nil
				}
				continue
			}
		}
	}
	// for i := int32(height); i >= 0; i-- {
	// 	blk, err := this.Dsp.Chain.GetBlockByHeight(uint32(i))
	// 	if err != nil || blk == nil {
	// 		continue
	// 	}
	// 	for _, t := range blk.Transactions {
	// 		hash := t.Hash()
	// 		event, err := this.Dsp.Chain.GetSmartContractEvent(hash.ToHexString())
	// 		if err != nil || event == nil {
	// 			continue
	// 		}
	// 		for _, n := range event.Notify {
	// 			states, ok := n.States.([]interface{})
	// 			if !ok {
	// 				continue
	// 			}
	// 			if len(states) != 4 || states[0] != "transfer" {
	// 				continue
	// 			}
	// 			from := states[1].(string)
	// 			to := states[2].(string)
	// 			if asset == "save" && n.ContractAddress == usdt.USDT_CONTRACT_ADDRESS.ToHexString() {
	// 				if txType == TxTypeAll && (from != addr && to != addr) {
	// 					continue
	// 				}
	// 				if txType == TxTypeSend && from != addr {
	// 					continue
	// 				}
	// 				if txType == TxTypeReceive && to != addr {
	// 					continue
	// 				}
	// 				amountFormat := utils.FormatUsdt(states[3].(uint64))
	// 				sendType := TxTypeSend
	// 				if to == addr {
	// 					sendType = TxTypeReceive
	// 				}
	// 				tx := &TxResp{
	// 					Txid:         hash.ToHexString(),
	// 					From:         from,
	// 					To:           to,
	// 					Type:         uint(sendType),
	// 					Asset:        "save",
	// 					Amount:       states[3].(uint64),
	// 					AmountFormat: amountFormat,
	// 					FeeFormat:    utils.FormatUsdt(10000000),
	// 					BlockHeight:  uint32(i),
	// 				}
	// 				tx.Timestamp = blk.Header.Timestamp
	// 				txs = append(txs, tx)
	// 				if limit > 0 && uint32(len(txs)) >= limit {
	// 					return txs, nil
	// 				}
	// 				continue
	// 			}
	// 		}
	// 	}
	// 	if i == 0 {
	// 		break
	// 	}
	// }
	return txs, nil
}

//asset transfer direct
func (this *Endpoint) AssetTransferDirect(to, asset, amountStr string) (string, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if derr != nil {
		return "", derr
	}
	if strings.ToLower(asset) == "save" {
		asset = "usdt"
	}
	var amount float64
	temp, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	amount = temp
	realAmount := uint64(amount * 1000000000)
	log.Debugf("amount :%v", realAmount)
	if asset == "usdt" {
		toAddr, err := common.AddressFromBase58(to)
		if err != nil {
			return "", &DspErr{Code: INVALID_PARAMS, Error: err}
		}
		txHash, err := this.Dsp.Chain.Native.Usdt.Transfer(chainCfg.DEFAULT_GAS_PRICE, chainCfg.DEFAULT_GAS_LIMIT, acc, toAddr, realAmount)
		if err != nil {
			return "", &DspErr{Code: CHAIN_TRANSFER_ERROR, Error: err}
		}
		tx := hex.EncodeToString(common.ToArrayReverse(txHash[:]))
		return tx, nil
	}
	return "", &DspErr{Code: CHAIN_UNKNOWN_ASSET, Error: ErrMaps[CHAIN_UNKNOWN_ASSET]}
}

func (this *Endpoint) InvokeNativeContract(version byte, contractAddr, method string, params []interface{}) (string, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if derr != nil {
		return "", derr
	}
	contractAddress, err := common.AddressFromBase58(contractAddr)
	if err != nil {
		return "", &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	newParams, derr := getNativeContractParams(method, params)
	if derr != nil {
		return "", derr
	}
	txHash, err := this.Dsp.Chain.InvokeNativeContract(chainCfg.DEFAULT_GAS_PRICE, chainCfg.DEFAULT_GAS_LIMIT, acc, version, contractAddress, method, newParams)
	if err != nil {
		return "", &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	return hex.EncodeToString(common.ToArrayReverse(txHash[:])), nil
}

func (this *Endpoint) PreInvokeNativeContract(version byte, contractAddr, method string, params []interface{}) (interface{}, *DspErr) {
	contractAddress, err := common.AddressFromBase58(contractAddr)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	newParams, derr := getNativeContractParams(method, params)
	if derr != nil {
		return nil, derr
	}
	ret, err := this.Dsp.Chain.PreExecInvokeNativeContract(contractAddress, version, method, newParams)
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	result, err := convertResult(method, data)
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	return result, nil
}

func getNativeContractParams(method string, params []interface{}) ([]interface{}, *DspErr) {
	switch method {
	case film.FILM_PUBLISH:
		if len(params) != 10 {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
		}
		newParams, err := convertParam(params, []string{"string", "string", "string", "string", "float64", "float64", "string", "string", "float64", "float64"})
		if err != nil {
			return nil, err
		}
		return []interface{}{&film.PublishFilmParam{
			Cover:       []byte(newParams[0].(string)),
			Url:         []byte(newParams[1].(string)),
			Name:        []byte(newParams[2].(string)),
			Desc:        []byte(newParams[3].(string)),
			Type:        uint64(newParams[4].(float64)),
			ReleaseYear: uint64(newParams[5].(float64)),
			Language:    []byte(newParams[6].(string)),
			Region:      []byte(newParams[7].(string)),
			Price:       uint64(newParams[8].(float64)),
			CreatedAt:   uint64(newParams[9].(float64)),
		}}, nil
	case film.FILM_GETINFO:
		if len(params) != 2 {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
		}
		newParams, err := convertParam(params, []string{"string", "address"})
		if err != nil {
			return nil, err
		}
		param := &film.FilmHashAddrParams{
			Hash:       []byte(newParams[0].(string)),
			WalletAddr: newParams[1].(common.Address),
		}
		return []interface{}{param}, nil
	case film.FILM_SEARCH:
		if len(params) != 8 {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
		}
		newParams, err := convertParam(params, []string{"string", "bool", "float64", "float64", "string", "address", "float64", "float64"})
		if err != nil {
			return nil, err
		}
		avaiable := 0
		if newParams[1].(bool) {
			avaiable = 1
		}
		param := &film.SearchFilmParams{
			Name:        []byte(newParams[0].(string)),
			Avaiable:    uint64(avaiable),
			Type:        uint64(newParams[2].(float64)),
			ReleaseYear: uint64(newParams[3].(float64)),
			Region:      []byte(newParams[4].(string)),
			WalletAddr:  common.Address(newParams[5].(common.Address)),
			Offset:      uint64(newParams[6].(float64)),
			Limit:       uint64(newParams[7].(float64)),
		}
		return []interface{}{param}, nil
	default:
		return params, nil
	}
}

func convertParam(param []interface{}, valType []string) ([]interface{}, *DspErr) {
	if len(param) != len(valType) {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	result := make([]interface{}, 0, len(param))
	for i, p := range param {
		switch valType[i] {
		case "string":
			str, ok := p.(string)
			if !ok {
				return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
			}
			result = append(result, str)
		case "float64":
			val, ok := p.(float64)
			if !ok {
				return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
			}
			result = append(result, val)
		case "bool":
			val, ok := p.(bool)
			if !ok {
				return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
			}
			fmt.Printf("Val %t\n", val)
			result = append(result, val)
		case "address":
			str, ok := p.(string)
			if !ok {
				return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
			}
			addr, err := common.AddressFromBase58(str)
			if err != nil {
				return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
			}
			result = append(result, addr)
		}
	}
	return result, nil
}

func convertResult(method string, data []byte) (interface{}, error) {
	var e reflect.Value
	switch method {
	case film.FILM_GETINFO:
		var filmInfo film.FilmInfo
		bf := bytes.NewReader(data)
		err := filmInfo.Deserialize(bf)
		if err != nil {
			return nil, err
		}
		e = reflect.ValueOf(&filmInfo).Elem()
	case film.FILM_SEARCH:
		var list film.KeyList
		bf := bytes.NewReader(data)
		err := list.Deserialize(bf)
		if err != nil {
			return nil, err
		}
		e = reflect.ValueOf(&list).Elem()
	default:
		return data, nil
	}

	return edgeUtils.ConvertStructToMap(e), nil
}
