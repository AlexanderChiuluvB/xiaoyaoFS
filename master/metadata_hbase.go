package master

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/filter"
	"github.com/tsuna/gohbase/hrpc"
	"io"
	"strings"
)

type HbaseStore struct {
	client gohbase.Client
	adminClient gohbase.AdminClient
}

func NewHbaseStore(config *config.Config) (store *HbaseStore, err error) {
	store = new(HbaseStore)
	store.client = gohbase.NewClient(config.HbaseHost)
	store.adminClient = gohbase.NewAdminClient(config.HbaseHost)


	dit := hrpc.NewDisableTable(context.Background(), []byte("filemeta"))
	err = store.adminClient.DisableTable(dit)
	if err != nil {
		if !strings.Contains(err.Error(), "TableNotEnabledException") {
			panic(err)
		}
	}

	det := hrpc.NewDeleteTable(context.Background(), []byte("filemeta"))
	err = store.adminClient.DeleteTable(det)
	if err != nil {
		panic(err)
	}


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

func (store *HbaseStore) Get(filePath string) (entry *Entry, err error) {
	families := map[string][]string{"cf": nil}
	key := []byte(filePath)
	getRequest, err := hrpc.NewGetStr(context.Background(), "filemeta", string(key), hrpc.Families(families))
	if err != nil {
		return nil, err
	}
	getResp, err := store.client.Get(getRequest)
	if err != nil {
		return nil, err
	}
	if len(getResp.Cells) == 0 {
		return nil, err
	}
	value := getResp.Cells[0].Value
	entry = new(Entry)
 	err = json.Unmarshal(value, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (store *HbaseStore) GetEntries(filePathPrefix string)(entries []*Entry, err error) {

	prefix := []byte(filePathPrefix)
	pFilter := filter.NewPrefixFilter(prefix)
	scanRequest, err := hrpc.NewScanStr(context.Background(), "filemeta",
		hrpc.Filters(pFilter))
	if err != nil {
		return nil, err
	}
	scanner := store.client.Scan(scanRequest)
	for {
		result, err := scanner.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("scan failed withe error %v", err)
		}
		for _, cell := range result.Cells {
			value := cell.Value
			entry := new(Entry)
			err = json.Unmarshal(value, entry)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (store *HbaseStore) Set(entry *Entry) error {
	key := []byte(entry.FilePath)
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	values := map[string]map[string][]byte{"cf": map[string][]byte{"value": value}}
	putRequest, err := hrpc.NewPutStr(context.Background(), "filemeta", string(key), values)
	if err != nil {
		return fmt.Errorf("failed to create put request: %s", err)
	}
	_, err = store.client.Put(putRequest)
	if err != nil {
		return fmt.Errorf("put key %s and value %s error : %v", entry.FilePath, string(value), err)
	}
	return nil
}

func (store *HbaseStore) Delete(filePath string) error {
	key := []byte(filePath)
	deleteRequest, err := hrpc.NewDelStr(context.Background(), "filemeta", string(key), map[string]map[string][]byte{
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
	//disableRpc := hrpc.NewDisableTable(context.Background(), []byte("filemeta"))
	//err := store.adminClient.DisableTable(disableRpc)
	//if err != nil {
		//return fmt.Errorf("diable table filemeta failed %v", err)
	//}
	//deleteRpc := hrpc.NewDeleteTable(context.Background(), []byte("filemeta"))
	//err = store.adminClient.DeleteTable(deleteRpc)
	//if err != nil {
	//	return fmt.Errorf("delete table filemeta failed %v", err)
	//}
	return nil
}

