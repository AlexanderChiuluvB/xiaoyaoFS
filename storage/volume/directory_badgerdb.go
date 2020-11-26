package volume

import (
	"encoding/binary"
	"github.com/dgraph-io/badger/v2"
	"path/filepath"
)

type BadgerDBDirectory struct {
	path string
	db   *badger.DB
}

func NewBadgerDBDirectory(dir string) (d *BadgerDBDirectory, err error) {
	d = new(BadgerDBDirectory)
	d.path = filepath.Join(dir, "index")
	d.db, err = badger.Open(badger.DefaultOptions(d.path).WithSyncWrites(false))
	return
}

func (d *BadgerDBDirectory) Get(vid,nid uint64) (n *Needle, err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	var data[]byte
	if err = d.db.View(func (txn *badger.Txn) error {
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
		return nil, err
	}
	return UnMarshalBinary(data)
}

func (d *BadgerDBDirectory) Has(vid,nid uint64) (has bool) {
	_, err := d.Get(vid, nid)
	return err == nil
}

func (d *BadgerDBDirectory) Set(vid, nid uint64, needle *Needle) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := MarshalBinary(needle)
	if err != nil {
		return err
	}
	if err = d.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, data)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (d *BadgerDBDirectory) Del(vid, nid uint64) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	if err = d.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (d *BadgerDBDirectory) Iter() (iter Iterator) {
	panic("implement me")
}

func (d *BadgerDBDirectory) Close() {
	d.db.Close()
}
