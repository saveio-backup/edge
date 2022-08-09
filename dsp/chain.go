package dsp

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkCom "github.com/saveio/themis-go-sdk/common"

	edgeCom "github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	hComm "github.com/saveio/edge/http/common"
	"github.com/saveio/themis-go-sdk/usdt"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/payload"
	bCom "github.com/saveio/themis/http/base/common"
	cUsdt "github.com/saveio/themis/smartcontract/service/native/usdt"
	sUtils "github.com/saveio/themis/smartcontract/service/native/utils"
)

const (
	TxTypeAll     = 0
	TxTypeSend    = 1
	TxTypeReceive = 2
)

var Version string

func (this *Endpoint) GetNodeVersion() (string, *DspErr) {
	chainVer := ""
	if dsp := this.getDsp(); dsp != nil {
		chainVer, _ = dsp.GetChainVersion()
	}
	max := 6
	if len(Version) < max {
		max = len(Version)
	}
	return fmt.Sprintf("%s-%s", chainVer, Version[:max]), nil
}

func (this *Endpoint) GetChainId() string {
	return config.Parameters.BaseConfig.ChainId
}

// get networkId
func (this *Endpoint) GetNetworkId() string {
	return fmt.Sprintf("%d", config.Parameters.BaseConfig.NetworkId)
}

//get block height
func (this *Endpoint) GetBlockHeight() (uint32, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	height, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return 0, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return height, nil
}

//get block hash by height
func (this *Endpoint) GetBlockHash(height uint32) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	hash, err := dsp.GetBlockHash(uint32(height))
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return hash, nil
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
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	block, err := dsp.GetBlockByHash(hash)
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
		w := common.ZeroCopySink{}
		block.Serialization(&w)
		return common.ToHexString(w.Bytes()), nil
	}
	return bCom.GetBlockInfo(block), nil
}

//get block height by transaction hash
func (this *Endpoint) GetBlockHeightByTxHash(str string) (uint32, *DspErr) {
	if len(str) == 0 {
		return 0, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	height, err := dsp.GetBlockHeightByTxHash(str)
	if err != nil {
		return height, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return height, nil
}

//get block transaction hashes by height
func (this *Endpoint) GetBlockTxsByHeight(height uint32) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	data, err := dsp.GetBlockTxHashesByHeight(uint32(height))
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	res := hComm.GetBlockTransactions(data)
	return res, nil
}

//get block by height
func (this *Endpoint) GetBlockByHeight(height uint32, raw string) (interface{}, *DspErr) {
	var getTxBytes = false
	if raw == "1" {
		getTxBytes = true
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	block, err := dsp.GetBlockByHeight(height)
	if err != nil || block == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_BLOCK, Error: ErrMaps[CHAIN_UNKNOWN_BLOCK]}
	}
	if getTxBytes {
		w := common.ZeroCopySink{}
		block.Serialization(&w)
		return common.ToHexString(w.Bytes()), nil
	} else {
		return bCom.GetBlockInfo(block), nil
	}
}

//get transaction by hash
func (this *Endpoint) GetTransactionByHash(hash, raw string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.GetTransaction(hash)
	if tx == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_TX, Error: ErrMaps[CHAIN_UNKNOWN_TX]}
	}
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if raw == "1" {
		w := common.ZeroCopySink{}
		tx.Serialization(&w)
		return common.ToHexString(w.Bytes()), nil
	}
	tran := bCom.TransArryByteToHexString(tx)
	//[TODO] need support height later ？
	var height uint32
	tran.Height = height
	return tran, nil
}

//get smartcontract event by height
func (this *Endpoint) GetSmartCodeEventTxsByHeight(height uint32) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	eInfos, err := dsp.GetSmartContractEventByBlock(uint32(height))
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if eInfos == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_SMARTCONTRACT_EVENT,
			Error: ErrMaps[CHAIN_UNKNOWN_SMARTCONTRACT_EVENT]}
	}
	return eInfos, nil
}

//get smartcontract event by transaction hash
func (this *Endpoint) GetSmartCodeEventByTxHash(hash string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	notify, err := dsp.GetSmartContractEvent(hash)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	if notify == nil {
		return nil, &DspErr{Code: CHAIN_UNKNOWN_SMARTCONTRACT_EVENT,
			Error: ErrMaps[CHAIN_UNKNOWN_SMARTCONTRACT_EVENT]}
	}
	return notify, nil
}

//get contract state
func (this *Endpoint) GetContractState(hash, raw string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	smartContract, err := dsp.GetSmartContract(hash)
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
	return bCom.TransPayloadToHex(contract), nil
}

//get storage from contract
func (this *Endpoint) GetStorage(hash, key string) (string, *DspErr) {
	item, err := common.HexToBytes(key)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	value, err := dsp.GetStorage(hash, item)
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
	if len(address) == 0 {
		address = this.getDspWalletAddress()
	}
	addr, err := common.AddressFromBase58(address)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	usdt, err := dsp.BalanceOf(addr)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	balances := make([]*BalanceResp, 0)
	balances = append(balances, &BalanceResp{
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
	return balances, nil
}

type BalanceHistoryResp struct {
	DateAt          int64
	TxsCount        uint32
	TxsSendCount    uint32
	TxsReceiveCount uint32
	Asset           string
	Balance         uint64
	BalanceFormat   string
}

//get balance history of address
func (this *Endpoint) GetBalanceHistory(address, limitStr string) ([]*BalanceHistoryResp, *DspErr) {
	addr, err := common.AddressFromBase58(address)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil || limit < 0 || limit > 31 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.BalanceOf(addr)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}

	var balanceHistoryDates []string
	balanceHistoryMap := make(map[string]*BalanceHistoryResp)
	var balanceHistoryArr []*BalanceHistoryResp

	index := 0
	t := time.Now()
	zeroDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	for index > int(-limit) {
		// dateStr := time.Now().AddDate(0, 0, index).Format("2006-01-02")
		dateT := zeroDate.AddDate(0, 0, index)
		dateStr := dateT.Format("2006-01-02")
		balanceHistoryDates = append(balanceHistoryDates, dateStr)
		balanceHistoryMap[dateStr] = &BalanceHistoryResp{
			DateAt:          dateT.Unix(),
			TxsCount:        0,
			TxsSendCount:    0,
			TxsReceiveCount: 0,
			Asset:           "save",
			Balance:         bal,
			BalanceFormat:   utils.FormatUsdt(bal),
		}
		index--
	}

	heightForRequest := ""
	limitForRequest := "" // can use for paging, ex. "30", "" means getall
	skipForRequest := ""
	flagForRequest := true
	filterFeeWithSameTxId := ""
	for flagForRequest {
		// TODO wangyu there will error
		txs, derr := this.GetTxByHeightAndLimit(address, "save", TxTypeAll, string(heightForRequest),
			limitForRequest, string(skipForRequest), false)
		if derr != nil {
			return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: derr.Error}
		}
		if len(txs) == 0 {
			flagForRequest = false
		}

		tEndStr := zeroDate.AddDate(0, 0, int(-limit)).Format("2006-01-02")
		for _, tx := range txs {
			tTxStr := time.Unix(int64(tx.Timestamp), 0).Format("2006-01-02")
			txToHexStr, err := common.AddressFromBase58(tx.To)
			if err != nil {
				return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
			}

			if tTxStr < tEndStr {
				flagForRequest = false
				break
			}

			txBlockHeightStr := strconv.FormatUint(uint64(tx.BlockHeight), 10)
			if heightForRequest != txBlockHeightStr {
				heightForRequest = txBlockHeightStr
				skipForRequest = "1"
			} else {
				skipForRequestI, err := strconv.ParseUint(skipForRequest, 10, 32)
				if err != nil {
					return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
				}
				skipForRequestI++
				skipForRequest = strconv.FormatUint(skipForRequestI, 10)
			}

			// filter tx with same txid , only process once
			if filterFeeWithSameTxId == tx.Txid {
				continue
			}

			for itemI, dateItemStr := range balanceHistoryDates {
				// Balance calculate time until dateItemStr
				if dateItemStr <= tTxStr {
					if tx.Type == TxTypeSend || tx.From == address {
						if tx.Amount == 10000000 &&
							strings.Contains(txToHexStr.ToHexString(), "0000000000000000000000000000000000000") {
							balanceHistoryMap[dateItemStr].Balance += tx.Amount
						} else {
							balanceHistoryMap[dateItemStr].Balance += tx.Amount + utils.ParseUsdt(tx.FeeFormat)
						}
						balanceHistoryMap[dateItemStr].BalanceFormat = utils.FormatUsdt(
							balanceHistoryMap[dateItemStr].Balance)
					} else if tx.Type == TxTypeReceive || tx.To == address {
						if balanceHistoryMap[dateItemStr].Balance > tx.Amount {
							balanceHistoryMap[dateItemStr].Balance -= tx.Amount
						} else {
							balanceHistoryMap[dateItemStr].Balance = 0
						}
						balanceHistoryMap[dateItemStr].BalanceFormat = utils.FormatUsdt(
							balanceHistoryMap[dateItemStr].Balance)
					} else {
					}
				}
				// Txs count calculate, without the last day
				if dateItemStr == tTxStr {
					if itemI > 0 {
						dateItemStrYsdt := balanceHistoryDates[itemI-1]
						if tx.Type == TxTypeSend {
							balanceHistoryMap[dateItemStrYsdt].TxsSendCount++
						} else if tx.Type == TxTypeReceive {
							balanceHistoryMap[dateItemStrYsdt].TxsReceiveCount++
						}
						balanceHistoryMap[dateItemStrYsdt].TxsCount++
					}
				}

				filterFeeWithSameTxId = tx.Txid
			}

			// process the last day, calcalate txs count
			if tTxStr == tEndStr {
				if tx.Type == TxTypeSend {
					balanceHistoryMap[balanceHistoryDates[limit-1]].TxsSendCount++
				} else if tx.Type == TxTypeReceive {
					balanceHistoryMap[balanceHistoryDates[limit-1]].TxsReceiveCount++
				}
				balanceHistoryMap[balanceHistoryDates[limit-1]].TxsCount++
			}
		}
		// use for debug paging, if there are difference with no paging
		// flagForRequest = false
	}

	for limit > 0 {
		dateT := zeroDate.AddDate(0, 0, int(-limit+1))
		dateStr := dateT.Format("2006-01-02")
		balanceHistoryArr = append(balanceHistoryArr, balanceHistoryMap[dateStr])
		limit--
	}

	return balanceHistoryArr, nil
}

//get merkle proof by transaction hash
func (this *Endpoint) GetMerkleProof(hash string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	proof, err := dsp.GetMerkleProof(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return proof, nil
}

//get avg gas price in block
//[TODO] need change themis hCom.GetGasPrice return gasprice and height as string
//[TODO] or just return gasprice
func (this *Endpoint) GetGasPrice() (uint64, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	price, err := dsp.GetGasPrice()
	if err != nil {
		return 0, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return price, nil
}

//get allowance
func (this *Endpoint) GetAllowance(asset, from, to string) (string, *DspErr) {
	fromAddr, err := bCom.GetAddress(from)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	toAddr, err := bCom.GetAddress(to)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	tx, err := bCom.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return "", &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return tx, nil
}

//get memory pool transaction count
func (this *Endpoint) GetMemPoolTxCount() (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	cnt, err := dsp.GetMemPoolTxCount()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return cnt, nil
}

//get memory poll transaction state
func (this *Endpoint) GetMemPoolTxState(hash string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	state, err := dsp.GetMemPoolTxState(hash)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	return state, nil
}

type TxInvokeType int

const (
	TxInvokeUsdtContract TxInvokeType = iota
	TxInvokeOtherContract
)

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
	ContractAddr string
	ContractType TxInvokeType
}

func (this *Endpoint) GetTxByHeightAndLimit(addr, asset string, txType uint64,
	heightStr, limitStr, skipTxCntStr string, ignoreOtherCont bool) ([]*TxResp, *DspErr) {
	log.Debugf("GetTxByHeightAndLimit %v %v %v %v %v %v",
		addr, asset, txType, heightStr, limitStr, skipTxCntStr)
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
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	currentHeight, err := dsp.GetCurrentBlockHeight()
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
	events, err := dsp.GetSmartContractEventByEventIdAndHeights(usdt.USDT_CONTRACT_ADDRESS.ToBase58(),
		addr, cUsdt.EVENT_USDT_STATE_CHANGE, 1, height)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	log.Debugf("events len: %d", len(events))
	hasSkip := uint64(0)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		tx := &TxResp{
			Txid: event.TxHash,
			// BlockHeight:  uint32(blockHeight),
			FeeFormat: "0",
			// Timestamp:    blk.Header.Timestamp,
			Amount:       0,
			AmountFormat: "0",
			Asset:        edgeCom.SAVE_ASSET,
			Type:         TxTypeSend,
			ContractType: TxInvokeUsdtContract,
		}
		sendToSelf := false
		for _, n := range event.Notify {
			addrFromHex, err := common.AddressFromHexString(n.ContractAddress)
			if err != nil {
				continue
			}
			if addrFromHex.ToBase58() != sUtils.UsdtContractAddress.ToBase58() {
				tx.ContractAddr = addrFromHex.ToBase58()
				tx.To = tx.ContractAddr
				tx.ContractType = TxInvokeOtherContract
				break
			}
		}
		if len(tx.ContractAddr) == 0 {
			tx.ContractAddr = sUtils.UsdtContractAddress.ToBase58()
		}
		statesSli := make([][]interface{}, 0, len(event.Notify))
		for _, n := range event.Notify {
			states, ok := n.States.([]interface{})
			if !ok {
				continue
			}
			if len(states) < 3 {
				continue
			}
			if _, ok := states[3].(uint64); !ok {
				continue
			}
			from := states[1].(string)
			to := states[2].(string)
			if from != addr && to != addr {
				continue
			}
			statesSli = append(statesSli, states)
		}
		var in, out uint64
		partner := ""
		for _, states := range statesSli {
			from := states[1].(string)
			to := states[2].(string)
			if from == addr && to != sUtils.GovernanceContractAddress.ToBase58() {
				out += states[3].(uint64)
				partner = to
			} else if to == addr {
				in += states[3].(uint64)
				partner = from
			}
			// set up amount
			if to == sUtils.GovernanceContractAddress.ToBase58() {
				tx.FeeFormat = utils.FormatUsdt(states[3].(uint64))
				if len(statesSli) == 1 {
					tx.From = from
					tx.To = to
					tx.ContractType = TxInvokeOtherContract
				}
			}
		}
		if out > in {
			tx.From = addr
			tx.To = partner
			tx.Amount = (out - in)
			tx.AmountFormat = utils.FormatUsdt(out - in)
		} else if in > out {
			tx.From = partner
			tx.To = addr
			tx.Amount = in - out
			tx.AmountFormat = utils.FormatUsdt(in - out)
		}
		if tx.From != addr && tx.To == addr {
			tx.Type = TxTypeReceive
		}
		sendToSelf = (tx.From == addr && tx.To == addr)
		if !sendToSelf && txType != TxTypeAll &&
			((txType == TxTypeSend && tx.Type != TxTypeSend) ||
				(txType == TxTypeReceive && tx.Type != TxTypeReceive)) {
			continue
		}
		if ignoreOtherCont && (tx.To == sUtils.GovernanceContractAddress.ToBase58() ||
			tx.ContractAddr != sUtils.UsdtContractAddress.ToBase58()) && tx.Amount == 0 {
			continue
		}
		toAppend := make([]*TxResp, 0)
		if !sendToSelf {
			toAppend = append(toAppend, tx)
		} else {
			if txType == TxTypeAll {
				tx.Type = TxTypeSend
				toAppend = append(toAppend, tx)
				tx2 := *tx
				tx2.Type = TxTypeReceive
				toAppend = append(toAppend, &tx2)
			} else {
				tx.Type = uint(txType)
				toAppend = append(toAppend, tx)
			}
		}
		if skipTxCnt > 0 && !sendToSelf && skipTxCnt > hasSkip {
			hasSkip++
			continue
		}
		if skipTxCnt > 0 && sendToSelf {
			if skipTxCnt >= hasSkip+2 {
				hasSkip += 2
				continue
			}
			if skipTxCnt == hasSkip+1 {
				hasSkip++
				toAppend = toAppend[1:]
			}
		}
		blockHeight := this.cache.ChainCache.HeightFromTxHash(event.TxHash)
		if blockHeight == 0 {
			blockHeight, err = dsp.GetBlockHeightByTxHash(event.TxHash)
			if err != nil {
				continue
			}
			this.cache.ChainCache.SetTxHashToHeight(event.TxHash, blockHeight)
		}
		if blockHeight > height {
			continue
		}
		timestamp := this.cache.ChainCache.TimestampFromHeight(blockHeight)
		if timestamp == 0 {
			blk, err := dsp.GetBlockByHeight(blockHeight)
			if err != nil {
				continue
			}
			timestamp = blk.Header.Timestamp
			this.cache.ChainCache.SetTimestampToHeight(blockHeight, timestamp)
		}
		for _, txItem := range toAppend {
			txItem.Timestamp = timestamp
			tx.BlockHeight = blockHeight
		}
		txs = append(txs, toAppend...)
		if limit > 0 && uint32(len(txs)) >= limit {
			if uint32(len(txs)) > limit {
				return txs[:limit], nil
			}
			return txs, nil
		}
	}
	return txs, nil
}

func (this *Endpoint) GetSmartContractEventByEventId(contractAddr, addr string, eventId uint32) (
	[]*sdkCom.SmartContactEvent, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	events, err := dsp.GetSmartContractEventByEventId(contractAddr, addr, eventId)
	if err != nil {
		return nil, &DspErr{CHAIN_INTERNAL_ERROR, err}
	}
	return events, nil
}

// GetAccountSmartContractEventByBlock. get smartcontract event for current account by block height
func (this *Endpoint) GetAccountSmartContractEventByBlock(height uint32) (*sdkCom.SmartContactEvent, error) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, errors.New("no dsp")
	}
	txs, err := dsp.GetBlockTxHashesByHeight(height)
	if err != nil {
		return nil, err
	}
	for _, tx := range txs.Transactions {
		event, err := dsp.GetSmartContractEvent(tx.ToHexString())
		if err != nil {
			continue
		}
		if event == nil {
			continue
		}
		log.Debugf("Events :%v, err %s", event.TxHash, err)
		for _, n := range event.Notify {
			contractAddr, err := common.AddressFromHexString(n.ContractAddress)
			if err != nil {
				continue
			}
			switch contractAddr.ToBase58() {
			case usdt.USDT_CONTRACT_ADDRESS.ToBase58():
				states, ok := n.States.([]interface{})
				if !ok || states[0].(string) != "transfer" || len(states) < 4 {
					continue
				}
				curWalletAddr := dsp.WalletAddress()
				if states[1] != curWalletAddr && states[2] != curWalletAddr {
					continue
				}
				return event, nil
			}
		}
	}
	return nil, nil
}

//asset transfer direct
func (this *Endpoint) AssetTransferDirect(to, asset, amountStr string) (string, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), this.getDspPassword())
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
	log.Debugf("transfer amount :%v", realAmount)
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if asset == "usdt" {
		toAddr, err := common.AddressFromBase58(to)
		if err != nil {
			return "", &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
		}
		balance, err := dsp.BalanceOf(acc.Address)
		if err != nil {
			return "", &DspErr{Code: CHAIN_TRANSFER_ERROR, Error: err}
		}
		if balance < realAmount+chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_LIMIT {
			return "", &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
		}
		tx, err := dsp.Transfer(chainCfg.DEFAULT_GAS_PRICE, chainCfg.DEFAULT_GAS_LIMIT, acc, toAddr, realAmount)
		if err != nil {
			return "", &DspErr{Code: CHAIN_TRANSFER_ERROR, Error: err}
		}
		return tx, nil
	}
	return "", &DspErr{Code: CHAIN_UNKNOWN_ASSET, Error: ErrMaps[CHAIN_UNKNOWN_ASSET]}
}

//asset transfer direct
func (this *Endpoint) BatchAssetTransferDirect(to, asset, amountStr string, times int) ([]string, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), this.getDspPassword())
	if derr != nil {
		return nil, derr
	}
	if strings.ToLower(asset) == "save" {
		asset = "usdt"
	}
	var amount float64
	temp, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, &DspErr{Code: CHAIN_INTERNAL_ERROR, Error: err}
	}
	amount = temp
	realAmount := uint64(amount * 1000000000)
	log.Debugf("transfer amount :%v", realAmount)
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	allTxs := make([]string, 0)
	if asset == "usdt" {
		toAddr, err := common.AddressFromBase58(to)
		if err != nil {
			return nil, &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
		}
		balance, err := dsp.BalanceOf(acc.Address)
		if err != nil {
			return nil, &DspErr{Code: CHAIN_TRANSFER_ERROR, Error: err}
		}
		if balance < realAmount*uint64(times)+chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_LIMIT {
			return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
		}
		wg := new(sync.WaitGroup)
		for i := 0; i < times; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tx, err := dsp.Transfer(chainCfg.DEFAULT_GAS_PRICE, chainCfg.DEFAULT_GAS_LIMIT, acc, toAddr, realAmount)
				if err != nil {
					log.Errorf("transfer err %s", err)
					return
				}
				allTxs = append(allTxs, tx)
			}()
		}
		wg.Wait()

		return allTxs, nil
	}
	return nil, &DspErr{Code: CHAIN_UNKNOWN_ASSET, Error: ErrMaps[CHAIN_UNKNOWN_ASSET]}
}

func (this *Endpoint) SwitchChain(chainId, configFileName string) *DspErr {
	log.Debug("chainId, configName: %s %s", chainId, configFileName)
	if config.Parameters.BaseConfig.ChainId == chainId {
		return nil
	}
	dsp := this.getDsp()
	if this != nil && dsp != nil {
		syncing, _ := this.IsChannelProcessBlocks()
		log.Debugf("SwitchChain syncing: %t", syncing)
		if syncing {
			return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
		}
	}
	cfgName := configFileName
	if len(cfgName) == 0 {
		cfgName = fmt.Sprintf("config-%s.json", chainId)
	}
	newCfg := config.GetConfigFromFile(cfgName)
	if newCfg == nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("config file not found: %s", cfgName)}
	}
	if newCfg.BaseConfig.ChainId != chainId {
		return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("chainId: %s not match id: %s from config file",
			chainId, newCfg.BaseConfig.ChainId)}
	}
	if this != nil && dsp != nil {
		if err := this.Stop(); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	}
	if err := config.SwitchConfig(cfgName); err != nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if err := config.Save(); err != nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if this == nil {
		return nil
	}
	this.initLog()
	if this.GetDspAccount() == nil {
		return nil
	}
	log.Debugf("restart dsp")
	if err := StartDspNode(this, true, true, true); err != nil {
		log.Errorf("Start dsp node err : %s", err)
		return &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	return nil
}

func (this *Endpoint) GetChainIdList() ([]string, *DspErr) {
	infos, err := ioutil.ReadDir(config.Parameters.BaseConfig.BaseDir)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	ids := make([]string, 0)
	idsM := make(map[string]struct{})
	for _, i := range infos {
		name := i.Name()
		if i.IsDir() || !strings.Contains(name, ".json") {
			continue
		}
		cfg := config.GetConfigFromFile(name)
		if cfg == nil {
			continue
		}
		if len(cfg.BaseConfig.ChainId) == 0 {
			continue
		}
		if _, ok := idsM[cfg.BaseConfig.ChainId]; ok {
			continue
		}
		ids = append(ids, cfg.BaseConfig.ChainId)
		idsM[cfg.BaseConfig.ChainId] = struct{}{}
	}
	return ids, nil
}

func (this *Endpoint) InvokeNativeContract(version byte, contractAddr, method string, params []interface{},
	gasPrice, gasLimit uint64) (string, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), this.getDspPassword())
	if derr != nil {
		return "", derr
	}
	contractAddress, err := common.AddressFromBase58(contractAddr)
	if err != nil {
		return "", &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	buf, err := json.Marshal(params)
	if err != nil {
		return "", &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	if gasPrice == 0 {
		gasPrice = chainCfg.DEFAULT_GAS_PRICE
	}
	if gasLimit == 0 {
		gasLimit = chainCfg.DEFAULT_GAS_LIMIT
	}
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.InvokeNativeContract(gasPrice, gasLimit, acc, version,
		contractAddress, method, []interface{}{buf})
	return tx, nil
}

func (this *Endpoint) PreExecNativeContract(version byte, contractAddr, method string, params []interface{}) (
	interface{}, *DspErr) {
	contractAddress, err := common.AddressFromBase58(contractAddr)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	buf, err := json.Marshal(params)
	if err != nil {
		return "", &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	ret, err := dsp.PreExecInvokeNativeContract(contractAddress, version, method, []interface{}{buf})
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	data, err := ret.Result.ToByteArray()
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	str := hex.EncodeToString(data)
	return str, nil
}

func (this *Endpoint) RegisterHeader(header, desc string, ttl uint64) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.RegisterHeader(header, desc, ttl)
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	return m, nil
}
