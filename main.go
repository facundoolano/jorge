package main

import (
	"flag"
	"fmt"
	"os"
)

// TODO consider using cobra or something else to make cli more declarative
// and get a better ux out of the box
func main() {

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)

	if len(os.Args) < 2 {
		printAndExit()
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
		buildCmd.Parse(os.Args[2:])
		// delete target if exist
		// create target dir
		// walk through files in src dir
		// copy them over to target
		// (later render templates and org)
		// (later minify)
		fmt.Println("not implemented yet")
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
		printAndExit()
	}
}

func printAndExit() {
	// TODO print usage
	fmt.Println("expected a subcommand")
	os.Exit(1)
}
