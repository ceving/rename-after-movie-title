package main

import "fmt"
import "golang.org/x/net/html"
import "io/ioutil"
import "net/http"
import "os"
import "path/filepath"
import "regexp"

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

func main() {
	for _, dir := range os.Args[1:] {
		// Generate absolute path.
		dir, err := filepath.Abs(dir)
		if err != nil {
			panic(err)
		}
		// Search NFO files.
		files, _ := filepath.Glob(filepath.Join(dir, "*.nfo"))
		for _, filename := range files {
			// Read a IMDB movie ID from the file.
			file, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			match := imdb_title_rx.FindSubmatch(file)
			if match != nil {
				id := string(match[1])
				url := imdb_title_url + id + "/"

				// Get HTML.
				resp, err := http.Get(url)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				// Parse HTML.
				doc, err := html.Parse(resp.Body)
				if err != nil {
					panic(err)
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
					fmt.Println(filepath.Dir(dir))
					fmt.Println("  " + filepath.Base(dir) + " -> " + title)
					err := os.Rename(dir, filepath.Join(filepath.Dir(dir), title))
					if err != nil {
						panic(err)
					}
					goto next
				}
			}
		}
	next:
	}
}

// Local Variables:
// compile-command: "go build rename-after-movie-title.go"
// End:
