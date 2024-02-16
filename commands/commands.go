package commands

import (
	"fmt"

	"github.com/facundoolano/blorg/config"
	"github.com/facundoolano/blorg/site"
)

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
