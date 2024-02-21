package commands

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"embed"

	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/site"
)

//go:embed all:initfiles
var initfiles embed.FS
var initConfig string = `name: "%s"
author: "%s"
url: "%s"
`
var initReadme string = `
# %s

A jorge blog by %s.
`

const FILE_RW_MODE = 0777

func Init(projectDir string) error {
	if err := ensureEmptyProjectDir(projectDir); err != nil {
		return err
	}

	siteName := prompt("site name")
	siteUrl := prompt("site url")
	siteAuthor := prompt("author")

	// creating config and readme files manually, since I want to use the supplied config values in their
	// contents. (I don't want to render liquid templates in the WalkDir below since some of the initfiles
	// are actual templates that should be left as is).
	configFile := fmt.Sprintf(initConfig, siteName, siteAuthor, siteUrl)
	readmeFile := fmt.Sprintf(initReadme, siteName, siteAuthor)
	os.WriteFile(filepath.Join(projectDir, "config.yml"), []byte(configFile), site.FILE_RW_MODE)
	os.WriteFile(filepath.Join(projectDir, "README.md"), []byte(readmeFile), site.FILE_RW_MODE)

	// walk over initfiles fs
	// copy create directories and copy files at target

	initfilesRoot := "initfiles"
	return fs.WalkDir(initfiles, initfilesRoot, func(path string, entry fs.DirEntry, err error) error {
		if path == initfilesRoot {
			return nil
		}
		subpath, _ := filepath.Rel(initfilesRoot, path)
		targetPath := filepath.Join(projectDir, subpath)

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

// Prompt the user for a string value
func prompt(label string) string {
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

func Post() error {
	// prompt for title
	// slugify
	// fail if file already exist
	// create a new .org file with the slug
	// add front matter and org options
	fmt.Println("not implemented yet")
	return nil
}

// Read the files in src/ render them and copy the result to target/
func Build(root string) error {
	config, err := config.Load(root)
	if err != nil {
		return err
	}

	site, err := site.Load(*config)
	if err != nil {
		return err
	}

	return site.Build()
}
