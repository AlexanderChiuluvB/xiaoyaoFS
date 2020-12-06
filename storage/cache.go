package storage

import (
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/dgraph-io/ristretto"
	"strconv"
)

type NeedleCache struct {
	c *ristretto.Cache
}

func newNeedleCache(config *config.Config) (*NeedleCache, error) {
	metaCache := new(NeedleCache)
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

func NeedleKey(vid, nid uint64) string {
	return fmt.Sprintf("n/%s/%s", strconv.FormatUint(vid, 10), strconv.FormatUint(nid, 10))
}

func FileKey(vid, nid uint64) string {
	return fmt.Sprintf("f/%s/%s", strconv.FormatUint(vid, 10), strconv.FormatUint(nid, 10))
}

func (n *NeedleCache) GetNeedle(vid, nid uint64) (needle *volume.Needle, err error) {
	if data, found := n.c.Get(NeedleKey(vid, nid)); found {
		return volume.UnMarshalBinary(data.([]byte))
	} else {
		return nil, nil
	}
}

func (n *NeedleCache) SetNeedle(vid, nid uint64, needle *volume.Needle) (err error) {
	key := NeedleKey(vid, nid)
	data, err := volume.MarshalBinary(needle)
	if err != nil {
		return errors.New(fmt.Sprintf("cache setNeedle Marshal error(%v)", err))
	}
	n.c.Set(key, data, 1)
	return
}

// DelMeta del meta from cache.
func (n *NeedleCache) DelNeedle(vid, nid uint64) (err error) {
	key := NeedleKey(vid, nid)
	n.c.Del(key)
	return nil
}
