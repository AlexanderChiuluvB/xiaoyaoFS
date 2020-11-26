package storage

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/disk"
	"io/ioutil"
	"log"
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
	NeedleLock      sync.RWMutex

	StoreDir string //Store对应的目录，该目录下存放着各个Volume File

	//get/upload/delete file
	ApiServer *http.ServeMux
	ApiHost string
	ApiPort int

	// each store server connects to a master server
	MasterHost string
	MasterPort int

	Directory *volume.LeveldbDirectory
	Cache *NeedleCache
}

func NewStore(config *config.Config) (*Store, error) {
	// create if not exists
	_, err := os.Stat(config.StoreDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(config.StoreDir, os.ModePerm)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	store := new(Store)
	store.StoreDir = config.StoreDir
	volumeInfos, err := ioutil.ReadDir(config.StoreDir)
	if err != nil {
		panic(err)
	}

	store.Volumes = make(map[uint64]*volume.Volume)

	for _, volumeFile := range volumeInfos {
		volumeFileName := volumeFile.Name()
		if strings.HasSuffix(volumeFileName, ".volume") {
			volumeId, err := strconv.ParseUint(volumeFileName[:len(volumeFileName)-7], 10, 64)
			if err != nil {
				return nil, err
			}
			store.Volumes[volumeId], err = volume.NewVolume(volumeId, config.StoreDir)
			if err != nil {
				return nil, err
			}
		}
	}

	if config.StoreApiHost == "" {
		store.ApiHost = "localhost"
	} else {
		store.ApiHost = config.StoreApiHost
	}

	if config.StoreApiPort == 0 {
		store.ApiPort = 7900
	} else {
		store.ApiPort = config.StoreApiPort
	}

	if config.MasterHost == "" {
		store.MasterHost = "localhost"
	} else {
		store.MasterHost = config.MasterHost
	}

	if config.MasterPort == 0 {
		store.MasterPort = 8888
	} else {
		store.MasterPort = config.MasterPort
	}

	store.Directory, err = volume.NewLeveldbDirectory(config.StoreDir)

	store.ApiServer = http.NewServeMux()

	store.ApiServer.HandleFunc("/add_volume", store.AddVolume)
	store.ApiServer.HandleFunc("/get", store.Get)
	//store.ApiServer.HandleFunc("/getNeedle", store.GetNeedle)
	store.ApiServer.HandleFunc("/put", store.Put)
	store.ApiServer.HandleFunc("/del", store.Del)

	return store, nil
}

func (store *Store) Start() {
	go store.HeartBeat()

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", store.ApiHost, store.ApiPort), store.ApiServer)
	if err != nil {
		panic(err)
	}
}

func (store *Store) Close() {
	store.Directory.Close()
}

func (store *Store) HeartBeat() {

	tick := time.NewTicker(HeartBeatInterval)
	defer tick.Stop()

	for {
		ss := new(master.StorageStatus)
		ss.ApiHost = store.ApiHost
		ss.ApiPort = store.ApiPort
		ss.VStatusList = make([]*master.VolumeStatus, 0, len(store.Volumes))
		
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
			volumeStatus := new(master.VolumeStatus)
			volumeStatus.VolumeId = vid
			volumeStatus.VolumeSize = v.GetVolumeSize()
			volumeStatus.Writable = v.Writeable
			volumeStatus.VolumeMaxFreeSize = v.RemainingSpace()
			ss.VStatusList = append(ss.VStatusList, volumeStatus)
		}

		err := api.Heartbeat(store.MasterHost, store.MasterPort, ss)
		if err != nil {
			panic(err)
		}
		<- tick.C
	}
}


