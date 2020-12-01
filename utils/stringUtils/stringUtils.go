package stringUtils

import (
	"path/filepath"
	"strings"
)

func DirAndName(fullPath string) (string, string) {
	dir, name := filepath.Split(fullPath)
	name = strings.ToValidUTF8(name, "?")
	if dir == "/" {
		return dir, name
	}
	if len(dir) < 1 {
		return "/", ""
	}
	return dir[:len(dir)-1], name
}

func FullPath(dir, name string) string {
	if strings.HasSuffix(dir, "/") {
		return dir+name
	}
	return dir + "/" + name
}