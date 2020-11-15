package mount

import (
	"context"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
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
	Entry *master.Entry
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

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) (err error) {
	if f.isOpen <= 0 {
		f.Entry, err = api.GetEntry(f.XiaoyaoFs.MasterHost, f.XiaoyaoFs.MasterPort, f.fullpath())
		if err != nil {
			return err
		}
	}
	attr.Inode = AsInode(f.fullpath())
	attr.Valid = time.Second
	attr.Gid = f.Entry.Gid
	attr.Uid = f.Entry.Uid
	attr.Size = f.Entry.FileSize
	attr.Crtime = f.Entry.Ctime
	attr.Mode = os.FileMode(f.Entry.Mode)
	attr.Blocks = attr.Size/blockSize + 1
	return nil
}

func (file *File) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	return nil
}

func (file *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	return nil
}

func (file *File) Setxattr(ctx context.Context, req *fuse.SetxattrRequest) error {

	return nil

}

func (file *File) Removexattr(ctx context.Context, req *fuse.RemovexattrRequest) error {

	return nil

}

func (file *File) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {


	return nil

}

func (file *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	// fsync works at OS level
	// write the file chunks to the filerGrpcAddress

	return nil
}

func (file *File) Forget() {

}

var _ = fs.Node(&File{})
var _ = fs.NodeOpener(&File{})
var _ = fs.NodeFsyncer(&File{})
var _ = fs.NodeSetattrer(&File{})
var _ = fs.NodeGetxattrer(&File{})
var _ = fs.NodeSetxattrer(&File{})
var _ = fs.NodeRemovexattrer(&File{})
var _ = fs.NodeListxattrer(&File{})
var _ = fs.NodeForgetter(&File{})