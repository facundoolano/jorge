package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/site"
)

const FILE_RW_MODE = 0666
const DIR_RWE_MODE = 0777

type Build struct {
	ProjectDir string `arg:"" name:"path" optional:"" default:"." help:"Path to the website project to build."`
}

// Read the files in src/ render them and copy the result to target/
func (cmd *Build) Run(ctx *kong.Context) error {
	config, err := config.Load(cmd.ProjectDir)
	if err != nil {
		return err
	}

	return site.Build(*config)
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
