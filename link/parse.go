package link

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Link represents a link (<a href="...">) in an HTML
// document.
type Link struct {
	Href string
	Text string
}

// Parse generates a slice of Link objects by parsing <a> element nodes from the given io.Reader.
func Parse(r io.Reader) ([]Link, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	nodes := linkNodes(doc)
	var links []Link
	for _, node := range nodes {
		links = append(links, buildLink(node))
	}
	return links, nil
}

// buildLink creates a Link struct from an HTML node. It set the Href to the value of the href attribute 
// and the Text to the concatenated text content of the node.
func buildLink(n *html.Node) Link {
	var ret Link
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			ret.Href = attr.Val
			break
		}
	}
	ret.Text = text(n)
	return ret
}

// text extracts and returns the concatenated text content from an HTML node and its descendants.
// It recursively traverses the node tree, collecting all text nodes, and removes extra whitespace.
func text(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type != html.ElementNode {
		return ""
	}
	var ret string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret += text(c)
	}

	return strings.Join(strings.Fields(ret), " ")
}

// linkNodes traverses an HTML node tree and returns a slice of all <a> element nodes found.
// It performs a recursive depth-first search to locate and collect all anchor tags.
//
// Notes:
//   - The function only considers nodes of type html.ElementNode with a tag name of "a".
//   - It recurses through all child nodes to ensure a thorough traversal of the tree.
func linkNodes(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "a" {
		return []*html.Node{n}
	}

	var ret []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret = append(ret, linkNodes(c)...)
	}
	return ret
}
