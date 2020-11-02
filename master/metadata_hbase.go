package master

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
	"strconv"
	"strings"
)

type HbaseStore struct {
	client gohbase.Client
	adminClient gohbase.AdminClient
}

func NewHbaseStore(config *storage.Config) (store *HbaseStore, err error) {
	store = new(HbaseStore)
	store.client = gohbase.NewClient(config.HbaseHost)
	store.adminClient = gohbase.NewAdminClient(config.HbaseHost)
	cFamilies := map[string]map[string]string{
		"cf": nil,
	}
	createTableRpc := hrpc.NewCreateTable(context.Background(), []byte("filemeta"), cFamilies)
	err = store.adminClient.CreateTable(createTableRpc)
	if err != nil {
		panic(err)
	}
	return store, nil
}

func (store *HbaseStore) Get(filePath string) (vid uint64, fid uint64, err error) {
	families := map[string][]string{"cf": nil}
	getRequest, err := hrpc.NewGetStr(context.Background(), "filemeta", filePath, hrpc.Families(families))
	if err != nil {
		return -1, -1, err
	}
	getResp, err := store.client.Get(getRequest)
	if err != nil {
		return -1, -1, err
	}
	if len(getResp.Cells) == 0 {
		return -1, -1, err
	}
	value := getResp.Cells[0].Value
	var valueInStr string
 	err = json.Unmarshal(value, &valueInStr)
	if err != nil {
		return -1, -1, err
	}
	parts := strings.Split(valueInStr, "/")
	vid, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	fid, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return vid, fid, nil
}

func (store *HbaseStore) Set(filePath string, vid uint64, fid uint64) error {
	value, err := json.Marshal(fmt.Sprintf("%d/%d", vid, fid))
	if err != nil {
		return err
	}
	values := map[string]map[string][]byte{"cf": map[string][]byte{"value": value}}
	putRequest, err := hrpc.NewPutStr(context.Background(), "filemeta", filePath, values)
	if err != nil {
		return fmt.Errorf("failed to create put request: %s", err)
	}
	_, err = store.client.Put(putRequest)
	if err != nil {
		return fmt.Errorf("put key %s and value %s error : %v", filePath, string(value), err)
	}
	return nil
}

func (store *HbaseStore) Delete(filePath string) error {
	deleteRequest, err := hrpc.NewDelStr(context.Background(), "filemeta", filePath, map[string]map[string][]byte{
		"cf": nil,
	})
	if err != nil {
		return fmt.Errorf("failed to create delete request: %s", err)
	}
	_, err = store.client.Delete(deleteRequest)
	if err != nil {
		return fmt.Errorf("delete key %s error : %v", filePath, err)
	}
	return nil
}


func (store *HbaseStore) Close() error {
	store.client.Close()
	disableRpc := hrpc.NewDisableTable(context.Background(), []byte("filemeta"))
	err := store.adminClient.DisableTable(disableRpc)
	if err != nil {
		return fmt.Errorf("diable table filemeta failed %v", err)
	}
	deleteRpc := hrpc.NewDeleteTable(context.Background(), []byte("filemeta"))
	err = store.adminClient.DeleteTable(deleteRpc)
	if err != nil {
		return fmt.Errorf("delete table filemeta failed %v", err)
	}
	return nil
}

