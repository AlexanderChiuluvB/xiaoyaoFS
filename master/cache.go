package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/cacheUtils"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/cacheUtils/gomemcache/memcache"
	"time"
)

type EntryCache struct {
	mc     *cacheUtils.Pool
	expire int32
}

// New new cache instance.
func New(c *cacheUtils.Config, expire time.Duration) (cache *EntryCache) {
	cache = &EntryCache{}
	cache.expire = int32(time.Duration(expire) / time.Second)
	cache.mc = cacheUtils.NewPool(c)
	return
}

// Ping check cacheUtils health
func (c *EntryCache) Ping() (err error) {
	conn := c.mc.Get()
	err = conn.Store("set", "ping", []byte{1}, 0, c.expire, 0)
	conn.Close()
	return
}

func MetaKey(filepath string) string {
	return fmt.Sprintf("m/%s", filepath)
}

func (c *EntryCache) GetMeta(filepath string) (entry *Entry, err error) {
	key := MetaKey(filepath)
	bytes, err := c.get(key)
	if err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return nil, err
		}
		return nil, err
	}
	entry = new(Entry)
	if err = json.Unmarshal(bytes, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (c *EntryCache) SetMeta(filepath string, entry *Entry) (err error) {
	key := MetaKey(filepath)
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err = c.set(key, entryBytes, c.expire); err != nil {
		return err
	}
	return
}

func (c *EntryCache) DelMeta(filepath string) (err error) {
	key := MetaKey(filepath)
	if err = c.del(key); err != nil {
		return err
	}
	return
}

func (c *EntryCache) set(key string, bs []byte, expire int32) (err error) {
	conn := c.mc.Get()
	defer conn.Close()
	return conn.Store("set", key, bs, 0, expire, 0)
}

func (c *EntryCache) get(key string) (bs []byte, err error) {
	var (
		conn = c.mc.Get()
	)
	defer conn.Close()
	if bs, err = conn.Get2("get", key); err != nil {
		return
	}
	return
}

func (c *EntryCache) del(key string) (err error) {
	conn := c.mc.Get()
	defer conn.Close()
	return conn.Delete(key)
}
