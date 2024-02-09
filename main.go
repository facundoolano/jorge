package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
	"gopkg.in/yaml.v3"
)

func main() {
	// TODO consider using cobra or something else to make cli more declarative
	// and get a better ux out of the box
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)

	if len(os.Args) < 2 {
		// TODO print usage
		exit("expected subcommand")
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		// get working directory
		// default to .
		// if not exist, create directory
		// copy over default files
		fmt.Println("not implemented yet")
	case "build":
		build()
	case "new":
		newCmd.Parse(os.Args[2:])
		// prompt for title
		// slugify
		// fail if file already exist
		// create a new .org file with the slug
		// add front matter and org options
		fmt.Println("not implemented yet")
	case "serve":
		// build
		// serve target with file server
		// (later watch and live reload)
		serveCmd.Parse(os.Args[2:])
		fmt.Println("not implemented yet")
	default:
		// TODO print usage
		exit("unknown subcommand")
	}

}

// Read the files in src/ render them and copy the result to target/
// TODO move elsewhere ?
func build() {
	const FILE_MODE = 0777

	// fail if no src dir
	_, err := os.ReadDir("src")
	if os.IsNotExist(err) {
		exit("missing src/ directory")
	} else if err != nil {
		panic("couldn't read src")
	}

	// clear previous target contents
	os.RemoveAll("target")
	os.Mkdir("target", FILE_MODE)

	// render each source file and copy it over to target
	filepath.WalkDir("src", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			subpath, _ := filepath.Rel("src", path)
			targetSubpath := filepath.Join("target", subpath)
			os.MkdirAll(targetSubpath, FILE_MODE)
		} else {

			// FIXME what if non text file?
			data, targetPath, err := render(path)

			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetPath))
			}

			// write the file contents over to target at the same location
			err = os.WriteFile(targetPath, []byte(data), FILE_MODE)
			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetPath))
			}
			fmt.Printf("wrote %v\n", targetPath)
		}

		return nil
	})

}

// TODO move elsewhere?
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

func exit(message string) {
	fmt.Println(message)
	os.Exit(1)
}
