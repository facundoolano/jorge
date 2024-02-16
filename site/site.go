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

	"github.com/facundoolano/blorg/config"
	"github.com/facundoolano/blorg/templates"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

const FILE_RW_MODE = 0777

type Site struct {
	Config  config.Config
	layouts map[string]templates.Template
	posts   []map[string]interface{}
	pages   []map[string]interface{}
	tags    map[string][]map[string]interface{}
	data    map[string]interface{}

	templateEngine *templates.Engine
	templates      map[string]*templates.Template
}

func Load(config config.Config) (*Site, error) {
	site := Site{
		layouts:        make(map[string]templates.Template),
		templates:      make(map[string]*templates.Template),
		Config:         config,
		tags:           make(map[string][]map[string]interface{}),
		data:           make(map[string]interface{}),
		templateEngine: templates.NewEngine(config.SiteUrl),
	}

	if err := site.loadDataFiles(); err != nil {
		return nil, err
	}

	if err := site.loadLayouts(); err != nil {
		return nil, err
	}

	if err := site.loadTemplates(); err != nil {
		return nil, err
	}

	return &site, nil
}

func (site *Site) loadLayouts() error {
	files, err := os.ReadDir(site.Config.LayoutsDir)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(site.Config.LayoutsDir, filename)
			templ, err := templates.Parse(site.templateEngine, path)
			if err != nil {
				return err
			}

			layout_name := strings.TrimSuffix(filename, filepath.Ext(filename))
			site.layouts[layout_name] = *templ
		}
	}

	return nil
}

func (site *Site) loadDataFiles() error {
	files, err := os.ReadDir(site.Config.DataDir)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(site.Config.DataDir, filename)

			yamlContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var data interface{}
			err = yaml.Unmarshal(yamlContent, &data)
			if err != nil {
				return err
			}

			data_name := strings.TrimSuffix(filename, filepath.Ext(filename))
			site.data[data_name] = data
		}
	}

	return nil
}

func (site *Site) loadTemplates() error {
	_, err := os.ReadDir(site.Config.SrcDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("missing %s directory", site.Config.SrcDir)
	} else if err != nil {
		return fmt.Errorf("couldn't read %s", site.Config.SrcDir)
	}

	err = filepath.WalkDir(site.Config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			templ, err := templates.Parse(site.templateEngine, path)
			// if something fails or this is not a template, skip
			if err != nil || templ == nil {
				return err
			}

			// set site related (?) metadata. Not sure if this should go elsewhere
			relPath, _ := filepath.Rel(site.Config.SrcDir, path)
			relPath = strings.TrimSuffix(relPath, filepath.Ext(relPath)) + templ.Ext()
			templ.Metadata["path"] = relPath
			templ.Metadata["url"] = "/" + strings.TrimSuffix(strings.TrimSuffix(relPath, "index.html"), ".html")
			templ.Metadata["dir"] = "/" + filepath.Dir(relPath)

			// posts are templates that can be chronologically sorted --that have a date.
			// the rest are pages.
			if _, ok := templ.Metadata["date"]; ok {

				// NOTE: getting the excerpt if not set at the front matter requires rendering the template
				// which could be too onerous for this stage. Consider postponing setting this and/or caching the
				// template render result
				templ.Metadata["excerpt"] = getExcerpt(templ)
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

func (site *Site) Build() error {
	// clear previous target contents
	os.RemoveAll(site.Config.TargetDir)
	os.Mkdir(site.Config.SrcDir, FILE_RW_MODE)

	// walk the source directory, creating directories and files at the target dir
	return filepath.WalkDir(site.Config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		subpath, _ := filepath.Rel(site.Config.SrcDir, path)
		targetPath := filepath.Join(site.Config.TargetDir, subpath)

		// if it's a directory, just create the same at the target
		if entry.IsDir() {
			return os.MkdirAll(targetPath, FILE_RW_MODE)
		}

		var contentReader io.Reader
		templ, found := site.templates[path]
		if !found {
			// if no template found at location, treat the file as static write its contents to target
			if site.Config.LinkStatic {
				// dev optimization: link static files instead of copying them
				abs, _ := filepath.Abs(path)
				return os.Symlink(abs, targetPath)
			}

			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			contentReader = srcFile
		} else {
			content, err := site.render(templ)
			if err != nil {
				return err
			}

			targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + templ.Ext()
			contentReader = bytes.NewReader(content)
		}

		targetExt := filepath.Ext(targetPath)
		// if live reload is enabled, inject the reload snippet to html files
		if site.Config.LiveReload && targetExt == ".html" {
			// TODO inject live reload snippet
		}

		// if enabled, minify web files
		if site.Config.Minify && (targetExt == ".html" || targetExt == ".css" || targetExt == ".js") {
			// TODO minify output
		}

		// write the file contents over to target
		fmt.Println("writing", targetPath)
		return writeToFile(targetPath, contentReader)
	})
}

func (site Site) render(templ *templates.Template) ([]byte, error) {
	ctx := map[string]interface{}{
		"site": map[string]interface{}{
			"config": site.Config.AsContext(),
			"posts":  site.posts,
			"tags":   site.tags,
			"pages":  site.pages,
			"data":   site.data,
		},
	}

	ctx["page"] = templ.Metadata
	content, err := templ.Render(ctx)
	if err != nil {
		return nil, err
	}

	// recursively render parent layouts
	layout := templ.Metadata["layout"]
	for layout != nil && err == nil {
		if layout_templ, ok := site.layouts[layout.(string)]; ok {
			ctx["layout"] = layout_templ.Metadata
			ctx["content"] = content
			content, err = layout_templ.Render(ctx)
			if err != nil {
				return nil, err
			}
			layout = layout_templ.Metadata["layout"]
		} else {
			return nil, fmt.Errorf("layout '%s' not found", layout)
		}
	}

	return content, nil
}

func writeToFile(targetPath string, source io.Reader) error {
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, source)
	if err != nil {
		return err
	}

	return targetFile.Sync()
}

// Assuming the given template is a post, try to generating an excerpt of it.
// If it contains an `excerpt` key in its metadata use that, otherwise try
// to render it as HTML and extract the text of its first <p>
func getExcerpt(templ *templates.Template) string {
	if excerpt, ok := templ.Metadata["excerpt"]; ok {
		return excerpt.(string)
	}

	// if we don't expect this to render to html don't bother parsing it
	if templ.Ext() != ".html" {
		return ""
	}

	ctx := map[string]interface{}{
		"page": templ.Metadata,
	}
	content, err := templ.Render(ctx)
	if err != nil {
		return ""
	}

	html, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return ""
	}

	ptag := findFirstParagraph(html)
	return getTextContent(ptag)
}

func findFirstParagraph(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "p" {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if p := findFirstParagraph(c); p != nil {
			return p
		}
	}
	return nil
}

func getTextContent(node *html.Node) string {
	var textContent string
	if node.Type == html.TextNode {
		textContent = node.Data
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		textContent += getTextContent(c)
	}
	return textContent
}
