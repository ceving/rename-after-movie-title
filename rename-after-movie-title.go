package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

var imdb_title_url string
var imdb_title_rx *regexp.Regexp

func init() {
	imdb_title_url = "http://www.imdb.com/title/"
	imdb_title_rx = regexp.MustCompile(imdb_title_url + "(tt[0-9]+)/")
}

type action func(node *html.Node) bool
type predicate func(node *html.Node) bool

func element(name string) predicate {
	return func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == name
	}
}

func attribute(name, value string) predicate {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == name && attr.Val == value {
				return true
			}
		}
		return false
	}
}

func text() predicate {
	return func(node *html.Node) bool {
		return node.Type == html.TextNode
	}
}

func and(a, b predicate) predicate {
	return func(node *html.Node) bool {
		return a(node) && b(node)
	}
}

func not(a predicate) predicate {
	return func(node *html.Node) bool {
		if a(node) {
			return false
		} else {
			return true
		}
	}
}

func matcher(matches predicate, continuation action) action {
	return func(node *html.Node) bool {
		if matches(node) {
			return continuation(node)
		}
		return true
	}
}

func walker(continuation action) action {
	var walk action
	walk = func(node *html.Node) bool {
		if continuation(node) {
			cont := true
			for child := node.FirstChild; child != nil && cont; child = child.NextSibling {
				cont = walk(child)
			}
			return true
		} else {
			return false
		}
	}
	return walk
}

func attrval(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

type Rename struct {
	old, new string
}

func main() {
	var genreg bool
	flag.BoolVar(&genreg, "g", false, "Generate registry file.")
	var recurse bool
	flag.BoolVar(&recurse, "r", true, "Scan directories recursively.")
	flag.Parse()
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
		return
	}
	var question string = ""
	var todo = []Rename{}

	// Search directories which could be renamed.
	search := func(path string) {
		// Generate absolute path.
		dir, err := filepath.Abs(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Search NFO files.
		files, _ := filepath.Glob(filepath.Join(dir, "*.nfo"))
		for _, filename := range files {
			// Read a IMDB movie ID from the file.
			file, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Println(err)
				return
			}
			match := imdb_title_rx.FindSubmatch(file)
			if match != nil {
				id := string(match[1])
				url := imdb_title_url + id + "/"

				// Get HTML.
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()

				// Parse HTML.
				doc, err := html.Parse(resp.Body)
				if err != nil {
					fmt.Println(err)
					return
				}

				// Find title.
				var title string
				find := walker(
					matcher(
						and(element("meta"), attribute("property", "og:title")),
						func(node *html.Node) bool {
							title = attrval(node, "content")
							return false
						}))
				find(doc)
				if title != "" {
					// Rename directory.
					question += filepath.Base(dir) + " → " + title + "\n"

					// Store in the todo list.
					ren := Rename{dir, filepath.Join(filepath.Dir(dir), title)}
					fmt.Printf("Adding: %#v\n", ren)
					todo = append(todo, ren)

					// We are done.
					return
				}
			}
		}
	}

	// Process directory
	for _, dir := range os.Args[1:] {
		if recurse {
			filepath.Walk(dir,
				func(path string, info os.FileInfo, err error) error {
					if info.IsDir() {
						search(path)
					}
					return nil
				})
		} else {
			search(dir)
		}
	}

	// Language codes are documented here:
	// http://msdn.microsoft.com/en-us/library/dd318693%28v=vs.85%29.aspx
	var LANG_GERMAN uint16 = 0x7

	// Get language and set the headline.
	var kernel32 = syscall.NewLazyDLL("kernel32.dll")
	var GetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
	langid, _, _ := GetUserDefaultUILanguage.Call()
	lang := uint16(langid) & 0xF
	var headline string
	switch {
	case lang == LANG_GERMAN:
		headline = "Die folgenden Verzeichnisse umbenennen?"
	case true:
		headline = "Rename the following directories?"
	}

	// Ask if the directories found should be renamed.
	var user32 = syscall.NewLazyDLL("user32.dll")
	var MessageBoxW = user32.NewProc("MessageBoxW")
	var MB_ICONQUESTION = 0x00000020
	var MB_YESNO = 0x00000004
	var IDYES uint = 6
	ret, _, _ := MessageBoxW.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(question))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(headline))),
		uintptr(MB_ICONQUESTION | MB_YESNO));
	if uint(ret) == IDYES {
		// Rename them
		for _, ren := range todo {
			fmt.Printf("Renaming: %s -> %s\n", ren.old, ren.new);
			err := os.Rename(ren.old, ren.new)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui \
//   rename-after-movie-title.go"
// End:
