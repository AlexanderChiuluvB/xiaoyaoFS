package storage

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewVolume(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	t.Logf("%+v", v)
}

func TestNewVolume_NewNeedle(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	os.Mkdir(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, "test.txt", uint64(len([]byte("test"))))
	defer v.DelNeedle(1)
	assert.NoError(t, err)
	assert.Equal(t, "test.txt", n.FileName)
	assert.Equal(t, uint64(4), n.FileSize)
	assert.Equal(t, uint64(8), n.NeedleOffset)
	t.Logf("%+v", n)

	n2, err := v.NewNeedle(2, "test.jpg", uint64(len([]byte("test"))))
	defer v.DelNeedle(2)
	assert.NoError(t, err)
	assert.Equal(t, "test.jpg", n2.FileName)
	assert.Equal(t, uint64(4), n2.FileSize)
	assert.Equal(t, uint64(60), n2.NeedleOffset)
	t.Logf("%+v", n2)
}

