package mount

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDirPath(t *testing.T) {

	p := &Dir{Name: "/some"}
	p = &Dir{Name: "path", parent: p}
	p = &Dir{Name: "to", parent: p}
	p = &Dir{Name: "a", parent: p}
	p = &Dir{Name: "file", parent: p}

	assert.Equal(t, "/some/path/to/a/file", p.FullPath())

	p = &Dir{Name: "/some"}
	assert.Equal(t, "/some", p.FullPath())

	p = &Dir{Name: "/"}
	assert.Equal(t, "/", p.FullPath())

	p = &Dir{Name: "/"}
	p = &Dir{Name: "path", parent: p}
	assert.Equal(t, "/path", p.FullPath())

	p = &Dir{Name: "/"}
	p = &Dir{Name: "path", parent: p}
	p = &Dir{Name: "to", parent: p}
	assert.Equal(t, "/path/to", p.FullPath())

}
