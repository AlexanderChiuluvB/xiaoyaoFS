package volume

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"path/filepath"
	"strconv"
)

type LeveldbDirectory struct {
	db   *leveldb.DB
	path string // leveldb 文件存放路径
}

func NewLeveldbDirectory(dir string, vid uint64) (d *LeveldbDirectory, err error) {
	d = new(LeveldbDirectory)
	d.path = filepath.Join(dir, strconv.FormatUint(vid, 10) + ".index")
	opts := &opt.Options{
		//Filter: filter.NewBloomFilter(64),
		BlockCacheCapacity:            32*1024*1024, // default value is 8MiB
		WriteBuffer:                   8*1024*1024, // default value is 4MiB
		CompactionTableSizeMultiplier: 4,
	}
	//d.db, err = leveldb.Open(storage.NewMemStorage(), opts)
	d.db, err = leveldb.OpenFile(d.path, opts)
	if err != nil {
		return nil, err
	}
	return
}

func (d *LeveldbDirectory) Get(id uint64) (n *Needle, err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	data, err := d.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return UnMarshalBinary(data)
}

func (d *LeveldbDirectory) New(n *Needle) (err error) {
	data, err := MarshalBinary(n)
	if err != nil {
		return err
	}
	return d.db.Put(data[:8], data, nil)
}

func (d *LeveldbDirectory) Has(id uint64) (has bool) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	_, err := d.db.Get(key, nil)
	return err == nil
}

func (d *LeveldbDirectory) Set(id uint64, needle *Needle) (err error) {
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

func (d *LeveldbDirectory) Del(id uint64) (err error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
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
