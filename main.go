package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
func build() {
	// fail if no src dir
	_, err := os.ReadDir("src")
	if os.IsNotExist(err) {
		exit("missing src/ directory")
	} else if err != nil {
		panic("couldn't read src")
	}

	// clear previous target contents
	const FILE_MODE = 0777
	os.RemoveAll("target")
	os.Mkdir("target", FILE_MODE)

	// render each source file and copy it over to target
	filepath.WalkDir("src", func(path string, entry fs.DirEntry, err error) error {
		subpath, _ := filepath.Rel("src", path)
		targetSubpath := filepath.Join("target", subpath)

		if entry.IsDir() {
			os.MkdirAll(targetSubpath, FILE_MODE)
		} else {
			// read file contents
			data, err := os.ReadFile(path)
			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetSubpath))
			}

			// TODO render templates and org
			// TODO minify

			// write the file contents over to target at the same location
			err = os.WriteFile(targetSubpath, data, FILE_MODE)
			if err != nil {
				panic(fmt.Sprintf("failed to load %s", targetSubpath))
			}
			fmt.Printf("wrote %v", targetSubpath)
		}

		return nil
	})
}

func exit(message string) {
	fmt.Println(message)
	os.Exit(1)
}
