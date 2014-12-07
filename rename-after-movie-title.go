package main

import "fmt"
import "golang.org/x/net/html"
import "net/http"
import "unicode"

type action func(node *html.Node)
type predicate func(node *html.Node) (bool)

func element(name string) (predicate) {
	return func (node *html.Node) (bool) {
		return node.Type == html.ElementNode && node.Data == name
	}
}

func attribute(name, value string) (predicate) {
	return func (node *html.Node) (bool) {
		for _,attr := range node.Attr {
			if attr.Key == name && attr.Val == value {
				return true
			}
		}
		return false
	}
}

func text() (predicate) {
	return func (node *html.Node) (bool) {
		return node.Type == html.TextNode
	}
}

func and(a, b predicate) (predicate) {
	return func (node *html.Node) (bool) {
		return a(node) && b(node)
	}
}

func matcher(matches predicate, continuation action) (action) {
	return func (node *html.Node) {
		if matches(node) {
			continuation(node);
		}
	}
}

func walker(continuation action) (action) {
	var walk action
	walk = func (node *html.Node) {
		continuation(node)
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	return walk
}

func normalizer(dst *string) (func(string)) {
	beginning := true
	whitespace := true
	return func (src string) {
		for _, r := range src {
			if unicode.IsSpace(r) {
				if !whitespace {
					whitespace = true
				}
				continue
			} else {
				if whitespace {
					whitespace = false
					if beginning {
						beginning = false
					} else {
						*dst += " ";
					}
				}
				*dst += string(r);
			}
		}
	}
}

func main() {

	// Get HTML
	resp, err := http.Get("http://www.imdb.com/title/tt0093191/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// Collect title text
	title := ""
	append_title := func() (func(*html.Node)) {
		normalize := normalizer(&title)
		return func(node *html.Node) {
			normalize(node.Data)
		}
	}()

	// Find title
	find := walker(matcher(and(element("h1"), attribute("class", "header")),
		walker(matcher(text(), append_title))))
	find(doc);

	// Report title
	fmt.Println(title)

}

// Local Variables:
// compile-command: "go run rename-after-movie-title.go"
// End:
