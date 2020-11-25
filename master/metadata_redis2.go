package master

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"gopkg.in/redis.v2"
	"strconv"
	"strings"
)

type MetadataRedis2 struct {
	client *redis.Client
}

func (m *MetadataRedis2) GetEntries(filePath string) (Entries []*Entry, err error) {
	panic("implement me")
}

func NewRedis2Store(config *config.Config)(*MetadataRedis2, error) {
	mr := new(MetadataRedis2)
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

func (m *MetadataRedis2)Get(filePath string) (vid, nid uint64, err error) {
	data, err := m.client.Get(filePath).Result()
	if err != nil {
		return 0, 0, err
	}
	dataParts := strings.Split(data, "/")
	vid, _ = strconv.ParseUint(dataParts[0], 10, 64)
	nid, _ = strconv.ParseUint(dataParts[1], 10, 64)
	return
}

func (m *MetadataRedis2)Set(filePath string, vid, nid uint64) error {
	_, err := m.client.Set(filePath, fmt.Sprintf("%d/%d", vid, nid)).Result()
	return err
}

func (m *MetadataRedis2)Delete(filePath string) error {
	_, err := m.client.Del(filePath).Result()
	return err
}

func (m *MetadataRedis2)Close() error {
	return m.client.Close()
}


