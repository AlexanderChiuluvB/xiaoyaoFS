package master

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type VolumeStatus struct {
	VolumeId uint64
	VolumeSize uint64
	VolumeMaxFreeSize uint64

	Writable bool

	StoreStatus *StorageStatus `json:"-"`

}

func (vs *VolumeStatus) GetFileUrl(fid uint64) string {
	return fmt.Sprintf("http://%s:%d/get?vid=%d&fid=%d", vs.StoreStatus.ApiHost, vs.StoreStatus.ApiPort,
		vs.VolumeId, fid)
}

func (vs *VolumeStatus) IsWritable(size uint64) bool {
	return vs.Writable && vs.VolumeMaxFreeSize > size
}

func (vs *VolumeStatus) Delete(fid uint64) error {
	return Delete(vs.StoreStatus.ApiHost, vs.StoreStatus.ApiPort, vs.VolumeId, fid)
}

func (vs *VolumeStatus) UploadFile(fid uint64, data *[]byte, fileName string) error {
	writerBuf := &bytes.Buffer{}
	mPart := multipart.NewWriter(writerBuf)
	filePart, err := mPart.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	_, err = filePart.Write(*data)
	if err != nil {
		return err
	}
	mPart.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/put?vid=%d&fid=%d", vs.StoreStatus.ApiHost,
		vs.StoreStatus.ApiPort, vs.VolumeId, fid), writerBuf)
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

func (vs *VolumeStatus) HasEnoughSpace() bool {
	return vs.VolumeSize / vs.VolumeMaxFreeSize < 100
}

