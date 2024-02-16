package site

import (
	"io"
	"slices"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/xml"
)

var SUPPORTED_MINIFIERS = []string{".css", ".html", ".js", ".xml"}

type Minifier = minify.M

func (site *Site) loadMinifier() {
	site.minifier = *minify.New()
	site.minifier.AddFunc(".css", css.Minify)
	site.minifier.AddFunc(".html", html.Minify)
	site.minifier.AddFunc(".js", js.Minify)
	site.minifier.AddFunc(".xml", xml.Minify)
}

func (site *Site) minify(extension string, contentReader io.Reader) io.Reader {

	if !site.Config.Minify || !slices.Contains(SUPPORTED_MINIFIERS, extension) {
		return contentReader
	}
	return site.minifier.Reader(extension, contentReader)
}
