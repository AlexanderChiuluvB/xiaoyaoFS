package master

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

type MetaCache struct {
	c *gocache.Cache
}

func newMetaCache(config *config.Config) *MetaCache {
	metaCache := new(MetaCache)
	metaCache.c = gocache.New(time.Duration(config.ExpireTime), time.Duration(config.PurgeTime))
	return metaCache
}

