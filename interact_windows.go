package main

import (
	"syscall"
	"unsafe"
)

var windows struct {
	user32 *syscall.LazyDLL
	MessageBoxW *syscall.LazyProc
	MB_ICONQUESTION int
	MB_YESNO int
	IDYES uint
}

func init() {
	windows.user32 = syscall.NewLazyDLL("user32.dll")
	windows.MessageBoxW = windows.user32.NewProc("MessageBoxW")
	windows.MB_ICONQUESTION = 0x00000020
	windows.MB_YESNO = 0x00000004
	windows.IDYES = 6
}

func ask(headline, question string) bool {
	ret, _, _ := windows.MessageBoxW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(question))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(headline))),
		uintptr(windows.MB_ICONQUESTION | windows.MB_YESNO))
	return uint(ret) == windows.IDYES
}

func fail(headline, message string) {
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
