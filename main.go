package main

import (
	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/commands"
	"github.com/facundoolano/jorge/config"
)

// TODO move to commands/init
type Init struct {
	ProjectDir string `arg:"" name:"path" help:"directory where to initialize the website project."`
}

func (cmd *Init) Run(ctx *kong.Context) error {
	return commands.Init(cmd.ProjectDir)
}

type Build struct {
	ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"path to the website project to build."`
}

func (cmd *Build) Run(ctx *kong.Context) error {
	config, err := config.Load(cmd.ProjectDir)
	if err != nil {
		return err
	}
	return commands.Build(config)
}

type Post struct {
	Title string `arg:"" optional:""`
}

func (cmd *Post) Run(ctx *kong.Context) error {
	title := cmd.Title
	if title == "" {
		title = commands.Prompt("title")
	}
	config, err := config.Load(".")
	if err != nil {
		return err
	}
	return commands.Post(config, title)
}

type Serve struct {
	ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"path to the website project to serve."`
}

func (cmd *Serve) Run(ctx *kong.Context) error {
	// FIXME add flags
	config, err := config.LoadDev(cmd.ProjectDir, "localhost", 4001, true)
	if err != nil {
		return err
	}
	return commands.Serve(config)
}

var cli struct {
	Init    Init             `cmd:"" help:"Initialize a new website project." aliases:"i"`
	Build   Build            `cmd:"" help:"Build a website project." aliases:"b"`
	Post    Post             `cmd:"" help:"Initialize a new post template file." help:"title of the new post." aliases:"p"`
	Serve   Serve            `cmd:"" help:"Run a local server for the website." aliases:"s"`
	Version kong.VersionFlag `short:"v"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.UsageOnError(),
		kong.HelpOptions{FlagsLast: true},
		kong.Vars{"version": "jorge v.0.1.2"},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
