package storage

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strconv"
)

type Size interface {
	Size() int64
}

func (s *Store) Get(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		vid uint64
		fid uint64
	)
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("fid"), err), http.StatusBadRequest)
		return
	}

	n, err := s.Cache.GetNeedle(vid, fid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get Cache Needle of fid %d of volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
		return
	}
	if n == nil {
		n, err = s.Directory.Get(vid, fid)
		if err != nil {
			http.Error(w, fmt.Sprintf("Get Needle of fid %d of volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
			return
		}
		err = s.Cache.SetNeedle(vid, fid, n)
		if err != nil {
			http.Error(w, fmt.Sprintf("Set Needle of fid %d of volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
			return
		}
	}

	n.File = s.Volumes[vid].File

	w.Header().Set("Content-Type", get_content_type(n.FileName))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf("%d", fid))
	w.Header().Set("Content-Length", strconv.FormatUint(n.FileSize, 10))
	etagMatch := false
	if r.Header.Get("If-None-Match") != "" {
		s := r.Header.Get("If-None-Match")
		if etag, err := strconv.ParseUint(s[1:len(s) - 1], 10, 64); err == nil && etag == fid {
			etagMatch = true
		}
	}
	if etagMatch {
		w.WriteHeader(http.StatusNotModified)
	} else {
		_, err = io.CopyN(w, n, int64(n.FileSize))
		if err != nil {
			http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
			return
		}
	}
}

/*
func (s *Store) GetNeedle(w http.ResponseWriter, r *http.Request){
	var (
		err              error
		vid uint64
		fid uint64
	)
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("fid"), err), http.StatusBadRequest)
		return
	}

	v := s.Volumes[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	n, err := v.GetNeedle(fid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get Needle of fid %d of volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
		return
	}
	needleBytes, err := json.Marshal(n)
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal needle data of fid %d of volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
	}

	reader := bytes.NewReader(needleBytes)
	w.Header().Set("Content-Type", get_content_type(n.FileName))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf("%d", fid))
	w.Header().Set("Content-Length", strconv.FormatUint(n.FileSize, 10))
	etagMatch := false
	if r.Header.Get("If-None-Match") != "" {
		s := r.Header.Get("If-None-Match")
		if etag, err := strconv.ParseUint(s[1:len(s) - 1], 10, 64); err == nil && etag == fid {
			etagMatch = true
		}
	}
	if etagMatch {
		w.WriteHeader(http.StatusNotModified)
	} else {
		_, err = reader.WriteTo(w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
			return
		}
	}
}
*/

func (s *Store) Put(w http.ResponseWriter, r *http.Request) {
	var (
		v *volume.Volume
		vid uint64
		fid uint64
		err error
	)

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}
	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("fid"), err), http.StatusBadRequest)
		return
	}
	v = s.Volumes[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// todo use io.Copy to speed up bytes copy
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	needle, err := v.NewFile(fid, &data, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Directory.Set(vid, fid, needle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Cache.SetNeedle(vid, fid, needle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Store) Del(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		fid, vid uint64
	)
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusNotFound)
		return
	}

	err = s.Directory.Del(vid, fid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	s.Cache.DelNeedle(vid, fid)
}


func (s *Store) AddVolume(w http.ResponseWriter, r *http.Request) {
	var (err error
		vid uint64
		v *volume.Volume
	)
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}
	if v, err = volume.NewVolume(vid, s.StoreDir); err != nil {
		http.Error(w, fmt.Sprintf("create new volume for vid %s in dir %s error(%v)", r.FormValue("vid"), s.StoreDir,err),
			http.StatusInternalServerError)
		return
	}
	s.Volumes[vid] = v
	w.WriteHeader(http.StatusCreated)
	return
}

func get_content_type(filepath string) string {
	content_type := "application/octet-stream"
	ext := path.Ext(filepath)
	if ext != "" && ext != "." {
		content_type_ := mime.TypeByExtension(ext)
		if content_type_ != "" {
			content_type = content_type_
		}
	}
	return content_type
}
