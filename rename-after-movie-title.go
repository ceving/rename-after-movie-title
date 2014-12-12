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
)

var imdb_title_url string
var imdb_title_rx *regexp.Regexp

func init() {
	imdb_title_url = "http://www.imdb.com/title/"
	imdb_title_rx = regexp.MustCompile(imdb_title_url + "(tt[0-9]+)/")
}

type NameChange struct {
	old, new string
}

type Main func()
type Mains []Main

func (ms *Mains) add (m Main) {
	*ms = append(*ms, m)
}

func (ms *Mains) run () {
	for _, m := range *ms {
		m()
	}
}

var mains Mains
var recurse bool

func main() {
	flag.BoolVar(&recurse, "r", true, "Scan directories recursively.")
	flag.Parse()
	mains.run()

	args := flag.Args()

	if len(args) == 0 {
		args = append(args, ".")
	}

	var question string
	var todo []NameChange

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
					question += "● " + filepath.Dir(dir) + "\n" +
						"\u2003- " + filepath.Base(dir) + "\n" +
						"\u2003- " + title + "\n"

					// Store in the todo list.
					ren := NameChange{dir, filepath.Join(filepath.Dir(dir), title)}
					fmt.Printf("Adding: %#v\n", ren)
					todo = append(todo, ren)

					// We are done.
					return
				}
			}
		}
	}

	// Process directory
	for _, dir := range args {
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

	if (len(todo) == 0) {
		info(
			l15n[lang][rename_after_movie_title],
			l15n[lang][can_not_find_any_movies])
	} else {

		// Rename directories after confirmation
		if ask(
			l15n[lang][rename_after_movie_title],
			l15n[lang][rename_the_following_directories] + "\n\n" + question) {
			for _, ren := range todo {
				fmt.Printf("Renaming: %s -> %s\n", ren.old, ren.new);
				err := os.Rename(ren.old, ren.new)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
