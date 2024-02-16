package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/facundoolano/jorge/commands"
)

func main() {
	err := run(os.Args)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// TODO consider using cobra or something else to make cli more declarative
	// and get a better ux out of the box
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)

	if len(os.Args) < 2 {
		// TODO print usage
		return errors.New("expected subcommand")
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		return commands.Init()
	case "build":
		rootDir := "."
		if len(os.Args) > 2 {
			rootDir = os.Args[2]
		}
		return commands.Build(rootDir)
	case "new":
		newCmd.Parse(os.Args[2:])
		return commands.New()
	case "serve":
		rootDir := "."
		if len(os.Args) > 2 {
			rootDir = os.Args[2]
		}
		return commands.Serve(rootDir)
	default:
		// TODO print usage
		return errors.New("unknown subcommand")
	}
}
