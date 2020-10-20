package volume

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	DEFAULT_DIR string = "/tmp/fs"
	MAX_VOLUME_SIZE uint64 = 1 << 36 //64GB
	VOLUME_INDEX_SIZE uint64 = 8 //每个Volume File用1byte来存储当前的offset的偏移量
)

type Volume struct {
	ID uint64
	MaxSize uint64
	CurrentOffset uint64 //当前最新文件所在的offset
	Path string
	File *os.File // Volume File 用一个Byte来存储当前的Offset
	Directory Directory
	lock sync.Mutex
}

func NewVolume(vid uint64, dir string) (v *Volume, err error) {
	if dir == "" {
		dir = DEFAULT_DIR
	}
	volumeFilePath := filepath.Join(dir, strconv.FormatUint(vid, 10) + ".data")
	v = new(Volume)
	v.ID = vid
	v.Path = dir
	v.File, err = os.OpenFile(volumeFilePath, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file :%v", err)
	}
	v.Directory, err = NewLeveldbDirectory(dir)
	if err != nil {
		return nil, fmt.Errorf("new leveldb directory :%v", err)
	}
	v.lock = sync.Mutex{}
	v.MaxSize = MAX_VOLUME_SIZE

	//每一次Volume重启,都需要读取其第一个byte以获得当前的offset
	var oldOffsetBytes = make([]byte, VOLUME_INDEX_SIZE)
	_, err = v.File.ReadAt(oldOffsetBytes, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	oldOffset := binary.BigEndian.Uint64(oldOffsetBytes)
	if oldOffset > VOLUME_INDEX_SIZE {
		v.SetCurrentIndex(oldOffset)
	} else {
		v.SetCurrentIndex(VOLUME_INDEX_SIZE)
	}

	return v, nil
}

func (v *Volume) GetNeedle(id uint64) (needle *Needle, err error) {
	needle, err = v.Directory.Get(id)
	if err != nil {
		return nil, err
	}
	needle.File = v.File
	return
}




func (v *Volume) SetCurrentIndex(currentIndex uint64) (err error) {
	v.CurrentOffset = currentIndex
	var offsetBytes = make([]byte, VOLUME_INDEX_SIZE)
	binary.BigEndian.PutUint64(offsetBytes, currentIndex)
	_, err = v.File.WriteAt(offsetBytes, 0)
	return
}
