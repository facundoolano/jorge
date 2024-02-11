package commands

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/facundoolano/blorg/templates"
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
// FIXME pass src and target by arg
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
		subpath, _ := filepath.Rel("src", path)
		targetPath := filepath.Join("target", subpath)

		if entry.IsDir() {
			os.MkdirAll(targetPath, FILE_MODE)
		} else {
			template, err := templates.Parse(path)
			if err != nil {
				return err
			}

			if template != nil {
				// if a template was found at source, render it
				targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + template.Ext()

				content, err := template.Render()
				if err != nil {
					return err
				}

				// write the file contents over to target at the same location
				fmt.Println("writing ", targetPath)
				return os.WriteFile(targetPath, content, FILE_MODE)
			} else {
				// if a non template was found, copy file as is
				fmt.Println("writing ", targetPath)
				return copyFile(path, targetPath)
			}
		}

		return nil
	})

	return err
}

func copyFile(source string, target string) error {
	// does this need to be so verbose?
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	targetFile, _ := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, srcFile)
	if err != nil {
		return err
	}

	return targetFile.Sync()
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
