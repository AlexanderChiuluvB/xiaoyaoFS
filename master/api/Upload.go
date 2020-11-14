package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func Upload(host string, port int, filePath string) error {

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(filePart, file)
	if err != nil {
		return err
	}

	mPart.Close()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/uploadFile?filepath=%s",
		host, port, filePath), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return nil
}

// Entry of filePath must be created first
func WriteData(host string, port int, filePath string, data *[]byte) error {

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = filePart.Write(*data)
	if err != nil {
		return err
	}

	mPart.Close()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/writeData?filepath=%s",
		host, port, filePath), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return nil
}


func InsertEntry(host string, port int, entry *master.Entry) error {

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("entry", entry.FilePath)
	if err != nil {
		return err
	}

	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = filePart.Write(entryBytes)
	if err != nil {
		return err
	}

	mPart.Close()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/insertEntry",
		host, port), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return nil
}

