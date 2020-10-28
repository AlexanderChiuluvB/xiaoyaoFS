package storage

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestStorageAPI(t *testing.T) {

	var (
		config *Config
		store *Store
		body []byte
		resp *http.Response
		buf = &bytes.Buffer{}
		err error
	)

	if config, err = NewConfig("./store.toml"); err != nil {
		t.Errorf("NewConfig() error(%v)", err)
		t.FailNow()
	}

	if store, err = NewStore(config); err != nil {
		t.Errorf("NewStore() error(%v)", err)
		t.FailNow()
	}
	assert.Equal(t,  "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/storeDir", store.StoreDir)

	go store.Start()

	buf.Reset()
	if resp, err = http.Post("http://localhost:7800/add_volume?vid=1", "application/x-www-form-urlencoded", nil); err != nil {
		t.Errorf("http.Post error(%v)", err)
		t.FailNow()
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("ioutil.ReadAll error(%v)", err)
		t.FailNow()
	}

	fmt.Println(len(body))

	buf.Reset()
	buf.WriteString("test-string")
	if resp, err = http.Post("http://localhost:7900/put", "application/x-www-form-urlencoded", buf); err != nil {
		t.Errorf("http.Post error(%v)", err)
		t.FailNow()
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("ioutil.ReadAll error(%v)", err)
		t.FailNow()
	}

	fmt.Println(len(body))
}
