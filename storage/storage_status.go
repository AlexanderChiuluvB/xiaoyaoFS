package storage

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/api"
	"time"
)

var MaxHeartbeatDuration  = time.Second * 10 //如果超过这个时间间隔仍然没有心跳认定失联
const DEFAULT_VOLUME_MAX_FREE_SIZE uint64 = 10 * (1 << 30)

type StorageStatus struct {
	ApiHost string
	ApiPort int

	DiskSize uint64
	DiskUsed uint64
	DiskFree uint64
	CanCreateVolume bool

	VolumeMaxSize   uint64

	LastHeartbeat   time.Time `json:"-"`

	VStatusList     []*VolumeStatus
}

func (ss *StorageStatus) IsAlive() bool {
	return ss.LastHeartbeat.Add(MaxHeartbeatDuration).After(time.Now())
}

func (ss *StorageStatus) CreateVolume(volumeId uint64) error {
	err := api.CreateVolume(ss.ApiHost, ss.ApiPort, volumeId)
	if err != nil {
		return err
	}

	ss.VStatusList = append(ss.VStatusList, &VolumeStatus{VolumeId: volumeId,
		StoreStatus: ss, Writable: true, VolumeMaxFreeSize: DEFAULT_VOLUME_MAX_FREE_SIZE})
	return nil
}
