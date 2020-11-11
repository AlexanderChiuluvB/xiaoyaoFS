package volume

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

var (
	errSmallNeedle = errors.New("needle预分配的空间太小")
)
const FixedNeedleSize uint64 = 52

type Needle struct {
	Id uint64
	FileSize uint64
	NeedleOffset uint64
	Uid uint32
	Gid uint32
	Mode uint32
	FileName string
	Ctime time.Time //8 bytes
	Mtime time.Time //8 bytes
	File *os.File
	CurrentOffset uint64 //当前读写操作的offset
}

func MarshalBinary(N *Needle) ([]byte, error) {
	if N == nil {
		return nil, errors.New("Nil needle")
	}
	data := make([]byte, FixedNeedleSize+ uint64(len(N.FileName)))
	binary.BigEndian.PutUint64(data[0:8], N.Id)
	binary.BigEndian.PutUint64(data[8:16], N.FileSize)
	binary.BigEndian.PutUint64(data[16:24], N.NeedleOffset)
	binary.BigEndian.PutUint32(data[24:28], N.Uid)
	binary.BigEndian.PutUint32(data[28:32], N.Gid)
	binary.BigEndian.PutUint32(data[32:36], N.Mode)
	binary.BigEndian.PutUint64(data[36:44], uint64(N.Ctime.Unix()))
	binary.BigEndian.PutUint64(data[44:52], uint64(N.Mtime.Unix()))
	copy(data[FixedNeedleSize:], []byte(N.FileName))
	return data, nil
}

func UnMarshalBinary(data []byte) (N *Needle, err error){

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	N = new(Needle)
	N.Id = binary.BigEndian.Uint64(data[0:8])
	N.FileSize = binary.BigEndian.Uint64(data[8:16])
	N.NeedleOffset = binary.BigEndian.Uint64(data[16:24])
	N.Uid = binary.BigEndian.Uint32(data[24:28])
	N.Gid = binary.BigEndian.Uint32(data[28:32])
	N.Mode = binary.BigEndian.Uint32(data[32:36])
	N.Ctime = time.Unix(int64(binary.BigEndian.Uint64(data[36:44])), 0)
	N.Mtime = time.Unix(int64(binary.BigEndian.Uint64(data[44:52])), 0)
	N.FileName = string(data[FixedNeedleSize:])
	return
}

func (f *Needle)Read(b []byte) (n int, err error) {
	start := f.NeedleOffset + FixedNeedleSize + uint64(len(f.FileName)) + f.CurrentOffset
	end := start + f.FileSize
	length := end - start
	if len(b) > int(length) {
		b = b[:length]
	}
	n, err = f.File.ReadAt(b, int64(start))
	f.CurrentOffset += uint64(n)
	if f.CurrentOffset >= f.FileSize {
		err = io.EOF
	}
	return n, nil
}

func (f *Needle)Write(b []byte) (n int, err error) {
	start := f.NeedleOffset + FixedNeedleSize + uint64(len(f.FileName)) + f.CurrentOffset
	end := start + f.FileSize
	length := end - start

	if len(b) > int(length) {
		return 0, errSmallNeedle
	} else {
		num, err := f.File.WriteAt(b, int64(start))
		if err != nil {
			return num, err
		}
		f.CurrentOffset += uint64(num)
		return num, nil
	}
}