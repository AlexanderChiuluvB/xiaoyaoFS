package storage

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNeedleMarshal(t *testing.T) {

	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, "text.txt", 2)
	assert.NoError(t, err)
	defer v.DelNeedle(1)
	data, err := MarshalBinary(n)
	assert.NoError(t, err)
	newN, err := UnMarshalBinary(data)
	assert.NoError(t, err)
	assert.NotNil(t, newN)
	assert.Equal(t, n.Id, newN.Id)
	assert.Equal(t, n.NeedleOffset, newN.NeedleOffset)
	assert.Equal(t, n.FileSize, newN.FileSize)
	assert.Equal(t, n.CurrentOffset, newN.CurrentOffset)
	// TODO 时间戳序列化和反序列化会有精度差异，统一用uin64来保存时间戳
	// 同时反序列化出来的needle是没有包含其os.File指针的
	// 反序列化出来的needle的os.File指针需要另外自己设置
	t.Logf("%+v", n)
	t.Logf("%+v", newN)
}

func TestNeedleReadWrite(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	data := []byte("20")
	n, err := v.NewNeedle(1,"test.jpg", 100)
	assert.NoError(t, err)

	num, err := n.Write(data)
	assert.NoError(t, err)
	var readByte []byte = make([]byte, num)
	n, err = v.GetNeedle(1)
	assert.NoError(t, err)
	num, err = n.Read(readByte)
	assert.NoError(t, err)
	t.Log(string(readByte))
	assert.Equal(t,data,readByte)
}

func TestNeedle_MultiReadWrite(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		n, err := v.NewNeedle(uint64(i), fmt.Sprintf("%d.jpg", i), 4)
		assert.NoError(t, err)
		var data []byte = make([]byte, 4)
		binary.BigEndian.PutUint32(data, uint32(i*i))
		num, err := n.Write(data)
		assert.NoError(t, err)
		assert.Equal(t,num,len(data))
	}
	for i := 0; i < 10; i++ {
		n, err := v.GetNeedle(uint64(i))
		assert.NoError(t, err)
		t.Logf("%+v", n)
		var readByte []byte = make([]byte, 4)
		_, err = n.Read(readByte)
		assert.NoError(t, err)
		origin := binary.BigEndian.Uint32(readByte)
		t.Log(origin)
		assert.Equal(t,uint32(i*i),origin)
	}
}


func TestVolume_DelNeedle(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	data := []byte("aaa")
	err = v.NewFile(1, &data, "1.txt")
	assert.NoError(t, err)
	t.Log("New:", 1)
	assert.True(t, v.Directory.Has(1))
	err = v.DelNeedle(1)
	assert.NoError(t, err)
	assert.False(t, v.Directory.Has(1))

}
