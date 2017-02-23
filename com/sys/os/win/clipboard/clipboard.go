package clipboard

import (
	"github.com/sail-services/sail-go/com/sys/os/win/api"
	"syscall"
	"unsafe"
)

func Set(str string) {
	str_utf16, _ := syscall.UTF16FromString(str)
	mem := api.Kernel("GlobalAlloc", api.GMEM_MOVEABLE, uintptr(len(str_utf16)*2))
	mem_lock := api.Kernel("GlobalLock", mem)
	api.Kernel("RtlMoveMemory", mem_lock, uintptr(unsafe.Pointer(&str_utf16[0])), uintptr(len(str_utf16)*2))
	api.Kernel("GlobalUnlock", mem)
	api.User("OpenClipboard", 0)
	api.User("EmptyClipboard")
	api.User("SetClipboardData", api.CF_UNICODETEXT, mem)
	api.User("CloseClipboard")
}

func Get() string {
	api.User("OpenClipboard", 0)
	text_src := api.User("GetClipboardData", api.CF_UNICODETEXT)
	api.User("CloseClipboard")
	return syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(text_src))[:])
}
