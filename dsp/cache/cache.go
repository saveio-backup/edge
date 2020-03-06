package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/saveio/edge/common"
)

type EdgeCache struct {
	ChainCache *ChainCache
}

func NewEdgeCache() *EdgeCache {
	lruCache, err := lru.NewARC(common.MAX_CACHE_SIZE)
	if err != nil {
		return nil
	}
	cache := &EdgeCache{
		ChainCache: &ChainCache{
			c: lruCache,
		},
	}
	return cache
}
