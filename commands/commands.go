package commands

import (
	"fmt"
	"github.com/facundoolano/blorg/site"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const SRC_DIR = "src"
const TARGET_DIR = "target"
const LAYOUTS_DIR = "layouts"
const FILE_RW_MODE = 0777

func Init() error {
	// get working directory
	// default to .
	// if not exist, create directory
	// copy over default files
	fmt.Println("not implemented yet")
	return nil
}

// Read the files in src/ render them and copy the result to target/
// TODO add root dir override support
func Build() error {
	site, err := site.Load(SRC_DIR, LAYOUTS_DIR)
	if err != nil {
		return err
	}

	// clear previous target contents
	os.RemoveAll(TARGET_DIR)
	os.Mkdir(TARGET_DIR, FILE_RW_MODE)

	// walk the source directory, creating directories and files at the target dir
	return filepath.WalkDir(SRC_DIR, func(path string, entry fs.DirEntry, err error) error {
		subpath, _ := filepath.Rel(SRC_DIR, path)
		targetPath := filepath.Join(TARGET_DIR, subpath)

		if entry.IsDir() {
			os.MkdirAll(targetPath, FILE_RW_MODE)
		} else {

			if templ, ok := site.TemplateIndex[path]; ok {
				// if a template was found at source, render it
				content, err := site.Render(templ)
				if err != nil {
					return err
				}

				// write the file contents over to target at the same location
				targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + templ.Ext()
				fmt.Println("writing", targetPath)
				return os.WriteFile(targetPath, []byte(content), FILE_RW_MODE)
			} else {
				// if a non template was found, copy file as is
				fmt.Println("writing", targetPath)
				return copyFile(path, targetPath)
			}
		}

		return nil
	})
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
