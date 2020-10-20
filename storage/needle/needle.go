package needle

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

var errSmallNeedle = errors.New("needle预分配的空间太小")

type NeedleInfo struct {
	Id uint64
	FileSize uint64
	InfoOffset uint64
	Ctime time.Time
	Mtime time.Time
	FileName string
}

type Needle struct {
	Info *NeedleInfo
	File *os.File
	FileOffset uint64
}

func (Ni *NeedleInfo) MarshalBinary() []byte {
	data := make([]byte, 40 + len(Ni.FileName))
	binary.BigEndian.PutUint64(data[0:8], Ni.Id)
	binary.BigEndian.PutUint64(data[8:16], Ni.FileSize)
	binary.BigEndian.PutUint64(data[16:24], Ni.InfoOffset)
	binary.BigEndian.PutUint64(data[24:32], uint64(Ni.Ctime.Unix()))
	binary.BigEndian.PutUint64(data[32:40], uint64(Ni.Mtime.Unix()))
	copy(data[40:], []byte(Ni.FileName))
	return data
}

func (Ni *NeedleInfo) UnMarshalBinary(data []byte) (err error){

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	Ni.Id = binary.BigEndian.Uint64(data[0:8])
	Ni.FileSize = binary.BigEndian.Uint64(data[16:24])
	Ni.InfoOffset = binary.BigEndian.Uint64(data[8:16])
	Ni.Ctime = time.Unix(int64(binary.BigEndian.Uint64(data[24:32])), 0)
	Ni.Mtime = time.Unix(int64(binary.BigEndian.Uint64(data[32:40])), 0)
	Ni.FileName = string(data[40:])
	return err
}

func (f *Needle)Read(b []byte) (n int, err error) {
	start := f.Info.InfoOffset + f.FileOffset
	end := f.Info.InfoOffset + f.Info.FileSize
	length := end - start
	if len(b) > int(length) {
		b = b[:length]
	}

	n, err = f.File.ReadAt(b, int64(start))
	f.FileOffset += uint64(n)
	if f.FileOffset >= f.Info.FileSize {
		err = io.EOF
	}
	return
}

func (f *Needle)Write(b []byte) (n int, err error) {
	start := f.Info.InfoOffset + f.FileOffset
	end := f.Info.InfoOffset + f.Info.FileSize
	length := end - start

	if len(b) > int(length) {
		return 0, errSmallNeedle
	} else {
		n, err := f.File.WriteAt(b, int64(start))
		if err != nil {
			return n, err
		}
		f.FileOffset += uint64(n)
		return
	}
}