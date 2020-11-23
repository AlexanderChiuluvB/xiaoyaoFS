package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/cacheUtils"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/cacheUtils/gomemcache/memcache"
	"strconv"
	"time"
)

type NeedleCache struct {
	mc     *cacheUtils.Pool
	expire int32
}

// New new cache instance.
func New(c *cacheUtils.Config, expire time.Duration) (cache *NeedleCache) {
	cache = &NeedleCache{}
	cache.expire = int32(time.Duration(expire) / time.Second)
	cache.mc = cacheUtils.NewPool(c)
	return
}

// Ping check cacheUtils health
func (c *NeedleCache) Ping() (err error) {
	conn := c.mc.Get()
	err = conn.Store("set", "ping", []byte{1}, 0, c.expire, 0)
	conn.Close()
	return
}

func NeedleKey(id uint64) string {
	return fmt.Sprintf("n/%s", strconv.FormatUint(id, 10))
}

func (c *NeedleCache) GetNeedle(id uint64) (n *volume.Needle, err error) {
	var (
		bs  []byte
		key = NeedleKey(id)
	)
	bs, err = c.get(key)
	if err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		return
	}
	n = new(volume.Needle)
	if err = json.Unmarshal(bs, n); err != nil {
		return nil, errors.New(fmt.Sprintf("cache Meta.Unmarshal(%d) error(%v)", id, err))
	}
	return
}

func (c *NeedleCache) SetNeedle(id uint64, n *volume.Needle) (err error) {
	key := NeedleKey(id)
	bs, err := json.Marshal(n)
	if err != nil {
		return errors.New(fmt.Sprintf("cache setMeta() Marshal(%s) error(%v)", bs, err))
	}
	if err = c.set(key, bs, c.expire); err != nil {
		return errors.New(fmt.Sprintf("cache setMeta() set(%s,%s) error(%v)", key, string(bs), err))
	}
	return
}

// DelMeta del meta from cache.
func (c *NeedleCache) DelNeedle(id uint64) (err error) {
	key := NeedleKey(id)
	if err = c.del(key); err != nil {
		return errors.New(fmt.Sprintf("cache DelMeta(%s) error(%v)", key, err))
	}
	return
}

func (c *NeedleCache) set(key string, bs []byte, expire int32) (err error) {
	conn := c.mc.Get()
	defer conn.Close()
	return conn.Store("set", key, bs, 0, expire, 0)
}

func (c *NeedleCache) get(key string) (bs []byte, err error) {
	var (
		conn = c.mc.Get()
	)
	defer conn.Close()
	if bs, err = conn.Get2("get", key); err != nil {
		return
	}
	return
}

func (c *NeedleCache) del(key string) (err error) {
	conn := c.mc.Get()
	defer conn.Close()
	return conn.Delete(key)
}
