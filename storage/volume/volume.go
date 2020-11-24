package volume

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	DefaultDir      string = "/tmp/fs"
	MaxVolumeSize   uint64 = 1 << 36 //64GB
	VolumeIndexSize uint64 = 8       //每个Volume File用1byte来存储当前的offset的偏移量
)

type Volume struct {
	ID            uint64
	MaxSize       uint64
	CurrentOffset uint64 //当前最新文件所在的offset
	Path          string
	File          *os.File // Volume File 用一个Byte来存储当前的Offset
	Directory     Directory
	rwlock        sync.RWMutex
	Writeable     bool
}

func NewVolume(vid uint64, dir string) (v *Volume, err error) {
	if dir == "" {
		dir = DefaultDir
	}
	volumeFilePath := filepath.Join(dir, strconv.FormatUint(vid, 10) + ".volume")
	v = new(Volume)
	v.ID = vid
	v.Path = dir
	v.File, err = os.OpenFile(volumeFilePath, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file :%v", err)
	}
	v.Directory, err = NewLeveldbDirectory(dir, vid)
	if err != nil {
		return nil, fmt.Errorf("new leveldb directory :%v", err)
	}
	//defer v.Directory.Close()
	v.rwlock = sync.RWMutex{}
	v.MaxSize = MaxVolumeSize
	v.Writeable = true
	//每一次Volume重启,都需要读取其第一个byte以获得当前的offset
	var oldOffsetBytes = make([]byte, VolumeIndexSize)
	_, err = v.File.ReadAt(oldOffsetBytes, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	oldOffset := binary.BigEndian.Uint64(oldOffsetBytes)
	if oldOffset > VolumeIndexSize {
		err = v.SetCurrentIndex(oldOffset)
	} else {
		err = v.SetCurrentIndex(VolumeIndexSize)
	}

	return v, nil
}

func (v *Volume) GetNeedle(id uint64) (needle *Needle, err error) {
	//v.rwlock.RLock()
	//defer v.rwlock.RUnlock()
	needle, err = v.Directory.Get(id)
	if err != nil {
		return nil, err
	}
	needle.File = v.File
	return
}

//删除needle的时候只是简单删除了db的metadata,并没有删除volume上的needle的真实数据
func (v *Volume) DelNeedle(id uint64) (err error) {
	v.rwlock.Lock()
	defer v.rwlock.Unlock()
	_, err = v.Directory.Get(id)
	if err != nil {
		return err
	}
	return v.Directory.Del(id)
}

func (v *Volume) GetNeedleBytes(id uint64) (data []byte, filename string, err error) {
	needle, err := v.Directory.Get(id)
	if err != nil {
		return nil, "", err
	}
	filename = needle.FileName
	data, err = ioutil.ReadAll(needle)
	if err != nil {
		return nil, "", err
	}
	return
}

func (v *Volume) SetCurrentIndex(currentIndex uint64) (err error) {
	v.CurrentOffset = currentIndex
	var offsetBytes = make([]byte, VolumeIndexSize)
	binary.BigEndian.PutUint64(offsetBytes, currentIndex)
	_, err = v.File.WriteAt(offsetBytes, 0)
	return
}

func (v *Volume) RemainingSpace() uint64 {
	return v.MaxSize - v.CurrentOffset
}

func (v *Volume) allocSpace(fileBodySize uint64, filenameSize uint64) (offset uint64, err error) {
	remainSize := v.RemainingSpace()
	totalSize := fileBodySize + filenameSize + uint64(FixedNeedleSize)
	if totalSize > remainSize {
		return v.CurrentOffset, errors.New(fmt.Sprintf("volume remain size too small, remainSize %d, allocSize %d",
			remainSize, totalSize))
	}
	offset = v.CurrentOffset
	err = v.SetCurrentIndex( offset + totalSize)
	return
}

// 1. alloc space
// 2. set needle's header
// 3. create meta info
func (v *Volume) NewNeedle(id uint64, fileName string, fileSize uint64) (n *Needle, err error) {
	v.rwlock.Lock()
	defer v.rwlock.Unlock()

	fileNameSize := uint64(len(fileName))
	offset, err := v.allocSpace(fileSize, fileNameSize)
	if err != nil {
		return nil, err
	}
	n = new(Needle)
	n.Id = id
	n.NeedleOffset = offset // needle 在 volume 的初始偏移量
	n.FileSize = fileSize
	//now := time.Now()
	//n.Uid = OS_UID
	//n.Gid = OS_GID
	//n.Mtime = now
	//n.Ctime = now
	n.File = v.File
	n.FileName = fileName
	//n.Mode = uint32(os.ModePerm)
	// 到这里初始化了一个新的 Needle

	// 然后把Needle的数据序列化
	needleData, err := MarshalBinary(n)
	if err != nil {
		return nil, err
	}

	// 然后在volume对应的文件的偏移量中写入needle的Data
	_, err = v.File.WriteAt(needleData, int64(n.NeedleOffset))
	if err != nil {
		return nil, err
	}

	// 在Directory层增加needle的meta
	err = v.Directory.New(n)
	return n, err
}

func (v *Volume) NewFile(id uint64, data *[]byte, fileName string) (needle *Needle, err error){

	needle, err = v.NewNeedle(id, fileName, uint64(len(*data)))
	if err != nil {
		return nil, fmt.Errorf("new needle : %v", err)
	}
	_, err = needle.Write(*data)
	if err != nil {
		return nil, fmt.Errorf("needle write error %v", err)
	}
	return needle, nil
}

func (v *Volume) GetVolumeSize() uint64 {
	fi, err := v.File.Stat()
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}



