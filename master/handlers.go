package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Size interface {
	Size() int64
}

func (m *Master) getFile(w http.ResponseWriter, r *http.Request) {
	vid, fid, err := m.Metadata.Get(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
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
			http.Redirect(w, r, vStatus.GetFileUrl(fid), http.StatusFound)
			return
		}
	}

	http.Error(w, "all volumes is dead", http.StatusInternalServerError)
}


func (m *Master) uploadFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "r.FromFile: " + err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var dst string
	if r.URL.Path[len(r.URL.Path) - 1] == '/' {
		dst = r.URL.Path + filepath.Base(header.Filename)
	} else {
		dst = r.URL.Path
	}
	fileName := filepath.Base(dst)

	var fileSize int64
	switch file.(type){
	case *os.File:
		s, _ := file.(*os.File).Stat()
		fileSize = s.Size()
	case Size:
		fileSize = file.(Size).Size()
	}

	writableVolumeStatusList, err := m.getWritableVolumes(uint64(fileSize))
	if err != nil {
		http.Error(w, "m.getWritableVolumes: " + err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "ioutil.Readall " + err.Error(), http.StatusInternalServerError)
		return
	}
	fid := storage.UniqueId()
	wg := sync.WaitGroup{}
	var uploadErr []error
	for _, vStatus := range writableVolumeStatusList {
		wg.Add(1)
		go func(vs *storage.VolumeStatus) {
			defer wg.Done()
			//给该vid对应的所有volume上传文件
			err = vs.UploadFile(fid, &data, fileName)
			if err != nil {
				uploadErr = append(uploadErr, fmt.Errorf("host: %s port: %d error: %s", vs.StoreStatus.ApiHost, vs.StoreStatus.ApiPort, err))
			}
		}(vStatus)
	}
	wg.Wait()

	if len(uploadErr) !=0 {
		for _, vStatus := range writableVolumeStatusList {
			go vStatus.Delete(fid)
		}
		errStr := ""
		for _, err := range uploadErr {
			errStr += err.Error() + "\n"
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	} else {
		//update meta data
		err = m.Metadata.Set(dst, writableVolumeStatusList[0].VolumeId, fid)
		if err != nil {
			http.Error(w, "m.Metadata.Set: " + err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}

}

func (m *Master) deleteFile(w http.ResponseWriter, r *http.Request) {
	vid, fid, err := m.Metadata.Get(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	m.MapMutex.RLock()
	vStatusList, ok := m.VolumeStatusListMap[vid]
	m.MapMutex.RUnlock()
	if !ok {
		http.Error(w, fmt.Sprintf("Cant find volume %d", vid), http.StatusNotFound)
		return
	} else if !m.isValidVolumes(vStatusList, 0) {
		http.Error(w, "can't delete file, because its readonly.", http.StatusNotAcceptable)
	}

	wg := sync.WaitGroup{}
	var deleteErr []error
	for _, vStatus := range vStatusList {
		wg.Add(1)
		go func(vStatus *storage.VolumeStatus) {
			e := vStatus.Delete(fid)
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

	err = m.Metadata.Delete(r.URL.Path)
	if err != nil {
		deleteErr = append(deleteErr, fmt.Errorf("m.Metadata.Delete(%s) %s", r.URL.Path, err.Error()))
	}

	if len(deleteErr) == 0 {
		w.WriteHeader(http.StatusAccepted)
	} else {
		errStr := ""
		for _, err := range deleteErr {
			errStr += err.Error() + "\n"
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

}

func (m *Master) heartbeat(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ioutil.Readall " + err.Error(), http.StatusInternalServerError)
		return
	}
	newStorageStatus := new(storage.StorageStatus)
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
