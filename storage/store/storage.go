package store

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/disk"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// per 5 secs the store server send a heartbeat http request to master
const HeartBeatInterval time.Duration = time.Second * 5
var MaxDiskUsedPercent uint = 90


// one store contains several volumes
type Store struct {
	Volumes map[uint64]*volume.Volume
	//TODO ZOOKEEPER
	VolumesLock 	sync.Mutex // protect Volumes map

	StoreDir string //Store对应的目录，该目录下存放着各个Volume File

	//add/del volume
	AdminServer *http.ServeMux
	AdminHost string
	AdminPort int

	//get/upload/delete file
	ApiServer *http.ServeMux
	ApiHost string
	ApiPort int

	// each store server connects to a master server
	MasterHost string
	MasterPort int
}

func NewStore(StoreDir string) (*Store, error) {
	f, err := os.OpenFile(StoreDir, os.O_RDWR | os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	store := new(Store)
	store.StoreDir = StoreDir

	volumeInfos, err := ioutil.ReadDir(StoreDir)
	if err != nil {
		panic(err)
	}

	store.Volumes = make(map[uint64]*volume.Volume)

	for _, volumeFile := range volumeInfos {
		volumeFileName := volumeFile.Name()
		if strings.HasSuffix(volumeFileName, ".volume") {
			volumeId, err := strconv.ParseUint(volumeFileName[:len(volumeFileName)-5], 10, 64)
			if err != nil {
				return nil, err
			}
			store.Volumes[volumeId], err = volume.NewVolume(volumeId, StoreDir)
			if err != nil {
				return nil, err
			}
		}
	}

	//TODO use config file to set
	store.AdminHost = "localhost"
	store.AdminPort = 7800

	store.ApiHost = "localhost"
	store.ApiPort = 7900

	store.MasterHost = "localhost"
	store.MasterPort = 8888

	store.AdminServer = http.NewServeMux()
	store.ApiServer = http.NewServeMux()

	store.AdminServer.HandleFunc("/add_volume", store.AddVolume)
	store.ApiServer.HandleFunc("/get", store.Get)
	store.ApiServer.HandleFunc("/put", store.Put)
	store.ApiServer.HandleFunc("/del", store.Del)

	return store, nil
}

func (store *Store) Start() {
	go store.HeartBeat()

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", store.AdminHost, store.AdminPort), store.AdminServer)
		if err != nil {
			panic(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", store.ApiHost, store.ApiPort), store.ApiServer)
	if err != nil {
		panic(err)
	}
}

func (store *Store) HeartBeat() {
	//TODO heartbeat with zookeeper
	
	tick := time.NewTicker(HeartBeatInterval)
	defer tick.Stop()

	for {
		ss := new(StorageStatus)
		ss.AdminHost = store.AdminHost
		ss.AdminHost = store.AdminHost
		ss.AdminPort = store.AdminPort
		ss.ApiHost = store.ApiHost
		ss.ApiPort = store.ApiPort
		ss.VStatusList = make([]*volume.VolumeStatus, 0, len(store.Volumes))
		
		diskUsage, _ := disk.DiskUsage(store.StoreDir)
		ss.DiskFree = diskUsage.Free
		ss.DiskSize = diskUsage.Size
		ss.DiskUsed = diskUsage.Used
		ss.VolumeMaxSize = volume.MaxVolumeSize

		diskUsedPercent := uint(float64(diskUsage.Used) / float64(diskUsage.Size) * 100)
		if diskUsedPercent >= MaxDiskUsedPercent {
			//禁止所有volume再进行truncate
			volume.MaxVolumeSize = 0
			ss.CanCreateVolume = false
		} else {
			ss.CanCreateVolume = true
		}

		//把更新后的status传回给master，由master来决定是否有必要创建新的volume
		for vid, v := range store.Volumes {
			volumeStatus := new(volume.VolumeStatus)
			volumeStatus.VolumeId = vid
			volumeStatus.VolumeSize = v.GetVolumeSize()
			volumeStatus.Writable = v.Writeable
			volumeStatus.VolumeMaxFreeSize = v.RemainingSpace()
			ss.VStatusList = append(ss.VStatusList, volumeStatus)
		}

		api.Heartbeat(store.MasterHost, store.MasterPort, ss)
		<- tick.C
	}
}


