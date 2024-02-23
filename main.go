package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/commands"
)

var cli struct {
	Init struct {
		Path string `arg:"" help:"directory where to initialize the website project."`
	} `cmd:"" help:"Initialize a new website project."`
	Build struct {
		Path string `arg:"" optional:"" default:"." help:"path to the website project to build."`
	} `cmd:"" help:"Build a website project."`
	Post struct {
		Title string `arg:"" optional:""`
	} `cmd:"" help:"Initialize a new post template file." help:"title of the new post."`
	Serve struct {
		Path string `arg:"" optional:"" default:"." help:"path to the website project to serve."`
	} `cmd:"" help:"Run a local server for the website."`
}

func main() {
	err := run()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := kong.Parse(&cli, kong.UsageOnError())
	switch ctx.Command() {
	case "init <path>":
		rootDir := ctx.Args[0]
		return commands.Init(rootDir)
	case "build":
		rootDir := ctx.Args[0]
		return commands.Build(rootDir)
	case "post":
		var title string
		if len(ctx.Args) > 0 {
			title = ctx.Args[0]
		} else {
			title = commands.Prompt("title")
		}
		rootDir := "."
		return commands.Post(rootDir, title)
	case "serve":
		rootDir := ctx.Args[0]
		return commands.Serve(rootDir)
	default:
		return nil
	}
}
