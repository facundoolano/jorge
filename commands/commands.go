package commands

import (
	"fmt"
	"path/filepath"

	"github.com/facundoolano/blorg/site"
)

const SRC_DIR = "src"
const TARGET_DIR = "target"
const LAYOUTS_DIR = "layouts"
const INCLUDES_DIR = "includes"
const DATA_DIR = "data"

func Init() error {
	// get working directory
	// default to .
	// if not exist, create directory
	// copy over default files
	fmt.Println("not implemented yet")
	return nil
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

// Read the files in src/ render them and copy the result to target/
func Build(root string) error {
	src := filepath.Join(root, SRC_DIR)
	target := filepath.Join(root, TARGET_DIR)
	layouts := filepath.Join(root, LAYOUTS_DIR)
	data := filepath.Join(root, DATA_DIR)

	site, err := site.Load(src, layouts, data)
	if err != nil {
		return err
	}

	return site.Build(src, target, true, false)
}
