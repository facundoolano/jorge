package markup

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

type Minifier struct {
	minifier *minify.M
}

func LoadMinifier() Minifier {
	minifier := minify.New()
	minifier.AddFunc(".css", css.Minify)
	minifier.AddFunc(".html", html.Minify)
	minifier.AddFunc(".js", js.Minify)
	minifier.AddFunc(".xml", xml.Minify)
	return Minifier{minifier}
}

// if enabled by config, minify web files
func (m *Minifier) Minify(extension string, contentReader io.Reader) io.Reader {

	if !slices.Contains(SUPPORTED_MINIFIERS, extension) {
		return contentReader
	}
	return m.minifier.Reader(extension, contentReader)
}
