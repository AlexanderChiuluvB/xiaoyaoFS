package storage

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *Store) AddVolume(w http.ResponseWriter, r *http.Request) {
	var (err error
		vid uint64
		v *Volume
		)
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}
	if v, err = NewVolume(vid, s.StoreDir); err != nil {
		http.Error(w, fmt.Sprintf("create new volume for vid %s in dir %s error(%v)", r.FormValue("vid"), s.StoreDir,err),
			http.StatusInternalServerError)
		return
	}
	s.Volumes[vid] = v
	w.WriteHeader(http.StatusCreated)
	return
}

