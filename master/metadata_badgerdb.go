package master

import (
	"encoding/binary"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/dgraph-io/badger/v2"
	"path/filepath"
)

type MetadataBadgerDB struct {
	path string
	db   *badger.DB
}

func (m *MetadataBadgerDB ) GetEntries(filePath string) (Entries []*Entry, err error) {
	panic("implement me")
}

func NewBadgerMetaStore(config *config.Config)(m *MetadataBadgerDB, err error) {
	m = new(MetadataBadgerDB)
	m.path = filepath.Join(config.StoreDir, "index")
	m.db, err = badger.Open(badger.DefaultOptions(m.path).WithSyncWrites(false))
	return
}

func (m *MetadataBadgerDB)Get(filePath string) (vid, nid uint64, err error) {
	key := []byte(filePath)
	var data[]byte
	if err = m.db.View(func (txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(data)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, 0, err
	}
	return binary.BigEndian.Uint64(data[:8]),  binary.BigEndian.Uint64(data[8:16]), nil
}

func (m *MetadataBadgerDB)Set(filePath string, vid, nid uint64) error {
	value := make([]byte, 16)
	binary.BigEndian.PutUint64(value[:8], vid)
	binary.BigEndian.PutUint64(value[8:16], nid)
	if err := m.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(filePath), value)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (m *MetadataBadgerDB)Delete(filePath string) error {
	if err := m.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(filePath))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (m *MetadataBadgerDB)Close() error {
	return m.db.Close()
}

