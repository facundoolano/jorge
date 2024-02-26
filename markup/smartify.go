package markup

// Implements a naive version of smart quote replacement, see https://daringfireball.net/projects/smartypants/
// The quote replacement code was adapted from gojekyll's smartify filter to work on entire HTML documents
// https://github.com/osteele/gojekyll/blob/f1794a874890bfb601cae767a0cce15d672e9058/filters/smartify.go
// MIT License: https://github.com/osteele/gojekyll/blob/f1794a874890bfb601cae767a0cce15d672e9058/LICENSE

import (
	"bytes"
	"io"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var SKIP_TAGS = []string{"pre", "code", "kbd", "script", "math"}

func Smartify(extension string, contentReader io.Reader) (io.Reader, error) {
	if extension != ".html" {
		return contentReader, nil
	}
	node, err := html.Parse(contentReader)
	if err != nil {
		return nil, err
	}

	smartifyHTMLNode(node)
	var buf bytes.Buffer
	html.Render(&buf, node)

	return &buf, nil
}

func smartifyHTMLNode(node *html.Node) {
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && slices.Contains(SKIP_TAGS, node.Data) {
			continue
		}
		if node.Type == html.TextNode {
			node.Data = smartifyString(node.Data)
		}
		smartifyHTMLNode(node)
	}
}

var smartifyTransforms = []struct {
	match *regexp.Regexp
	repl  string
}{
	{regexp.MustCompile("(^|[^[:alnum:]])``(.+?)''"), "$1“$2”"},
	{regexp.MustCompile(`(^|[^[:alnum:]])'`), "$1‘"},
	{regexp.MustCompile(`'`), "’"},
	{regexp.MustCompile(`(^|[^[:alnum:]])"`), "$1“"},
	{regexp.MustCompile(`"($|[^[:alnum:]])`), "”$1"},
	{regexp.MustCompile(`(^|\s)--($|\s)`), "$1–$2"},
	{regexp.MustCompile(`(^|\s)---($|\s)`), "$1—$2"},
}

var smartifyReplacer *strings.Replacer
var smartifyReplaceSpans = map[string]string{}

func init() {
	smartifyReplacer = strings.NewReplacer(
		"...", "…",
		"(c)", "©",
		"(r)", "®",
		"(tm)", "™",
	)
}

func smartifyString(s string) string {
	for _, rule := range smartifyTransforms {
		s = rule.match.ReplaceAllString(s, rule.repl)
	}
	return smartifyReplacer.Replace(s)
}
