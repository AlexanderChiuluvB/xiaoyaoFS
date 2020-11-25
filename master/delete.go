package master

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Delete(host string, port int, vid uint64, fid uint64) error {
	url := fmt.Sprintf("http://%s:%d/del?vid=%d&fid=%d", host, port, vid, fid)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}else {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
}
