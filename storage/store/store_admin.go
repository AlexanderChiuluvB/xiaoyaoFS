package store

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"net/http"
	"strconv"
)

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
		http.Error(w, fmt.Sprintf("create new volume for vid %d in dir %s error(%v)", r.FormValue("vid"),s.StoreDir,err),
			http.StatusInternalServerError)
		return
	}
	s.Volumes[vid] = v
	w.WriteHeader(http.StatusCreated)
	return
}