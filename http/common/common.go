package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saveio/edge/common/config"
	sdk "github.com/saveio/themis-go-sdk"
	sdkcom "github.com/saveio/themis-go-sdk/common"
	"github.com/saveio/themis/common"
	httpcom "github.com/saveio/themis/http/base/common"
)

func GetBalance(address common.Address) (*httpcom.BalanceOfRsp, error) {
	chain := sdk.NewChain()
	chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)
	usdtBal, err := chain.Native.Usdt.BalanceOf(address)
	if err != nil {
		return nil, err
	}
	balance := &httpcom.BalanceOfRsp{}
	balance.Usdt = strconv.FormatUint(usdtBal, 10)
	return balance, nil
}

func GetAllowance(asset string, from, to common.Address) (string, error) {
	chain := sdk.NewChain()
	chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)

	if strings.ToLower(asset) == "usdt" {
		allowance, err := chain.Native.Usdt.Allowance(from, to)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(allowance, 10), nil
	}

	return "", fmt.Errorf("unknown asset %s", asset)
}

func GetBlockTransactions(block *sdkcom.BlockTxHashes) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		t := block.Transactions[i]
		trans[i] = t.ToHexString()
	}

	b := sdkcom.BlockTxHashesStr{
		Hash:         block.Hash.ToHexString(),
		Height:       block.Height,
		Transactions: trans,
	}
	return b
}
