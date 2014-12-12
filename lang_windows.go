// +build windows

package main

import (
	"syscall"
)

func init() {
	// Get language and set the headline.
	var kernel32 = syscall.NewLazyDLL("kernel32.dll")
	var GetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
	langid, _, _ := GetUserDefaultUILanguage.Call()
	lid := uint8(uint16(langid) & 0xF)

	// Language codes are documented here:
	// http://msdn.microsoft.com/en-us/library/dd318693%28v=vs.85%29.aspx
	switch {
	case lid == uint8(0x7):
		lang = de
	}
}
