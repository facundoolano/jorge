package markup

import (
	"bytes"
	"io"

	"golang.org/x/net/html"
)

// Find the first p tag in the given html document and return its text content.
func ExtractFirstParagraph(htmlReader io.Reader) string {
	html, err := html.Parse(htmlReader)
	if err != nil {
		return ""
	}

	ptag := findFirstElement(html, "p")
	if ptag == nil {
		return ""
	}
	return getTextContent(ptag)
}

// Inject a <script> tag with the given JavaScript code into provided the HTML document
// and return the updated document as a new io.Reader
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

	head := findFirstElement(doc, "head")
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

// Finds the first occurrence of the specified element in the HTML document
func findFirstElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if element := findFirstElement(c, tagName); element != nil {
			return element
		}
	}
	return nil
}

// Finds the <head> element in the HTML document
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
