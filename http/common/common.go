package http

import (
	sdkcom "github.com/saveio/themis-go-sdk/common"
)

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
