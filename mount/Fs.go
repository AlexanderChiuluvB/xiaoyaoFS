package mount

import (
	"bazil.org/fuse/fs"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"sync"
)

var _ fs.FS = (*XiaoyaoFs)(nil)

type XiaoyaoFs struct {
	root fs.Node

	handlesLock sync.Mutex
	handles map[uint64]*FileHandle
}

func NewXiaoyaoFs(c *config.Config) *XiaoyaoFs{
	xiaoyaoFs := new(XiaoyaoFs)
	return &XiaoyaoFs{
		root: &Dir{Name: c.MountDir,
			XiaoyaoFs: xiaoyaoFs},
	}
}

func (x *XiaoyaoFs) Root() (fs.Node, error) {
	return x.root, nil
}


