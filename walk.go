package main

import (
	"golang.org/x/net/html"
)

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

// Local Variables:
// compile-command: "go build -ldflags -H=windowsgui"
// End:
