package commands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
	"gopkg.in/yaml.v3"
)

func Init() error {
	// get working directory
	// default to .
	// if not exist, create directory
	// copy over default files
	fmt.Println("not implemented yet")
	return nil
}

// Read the files in src/ render them and copy the result to target/
func Build() error {
	const FILE_MODE = 0777

	// fail if no src dir
	_, err := os.ReadDir("src")
	if os.IsNotExist(err) {
		return errors.New("missing src/ directory")
	} else if err != nil {
		return errors.New("couldn't read src")
	}

	// clear previous target contents
	os.RemoveAll("target")
	os.Mkdir("target", FILE_MODE)

	// render each source file and copy it over to target
	err = filepath.WalkDir("src", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			subpath, _ := filepath.Rel("src", path)
			targetSubpath := filepath.Join("target", subpath)
			os.MkdirAll(targetSubpath, FILE_MODE)
		} else {

			// FIXME what if non text file?
			data, targetPath, err := render(path)
			if err != nil {
				return fmt.Errorf("failed to render %s", path)
			}

			// write the file contents over to target at the same location
			err = os.WriteFile(targetPath, []byte(data), FILE_MODE)
			if err != nil {
				return fmt.Errorf("failed to load %s", targetPath)
			}
			fmt.Printf("wrote %v\n", targetPath)
		}

		return nil
	})

	return err
}

func New() error {
	// prompt for title
	// slugify
	// fail if file already exist
	// create a new .org file with the slug
	// add front matter and org options
	fmt.Println("not implemented yet")
	return nil
}

func Serve() error {
	// build
	// serve target with file server
	// (later watch and live reload)
	fmt.Println("not implemented yet")
	return nil
}

// TODO move elsewhere?
// TODO doc
func render(sourcePath string) (string, string, error) {
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
