package volume

import "github.com/AlexanderChiuluvB/xiaoyaoFS/storage/store"

type VolumeStatus struct {
	VolumeId uint64
	VolumeSize uint64
	VolumeMaxFreeSize uint64

	Writable bool

	StoreStatus *store.StorageStatus `json:"-"`

}
