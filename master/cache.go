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
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	return metaCache, nil
}

