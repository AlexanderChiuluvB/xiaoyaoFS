package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get(host string, port int, filePath string) ([]byte, error) {

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/getFile?filepath=%s", host,
		port, filePath))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	}else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}

}

func GetNeedle(host string, port int, filePath string) ([]byte, error) {

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/getNeedle?filepath=%s", host,
		port, filePath))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	}else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}
}