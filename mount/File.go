package mount

import (
	"context"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"os"
	"strings"
	"time"
)

const blockSize = 512


type File struct {
	Name string
	XiaoyaoFs *XiaoyaoFs
	Dir *Dir
	isOpen int
	Needle *volume.Needle
}

func (f *File) fullpath() string {
	dirFullPath := f.Dir.FullPath()
	if strings.HasSuffix(dirFullPath, "/") {
		return dirFullPath + f.Name
	}
	return dirFullPath + "/" + f.Name
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	handle, err := f.XiaoyaoFs.AcquireHandle(f, req.Uid, req.Gid)
	if err != nil {
		return nil, err
	}
	resp.Handle = fuse.HandleID(handle.Handle)
	return handle, nil
}

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = AsInode(f.fullpath())
	attr.Valid = time.Second
	attr.Mtime = f.Needle.Mtime
	attr.Ctime = f.Needle.Ctime
	attr.Size = f.Needle.FileSize
	attr.Gid = f.Needle.Gid
	attr.Uid = f.Needle.Uid
	attr.Mode = os.FileMode(f.Needle.Mode)
	attr.Blocks = attr.Size/blockSize + 1
	return nil
}

var _ fs.Node = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)