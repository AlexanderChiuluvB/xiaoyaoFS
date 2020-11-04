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

	file, err := os.Open("../test/logo.png")
	assert.NoError(t, err)
	defer file.Close()

	file2, err := os.Open("../test/logo.png")
	assert.NoError(t, err)
	defer file2.Close()
	expectedFileByte, err := ioutil.ReadAll(file2)
	assert.NoError(t, err)

	writerBuf := &bytes.Buffer{}
	mPart := multipart.NewWriter(writerBuf)
	filePart, err := mPart.CreateFormFile("file", "FduExchange.png")
	assert.NoError(t, err)

	_, err = io.Copy(filePart, file)
	assert.NoError(t, err)
	mPart.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/uploadFile?filepath=%s",
		m.MasterHost, m.MasterPort, file.Name()), writerBuf)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", mPart.FormDataContentType())
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

	//make sure the file we get From both storage is the same

	vid, fid, err := m.Metadata.Get(file.Name())
	assert.NoError(t, err)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", store1.ApiHost,
		store1.ApiPort, vid, fid))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte), len(body))
	assert.Equal(t, expectedFileByte, body)

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", store2.ApiHost,
		store2.ApiPort, vid, fid))
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