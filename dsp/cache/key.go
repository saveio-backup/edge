package cache

import (
	"fmt"
)

func TxHashHeightKey(txHash string) string {
	return fmt.Sprintf("CACHE_TX_HEIGHT: %s", txHash)
}

func BlockHeightTimestampKey(blockHeight uint32) string {
	return fmt.Sprintf("BLOCK_HEIGHT_TS: %d", blockHeight)
}
