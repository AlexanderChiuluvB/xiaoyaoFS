package mount

import (
	"bytes"
	"context"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"os"
	"strings"
)

type Dir struct {
	Name string
	XiaoyaoFs *XiaoyaoFs
	parent *Dir
	Entry *master.Entry

}

func (dir *Dir) FullPath() string {
	var parts []string
	for p := dir; p != nil; p = p.parent {
		if strings.HasPrefix(p.Name, "/") {
			if len(p.Name) > 1 {
				parts = append(parts, p.Name[1:])
			}
		} else {
			parts = append(parts, p.Name)
		}
	}

	if len(parts) == 0 {
		return "/"
	}

	var buf bytes.Buffer
	for i := len(parts) - 1; i >= 0; i-- {
		buf.WriteString("/")
		buf.WriteString(parts[i])
	}
	return buf.String()
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	panic("implement me")
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	panic("implement me")
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	isDirectory := req.Mode & os.ModeDir > 0
	if isDirectory {
		// create directory

	} else {
		// create file

	}
	return nil,nil,nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	return nil, nil
	//TODO 参考seaweadFS增加一层fsNode的缓存
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

func (d *Dir) newFile(name string) fs.Node {
	return &File{
		Name:           name,
		Dir:            d,
		XiaoyaoFs:      d.XiaoyaoFs,
	}
}

var _ fs.Node = (*Dir)(nil)
var _ fs.NodeRequestLookuper = (*Dir)(nil)
var _ fs.NodeCreater = (*Dir)(nil)
var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)
var _ fs.NodeRenamer = (*Dir)(nil)
