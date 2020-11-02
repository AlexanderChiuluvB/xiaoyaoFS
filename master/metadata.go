package master

import "io"

type metadata interface {
	Get(filePath string) (vid uint64, fid uint64, err error)
	Set(filePath string, vid uint64, fid uint64) error
	Delete(filePath string) error
	io.Closer
}
