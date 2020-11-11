package mount

import (
	"crypto/md5"
	"encoding/json"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/seaweedfs/fuse/fs"
	"io"
	"sync"
)

var _ fs.FS = (*XiaoyaoFs)(nil)

type XiaoyaoFs struct {
	root fs.Node

	handlesLock sync.Mutex
	handles map[uint64]*FileHandle

	MasterHost string
	MasterPort int
}

func NewXiaoyaoFs(c *config.Config) *XiaoyaoFs{
	xiaoyaoFs := new(XiaoyaoFs)
	return &XiaoyaoFs{
			root: &Dir{Name: c.MountDir,
			XiaoyaoFs: xiaoyaoFs},
			MasterHost: c.MasterHost,
			MasterPort: c.MasterPort,
	}
}

func (x *XiaoyaoFs) AcquireHandle(file *File, uid, gid uint32) (fileHandle *FileHandle, err error) {
	x.handlesLock.Lock()
	defer x.handlesLock.Unlock()

	inodeId := AsInode(file.fullpath())

	existingHandle, found := x.handles[inodeId]
	if found && existingHandle != nil {
		file.isOpen++
		return existingHandle, nil
	}

	fileHandle = NewFileHandle(file, uid, gid)
	needleBytes, err := api.GetNeedle(x.MasterHost, x.MasterPort, file.Name)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(needleBytes, file.Needle)
	if err != nil {
		return nil, err
	}
	file.isOpen++
	x.handles[inodeId] = fileHandle
	fileHandle.Handle = inodeId
	return
}

func AsInode(path string) uint64 {
	return uint64(HashStringToLong(string(path)))
}

func HashStringToLong(dir string) (v int64) {
	h := md5.New()
	io.WriteString(h, dir)

	b := h.Sum(nil)

	v += int64(b[0])
	v <<= 8
	v += int64(b[1])
	v <<= 8
	v += int64(b[2])
	v <<= 8
	v += int64(b[3])
	v <<= 8
	v += int64(b[4])
	v <<= 8
	v += int64(b[5])
	v <<= 8
	v += int64(b[6])
	v <<= 8
	v += int64(b[7])

	return
}

func (x *XiaoyaoFs) Root() (fs.Node, error) {
	return x.root, nil
}


