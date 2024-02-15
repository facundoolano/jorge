package templates

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
	"github.com/osteele/liquid"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

const FM_SEPARATOR = "---"

type Engine = liquid.Engine

type Template struct {
	SrcPath        string
	Metadata       map[string]interface{}
	liquidTemplate liquid.Template
}

func NewEngine() *Engine {
	e := liquid.NewEngine()
	loadJekyllFilters(e)
	return e
}

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

// Return the extension for the output format of this template
func (templ Template) Ext() string {
	ext := filepath.Ext(templ.SrcPath)
	if ext == ".org" || ext == ".md" {
		return ".html"
	}
	return ext
}

func (templ Template) Render(context map[string]interface{}) ([]byte, error) {
	content, err := templ.liquidTemplate.Render(context)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(templ.SrcPath)
	if ext == ".org" {
		doc := org.New().Parse(bytes.NewReader(content), templ.SrcPath)
		contentStr, err := doc.Write(org.NewHTMLWriter())
		if err != nil {
			return nil, err
		}
		content = []byte(contentStr)
	} else if ext == ".md" {
		var buf bytes.Buffer
		if err := goldmark.Convert(content, &buf); err != nil {
			return nil, err
		}
		content = buf.Bytes()
	}

	return content, nil
}
