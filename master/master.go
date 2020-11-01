package master

import (
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
	"net/http"
	"sync"
)

type Master struct {
	MasterServer *http.ServeMux
	MasterHost string
	MasterPort int

	StorageStatusList []*storage.StorageStatus
	// key: volume id  value: volume status List
	VolumeStatusListMap map[uint64][]*storage.VolumeStatus
	MapMutex sync.RWMutex
	Metadata metadata
}

func NewMaster(config *storage.Config) (*Master, error){
	m := new(Master)
	if config.MasterPort == 0 {
		m.MasterPort = 8888
	} else {
		m.MasterPort = config.MasterPort
	}

	if config.MasterHost == "" {
		m.MasterHost = "localhost"
	} else {
		m.MasterHost = config.MasterHost
	}

	m.StorageStatusList = make([]*storage.StorageStatus, 0, 1)
	m.VolumeStatusListMap = make(map[uint64][]*storage.VolumeStatus)

	m.MasterServer = http.NewServeMux()
	m.MasterServer.HandleFunc("/getFile", m.getFile)
	m.MasterServer.HandleFunc("/uploadFile", m.uploadFile)
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

func (m *Master) getWritableVolumes(size uint64) ([]*storage.VolumeStatus, error) {
	m.MapMutex.RLock()
	defer m.MapMutex.RUnlock()

	for _, vStatusList := range m.VolumeStatusListMap {
		if m.isValidVolumes(vStatusList, size){
			return vStatusList, nil
		}
	}
	return nil, errors.New("can't find writable volumes")
}

func (m *Master) isValidVolumes(vStatusList []*storage.VolumeStatus, size uint64) (bool) {
	for _, vs := range vStatusList {
		if !vs.StoreStatus.IsAlive() {
			return false
		}
		if !vs.IsWritable(size) {
			return false
		}
	}
	return len(vStatusList) != 0
}


