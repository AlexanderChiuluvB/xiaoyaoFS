package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/stringUtils"
	"gopkg.in/redis.v2"
)

const (
	DIR_LIST_MARKER = "\x00"
)

type MetadataRedis struct {
	client *redis.Client
}

func NewRedisStore(config *config.Config)(*MetadataRedis, error) {
	mr := new(MetadataRedis)
	mr.client = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr: fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password: config.Password,
		DB: int64(config.Database),
	})

	if _, err := mr.client.Get("__key").Result(); err != nil && err != redis.Nil {
		return nil, err
	}
	return mr, nil
}

func (m *MetadataRedis) GetEntries(dir string) (Entries []*Entry, err error) {
	members, err := m.client.SMembers(genDirectoryListKey(dir)).Result()
	if err != nil {
		return nil, err
	}
	for _, fileName := range members {
		fullPath := stringUtils.FullPath(dir, fileName)
		entry := new(Entry)
		data, err := m.client.Get(fullPath).Result()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(data), entry)
		if err != nil {
			return nil, err
		}
		Entries = append(Entries, entry)
	}
	return Entries, nil
}

func (m *MetadataRedis)Get(filePath string) (entry *Entry, err error) {
	entry = new(Entry)
	data, err := m.client.Get(filePath).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(data), entry)
	if err != nil {
		return nil, err
	}

	return
}

func (m *MetadataRedis)Set(entry *Entry) error {
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = m.client.Set(entry.FilePath, string(entryBytes)).Result()
	if err != nil {
		return err
	}
	dir, name := stringUtils.DirAndName(entry.FilePath)
	if name != "" {
		_, err = m.client.SAdd(genDirectoryListKey(dir), name).Result()
		if err != nil {
			return err
		}
	}
	return err
}

func (m *MetadataRedis)Delete(filePath string) error {
	dir, name := stringUtils.DirAndName(filePath)
	if name != "" {
		_, err := m.client.SRem(dir, name).Result()
		if err != nil {
			return err
		}
	}
	_, err := m.client.Del(filePath).Result()
	return err
}

func (m *MetadataRedis)Close() error {
	return m.client.Close()
}

func genDirectoryListKey(dir string) (dirList string) {
	return dir + DIR_LIST_MARKER
}


