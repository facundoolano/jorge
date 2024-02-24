package commands

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/site"
)

//go:embed all:initfiles
var initfiles embed.FS

var INIT_CONFIG string = `name: "%s"
author: "%s"
url: "%s"
`
var INIT_README string = `
# %s

A jorge blog by %s.
`

type Init struct {
	ProjectDir string `arg:"" name:"path" help:"Directory where to initialize the website project."`
}

// Initialize a new jorge project in the given directory,
// prompting for basic site config and creating default files.
func (cmd *Init) Run(ctx *kong.Context) error {
	if err := ensureEmptyProjectDir(cmd.ProjectDir); err != nil {
		return err
	}

	siteName := Prompt("site name")
	siteUrl := Prompt("site url")
	siteAuthor := Prompt("author")
	fmt.Println()

	// creating config and readme files manually, since I want to use the supplied config values in their
	// contents. (I don't want to render liquid templates in the WalkDir below since some of the initfiles
	// are actual templates that should be left as is).
	configPath := filepath.Join(cmd.ProjectDir, "config.yml")
	configFile := fmt.Sprintf(INIT_CONFIG, siteName, siteAuthor, siteUrl)
	os.WriteFile(configPath, []byte(configFile), site.FILE_RW_MODE)
	fmt.Println("added", configPath)

	readmePath := filepath.Join(cmd.ProjectDir, "README.md")
	readmeFile := fmt.Sprintf(INIT_README, siteName, siteAuthor)
	os.WriteFile(readmePath, []byte(readmeFile), site.FILE_RW_MODE)
	fmt.Println("added", readmePath)

	// walk over initfiles fs
	// copy create directories and copy files at target

	initfilesRoot := "initfiles"
	return fs.WalkDir(initfiles, initfilesRoot, func(path string, entry fs.DirEntry, err error) error {
		if path == initfilesRoot {
			return nil
		}
		subpath, _ := filepath.Rel(initfilesRoot, path)
		targetPath := filepath.Join(cmd.ProjectDir, subpath)

		// if it's a directory create it at the same location
		if entry.IsDir() {
			return os.MkdirAll(targetPath, FILE_RW_MODE)
		}

		// TODO duplicated in site, extract to somewhere else
		// if its a file, copy it over
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		source, err := initfiles.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()

		_, err = io.Copy(targetFile, source)
		if err != nil {
			return err
		}
		fmt.Println("added", targetPath)
		return targetFile.Sync()
	})
}

func ensureEmptyProjectDir(projectDir string) error {
	if err := os.Mkdir(projectDir, 0777); err != nil {
		// if it fails with dir already exist, check if it's empty
		// https://stackoverflow.com/a/30708914/993769
		if os.IsExist(err) {
			// check if empty
			dir, err := os.Open(projectDir)
			if err != nil {
				return err
			}
			defer dir.Close()

			// if directory is non empty, fail
			_, err = dir.Readdirnames(1)
			if err == nil {
				return fmt.Errorf("non empty directory %s", projectDir)
			}
			return err
		}
	}
	return nil
}
