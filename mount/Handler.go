package mount

import (
	"context"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"sync"
)

type FileHandle struct {
	F *File
	RequestId fuse.RequestID
	NodeId    fuse.NodeID
	Handle uint64
	Uid uint32
	Gid uint32
	sync.RWMutex
}

var _ fs.Handle = (*FileHandle)(nil)
var _ fs.HandleReleaser = (*FileHandle)(nil)
var _ fs.HandleReader = (*FileHandle)(nil)
var _ fs.HandleWriter = (*FileHandle)(nil)

func NewFileHandle(file *File, uid, gid uint32) *FileHandle {
	fh := &FileHandle{
		F: file,
		Uid: uid,
		Gid: gid,
	}
	return fh
}

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fh.Lock()
	defer fh.Unlock()

	data := make([]byte, len(req.Data))
	copy(data, req.Data)
	resp.Size = len(data)

	return api.Upload(fh.F.XiaoyaoFs.MasterHost, fh.F.XiaoyaoFs.MasterPort, fh.F.Name)
}

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error{
	fh.Lock()
	defer fh.Unlock()

	fh.F.isOpen--

	if fh.F.isOpen < 0 {
		fh.F.isOpen = 0
		return nil
	}

	if fh.F.isOpen == 0 {
		fh.F.XiaoyaoFs.ReleaseHandle(fh.F.fullpath(), fuse.HandleID(fh.Handle))
	}
	return nil
}

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error

	fh.RLock()
	defer fh.RUnlock()
	resp.Data, err = api.Get(fh.F.XiaoyaoFs.MasterHost, fh.F.XiaoyaoFs.MasterPort, fh.F.Name)
	if err != nil {
		return err
	}

	return nil
}


