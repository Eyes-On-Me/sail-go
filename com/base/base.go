package base

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func CallerInfoGet() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (strings.Index(file, "src/github") == -1) && (strings.Index(file, "src/prise/foundation") == -1) && (strings.Index(file, "src/net") == -1) && (strings.Index(file, "src/runtime") == -1) {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
			return fmt.Sprintf("%v:%v", file, line)
		}
	}
	return ""
}

func FuncNameGet(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
