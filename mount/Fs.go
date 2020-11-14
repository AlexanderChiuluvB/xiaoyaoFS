package mount

import (
	"crypto/md5"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"io"
	"sync"
)

var _ fs.FS = (*XiaoyaoFs)(nil)

type XiaoyaoFs struct {
	root fs.Node
	rootPath string
	handlesLock sync.Mutex
	handles map[uint64]*FileHandle

	MasterHost string
	MasterPort int
}

func NewXiaoyaoFs(c *config.Config) *XiaoyaoFs {
	xiaoyaoFs :=  &XiaoyaoFs{
		handles : make(map[uint64]*FileHandle),
		rootPath: c.MountDir,
		MasterHost: c.MasterHost,
		MasterPort: c.MasterPort,
	}
	xiaoyaoFs.root = &Dir{Name: c.MountDir,
		XiaoyaoFs: xiaoyaoFs}
	return xiaoyaoFs
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
	file.Entry, err = api.GetEntry(x.MasterHost, x.MasterPort, file.fullpath())
	if err != nil {
		return nil, err
	}
	file.isOpen++
	x.handles[inodeId] = fileHandle
	fileHandle.Handle = inodeId
	return
}

func (x *XiaoyaoFs) ReleaseHandle(fullpath string, id fuse.HandleID) {
	x.handlesLock.Lock()
	defer x.handlesLock.Unlock()
	delete(x.handles, AsInode(fullpath))
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


