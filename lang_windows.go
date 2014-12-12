package main

import (
	"syscall"
)

var iso_639_1 string = "en"

func init() {
	// Get language and set the headline.
	var kernel32 = syscall.NewLazyDLL("kernel32.dll")
	var GetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
	langid, _, _ := GetUserDefaultUILanguage.Call()
	lang := uint8(uint16(langid) & 0xF)

	// Language codes are documented here:
	// http://msdn.microsoft.com/en-us/library/dd318693%28v=vs.85%29.aspx
	switch {
	case lang == uint8(0x7):
		iso_639_1 = "de"
	}
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
