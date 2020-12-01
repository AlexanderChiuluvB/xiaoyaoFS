package master

import (
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/uuid"
	"net/http"
	"sync"
)

type Master struct {
	MasterServer *http.ServeMux
	MasterHost string
	MasterPort int

	StorageStatusList []*StorageStatus

	// key: volume id  value: volume status List
	VolumeStatusListMap map[uint64][]*VolumeStatus
	MapMutex sync.RWMutex
	Metadata metadata

	//只要现在的Volume少于MaxVolumeNum,就可以一直增加新的Volume
	MaxVolumeNum int

	Cache *EntryCache
}

func NewMaster(config *config.Config) (*Master, error){
	m := new(Master)
	if config.MasterPort == 0 {
		m.MasterPort = 8888
	} else {
		m.MasterPort = config.MasterPort
	}
	var err error

	switch config.MetaType {
	case "Hbase":
		m.Metadata, err = NewHbaseStore(config)
		if err != nil {
			panic(fmt.Errorf("NewHbase error %v", err))
		}
	case "Cassandra":
		m.Metadata, err = NewCassandraStore(config)
		if err != nil {
			panic(fmt.Errorf("NewCassandra error %v", err))
		}
	case "Redis":
		m.Metadata, err = NewRedisStore(config)
		if err != nil {
			panic(fmt.Errorf("NewRedis error %v", err))
		}
	}

	if config.MasterHost == "" {
		m.MasterHost = "localhost"
	} else {
		m.MasterHost = config.MasterHost
	}
	m.MaxVolumeNum = config.MaxVolumeNum

	m.StorageStatusList = make([]*StorageStatus, 0, 1)
	m.VolumeStatusListMap = make(map[uint64][]*VolumeStatus)

//	m.Cache = New(config.Mc, time.Duration(config.ExpireMc))

	m.MasterServer = http.NewServeMux()
	m.MasterServer.HandleFunc("/getFile", m.getFile)
	m.MasterServer.HandleFunc("/getEntry", m.getEntry)
	m.MasterServer.HandleFunc("/getEntries", m.getEntries)
	m.MasterServer.HandleFunc("/insertEntry", m.insertEntry)
	m.MasterServer.HandleFunc("/uploadFile", m.uploadFile)
	m.MasterServer.HandleFunc("/writeData", m.writeData)
	m.MasterServer.HandleFunc("/deleteFile", m.deleteFile)
	m.MasterServer.HandleFunc("/heartbeat", m.heartbeat)


	return m, nil
}

func (m *Master) Start() {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", m.MasterHost, m.MasterPort), m.MasterServer)
	if err != nil {
		panic(err)
	}
}

func (m *Master) Close() {
	m.Metadata.Close()
}

// TODO: 负载均衡现在是遍历随机选一个
func (m *Master) getWritableVolumes(size uint64) (uint64, []*VolumeStatus, error) {
	m.MapMutex.RLock()
	defer m.MapMutex.RUnlock()

	for vid, vStatusList := range m.VolumeStatusListMap {
		if m.isValidVolumes(vStatusList, size){
			return vid, vStatusList, nil
		}
	}
	return 0, nil, errors.New("can't find writable volumes")
}

func (m *Master) isValidVolumes(vStatusList []*VolumeStatus, size uint64) bool {
	for _, vs := range vStatusList {
		if !vs.StoreStatus.IsAlive() {
			return false
		}
		if !vs.IsWritable(size) {
			return false
		}
		if !vs.HasEnoughSpace() {
			return false
		}
	}
	return len(vStatusList) != 0
}

//给所有的storage server创建新的volume
func (m *Master) createNewVolume(status *StorageStatus) error {
	if !status.IsAlive() {
		return fmt.Errorf("%s:%d is dead", status.ApiHost, status.ApiPort)
	}

	storageStatusList, err := m.getStorageStatusList(status)
	if err != nil {
		return err
	}

	vid := uuid.UniqueId()
	for _, status := range storageStatusList {
		err := status.CreateVolume(vid)
		if err != nil {
			return err
		}
	}
	//构造vstatusListMap[vid] = vStatusList
	vStatusList := make([]*VolumeStatus, 0, len(storageStatusList))
	for _, storageStatus := range storageStatusList {
		for _, volumeStatus := range storageStatus.VStatusList {
			if volumeStatus.VolumeId == vid {
				vStatusList = append(vStatusList, volumeStatus)
				break
			}
		}
	}

	m.MapMutex.RLock()
	m.VolumeStatusListMap[vid] = vStatusList
	m.MapMutex.RUnlock()
	return nil

}

func (m *Master) needCreateVolume(status *StorageStatus) bool {
	m.MapMutex.RLock()
	defer m.MapMutex.RUnlock()

	need := true
	if len(m.VolumeStatusListMap) < m.MaxVolumeNum {
		return true
	}
	for _, vs := range status.VStatusList {
		if m.isValidVolumes(m.VolumeStatusListMap[vs.VolumeId], 0) {
			need = false
			break
		}
	}
	return need && status.CanCreateVolume
}

func (m *Master) updateStorageStatus(newStatus *StorageStatus) {
	m.MapMutex.RLock()
	defer m.MapMutex.RUnlock()

	//update storageStatusList 和 volumeStatusListMap

	for idx, oldStatus := range m.StorageStatusList {
		if newStatus.ApiHost == oldStatus.ApiHost && newStatus.ApiPort == oldStatus.ApiPort {
			m.StorageStatusList = append(m.StorageStatusList[:idx], m.StorageStatusList[idx+1:]...)
			//把volumeStatusListmap中所有是oldStatus的volumeStatus删除
			for _, vs := range oldStatus.VStatusList {
				vsList := m.VolumeStatusListMap[vs.VolumeId]
				if len(vsList) == 1 {
					delete(m.VolumeStatusListMap, vs.VolumeId)
					continue
				}
				for i, vs_ := range vsList {
					if vs_ == vs {
						vsList = append(vsList[:i], vsList[i+1:]...)
						break
					}
				}
				m.VolumeStatusListMap[vs.VolumeId] = vsList
			}
			break
		}
	}

	m.StorageStatusList = append(m.StorageStatusList, newStatus)

	//把newStorageStatus的Volume StatusList信息更新到volumeStatusListMap中
	for _, vs := range newStatus.VStatusList {
		vs.StoreStatus = newStatus
		vsList := m.VolumeStatusListMap[vs.VolumeId]
		if vsList == nil {
			vsList = []*VolumeStatus{vs}
		} else {
			vsList = append(vsList, vs)
		}
		m.VolumeStatusListMap[vs.VolumeId] = vsList
	}
}

func (m *Master) getStorageStatusList(newStatus *StorageStatus) ([]*StorageStatus, error) {
	m.MapMutex.RLock()
	defer m.MapMutex.RUnlock()

	resultStorageStatusList := []*StorageStatus{newStatus}
	for _, status := range m.StorageStatusList {
		resultStorageStatusList = append(resultStorageStatusList, status)
	}
	return resultStorageStatusList, nil
}




