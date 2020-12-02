package config

import (
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

	// master meta type
	MetaType string

	//Redis
	RedisHost string
	RedisPort int
	Password  string
	Database  int

	// Hbase
	HbaseHost string

	// Cassandra
	CassandraHosts []string
	Keyspace       string

	// ClickHouse
	ClickHouseHost string

	ExpireTime       time.Duration
	PurgeTime        time.Duration
	MaxVolumeNum     int

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
