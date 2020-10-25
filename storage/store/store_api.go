package store

import (
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	log "github.com/golang/glog"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"time"
)

func (s *Store) Get(w http.ResponseWriter, r *http.Request) {
	var (
		ret              = http.StatusOK
		err              error
		params           = r.URL.Query()
		vid uint64
		fid uint64
	)
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer HttpGetWriter(r, w, time.Now(), &err, &ret)

	if vid, err = strconv.ParseUint(params.Get("vid"), 10, 64); err != nil {
		log.Errorf("strconv.ParseInt(\"%s\") error(%v)", params.Get("vid"), err)
		ret = http.StatusBadRequest
		return
	}

	v := s.Volumes[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	if fid, err = strconv.ParseUint(params.Get("fid"), 10, 64); err != nil {
		log.Errorf("strconv.ParseInt(\"%s\") error(%v)", params.Get("vid"), err)
		ret = http.StatusBadRequest
		return
	}

	n, err := v.GetNeedle(fid)
	if err != nil {
		log.Errorf("Get needle of fid %d if volume vid %d error %v", fid, vid, err)
		ret = http.StatusBadRequest
		return
	}

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
	} else if r.Method != http.MethodHead {
		io.Copy(w, n.File)
	}

}

func (s *Store) Put(w http.ResponseWriter, r *http.Request) {
	var (
		v *volume.Volume
		vid uint64
		fid uint64
		err error
		res = map[string]interface{}{}
	)

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer HttpPostWriter(r, w, time.Now(), &err, res)
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		err = errors.New(fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err))
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

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		err = errors.New(fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("fid"), err))
		return
	}

	v.NewNeedle()



}

func (s *Store) Del(writer http.ResponseWriter, request *http.Request) {

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
