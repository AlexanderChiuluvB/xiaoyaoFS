package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"io/ioutil"
	"net/http"
)

func Heartbeat(host string, port int, ss *master.StorageStatus) error {
	url := fmt.Sprintf("http://%s:%d/heartbeat", host, port)
	body, err := json.Marshal(ss)
	reader := bytes.NewReader(body)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ = ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d != 200  body: %s", resp.StatusCode, body)
	}
	return nil
}
