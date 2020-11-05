package mount

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"io"
)

type File struct {
	Name string
	XiaoyaoFs *XiaoyaoFs
	Dir *Dir
	reader io.ReaderAt
}

func (f File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	panic("implement me")
}

func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
	panic("implement me")
}

var _ fs.Node = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)