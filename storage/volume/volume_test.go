package volume

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewVolume(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	dir, _ := ioutil.TempDir(testDir, "")
	defer os.RemoveAll(dir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	t.Logf("%+v", v)
}

func TestNewVolume_NewNeedle(t *testing.T) {
	testDir := "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test"
	dir, _ := ioutil.TempDir(testDir, "")
	defer os.RemoveAll(dir)
	v, err := NewVolume(1, testDir)
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, "test.txt", []byte("test"))
	defer v.DelNeedle(1)
	assert.NoError(t, err)
	t.Logf("%+v", n)
}

