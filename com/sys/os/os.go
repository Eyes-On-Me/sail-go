package os

import (
	"os"
	"runtime"
	"strings"
)

var (
	is_win_ bool
)

func init() {
	is_win_ = runtime.GOOS == "windows"
}

func IsWin() bool {
	return is_win_
}

// 获取 GOPATH
func GOPATHGet() []string {
	gopath := os.Getenv("GOPATH")
	var paths []string
	if is_win_ {
		gopath = strings.Replace(gopath, "\\", "/", -1)
		paths = strings.Split(gopath, ";")
	} else {
		paths = strings.Split(gopath, ":")
	}
	return paths
}
