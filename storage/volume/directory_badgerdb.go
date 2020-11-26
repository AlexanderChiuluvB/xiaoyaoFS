package volume

import (
	"encoding/binary"
	"github.com/dgraph-io/badger/v2"
	"path/filepath"
	"strconv"
)

type BadgerDBDirectory struct {
	path string
	db   *badger.DB
}

func NewBadgerDBDirectory(dir string, vid uint64) (d *BadgerDBDirectory, err error) {
	d = new(BadgerDBDirectory)
	d.path = filepath.Join(dir, strconv.FormatUint(vid, 10) + ".index")
	d.db, err = badger.Open(badger.DefaultOptions(d.path).WithSyncWrites(false))
	return
}

func (d *BadgerDBDirectory) Get(id uint64) (n *Needle, err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	var data[]byte
	if err = d.db.View(func (txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func (val []byte) error {
			data = append(data, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return UnMarshalBinary(data)
}

func (d *BadgerDBDirectory) New(n *Needle) (err error) {
	data, err := MarshalBinary(n)
	if err != nil {
		return err
	}
	if err = d.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(data[:8], data)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (d *BadgerDBDirectory) Has(id uint64) (has bool) {
	_, err := d.Get(id)
	return err == nil
}

func (d *BadgerDBDirectory) Set(id uint64, needle *Needle) (err error) {
	oldNeedle, err := d.Get(id)
	if err != nil {
		return
	}
	err = d.Del(id)
	if err != nil {
		return d.New(oldNeedle)
	}
	return d.New(needle)
}

func (d *BadgerDBDirectory) Del(id uint64) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
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
