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
		subpath, _ := filepath.Rel("src", path)
		targetSubpath := filepath.Join("target", subpath)

		if entry.IsDir() {
			os.MkdirAll(targetSubpath, FILE_MODE)
		} else {

			data, err := render(path)

			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetSubpath))
			}

			// write the file contents over to target at the same location
			err = os.WriteFile(targetSubpath, data, FILE_MODE)
			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetSubpath))
			}
			fmt.Printf("wrote %v\n", targetSubpath)
		}

		return nil
	})

}

// TODO move elsewhere?
func render(path string) ([]byte, error) {
	file, err := os.Open(path)
	defer file.Close()

	fileContent, frontMatter, err := extractFrontMatter(file)
	if err != nil {
		exit(fmt.Sprintf("error in %s: %s", path, err))
	}

	if len(frontMatter) > 0 {
		fmt.Println("Detected front matter:", frontMatter)
	}

	if filepath.Ext(path) == ".org" {
		// TODO produce html from org
	} else {
		// TODO render liquid template
	}

	// TODO minify

	return fileContent, err
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
			return nil, nil, errors.New(fmt.Sprint("invalid yaml: ", err))
		}
	}

	return outContent, frontMatter, nil
}

func exit(message string) {
	fmt.Println(message)
	os.Exit(1)
}
