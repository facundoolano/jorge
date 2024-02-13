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

type Template struct {
	SrcPath  string
	Metadata map[string]interface{}
}

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
		return nil, nil
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
	isFirstLine := true
	for scanner.Scan() {
		if isFirstLine {
			isFirstLine = false
			contents = scanner.Text()
		} else {
			contents += "\n" + scanner.Text()
		}
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
