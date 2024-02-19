package site

import (
	"io"

	"golang.org/x/net/html"
)

func ExtractFirstParagraph(doc io.Reader) string {
	html, err := html.Parse(doc)
	if err != nil {
		return ""
	}

	ptag := findFirstParagraph(html)
	if ptag == nil {
		return ""
	}
	return getTextContent(ptag)
}

func findFirstParagraph(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "p" {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if p := findFirstParagraph(c); p != nil {
			return p
		}
	}
	return nil
}

func getTextContent(node *html.Node) string {
	var textContent string
	if node.Type == html.TextNode {
		textContent = node.Data
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		textContent += getTextContent(c)
	}
	return textContent
}
