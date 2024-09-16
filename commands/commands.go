package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/site"
)

const FILE_RW_MODE = 0666
const DIR_RWE_MODE = 0777

type Build struct {
	ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"Path to the website project to build."`
	NoMinify   bool   `help:"Disable file minifying."`
}

// Read the files in src/ render them and copy the result to target/
func (cmd *Build) Run(ctx *kong.Context) error {
	start := time.Now()

	config, err := config.Load(cmd.ProjectDir)
	if err != nil {
		return err
	}
	config.Minify = !cmd.NoMinify

	err = site.Build(*config)
	fmt.Printf("done in %.2fs\n", time.Since(start).Seconds())
	return err
}

// Prompt the user for a string value
func Prompt(label string) string {
	// https://dev.to/tidalcloud/interactive-cli-prompts-in-go-3bj9
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+": ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

type Meta struct {
	Expression string `arg:"" name:"expression" default:"site" help:"liquid expression to be evaluated (what goes inside of {{ ... }} in templates)"`
}

// Load the site metadata and use it as context to evaluate a liquid expression
func (cmd *Meta) Run(ctx *kong.Context) error {

	config, err := config.Load(".")
	if err != nil {
		return err
	}

	// remove optional {{}} wrapper
	expression := strings.Trim(cmd.Expression, " {}")

	result, err := site.EvalMetadata(*config, expression)
	if err == nil {
		fmt.Println(result)
	}
	return err
}
