package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/commands"
	"github.com/facundoolano/jorge/config"
)

var cli struct {
	Init struct {
		ProjectDir string `arg:"" name:"path" help:"directory where to initialize the website project."`
	} `cmd:"" help:"Initialize a new website project." aliases:"i"`
	Build struct {
		ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"path to the website project to build."`
	} `cmd:"" help:"Build a website project." aliases:"b"`
	Post struct {
		Title string `arg:"" optional:""`
	} `cmd:"" help:"Initialize a new post template file." help:"title of the new post." aliases:"p"`
	Serve struct {
		ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"path to the website project to serve."`
	} `cmd:"" help:"Run a local server for the website." aliases:"s"`
	Version     struct{} `cmd:"" help:"print version information" aliases:"v"`
	VersionFlag bool     `name:"version" short:"v"`
}

func main() {
	err := run()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

// FIXME try to reduce duplication/boilerplate
func run() error {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.HelpOptions{FlagsLast: true})
	if cli.VersionFlag {
		printVersion()
		return nil
	}

	switch ctx.Command() {
	case "init <path>":
		return commands.Init(cli.Init.ProjectDir)
	case "build", "build <path>":
		config, err := config.Load(cli.Build.ProjectDir)
		if err != nil {
			return err
		}
		return commands.Build(config)
	case "post <title>":
		config, err := config.Load(".")
		if err != nil {
			return err
		}
		return commands.Post(config, cli.Post.Title)
	case "post":
		title := commands.Prompt("title")
		config, err := config.Load(".")
		if err != nil {
			return err
		}
		return commands.Post(config, title)
	case "serve", "serve <path>":
		// FIXME add flags
		config, err := config.LoadDev(cli.Serve.ProjectDir, "localhost", 4001, true)
		if err != nil {
			return err
		}
		return commands.Serve(config)
	case "version":
		printVersion()
		return nil
	default:
		return fmt.Errorf("unexpected input %s", ctx.Command())
	}
}

func printVersion() {
	// FIXME set this automatically when tagging in git
	fmt.Println("jorge v0.1.2")
}
