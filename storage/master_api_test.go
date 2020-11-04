package storage

import (
	"bytes"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func TestMasterAPI(t *testing.T) {

	var (
		config1 *config.Config
		config2 *config.Config
		m *master.Master
		store1 *Store
		store2 *Store
		body []byte
		resp *http.Response
		err error
	)

	config1, err = config.NewConfig("../store1.toml")
	assert.NoError(t, err)

	config2, err = config.NewConfig("../store2.toml")
	assert.NoError(t, err)

	m, err = master.NewMaster(config1)
	assert.NoError(t, err)

	m.Metadata, err = master.NewHbaseStore(config1)
	assert.NoError(t, err)

	go m.Start()

	store1, err = NewStore(config1)
	assert.NoError(t, err)
	go store1.Start()

	store2, err = NewStore(config2)
	assert.NoError(t, err)
	go store2.Start()

	file, err := os.Open("../test/nut.png")
	assert.NoError(t, err)
	defer file.Close()

	file2, err := os.Open("../test/nut.png")
	assert.NoError(t, err)
	defer file2.Close()
	expectedFileByte, err := ioutil.ReadAll(file2)
	assert.NoError(t, err)

	beanFile, err := os.Open("../test/bean.png")
	assert.NoError(t, err)
	defer beanFile.Close()

	beanFile2, err := os.Open("../test/bean.png")
	assert.NoError(t, err)
	defer beanFile2.Close()
	expectedFileByte2, err := ioutil.ReadAll(beanFile2)
	assert.NoError(t, err)


	writerBuf := &bytes.Buffer{}
	mPart := multipart.NewWriter(writerBuf)
	filePart, err := mPart.CreateFormFile("file", "nut.png")
	assert.NoError(t, err)

	_, err = io.Copy(filePart, file)
	assert.NoError(t, err)
	mPart.Close()



	writerBuf2 := &bytes.Buffer{}
	mPart2 := multipart.NewWriter(writerBuf2)
	filePart2, err := mPart2.CreateFormFile("file", "bean.png")
	assert.NoError(t, err)

	_, err = io.Copy(filePart2, beanFile)
	assert.NoError(t, err)
	mPart2.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/uploadFile?filepath=%s",
		m.MasterHost, m.MasterPort, file.Name()), writerBuf)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/uploadFile?filepath=%s",
		m.MasterHost, m.MasterPort, beanFile.Name()), writerBuf2)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", mPart2.FormDataContentType())
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/getFile?filepath=%s", m.MasterHost,
		m.MasterPort, file.Name()))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte), len(body))
	assert.Equal(t, expectedFileByte, body)

	err = ioutil.WriteFile("../test/download-nut.png", body, os.ModePerm)
	assert.NoError(t, err)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/getFile?filepath=%s", m.MasterHost,
		m.MasterPort, beanFile.Name()))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)


	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte2), len(body))
	assert.Equal(t, expectedFileByte2, body)

	err = ioutil.WriteFile("../test/download-bean.png", body, os.ModePerm)
	assert.NoError(t, err)

	//make sure the file we get From both storage is the same
	vid1, nid1, err := m.Metadata.Get(file.Name())
	assert.NoError(t, err)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", store1.ApiHost,
		store1.ApiPort, vid1, nid1))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte), len(body))
	assert.Equal(t, expectedFileByte, body)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", store2.ApiHost,
		store2.ApiPort, vid1, nid1))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte), len(body))
	assert.Equal(t, expectedFileByte, body)

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%d/deleteFile?filepath=%s",
		m.MasterHost, m.MasterPort, file.Name()), nil)
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/getFile?filepath=%s", m.MasterHost,
		m.MasterPort, file.Name()))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)


}