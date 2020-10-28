package storage

type VolumeStatus struct {
	VolumeId uint64
	VolumeSize uint64
	VolumeMaxFreeSize uint64

	Writable bool

	StoreStatus *StorageStatus `json:"-"`

}
