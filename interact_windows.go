// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var windows struct {
	user32 *syscall.LazyDLL
	MessageBoxW *syscall.LazyProc
	MB_ICONQUESTION,
	MB_ICONINFORMATION,
	MB_YESNO,
	MB_OK,
	IDYES uint
}

func init() {
	windows.user32 = syscall.NewLazyDLL("user32.dll")
	windows.MessageBoxW = windows.user32.NewProc("MessageBoxW")
	windows.MB_ICONINFORMATION = 0x40
	windows.MB_ICONQUESTION    = 0x20
	windows.MB_YESNO = 0x4
	windows.MB_OK    = 0x0
	windows.IDYES = 6
}

func ask(question string) bool {
	ret, _, _ := windows.MessageBoxW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(question))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(l15n[lang][application_title]))),
		uintptr(windows.MB_ICONQUESTION | windows.MB_YESNO))
	return uint(ret) == windows.IDYES
}

func info(message string) {
	windows.MessageBoxW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(l15n[lang][application_title]))),
		uintptr(windows.MB_ICONINFORMATION | windows.MB_OK))
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
