package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/uuid"
	gocache "github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
  OsUid = uint32(os.Getuid())
  OsGid = uint32(os.Getgid())
  )

type MetaID struct {
	VID uint64
	NID uint64
}

type Size interface {
	Size() int64
}

func (m *Master) getFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filePath := r.FormValue("filepath")
	var vid,nid uint64
	var err error

	if metaID, found := m.Cache.c.Get(filePath); found {
		meta := metaID.(*MetaID)
		vid, nid = meta.VID, meta.NID
	} else {
		vid, nid, err = m.Metadata.Get(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	if vid != 0 && nid != 0 {
		m.MapMutex.RLock()
		vStatusList, ok := m.VolumeStatusListMap[vid]
		m.MapMutex.RUnlock()
		if !ok {
			http.Error(w, fmt.Sprintf("Cant find volume %d", vid), http.StatusNotFound)
			return
		}
		length := len(vStatusList)
		for i:=0; i < length; i++ {
			vStatus := vStatusList[i]
			if vStatus.StoreStatus.IsAlive() {
				http.Redirect(w, r, vStatus.GetFileUrl(nid), http.StatusFound)
				return
			}
		}
		http.Error(w, "all volumes is dead", http.StatusInternalServerError)
	}
	http.Error(w, "entry is nil", http.StatusInternalServerError)
}

func (m *Master) insertEntry(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}


func (m *Master) writeData(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}


func (m *Master) getEntry(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (m *Master) getEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//TODO CACHE?
	entries, err := m.Metadata.GetEntries(r.FormValue("prefix"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	entriesBytes, err := json.Marshal(entries)
	if err != nil {
		http.Error(w, fmt.Sprintf("json marshal entry error: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(entriesBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("write marshaled bytes to writer error: %v", err), http.StatusInternalServerError)
		return
	}
}


func (m *Master) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "r.FromFile: " + err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filePath := r.FormValue("filepath")
	fileName := filepath.Base(filePath)

	var fileSize int64
	switch file.(type){
	case *os.File:
		s, _ := file.(*os.File).Stat()
		fileSize = s.Size()
	case Size:
		fileSize = file.(Size).Size()
	}

	vid, writableVolumeStatusList, err := m.getWritableVolumes(uint64(fileSize))
	if err != nil {
		http.Error(w, "m.getWritableVolumes: " + err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "ioutil.Readall " + err.Error(), http.StatusInternalServerError)
		return
	}
	nid := uuid.UniqueId()
	wg := sync.WaitGroup{}
	var uploadErr []error
	for _, vStatus := range writableVolumeStatusList {
		wg.Add(1)
		go func(vs *VolumeStatus) {
			defer wg.Done()
			//给该vid对应的所有volume上传文件
			err = vs.UploadFile(nid, &data, fileName)
			if err != nil {
				uploadErr = append(uploadErr, fmt.Errorf("host: %s port: %d error: %s", vs.StoreStatus.ApiHost, vs.StoreStatus.ApiPort, err))
			}
		}(vStatus)
	}
	wg.Wait()

	if len(uploadErr) !=0 {
		for _, vStatus := range writableVolumeStatusList {
			go vStatus.Delete(nid)
		}
		errStr := ""
		for _, err := range uploadErr {
			errStr += err.Error() + "\n"
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	} else {
		m.Cache.c.Set(filePath, &MetaID{NID: nid, VID: vid}, gocache.DefaultExpiration)
		err = m.Metadata.Set(filePath, vid, nid)
		if err != nil {
			http.Error(w, "m.Metadata.Set: " + err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}

}

func (m *Master) deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filePath := r.FormValue("filepath")
	vid, nid, err := m.Metadata.Get(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if vid != 0 && nid != 0 {
		m.MapMutex.RLock()
		vStatusList, ok := m.VolumeStatusListMap[vid]
		if !ok {
			http.Error(w, fmt.Sprintf("Cant find volume %d", vid), http.StatusNotFound)
			return
		}
		m.MapMutex.RUnlock()
		if !ok {
			http.Error(w, fmt.Sprintf("Cant find volume %d", vid), http.StatusNotFound)
			return
		} else if !m.isValidVolumes(vStatusList, 0) {
			http.Error(w, "can't delete file, because its readonly.", http.StatusNotAcceptable)
			return
		}

		wg := sync.WaitGroup{}
		var deleteErr []error
		for _, vStatus := range vStatusList {
			wg.Add(1)
			go func(vStatus *VolumeStatus) {
				e := vStatus.Delete(nid)
				if e != nil {
					deleteErr = append(
						deleteErr,
						fmt.Errorf("%s:%d %s", vStatus.StoreStatus.ApiHost, vStatus.StoreStatus.ApiPort, e),
					)
				}
				wg.Done()
			}(vStatus)
		}
		wg.Wait()

		//TODO: delMeta if exists
		m.Cache.c.Delete(filePath)
		err = m.Metadata.Delete(filePath)
		if err != nil {
			deleteErr = append(deleteErr, fmt.Errorf("m.Metadata.Delete(%s) %s", r.FormValue("filepath"), err.Error()))
		}

		if len(deleteErr) == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			errStr := ""
			for _, err := range deleteErr {
				errStr += err.Error() + "\n"
			}
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}
	}


}

func (m *Master) heartbeat(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ioutil.Readall " + err.Error(), http.StatusInternalServerError)
		return
	}
	newStorageStatus := new(StorageStatus)
	newStorageStatus.LastHeartbeat = time.Now()
	err = json.Unmarshal(body, newStorageStatus)
	if err != nil {
		http.Error(w, "json.Unmarshal " + err.Error(), http.StatusInternalServerError)
		return
	}

	remoteIP := r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
	if newStorageStatus.ApiHost == ""  {
		newStorageStatus.ApiHost = remoteIP
	}

	m.updateStorageStatus(newStorageStatus)

	if m.needCreateVolume(newStorageStatus) {
		go m.createNewVolume(newStorageStatus)
	}
}
