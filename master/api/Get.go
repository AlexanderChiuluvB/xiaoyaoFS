package api

import (
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
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

func GetEntry(host string, port int, filePath string) (entry *master.Entry, err error) {

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/getEntry?filepath=%s", host,
		port, filePath))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		entryBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		entry := new(master.Entry)
		err = json.Unmarshal(entryBytes, entry)
		if err != nil {
			return nil, err
		}
		return entry, nil
	}else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}
}

func GetEntries(host string, port int, filePathPrefix string) (entries []*master.Entry, err error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/getEntries?prefix=%s", host,
		port, filePathPrefix))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		entriesBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(entriesBytes, &entries)
		if err != nil {
			return nil, err
		}
		return entries, nil
	}else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}
}
