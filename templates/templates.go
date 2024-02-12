// TODO consider making this another package
package templates

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
	"gopkg.in/osteele/liquid.v1"
	"gopkg.in/yaml.v3"
)

const FM_SEPARATOR = "---"

type Type string

const (
	// a file that doesn't have a front matter header, and thus is not renderable.
	STATIC Type = "static"

	// Templates in the root /layouts/ can be used to wrap around other template's content
	// by setting the `layout` front matter field.
	LAYOUT Type = "layout"

	// A template that has a date, and thus can be ordered chronologically in a directory.
	// They can thus be arranged in archives, feeds, etc.
	// Posts are also assumed to have a title and can be excerpted.
	POST Type = "post"

	// The rest of the templates: no layout and no post
	PAGE Type = "page"
)

type Template struct {
	Type     Type
	SrcPath  string
	Metadata map[string]interface{}
}

// TODO think about knowledge boundaries
// should this know to tell if its a layout based on srcPath conventions?
// should it be able to detect its own type? does it still make sense to track a template type,
// separate from the site?
func Parse(path string) (*Template, error) {
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
		return &Template{Type: STATIC}, nil
	}

	// read and parse the yaml from the front matter
	var yamlContent []byte
	closed := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == FM_SEPARATOR {
			closed = true
			break
		}
		yamlContent = append(yamlContent, []byte(line+"\n")...)
	}
	if !closed {
		return nil, errors.New("front matter not closed")
	}

	var metadata map[string]interface{}
	if len(yamlContent) != 0 {
		err := yaml.Unmarshal([]byte(yamlContent), &metadata)
		if err != nil {
			return nil, fmt.Errorf("invalid yaml: %s", err)
		}
	}

	templ := Template{SrcPath: path, Metadata: metadata}

	// FIXME this also should check that it's in the root folder
	if strings.HasSuffix(filepath.Dir(templ.SrcPath), "layouts") {
		templ.Type = LAYOUT
	} else if _, ok := metadata["date"]; ok {
		templ.Type = POST
	} else {
		templ.Type = PAGE
	}

	return &templ, nil
}

func (templ Template) Ext() string {
	ext := filepath.Ext(templ.SrcPath)
	if ext == ".org" {
		ext = ".html"
	}
	return ext
}

func (templ Template) Render(context map[string]interface{}) (string, error) {
	file, _ := os.Open(templ.SrcPath)
	defer file.Close()
	scanner := bufio.NewScanner(file)

	// first line is the front matter delimiter, Scan to skip
	// and keep skipping until the closing delimiter
	scanner.Scan()
	scanner.Scan()
	for scanner.Text() != FM_SEPARATOR {
		scanner.Scan()
	}

	// now read the proper template contents to memory
	contents := ""
	for scanner.Scan() {
		contents += scanner.Text() + "\n"
	}

	if strings.HasSuffix(templ.SrcPath, ".org") {
		// if it's an org file, convert to html
		doc := org.New().Parse(strings.NewReader(contents), templ.SrcPath)
		return doc.Write(org.NewHTMLWriter())
	}

	// for other file types, assume a liquid template
	engine := liquid.NewEngine()
	return engine.ParseAndRenderString(contents, context)
}
