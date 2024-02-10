package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
