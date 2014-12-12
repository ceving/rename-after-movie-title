package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	var genreg bool
	flag.BoolVar(&genreg, "g", false, "Generate registry file.")

	mains.add(
		func() {
			if genreg {
				exe := os.Args[0]
				reg := filepath.Join(
					filepath.Dir(exe),
					strings.TrimSuffix(
						filepath.Base(exe),
						filepath.Ext(exe)) + ".reg")
				fmt.Println(reg)
				recopt := ""
				if !recurse {
					recopt = " -r=false"
				}
				data := `Windows Registry Editor Version 5.00

[HKEY_CLASSES_ROOT\Directory\shell\rename-after-movie-title]
@="Rename after movie title"

[HKEY_CLASSES_ROOT\Directory\shell\rename-after-movie-title\command]
@="\"` + strings.Replace(exe, `\`, `\\`, -1) + recopt + `\" \"%1\""

`
				ioutil.WriteFile(reg, []byte(data), 0644)
				os.Exit(0)
			}
		})
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
