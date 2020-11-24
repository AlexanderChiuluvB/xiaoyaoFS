package master

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/uuid"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
  OS_UID = uint32(os.Getuid())
  OS_GID = uint32(os.Getgid())
)
type Size interface {
	Size() int64
}

func (m *Master) getFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filePath := r.FormValue("filepath")

	entry, err := m.Metadata.Get(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if entry != nil {
		m.MapMutex.RLock()
		vStatusList, ok := m.VolumeStatusListMap[entry.Vid]
		m.MapMutex.RUnlock()
		if !ok {
			http.Error(w, fmt.Sprintf("Cant find volume %d", entry.Vid), http.StatusNotFound)
			return
		}
		length := len(vStatusList)
		for i:=0; i < length; i++ {
			vStatus := vStatusList[i]
			if vStatus.StoreStatus.IsAlive() {
				http.Redirect(w, r, vStatus.GetFileUrl(entry.Nid), http.StatusFound)
				return
			}
		}
		http.Error(w, "all volumes is dead", http.StatusInternalServerError)
	}
	http.Error(w, "entry is nil", http.StatusInternalServerError)
}

func (m *Master) insertEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("entry")
	if err != nil {
		http.Error(w, "r.FromFile: " + err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "ioutil.Readall " + err.Error(), http.StatusInternalServerError)
		return
	}
	entry := new(Entry)
	err = json.Unmarshal(data, entry)
	if err != nil {
		http.Error(w, "Unmarshal " + err.Error(), http.StatusInternalServerError)
		return
	}
	err = m.Cache.SetMeta(entry.FilePath, entry)
	if err != nil {
		http.Error(w, "m.Cache.SetMeta: " + err.Error(), http.StatusInternalServerError)
		return
	}
	err = m.Metadata.Set(entry)
	if err != nil {
		http.Error(w, "set entry " + err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}


func (m *Master) writeData(w http.ResponseWriter, r *http.Request) {
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
	fid := uuid.UniqueId()
	wg := sync.WaitGroup{}
	var uploadErr []error
	for _, vStatus := range writableVolumeStatusList {
		wg.Add(1)
		go func(vs *VolumeStatus) {
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
		entry, _ := m.Cache.GetMeta(filePath)
		if entry == nil {
			entry, err = m.Metadata.Get(filePath)
			if err != nil {
				http.Error(w, "m.Metadata.Get: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		entry.Nid = fid
		entry.Vid = vid
		entry.Mtime = time.Now()
		entry.FileSize = uint64(fileSize)
		err = m.Cache.SetMeta(entry.FilePath, entry)
		if err != nil {
			http.Error(w, "m.Cache.SetMeta: " + err.Error(), http.StatusInternalServerError)
			return
		}

		err = m.Metadata.Set(entry)
		if err != nil {
			http.Error(w, "m.Metadata.Set: " + err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}


func (m *Master) getEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filePath := r.FormValue("filepath")
	//cache
	entry, err:= m.Cache.GetMeta(filePath)
	if entry == nil {
		entry, err = m.Metadata.Get(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		http.Error(w, fmt.Sprintf("json marshal entry error: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(entryBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("write marshaled bytes to writer error: %v", err), http.StatusInternalServerError)
		return
	}
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
	fid := uuid.UniqueId()
	wg := sync.WaitGroup{}
	var uploadErr []error
	for _, vStatus := range writableVolumeStatusList {
		wg.Add(1)
		go func(vs *VolumeStatus) {
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
		entry := new(Entry)
		entry.Nid = fid
		entry.Vid = vid
		entry.FilePath = filePath
		entry.Mtime = time.Now()
		entry.Ctime = time.Now()
		entry.FileSize = uint64(fileSize)
		entry.Uid = OS_UID
		entry.Gid = OS_GID
		entry.Mode = uint32(os.ModePerm)

		/*err = m.Cache.SetMeta(entry.FilePath, entry)
		if err != nil {
			http.Error(w, "m.Cache.SetMeta: " + err.Error(), http.StatusInternalServerError)
			return
		}*/

		err = m.Metadata.Set(entry)
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
	entry, err := m.Metadata.Get(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if entry != nil {
		m.MapMutex.RLock()
		vStatusList, ok := m.VolumeStatusListMap[entry.Vid]
		m.MapMutex.RUnlock()
		if !ok {
			http.Error(w, fmt.Sprintf("Cant find volume %d", entry.Vid), http.StatusNotFound)
			return
		} else if !m.isValidVolumes(vStatusList, 0) {
			http.Error(w, "can't delete file, because its readonly.", http.StatusNotAcceptable)
		}

		wg := sync.WaitGroup{}
		var deleteErr []error
		for _, vStatus := range vStatusList {
			wg.Add(1)
			go func(vStatus *VolumeStatus) {
				e := vStatus.Delete(entry.Nid)
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
		//_ = m.Cache.DelMeta(filePath)
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
