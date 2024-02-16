package commands

import (
	"fmt"

	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/site"
)

func Init(rootDir string) error {
	// os.MkDir
	//   if already exist, check if empty
	//   https://stackoverflow.com/a/30708914/993769
	//   if not empty fail

	// prompt site name
	// prompt author
	// prompt url
	// build context with supplied answers

	// walk over initfiles dir
	// if directory: create at target
	// if file: read, render with context, write at target

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
	config, err := config.Load(root)
	if err != nil {
		return err
	}

	site, err := site.Load(*config)
	if err != nil {
		return err
	}

	return site.Build()
}
