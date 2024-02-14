package site

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/facundoolano/blorg/templates"
)

// TODO review build and other commands and think what can be brought over here.
// e.g. SRC and TARGET dir knowledge
type Site struct {
	config  map[string]string // may need to make this interface{} if config gets sophisticated
	layouts map[string]templates.Template
	posts   []map[string]interface{}
	pages   []map[string]interface{}
	tags    map[string][]map[string]interface{}

	templates map[string]*templates.Template
}

func Load(srcDir string, layoutsDir string) (*Site, error) {
	// TODO load config from config.yml
	site := Site{
		layouts:   make(map[string]templates.Template),
		templates: make(map[string]*templates.Template),
		config:    make(map[string]string),
		tags:      make(map[string][]map[string]interface{}),
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

	err = filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			templ, err := templates.Parse(path)
			// if sometime fails or this is not a template, skip
			if err != nil || templ == nil {
				return err
			}

			// set site related (?) metadata. Not sure if this should go elsewhere
			relPath, _ := filepath.Rel(srcDir, path)
			templ.Metadata["path"] = relPath
			templ.Metadata["url"] = "/" + strings.TrimSuffix(relPath, ".html")
			templ.Metadata["dir"] = "/" + filepath.Dir(relPath)

			// posts are templates that can be chronologically sorted --that have a date.
			// the rest are pages.
			if _, ok := templ.Metadata["date"]; ok {
				site.posts = append(site.posts, templ.Metadata)

				// also add to tags index
				if tags, ok := templ.Metadata["tags"]; ok {
					for _, tag := range tags.([]interface{}) {
						tag := tag.(string)
						site.tags[tag] = append(site.tags[tag], templ.Metadata)
					}
				}

			} else {
				// the index pages should be skipped from the page directory
				filename := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				if filename != "index" {
					site.pages = append(site.pages, templ.Metadata)
				}
			}
			site.templates[path] = templ
		}
		return nil
	})

	if err != nil {
		return err
	}

	// sort posts by reverse chronological order
	Compare := func(a map[string]interface{}, b map[string]interface{}) int {
		return b["date"].(time.Time).Compare(a["date"].(time.Time))
	}
	slices.SortFunc(site.posts, Compare)
	for _, posts := range site.tags {
		slices.SortFunc(posts, Compare)
	}
	return nil
}

func (site Site) RenderTemplate(path string) (io.Reader, bool, error) {
	templ, ok := site.templates[path]
	// if no known template at that location, return nil
	if !ok {
		return nil, false, nil
	}

	ctx := map[string]interface{}{
		"site": map[string]interface{}{
			"config": site.config,
			"posts":  site.posts,
			"tags":   site.tags,
			"pages":  site.pages,
		},
	}

	ctx["page"] = templ.Metadata
	content, err := templ.Render(ctx)
	if err != nil {
		return nil, true, err
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
			return nil, true, fmt.Errorf("layout '%s' not found", layout)
		}
	}

	return bytes.NewReader(content), true, nil
}
