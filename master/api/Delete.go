package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Delete(host string, port int, filepath string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/deleteFile?filepath=%s",
		host, port, filepath), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}else {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}

}
