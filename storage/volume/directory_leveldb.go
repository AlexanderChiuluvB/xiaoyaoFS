package volume

import (
	"encoding/binary"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"path/filepath"
)

type LeveldbDirectory struct {
	db   *leveldb.DB
	path string // leveldb 文件存放路径
}

func NewLeveldbDirectory(config *config.Config) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	d.path = filepath.Join(config.StoreDir, "index")
	opts := &opt.Options{
		BlockCacheCapacity:            config.BlockCacheCapacity, // default value is 8MiB
		WriteBuffer:                   config.WriteBuffer, // default value is 4MiB
		CompactionTableSizeMultiplier: config.CompactionTableSizeMultiplier,
	}
	d.db, err = leveldb.OpenFile(d.path, opts)
	if err != nil {
		return nil, err
	}
	return
}

func (d *LeveldbDirectory) Get(vid, nid uint64) (n *Needle, err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := d.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return UnMarshalBinary(data)
}

func (d *LeveldbDirectory) Has(vid, nid uint64) (has bool) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	_, err := d.db.Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(vid, nid uint64, needle *Needle) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	data, err := MarshalBinary(needle)
	if err != nil {
		return err
	}
	return d.db.Put(key, data, nil)
}

func (d *LeveldbDirectory) Del(vid, nid uint64) (err error) {
	key := make([]byte, 16)
	binary.BigEndian.PutUint64(key[:8], vid)
	binary.BigEndian.PutUint64(key[8:16], nid)
	return d.db.Delete(key, nil)
}

func (d *LeveldbDirectory) Iter() (iter Iterator) {
	it :=  d.db.NewIterator(nil, nil)
	levelIt := &LeveldbIterator{
		iter: it,
	}
	return levelIt
}

func (d *LeveldbDirectory) Close() {
	d.db.Close()
}

type LeveldbIterator struct {
	iter iterator.Iterator
}


func (it *LeveldbIterator) Next() (key []byte, exists bool) {
	exists = it.iter.Next()
	key = it.iter.Key()
	return
}

func (it *LeveldbIterator) Release() {
	it.iter.Release()
}
