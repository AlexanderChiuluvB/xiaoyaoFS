package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"gopkg.in/redis.v2"
)

type MetadataRedis struct {
	client *redis.Client
}

func (m *MetadataRedis) GetEntries(value string) (Entries []*Entry, err error) {
	panic("implement me")
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
	return err
}

func (m *MetadataRedis)Delete(filePath string) error {
	_, err := m.client.Del(filePath).Result()
	return err
}

func (m *MetadataRedis)Close() error {
	return m.client.Close()
}



