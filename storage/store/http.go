package store

import (
	"encoding/json"
	"net/http"
	log "github.com/golang/glog"

	"time"
)

const RetOK = 1

func HttpPostWriter(r *http.Request, wr http.ResponseWriter, start time.Time, err *error, result map[string]interface{}) {
	var (
		byteJson []byte
		err1     error
		errStr   string
		ret      = RetOK
	)
	if *err != nil {
		panic(err)
	}
	result["ret"] = ret
	if byteJson, err1 = json.Marshal(result); err1 != nil {
		log.Errorf("json.Marshal(\"%v\") failed (%v)", result, err1)
		return
	}
	wr.Header().Set("Content-Type", "application/json;charset=utf-8")
	if _, err1 = wr.Write(byteJson); err1 != nil {
		log.Errorf("http Write() error(%v)", err1)
		return
	}
	log.Infof("%s path:%s(params:%s,time:%f,ret:%v[%v])", r.Method,
		r.URL.Path, r.Form.Encode(), time.Now().Sub(start).Seconds(), ret, errStr)
}

func HttpGetWriter(r *http.Request, wr http.ResponseWriter, start time.Time, err *error, ret *int) {
	var errStr string
	if *ret != http.StatusOK {
		if *err != nil {
			errStr = (*err).Error()
		}
		http.Error(wr, errStr, *ret)
	}
	log.Infof("%s path:%s(params:%s,time:%f,err:%s,ret:%v[%v])", r.Method,
		r.URL.Path, r.URL.String(), time.Now().Sub(start).Seconds(), errStr, *ret, errStr)
}