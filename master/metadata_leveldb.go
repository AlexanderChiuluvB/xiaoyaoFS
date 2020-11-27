package master

import (
	"encoding/binary"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/syndtr/goleveldb/leveldb"
	"path/filepath"
)

type MetadataLevelDB struct {
	db   *leveldb.DB
	path string // leveldb 文件存放路径
}

func (m *MetadataLevelDB) GetEntries(filePath string) (Entries []*Entry, err error) {
	panic("implement me")
}

func NewLevelDBMetaStore(config *config.Config)(m *MetadataLevelDB, err error) {
	m = new(MetadataLevelDB)
	m.path = filepath.Join(config.StoreDir, "index")
	m.db, err = leveldb.OpenFile(m.path, nil)
	if err != nil {
		return nil, err
	}
	return
}

func (m *MetadataLevelDB)Get(filePath string) (vid, nid uint64, err error) {
	key := []byte(filePath)
	var data[]byte
	data, err = m.db.Get(key, nil)
	if err != nil {
		return 0,0 , err
	}
	return binary.BigEndian.Uint64(data[:8]),  binary.BigEndian.Uint64(data[8:16]), nil
}

func (m *MetadataLevelDB)Set(filePath string, vid, nid uint64) error {
	value := make([]byte, 16)
	binary.BigEndian.PutUint64(value[:8], vid)
	binary.BigEndian.PutUint64(value[8:16], nid)
	return m.db.Put([]byte(filePath), value, nil)

}

func (m *MetadataLevelDB)Delete(filePath string) error {
	return m.db.Delete([]byte(filePath), nil)
}

func (m *MetadataLevelDB)Close() error {
	return m.db.Close()
}