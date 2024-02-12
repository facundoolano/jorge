package commands

import (
	"fmt"

	"github.com/facundoolano/blorg/templates"
)

// TODO review build and other commands and think what can be brought over here.
// e.g. SRC and TARGET dir knowledge
type Site struct {
	config  map[string]string // may need to make this interface{} if config gets sophisticated
	layouts map[string]templates.Template
	posts   []templates.Template
	pages   []templates.Template
	tags    map[string]*templates.Template

	renderCache map[string]string
}

func (site Site) render(templ *templates.Template) (string, error) {
	ctx := site.contextFor(templ)
	content, err := templ.Render(ctx)
	if err != nil {
		return "", err
	}

	// recursively render parent layouts
	layout := templ.Metadata["layout"]
	for layout != nil && err == nil {
		if layout_templ, ok := site.layouts[layout.(string)]; ok {
			ctx["content"] = content
			content, err = layout_templ.Render(ctx)
			layout = layout_templ.Metadata["layout"]
		} else {
			return "", fmt.Errorf("layout '%s' not found", layout)
		}
	}

	return content, err
}

func (site Site) templateIndex() map[string]*templates.Template {
	templIndex := make(map[string]*templates.Template)
	for _, templ := range append(site.posts, site.pages...) {
		templIndex[templ.SrcPath] = &templ
	}
	return templIndex
}

func (site Site) contextFor(templ *templates.Template) map[string]interface{} {
	bindings := map[string]interface{}{
		"config": site.config,
		"posts":  site.posts,
		"tags":   site.tags,
	}
	if templ.Type == templates.LAYOUT {
		bindings["layout"] = templ.Metadata
	} else {
		// assuming that if it's not a layout then it must be a page
		bindings["page"] = templ.Metadata
	}
	return bindings
}
