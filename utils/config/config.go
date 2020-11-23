package config

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/cacheUtils"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/time"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type Config struct {
	StoreDir string

	StoreApiHost string
	StoreApiPort int

	MasterHost string
	MasterPort int

	MountDir string
	MetaType string

	MaxVolumeNum int

	// Hbase
	HbaseHost string

	// Cassandra
	CassandraHosts []string
	Keyspace       string

	Mc *cacheUtils.Config
	ExpireMc time.Duration
}

// NewConfig new a config.
func NewConfig(conf string) (c *Config, err error) {
	var (
		file *os.File
		blob []byte
	)
	c = new(Config)
	if file, err = os.Open(conf); err != nil {
		return
	}
	if blob, err = ioutil.ReadAll(file); err != nil {
		return
	}
	err = toml.Unmarshal(blob, c)
	return
}
