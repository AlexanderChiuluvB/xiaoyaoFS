package storage

import "fmt"

type VolumeStatus struct {
	VolumeId uint64
	VolumeSize uint64
	VolumeMaxFreeSize uint64

	Writable bool

	StoreStatus *StorageStatus `json:"-"`

}

func (s *VolumeStatus) GetFileUrl(fid uint64) string {
	return fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", s.StoreStatus.ApiHost, s.StoreStatus.ApiPort,
		s.VolumeId, fid)
}

