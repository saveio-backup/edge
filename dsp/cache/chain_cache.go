package cache

import lru "github.com/hashicorp/golang-lru"

type ChainCache struct {
	c *lru.ARCCache
}

func (this *ChainCache) SetTxHashToHeight(txHash string, height uint32) {
	this.c.Add(TxHashHeightKey(txHash), height)
}

func (this *ChainCache) HeightFromTxHash(txHash string) uint32 {
	value, ok := this.c.Get(TxHashHeightKey(txHash))
	if !ok {
		return 0
	}
	height, ok := value.(uint32)
	if !ok {
		return 0
	}
	return height
}

func (this *ChainCache) SetTimestampToHeight(height, ts uint32) {
	this.c.Add(BlockHeightTimestampKey(height), ts)
}

func (this *ChainCache) TimestampFromHeight(height uint32) uint32 {
	value, ok := this.c.Get(BlockHeightTimestampKey(height))
	if !ok {
		return 0
	}
	ts, ok := value.(uint32)
	if !ok {
		return 0
	}
	return ts
}
