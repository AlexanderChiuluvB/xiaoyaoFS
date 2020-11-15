package mount

import (
	"bytes"
	"context"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"os"
	"strings"
	"time"
)

type Dir struct {
	Name string
	XiaoyaoFs *XiaoyaoFs
	parent *Dir
	Entry *master.Entry

}

var _ = fs.Node(&Dir{})
var _ = fs.NodeCreater(&Dir{})
var _ = fs.NodeMkdirer(&Dir{})
var _ = fs.NodeFsyncer(&Dir{})
var _ = fs.NodeRequestLookuper(&Dir{})
var _ = fs.HandleReadDirAller(&Dir{})
var _ = fs.NodeRemover(&Dir{})
var _ = fs.NodeRenamer(&Dir{})
var _ = fs.NodeSetattrer(&Dir{})
var _ = fs.NodeGetxattrer(&Dir{})
var _ = fs.NodeSetxattrer(&Dir{})
var _ = fs.NodeRemovexattrer(&Dir{})
var _ = fs.NodeListxattrer(&Dir{})
var _ = fs.NodeForgetter(&Dir{})

func (dir *Dir) Removexattr(ctx context.Context, req *fuse.RemovexattrRequest) error {
	return nil
}

func (dir *Dir) Setxattr(ctx context.Context, req *fuse.SetxattrRequest) error {
	return nil
}


func (d *Dir) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	return nil
}


func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDirectory fs.Node) error {
	return nil
}

func (d *Dir) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	// fsync works at OS level
	// write the file chunks to the filerGrpcAddress

	return nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {

	if req.Valid.Mode() {
		d.Entry.Mode = uint32(req.Mode)
	}

	if req.Valid.Uid() {
		d.Entry.Uid = req.Uid
	}

	if req.Valid.Gid() {
		d.Entry.Gid = req.Gid
	}

	if req.Valid.Mtime() {
		d.Entry.Mtime = req.Mtime
	}

	err := api.InsertEntry(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, d.Entry)
	if err != nil {
		return err
	}
	return nil

}

func (dir *Dir) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	return nil
}

func (dir *Dir) Forget() {
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

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	if !req.Dir {
		return d.removeFile(req)
	}
	return d.removeFolder(req)
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	var node fs.Node
	entry := new(master.Entry)
	entry.IsDirectory = req.Mode & os.ModeDir > 0
	entry.Uid = req.Uid
	entry.Gid = req.Gid
	entry.Mode = uint32(req.Mode)
	entry.Mtime = time.Now()
	entry.Ctime = time.Now()
	entry.FilePath = d.FullPath()+"/"+req.Name

	err := api.InsertEntry(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, entry)
	if err != nil {
		return nil, nil, err
	}

	if entry.IsDirectory {
		node = d.newDirectory(entry, req.Name)
		return node, nil, nil
	}

	node = d.newFile(req.Name, entry)
	file := node.(*File)
	fh, err := d.XiaoyaoFs.AcquireHandle(file, req.Uid, req.Gid)
	if err != nil {
		return nil, nil, err
	}
	return file, fh, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	var node fs.Node
	entry := new(master.Entry)
	entry.IsDirectory = true
	entry.Uid = req.Uid
	entry.Gid = req.Gid
	entry.Mode = uint32(req.Mode)
	entry.Mtime = time.Now()
	entry.Ctime = time.Now()
	entry.FilePath = d.FullPath() + "/" + req.Name

	err := api.InsertEntry(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, entry)
	if err != nil {
		return nil, err
	}
	node = d.newDirectory(entry, req.Name)
	return node, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) (dirents []fuse.Dirent, err error) {
	entries, err := api.GetEntries(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, d.Name)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		entryName := strings.TrimPrefix(entry.FilePath, d.Name+"/")
		inode := AsInode(entry.FilePath)
		if entry.IsDirectory {
			dirent := fuse.Dirent{
				Inode: inode,
				Name: entryName,
				Type: fuse.DT_Dir,
			}
			dirents = append(dirents, dirent)
		} else {
			dirent := fuse.Dirent{
				Inode: inode,
				Name: entryName,
				Type: fuse.DT_File}
			dirents = append(dirents, dirent)
		}
	}
	return
}


func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (node fs.Node, err error) {
	fullPath := d.FullPath() + "/" + req.Name
	entry, err := api.GetEntry(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, fullPath)
	if err != nil {
		return nil, err
	}
	if entry != nil && entry.FilePath != ""{
		if entry.IsDirectory {
			node = d.newDirectory(entry, req.Name)
		} else {
			node = d.newFile(req.Name, entry)
		}
		resp.Attr.Inode = AsInode(fullPath)
		resp.Attr.Valid = time.Second
		resp.Attr.Mtime = entry.Mtime
		resp.Attr.Crtime = entry.Ctime
		resp.Attr.Mode = os.FileMode(entry.Mode)
		resp.Attr.Gid = entry.Gid
		resp.Attr.Uid = entry.Uid
		return node, nil
	}
	return nil, fuse.ENOENT
}

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {

	if d.FullPath() == d.XiaoyaoFs.rootPath {
		d.setRootDirAttr(attr)
		return nil
	}
	/*entry, err := api.GetEntry(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, d.FullPath())
	if err != nil {
		return err
	}*/
	attr.Inode = AsInode(d.FullPath())
	attr.Mode = os.ModeDir
	attr.Mtime = time.Now()
	attr.Crtime = time.Now()
	attr.Gid = master.OS_GID
	attr.Uid = master.OS_UID
	return nil
}

func (d *Dir) newFile(name string, entry *master.Entry) fs.Node {
	return &File{
		Name:           name,
		Dir:            d,
		XiaoyaoFs:      d.XiaoyaoFs,
		Entry: entry,
	}
}

func (d *Dir) newDirectory(entry *master.Entry, dirName string) fs.Node {
	return &Dir{Name: dirName, XiaoyaoFs: d.XiaoyaoFs,
		Entry: entry, parent: d}
}

func (d *Dir) setRootDirAttr(attr *fuse.Attr) {
	attr.Inode = 1
	attr.Valid = time.Hour
	attr.BlockSize = 1024*1024
	attr.Mode = os.ModeDir
	attr.Ctime = time.Now()
	attr.Crtime = time.Now()
	attr.Mtime = time.Now()
	attr.Uid = master.OS_UID
	attr.Gid = master.OS_GID
}

func (d *Dir) removeFile(req *fuse.RemoveRequest) error {
	fullPath := d.FullPath() + "/" + req.Name
	err := api.Delete(d.XiaoyaoFs.MasterHost, d.XiaoyaoFs.MasterPort, fullPath)
	if err != nil {
		return fuse.ENOENT
	}
	return nil
}

func (d *Dir) removeFolder(req *fuse.RemoveRequest) error {
	// Add a recursive delete method
	return nil
}

