package markup

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/niklasfasching/go-org/org"
	"github.com/osteele/liquid"
	"github.com/yuin/goldmark"
	gm_highlight "github.com/yuin/goldmark-highlighting/v2"
	"gopkg.in/yaml.v3"
)

const FM_SEPARATOR = "---"

type Engine = liquid.Engine

type Template struct {
	SrcPath        string
	Metadata       map[string]interface{}
	liquidTemplate liquid.Template
}

// Create a new template engine, with custom liquid filters.
// The `siteUrl` is necessary to provide context for the absolute_url filter.
func NewEngine(siteUrl string, includesDir string) *Engine {
	e := liquid.NewEngine()
	loadJekyllFilters(e, siteUrl, includesDir)
	return e
}

// Try to parse a liquid template at the given location.
// Files starting with front matter (--- sorrrounded yaml)
// are considered templates. If the given file is not headed by front matter
// return (nil, nil).
// The front matter contents are stored in the returned template's Metadata.
func Parse(engine *Engine, path string) (*Template, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	scanner.Scan()
	line := scanner.Text()

	// if the file doesn't start with a front matter delimiter, it's not a template
	if strings.TrimSpace(line) != FM_SEPARATOR {
		return nil, nil
	}

	// extract the yaml front matter and save the rest of the template content separately
	var yamlContent []byte
	var liquidContent []byte
	yamlClosed := false
	for scanner.Scan() {
		line := append(scanner.Bytes(), '\n')
		if yamlClosed {
			liquidContent = append(liquidContent, line...)
		} else {
			if strings.TrimSpace(scanner.Text()) == FM_SEPARATOR {
				yamlClosed = true
				continue
			}
			yamlContent = append(yamlContent, line...)
		}
	}
	liquidContent = bytes.TrimSuffix(liquidContent, []byte("\n"))

	if !yamlClosed {
		return nil, errors.New("front matter not closed")
	}

	metadata := make(map[string]interface{})
	if len(yamlContent) != 0 {
		err := yaml.Unmarshal([]byte(yamlContent), &metadata)
		if err != nil {
			return nil, fmt.Errorf("invalid yaml: %s", err)
		}
	}

	liquid, err := engine.ParseTemplateAndCache(liquidContent, path, 0)
	if err != nil {
		return nil, err
	}

	templ := Template{SrcPath: path, Metadata: metadata, liquidTemplate: *liquid}
	return &templ, nil
}

// Return the extension of this template's source file.
func (templ Template) SrcExt() string {
	return filepath.Ext(templ.SrcPath)
}

// Return the extension for the output format of this template
func (templ Template) TargetExt() string {
	ext := filepath.Ext(templ.SrcPath)
	if ext == ".org" || ext == ".md" {
		return ".html"
	}
	return ext
}

func (templ Template) IsDraft() bool {
	if draft, ok := templ.Metadata["draft"]; ok {
		return draft.(bool)
	}
	return false
}

func (templ Template) IsPost() bool {
	_, ok := templ.Metadata["date"]
	return ok
}

// Renders the liquid template with the given context as bindings.
// If the template source is org or md, convert them to html after the
// liquid rendering.
func (templ Template) Render(context map[string]interface{}, hlTheme string) ([]byte, error) {
	// liquid rendering
	content, err := templ.liquidTemplate.Render(context)
	if err != nil {
		return nil, err
	}

	if templ.SrcExt() == ".org" {
		// org-mode rendering
		doc := org.New().Parse(bytes.NewReader(content), templ.SrcPath)
		htmlWriter := org.NewHTMLWriter()

		// make * -> h1, ** -> h2, etc
		htmlWriter.TopLevelHLevel = 1
		htmlWriter.HighlightCodeBlock = highlightCodeBlock(hlTheme)

		contentStr, err := doc.Write(htmlWriter)
		if err != nil {
			return nil, err
		}
		content = []byte(contentStr)
	} else if templ.SrcExt() == ".md" {
		// markdown rendering
		var buf bytes.Buffer
		md := goldmark.New(goldmark.WithExtensions(
			gm_highlight.NewHighlighting(
				gm_highlight.WithStyle(hlTheme),
			),
		))
		if err := md.Convert(content, &buf); err != nil {
			return nil, err
		}
		content = buf.Bytes()
	}

	return content, nil
}

func highlightCodeBlock(hlTheme string) func(source string, lang string, inline bool, params map[string]string) string {
	// from https://github.com/niklasfasching/go-org/blob/a32df1461eb34a451b1e0dab71bd9b2558ea5dc4/blorg/util.go#L58
	return func(source, lang string, inline bool, params map[string]string) string {
		var w strings.Builder
		l := lexers.Get(lang)
		if l == nil {
			l = lexers.Fallback
		}
		l = chroma.Coalesce(l)
		it, _ := l.Tokenise(nil, source)
		options := []html.Option{}
		if params[":hl_lines"] != "" {
			ranges := org.ParseRanges(params[":hl_lines"])
			if ranges != nil {
				options = append(options, html.HighlightLines(ranges))
			}
		}
		_ = html.New(options...).Format(&w, styles.Get(hlTheme), it)
		if inline {
			return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
		}
		return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
	}
}
