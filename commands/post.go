package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/facundoolano/jorge/config"
	"golang.org/x/text/unicode/norm"
)

var DEFAULT_FRONTMATTER string = `---
title: %s
date: %s
layout: post
lang: %s
tags: []
---
`

var DEFAULT_ORG_DIRECTIVES string = `#+OPTIONS: toc:nil num:nil
#+LANGUAGE: %s
`

type Post struct {
	Title string `arg:"" optional:""`
}

// Create a new post template in the given site, with the given title,
// with pre-filled front matter.
func (cmd *Post) Run(ctx *kong.Context) error {
	title := cmd.Title
	if title == "" {
		title = Prompt("title")
	}
	config, err := config.Load(".")
	if err != nil {
		return err
	}
	now := time.Now()
	slug := slugify(title)
	filename := strings.ReplaceAll(config.PostFormat, ":title", slug)

	filename = strings.ReplaceAll(filename, ":year", fmt.Sprintf("%d", now.Year()))
	filename = strings.ReplaceAll(filename, ":month", fmt.Sprintf("%02d", now.Month()))
	filename = strings.ReplaceAll(filename, ":day", fmt.Sprintf("%02d", now.Day()))
	path := filepath.Join(config.SrcDir, filename)

	// ensure the dir already exists
	if err := os.MkdirAll(filepath.Dir(path), FILE_RW_MODE); err != nil {
		return err
	}

	// if file already exists, prompt user for a different one
	if _, err := os.Stat(path); os.IsExist(err) {
		fmt.Printf("%s already exists\n", path)
		filename = Prompt("filename")
		path = filepath.Join(config.SrcDir, filename)
	}

	// initialize the post front matter
	content := fmt.Sprintf(DEFAULT_FRONTMATTER, title, now.Format(time.DateTime), config.Lang)

	// org files need some extra boilerplate
	if filepath.Ext(path) == ".org" {
		content += fmt.Sprintf(DEFAULT_ORG_DIRECTIVES, config.Lang)
	}

	if err := os.WriteFile(path, []byte(content), FILE_RW_MODE); err != nil {
		return err
	}
	fmt.Println("added", path)
	return nil
}

var nonWordRegex = regexp.MustCompile(`[^\w-]`)
var whitespaceRegex = regexp.MustCompile(`\s+`)

func slugify(title string) string {
	slug := strings.ToLower(title)
	slug = strings.TrimSpace(slug)
	slug = norm.NFD.String(slug)
	slug = whitespaceRegex.ReplaceAllString(slug, "-")
	slug = nonWordRegex.ReplaceAllString(slug, "")

	return slug
}
