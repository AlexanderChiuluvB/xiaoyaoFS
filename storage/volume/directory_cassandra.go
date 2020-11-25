package volume

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/gocql/gocql"
	"time"
)

type CassandraDirectory struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func NewCassandraDirectory(config *config.Config) (c *CassandraDirectory, err error) {
	c = new(CassandraDirectory)
	c.cluster = gocql.NewCluster(config.CassandraHosts...)
	c.cluster.Consistency = gocql.Any
	c.cluster.Keyspace = config.Keyspace
	c.cluster.Timeout = time.Duration(time.Second*2)
	c.session, err = c.cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	//if err := c.session.Query(fmt.Sprintf(" USE %s; CREATE TABLE IF NOT EXISTS needle (vid bigint, nid bigint, meta blob, PRIMARY KEY (vid, nid))", config.Keyspace)).Exec(); err != nil {
	//	panic(err)
	//}
	return
}

func (c *CassandraDirectory) Get(vid, nid uint64) (n *Needle, err error) {
	var data []byte
	if err = c.session.Query("SELECT meta FROM needle WHERE vid = ? AND nid = ?", vid, nid).Consistency(gocql.One).Scan(&data); err != nil {
		return nil, err
	}
	return UnMarshalBinary(data)
}

func (c *CassandraDirectory) Has(vid, nid uint64) (has bool) {
	_, err := c.Get(vid, nid)
	return err == nil
}

func (c *CassandraDirectory) Set(vid, nid uint64, needle *Needle) (err error) {
	needleBytes, err := MarshalBinary(needle)
	if err != nil {
		return err
	}
	if err := c.session.Query("INSERT INTO needle (vid, nid, meta) VALUES (?, ?, ?)", vid, nid, needleBytes).Exec(); err != nil {
		return err
	}
	return
}

func (c *CassandraDirectory) Del(vid, nid uint64) (err error) {
	if err := c.session.Query("DELETE FROM needle WHERE vid = ? AND nid = ?", vid, nid).Exec(); err != nil {
		return err
	}
	return
}

func (c *CassandraDirectory) Close() {
	c.session.Close()
}