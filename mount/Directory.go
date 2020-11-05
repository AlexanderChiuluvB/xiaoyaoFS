package mount

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
)

type Dir struct {
	Name string
	XiaoyaoFs *XiaoyaoFs
	parent *Dir
	//todo entry
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	panic("implement me")
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	panic("implement me")
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	panic("implement me")
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	panic("implement me")
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	panic("implement me")
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	panic("implement me")
}

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	panic("implement me")
}

var _ fs.Node = (*Dir)(nil)
var _ fs.NodeRequestLookuper = (*Dir)(nil)
var _ fs.NodeCreater = (*Dir)(nil)
var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)
var _ fs.NodeRenamer = (*Dir)(nil)
