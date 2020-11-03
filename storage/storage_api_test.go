package storage


import (
	"bytes"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	config2 "github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func TestStorageAPI(t *testing.T) {

	var (
		config *config2.Config
		store *Store
		body []byte
		resp *http.Response
		buf = &bytes.Buffer{}
		err error
	)

	config, err = config2.NewConfig("../store1.toml")
	assert.NoError(t, err)

	m, err := master.NewMaster(config)
	assert.NoError(t, err)

	go m.Start()

	store, err = NewStore(config)
	assert.NoError(t, err)
	assert.Equal(t,  "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/storeDir1", store.StoreDir)
	defer store.Close()

	go store.Start()

	buf.Reset()
	//TODO 增加删除Volume的接口
	resp, err = http.Post("http://localhost:7900/add_volume?vid=3", "application/x-www-form-urlencoded", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	//put file
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

	req, err := http.NewRequest(http.MethodPost, "http://localhost:7900/put?vid=3&fid=1", writerBuf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	//test get
	resp, err = http.Get(fmt.Sprintf("http://localhost:7900/get?vid=3&fid=1"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFileByte), len(body))
	assert.Equal(t, expectedFileByte, body)

	err = ioutil.WriteFile("../test/logo2.png", body, os.ModePerm)
	assert.NoError(t, err)

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:7900/del?vid=3&fid=1"), nil)
	assert.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)

	resp, err = http.Get(fmt.Sprintf("http://localhost:7900/get?vid=3&fid=1"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

}
