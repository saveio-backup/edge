package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/saveio/edge/common/config"
	sdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	httpcom "github.com/saveio/themis/http/base/common"
	rpccommon "github.com/saveio/themis/http/base/common"
)

const (
	ASSET_USDT = "usdt"
)

var chain = sdk.NewChain()
var restClient = chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)

//Return balance of address in base58 code
func GetBalance(address string) (*httpcom.BalanceOfRsp, error) {
	addr, err := common.AddressFromBase58(address)
	if err != nil {
		return nil, err
	}

	ontBal, err := chain.Native.Usdt.BalanceOf(addr)
	if err != nil {
		return nil, err
	}
	balance := &httpcom.BalanceOfRsp{}
	balance.Usdt = strconv.FormatUint(ontBal, 10)
	return balance, nil
}

func GetAccountBalance(address, asset string) (uint64, error) {
	balances, err := GetBalance(address)
	if err != nil {
		return 0, err
	}
	var balance uint64
	switch strings.ToLower(asset) {
	case ASSET_USDT:
		balance, err = strconv.ParseUint(balances.Usdt, 10, 64)
	default:
		return 0, fmt.Errorf("unsupport asset:%s", asset)
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func GetAllowance(asset, from, to string) (string, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return "", fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return "", fmt.Errorf("to address:%s invalid:%s", to, err)
	}

	if strings.ToLower(asset) == ASSET_USDT {
		allowance, err := chain.Native.Usdt.Allowance(fromAddr, toAddr)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(allowance, 10), nil
	}
	return "", fmt.Errorf("unknown asset %s", asset)
}

//Transfer ont|ong from account to another account
func Transfer(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return "", fmt.Errorf("to address:%s invalid:%s", to, err)
	}

	if strings.ToLower(asset) == ASSET_USDT {
		tx, err := chain.Native.Usdt.Transfer(gasPrice, gasLimit, signer, toAddr, amount)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
	}
	return "", fmt.Errorf("unknown asset %s", asset)
}

func TransferFrom(gasPrice, gasLimit uint64, signer *account.Account, asset, sender, from, to string, amount uint64) (string, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return "", fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return "", fmt.Errorf("to address:%s invalid:%s", to, err)
	}
	if signer.Address.ToBase58() != sender {
		return "", fmt.Errorf("sender: %s is not signer: %s", sender, signer.Address.ToBase58())
	}

	if strings.ToLower(asset) == ASSET_USDT {
		tx, err := chain.Native.Usdt.TransferFrom(gasPrice, gasLimit, signer, fromAddr, toAddr, amount)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
	}
	return "", fmt.Errorf("unknown asset %s", asset)
}

func Approve(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return "", fmt.Errorf("to address:%s invalid:%s", to, err)
	}

	if strings.ToLower(asset) == ASSET_USDT {
		tx, err := chain.Native.Usdt.Approve(gasPrice, gasLimit, signer, toAddr, amount)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
	}
	return "", fmt.Errorf("unknown asset %s", asset)
}

//GetSmartContractEvent return smart contract event execute by invoke transaction by hex string code
func GetSmartContractEvent(txHash string) (*rpccommon.ExecuteNotify, error) {

	event, err := chain.GetSmartContractEvent(txHash)
	if err != nil {
		return nil, err
	}
	eventInfos := make([]rpccommon.NotifyEventInfo, 0)
	for _, enotify := range event.Notify {
		eventInfos = append(eventInfos, rpccommon.NotifyEventInfo{
			ContractAddress: enotify.ContractAddress,
			States:          enotify.States,
		})
	}
	notifies := &rpccommon.ExecuteNotify{
		TxHash:      event.TxHash,
		State:       event.State,
		GasConsumed: event.GasConsumed,
		Notify:      eventInfos,
	}
	return notifies, nil
}

func GetSmartContractEventInfo(txHash string) ([]byte, error) {

	event, err := chain.GetSmartContractEvent(txHash)
	if err != nil {
		return nil, err
	}
	return json.Marshal(event)
}

func GetRawTransaction(txHash string) ([]byte, error) {

	tx, err := chain.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}
	txInfo := httpcom.TransArryByteToHexString(tx)
	txInfo.Height, _ = chain.GetBlockHeightByTxHash(txHash)
	return json.Marshal(txInfo)
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {

	blockHash, ok := hashOrHeight.(string)
	if ok {
		block, err := chain.GetBlockByHash(blockHash)
		if err != nil {
			return nil, err
		}
		return json.Marshal(httpcom.GetBlockInfo(block))
	}
	blockHeight, ok := hashOrHeight.(int64)
	if ok {
		block, err := chain.GetBlockByHeight(uint32(blockHeight))
		if err != nil {
			return nil, err
		}
		return json.Marshal(httpcom.GetBlockInfo(block))
	}
	return nil, fmt.Errorf("invalid params %v", hashOrHeight)
}

func GetNetworkId() (uint32, error) {

	return chain.GetNetworkId()
}

func GetBlockData(hashOrHeight interface{}) ([]byte, error) {

	blockHash, ok := hashOrHeight.(string)
	if ok {
		block, err := chain.GetBlockByHash(blockHash)
		if err != nil {
			return nil, err
		}
		return json.Marshal(httpcom.GetBlockInfo(block))
	}
	blockHeight, ok := hashOrHeight.(uint32)
	if ok {
		block, err := chain.GetBlockByHeight(blockHeight)
		if err != nil {
			return nil, err
		}
		return json.Marshal(httpcom.GetBlockInfo(block))
	}
	return nil, fmt.Errorf("invalid params %v", hashOrHeight)
}

func GetBlockCount() (uint32, error) {

	return chain.GetCurrentBlockHeight()
}

func GetTxHeight(txHash string) (uint32, error) {

	return chain.GetBlockHeightByTxHash(txHash)
}
