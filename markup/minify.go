package markup

import (
	"io"
	"path/filepath"
	"slices"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/xml"
)

var SUPPORTED_MINIFIERS = []string{".css", ".html", ".js", ".xml"}

type Minifier struct {
	minifier   *minify.M
	exclusions []string
}

func LoadMinifier(exclusions []string) Minifier {
	minifier := minify.New()
	minifier.AddFunc(".css", css.Minify)
	minifier.AddFunc(".html", html.Minify)
	minifier.AddFunc(".js", js.Minify)
	minifier.AddFunc(".xml", xml.Minify)
	return Minifier{minifier, exclusions}
}

func (m *Minifier) Minify(path string, contentReader io.Reader) io.Reader {

	for _, exclusion := range m.exclusions {
		if matched, _ := filepath.Match(exclusion, path); matched {
			return contentReader
		}
	}

	extension := filepath.Ext(path)
	if !slices.Contains(SUPPORTED_MINIFIERS, extension) {
		return contentReader
	}
	return m.minifier.Reader(extension, contentReader)
}
