package site

import (
	"fmt"
	"github.com/facundoolano/blorg/templates"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TODO review build and other commands and think what can be brought over here.
// e.g. SRC and TARGET dir knowledge
type Site struct {
	config  map[string]string // may need to make this interface{} if config gets sophisticated
	layouts map[string]templates.Template
	posts   []templates.Template
	pages   []templates.Template
	tags    map[string]*templates.Template

	TemplateIndex map[string]*templates.Template
}

func Load(srcDir string, layoutsDir string) (*Site, error) {
	// TODO load config from config.yml
	site := Site{
		layouts:       make(map[string]templates.Template),
		TemplateIndex: make(map[string]*templates.Template),
		config:        make(map[string]string),
	}

	if err := site.loadLayouts(layoutsDir); err != nil {
		return nil, err
	}

	if err := site.loadTemplates(srcDir); err != nil {
		return nil, err
	}

	return &site, nil
}

func (site *Site) loadLayouts(layoutsDir string) error {
	files, err := os.ReadDir(layoutsDir)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(layoutsDir, filename)
			templ, err := templates.Parse(path)
			if err != nil {
				return err
			}

			layout_name := strings.TrimSuffix(filename, filepath.Ext(filename))
			site.layouts[layout_name] = *templ
		}
	}

	return nil
}

func (site *Site) loadTemplates(srcDir string) error {
	_, err := os.ReadDir(srcDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("missing %s directory", srcDir)
	} else if err != nil {
		return fmt.Errorf("couldn't read %s", srcDir)
	}

	return filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			templ, err := templates.Parse(path)
			// if sometime fails or this is not a template, skip
			if err != nil || templ == nil {
				return err
			}

			// posts are templates that can be chronologically sorted --that have a date.
			// the rest are pages.
			if _, ok := templ.Metadata["date"]; ok {
				site.posts = append(site.posts, *templ)
			} else {
				site.pages = append(site.pages, *templ)
			}
			site.TemplateIndex[path] = templ

			// TODO load tags
		}
		return nil
	})
}

func (site Site) Render(templ *templates.Template) (string, error) {
	ctx := site.baseContext()
	ctx["page"] = templ.Metadata
	content, err := templ.Render(ctx)
	if err != nil {
		return "", err
	}

	// recursively render parent layouts
	layout := templ.Metadata["layout"]
	for layout != nil && err == nil {
		if layout_templ, ok := site.layouts[layout.(string)]; ok {
			ctx["layout"] = layout_templ.Metadata
			ctx["content"] = content
			content, err = layout_templ.Render(ctx)
			layout = layout_templ.Metadata["layout"]
		} else {
			return "", fmt.Errorf("layout '%s' not found", layout)
		}
	}

	return content, err
}

func (site Site) baseContext() map[string]interface{} {
	return map[string]interface{}{
		"config": site.config,
		"posts":  site.posts,
		"tags":   site.tags,
	}
}
