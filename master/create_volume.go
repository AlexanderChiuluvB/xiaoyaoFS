package master

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func CreateVolume(host string, port int, vid uint64) error {
	resp, err := http.Post(fmt.Sprintf("http://%s:%d/add_volume?vid=%d", host, port, vid), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusCreated {
		return nil
	}else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(body))
	}
}
