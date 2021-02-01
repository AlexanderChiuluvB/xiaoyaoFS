package master

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/gocql/gocql"
	"time"
)

type MetadataCassandra struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func NewCassandraMetaStore(config *config.Config) (c *MetadataCassandra, err error) {
	c = new(MetadataCassandra)
	c.cluster = gocql.NewCluster(config.CassandraHosts...)
	c.cluster.Consistency = gocql.Any
	c.cluster.Keyspace = config.Keyspace
	c.cluster.Timeout = time.Duration(time.Second*2)
	c.session, err = c.cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	return c, nil
}

func (c *MetadataCassandra) GetEntries(filePath string) (Entries []*Entry, err error) {
	panic("implement me")
}

func (c *MetadataCassandra) Set(filePath string, vid, nid uint64) error {
	return c.session.Query("INSERT INTO metadata (vid, nid, filePath) VALUES (?, ?, ?)", vid, nid, filePath).Exec()

}

func (c *MetadataCassandra) Delete(filePath string) error {
	return c.session.Query("DELETE FROM metadata WHERE filePath = ?", filePath).Exec()
}

func (c *MetadataCassandra) Get(filePath string) (vid, nid uint64, err error) {
	var result struct {
		vid uint64
		nid uint64
	}
	if err := c.session.Query("SELECT vid,nid FROM metadata WHERE filePath = ?", filePath).Consistency(gocql.One).Scan(&result);err!=nil{
		return 0,0, err
	}
	return result.vid, result.nid, nil
}

func (c *MetadataCassandra) Close() error {
	c.session.Close()
	return nil
}
