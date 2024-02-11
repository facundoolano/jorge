// TODO consider making this another package
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
	"gopkg.in/yaml.v3"
)

const FM_SEPARATOR = "---"

type Template struct {
	srcPath  string
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

	return &Template{srcPath: path, Metadata: metadata}, nil
}

func (templ Template) Ext() string {
	return filepath.Ext(templ.srcPath)
}

func (templ Template) Render() ([]byte, error) {
	file, _ := os.Open(templ.srcPath)
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
	var contents []byte
	for scanner.Scan() {
		contents = append(contents, scanner.Text()+"\n"...)
	}

	if templ.Ext() == ".org" {
		// if it's an org file, convert to html
		doc := org.New().Parse(bytes.NewReader(contents), templ.srcPath)
		html, err := doc.Write(org.NewHTMLWriter())
		contents = []byte(html)
		if err != nil {
			return nil, err
		}

	} else {
		// TODO for other file types, assume a liquid template
	}

	// TODO: if layout in metadata, pass the result to the rendered parent

	return contents, nil
}
