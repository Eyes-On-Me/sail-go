package api

import (
	"syscall"
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 2
)

var (
	u32 = syscall.MustLoadDLL("user32")
	k32 = syscall.MustLoadDLL("kernel32")
)

func User(api string, a ...uintptr) uintptr {
	b, _, _ := u32.MustFindProc(api).Call(a...)
	return b
}

func Kernel(api string, a ...uintptr) uintptr {
	b, _, _ := k32.MustFindProc(api).Call(a...)
	return b
}
