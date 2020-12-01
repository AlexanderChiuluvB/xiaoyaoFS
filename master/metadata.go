package master

import (
	"io"
	"time"
)

type metadata interface {
	Get(filePath string) (entry *Entry, err error)
	GetEntries(filePath string) (Entries []*Entry, err error)
	Set(entry *Entry) error
	Delete(filePath string) error
	io.Closer
}

type Entry struct {
	FilePath string
	FileSize uint64
	Vid      uint64
	Nid      uint64
	Uid      uint32
	Gid      uint32
	Mode     uint32
	Ctime    time.Time //8 bytes
	Mtime    time.Time //8 bytes
	IsDirectory bool
}