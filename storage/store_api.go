package storage

import (
	"encoding/json"
	"fmt"
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

	v := s.Volumes[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("fid"), err), http.StatusBadRequest)
		return
	}

	n, err := v.GetNeedle(fid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get Needle of fid %d if volume vid %d error %v", fid, vid, err), http.StatusBadRequest)
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
		//TODO: io.Copy
		var needleData []byte
		n.File.Seek(int64(n.NeedleOffset + FixedNeedleSize + uint64(len(n.FileName)) + n.CurrentOffset),0)
		needleData, err = ioutil.ReadAll(n.File)
		if err != nil {
			http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusBadRequest)
			return
		}
		_, err = w.Write(needleData)
		if err != nil {
			http.Error(w, fmt.Sprintf("write to http.writer error %v", err), http.StatusBadRequest)
			return
		}
	}

}

func (s *Store) Put(w http.ResponseWriter, r *http.Request) {
	var (
		v *Volume
		vid uint64
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

	fid, err := v.NewFile(&data, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fidBytes, err := json.Marshal(fid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(fidBytes)
}

func (s *Store) Del(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		fid, vid uint64
		v        *Volume
	)
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}

	v = s.Volumes[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	if fid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusNotFound)
		return
	}

	err = v.DelNeedle(fid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
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
