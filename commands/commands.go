package commands

import (
	"fmt"

	"github.com/facundoolano/blorg/site"
)

const SRC_DIR = "src"
const TARGET_DIR = "target"
const LAYOUTS_DIR = "layouts"

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
// TODO add root dir override support
func Build() error {
	site, err := site.Load(SRC_DIR, LAYOUTS_DIR)
	if err != nil {
		return err
	}

	return site.Build(SRC_DIR, TARGET_DIR, true, false)
}
