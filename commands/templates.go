package commands

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

// TODO move elsewhere?
// TODO doc
func render(sourcePath string) (string, string, error) {
	// FIXME remove src target knowledge
	subpath, _ := filepath.Rel("src", sourcePath)
	targetPath := filepath.Join("target", subpath)
	isOrgFile := filepath.Ext(sourcePath) == ".org"
	if isOrgFile {
		targetPath = strings.TrimSuffix(targetPath, "org") + "html"
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	fileContent, _, err := extractFrontMatter(file)
	if err != nil {
		return "", "", (fmt.Errorf("error in %s: %s", sourcePath, err))
	}

	var html string
	// FIXME this should be renamed to .html
	// (in general, the render process should be able to instruct a differnt target path)
	if isOrgFile {
		doc := org.New().Parse(bytes.NewReader(fileContent), sourcePath)
		html, err = doc.Write(org.NewHTMLWriter())
		if err != nil {
			return "", "", err
		}

	} else {
		// TODO render liquid template
	}

	// TODO if yaml contains layout, pass to parent

	// TODO minify

	return html, targetPath, nil
}

func extractFrontMatter(file *os.File) ([]byte, map[string]interface{}, error) {
	const FM_SEPARATOR = "---"

	var outContent, yamlContent []byte

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// if line starts front matter, write lines to yaml until front matter is closed
		if strings.TrimSpace(line) == FM_SEPARATOR {
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
				return nil, nil, errors.New("front matter not closed")
			}
		} else {
			// non front matter/yaml content goes to the output slice
			outContent = append(outContent, []byte(line+"\n")...)
		}
	}
	// drop the extraneous last line break
	outContent = bytes.TrimRight(outContent, "\n")

	var frontMatter map[string]interface{}
	if len(yamlContent) != 0 {
		err := yaml.Unmarshal([]byte(yamlContent), &frontMatter)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid yaml: ", err)
		}
	}

	return outContent, frontMatter, nil
}
