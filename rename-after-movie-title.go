package main

import "fmt"
import "golang.org/x/net/html"
import "net/http"

type action func(node *html.Node)
type predicate func(node *html.Node) (bool)

func print_node(node *html.Node) {
	fmt.Println(node)
}

func is_element(name string) (predicate) {
	return func (node *html.Node) (bool) {
		return node.Type == html.ElementNode && node.Data == name
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

	// Find title
	find := walker(matcher(is_element("h1"), print_node))
	find(doc);

//	fmt.Println(doc)
}
