package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/gocql/gocql"
	"path/filepath"
	"time"
)

type CassandraStore struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func (c CassandraStore) Get(filePath string) (entry *Entry, err error) {
	dir, name := getDirAndName(filePath)
	var data []byte
	if err := c.session.Query("SELECT meta FROM filemeta WHERE directory=? AND name=?",
		dir, name).Consistency(gocql.One).Scan(&data); err != nil {
		if err != gocql.ErrNotFound {
			return nil, errors.New(fmt.Sprintf("gocql SELECT query error with %v", err))
		}
	}
	if len(data) == 0 {
		return nil, gocql.ErrNotFound
	}
	entry = new(Entry)
	err = json.Unmarshal(data, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (c CassandraStore) GetEntries(filePathPrefix string) (Entries []*Entry, err error) {
	cqlStr := "SELECT NAME, meta FROM filemeta WHERE directory=? ORDER BY NAME ASC"

	var data []byte
	var name string
	iter := c.session.Query(cqlStr, filePathPrefix).Iter()
	for iter.Scan(&name, &data) {
		entry := new(Entry)
		err = json.Unmarshal(data, entry)
		if err != nil {
			return nil, err
		}
		Entries = append(Entries, entry)
	}
	if err := iter.Close(); err != nil {
		panic(err)
	}

	return Entries, err
}

func (c CassandraStore) Set(entry *Entry) error {
	dir, name := getDirAndName(entry.FilePath)
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := c.session.Query(
		"INSERT INTO filemeta (directory,name,meta) VALUES(?,?,?)",
		dir, name, value).Exec(); err != nil {
		return fmt.Errorf("insert %s: %s", entry.FilePath, err)
	}

	return nil
}

//point delete
func (c CassandraStore) Delete(filePath string) error {

	dir, name := getDirAndName(filePath)

	if err := c.session.Query(
		"DELETE FROM filemeta WHERE directory=? AND name=?",
		dir, name).Exec(); err != nil {
		return fmt.Errorf("delete %s : %v", filePath, err)
	}

	return nil
}

//TODO Add range delete, delete a whole directory
func (c CassandraStore) Close() error {
	c.session.Close()
	return nil
}

func NewCassandraStore(config *config.Config) (c *CassandraStore, err error) {
	c = new(CassandraStore)
	c.cluster = gocql.NewCluster(config.CassandraHosts...)
	c.cluster.Consistency = gocql.Any
	c.cluster.Keyspace = "xiaoyaofs"
	c.cluster.Timeout = time.Duration(time.Second*2)
	c.session, err = c.cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	return 
}

func getDirAndName(fullPath string) (dir, name string) {
	dir, name = filepath.Split(fullPath)
	if dir == "/" {
		return dir, name
	}
	if len(dir) < 1 {
		return "/", ""
	}
	return dir[:len(dir)-1], name
}