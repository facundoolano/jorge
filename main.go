package main

import (
	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/commands"
)

var cli struct {
	Init    commands.Init    `cmd:"" help:"Initialize a new website project." aliases:"i"`
	Build   commands.Build   `cmd:"" help:"Build a website project." aliases:"b"`
	Post    commands.Post    `cmd:"" help:"Initialize a new post template file." aliases:"p"`
	Serve   commands.Serve   `cmd:"" help:"Run a local server for the website." aliases:"s"`
	Meta    commands.Meta    `cmd:"" help:"Get the JSON results from evaluating a liquid template expression within the site context." aliases:"m"`
	Version kong.VersionFlag `short:"v"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.UsageOnError(),
		kong.HelpOptions{FlagsLast: true},
		kong.Vars{"version": "jorge v0.10.0"},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
