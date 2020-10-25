package store

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage/volume"
	"net/http"
	log "github.com/golang/glog"
	"strconv"
	"time"
)

func (s *Store) AddVolume(wr http.ResponseWriter, r *http.Request) {
	var (err error
		res = map[string]interface{}{}
		vid uint64
		v *volume.Volume
		)
	if r.Method != "POST" {
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer HttpPostWriter(r, wr, time.Now(), &err, res)

	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		log.Errorf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err)
		return
	}
	log.Infof("add volume: %d", vid)
	if v, err = volume.NewVolume(vid, s.StoreDir); err != nil {
		log.Errorf("create new volume for vid %d in dir %s error(%v)", r.FormValue("vid"),s.StoreDir,err)
		return
	}
	s.Volumes[vid] = v
	return
}