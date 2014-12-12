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
)

var imdb_title_url string
var imdb_title_rx *regexp.Regexp

func init() {
	imdb_title_url = "http://www.imdb.com/title/"
	imdb_title_rx = regexp.MustCompile(
		"(?:http://)?(?:www.)?imdb.[^./]+/title/(tt[0-9]+)/?")
}

type NameChange struct {
	old, neu string
}

func sanitize(name string) string {
	for _, invalid := range []string{`\`, `/`, `:`, `*`, `?`, `<`, `>`, `|`, `"`} {
		name = strings.Replace(name, invalid, "", -1)
	}
	return name
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
	var recurse bool
	flag.BoolVar(&recurse, "r", true, "Scan directories recursively.")
	flag.Parse()
	mains.run()

	args := flag.Args()

	if len(args) == 0 {
		args = append(args, ".")
	}

	var question string
	var todo []*NameChange

	// Search directories which could be renamed.
	search := func(path string) {
		// Generate absolute path.
		dir, err := filepath.Abs(path)
		if err != nil {
			failure(fmt.Sprintf("%s", err))
			return
		}
		// Search NFO files.
		files, _ := filepath.Glob(filepath.Join(dir, "*.nfo"))
		for _, filename := range files {
			// Read a IMDB movie ID from the file.
			file, err := ioutil.ReadFile(filename)
			if err != nil {
				failure(fmt.Sprintf("%s", err))
				return
			}
			match := imdb_title_rx.FindSubmatch(file)
			if match != nil {
				id := string(match[1])
				url := imdb_title_url + id + "/"

				// Get HTML.
				resp, err := http.Get(url)
				if err != nil {
					failure(fmt.Sprintf("%s", err))
					return
				}
				defer resp.Body.Close()

				// Parse HTML.
				doc, err := html.Parse(resp.Body)
				if err != nil {
					failure(fmt.Sprintf("%s", err))
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
					title = sanitize(title)

					// Rename directory.
					question += "●\u00A0" + filepath.Dir(dir) + "\n" +
						"\u2003\u2003-\u00A0" + filepath.Base(dir) + "\n" +
						"\u2003\u2003-\u00A0" + title + "\n"

					// Store in the todo list.
					ren := NameChange{dir, filepath.Join(filepath.Dir(dir), title)}
					todo = append(todo, &ren)

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
		info(l15n[lang][can_not_find_any_movies])
	} else {

		// Rename directories after confirmation
		if ask(l15n[lang][rename_the_following_directories] + "\n\n" + question) {
			for _, ren := range todo {
				err := os.Rename(ren.old, ren.neu)
				if err != nil {
					failure(fmt.Sprintf("%s", err))
				}
			}
		}
	}
}
