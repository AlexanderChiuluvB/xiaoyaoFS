package master

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/dgraph-io/ristretto"
)

type MetaCache struct {
	c *ristretto.Cache
}

func newMetaCache(config *config.Config) (*MetaCache, error) {
	metaCache := new(MetaCache)
	var err error
	metaCache.c, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: config.NumCounters,     // number of keys to track frequency of (10M).
		MaxCost:     config.MaxCost, // maximum cost of cache (1GB).
		BufferItems: config.BufferItem,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	return metaCache, nil
}

