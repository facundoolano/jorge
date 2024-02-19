package site

import (
	"bytes"
	"io"

	"golang.org/x/net/html"
)

func ExtractFirstParagraph(doc io.Reader) string {
	html, err := html.Parse(doc)
	if err != nil {
		return ""
	}

	ptag := FindFirstElement(html, "p")
	if ptag == nil {
		return ""
	}
	return getTextContent(ptag)
}

// InjectScriptIntoHTML injects a <script> tag with the given JavaScript code into the HTML document
// provided as an io.Reader. It returns the modified HTML content as an io.Reader.
func InjectScript(htmlReader io.Reader, jsCode string) (io.Reader, error) {
	doc, err := html.Parse(htmlReader)
	if err != nil {
		return nil, err
	}

	scriptNode := &html.Node{
		Type: html.ElementNode,
		Data: "script",
		Attr: []html.Attribute{
			{Key: "type", Val: "text/javascript"},
		},
	}

	// insert the script code inside the script tag
	scriptTextNode := &html.Node{
		Type: html.TextNode,
		Data: jsCode,
	}
	scriptNode.AppendChild(scriptTextNode)

	head := FindFirstElement(doc, "head")
	if head == nil {
		// If <head> element not found, create one and append it to the document
		head = &html.Node{
			Type: html.ElementNode,
			Data: "head",
		}
		doc.InsertBefore(head, doc.FirstChild)
	}

	// Append the <script> element to the <head> element
	head.AppendChild(scriptNode)

	// Serialize the modified HTML document to a buffer
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	// Return a reader for the modified HTML content
	return &buf, nil
}

// FindFirstElement finds the first occurrence of the specified HTML element in the HTML document
func FindFirstElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if element := FindFirstElement(c, tagName); element != nil {
			return element
		}
	}
	return nil
}

// findHead finds the <head> element in the HTML document
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
