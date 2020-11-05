package mount

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"sync"
)

type FileHandle struct {
	F *File
	RequestId fuse.RequestID
	NodeId    fuse.NodeID
	Uid uint32
	Gid uint32
	sync.RWMutex
}

var _ fs.Handle = (*FileHandle)(nil)
var _ fs.HandleReleaser = (*FileHandle)(nil)
var _ fs.HandleReader = (*FileHandle)(nil)
var _ fs.HandleWriter = (*FileHandle)(nil)

func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fh.Lock()
	defer fh.Unlock()

	data := make([]byte, len(req.Data))
	copy(data, req.Data)
	resp.Size = len(data)

	return nil
}

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error{
	panic("implement me")
}

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fh.RLock()
	defer fh.RUnlock()

	buff := make([]byte, req.Size)
	n, err := fh.F.reader.ReadAt(buff, req.Offset)
	if err != nil {
		return err
	}
	resp.Data = buff[:n]
	return nil
}


