package site

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/templates"
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

	minifier Minifier
}

func Load(config config.Config) (*Site, error) {
	site := Site{
		layouts:        make(map[string]templates.Template),
		templates:      make(map[string]*templates.Template),
		Config:         config,
		tags:           make(map[string][]map[string]interface{}),
		data:           make(map[string]interface{}),
		templateEngine: templates.NewEngine(config.SiteUrl, config.IncludesDir),
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

	site.loadMinifier()

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
				return checkFileError(err)
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
	if _, err := os.Stat(site.Config.SrcDir); os.IsNotExist(err) {
		return fmt.Errorf("missing src directory")
	}

	err := filepath.WalkDir(site.Config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			templ, err := templates.Parse(site.templateEngine, path)
			// if something fails or this is not a template, skip
			if err != nil || templ == nil {
				return checkFileError(err)
			}

			// set site related (?) metadata. Not sure if this should go elsewhere
			relPath, _ := filepath.Rel(site.Config.SrcDir, path)
			srcPath, _ := filepath.Rel(site.Config.RootDir, path)
			relPath = strings.TrimSuffix(relPath, filepath.Ext(relPath)) + templ.Ext()
			templ.Metadata["src_path"] = srcPath
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

	wg, files := spawnBuildWorkers(site)
	defer wg.Wait()
	defer close(files)

	// walk the source directory, creating directories and files at the target dir
	err := filepath.WalkDir(site.Config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		subpath, _ := filepath.Rel(site.Config.SrcDir, path)
		targetPath := filepath.Join(site.Config.TargetDir, subpath)

		// if it's a directory, just create the same at the target
		if entry.IsDir() {
			return os.MkdirAll(targetPath, FILE_RW_MODE)
		}
		// if it's a file (either static or template) send the path to a worker to build in target
		files <- path
		return nil
	})

	return err
}

// Create a channel to send paths to build and a worker pool to handle them concurrently
func spawnBuildWorkers(site *Site) (*sync.WaitGroup, chan string) {

	var wg sync.WaitGroup
	files := make(chan string, 20)

	for range runtime.NumCPU() {
		wg.Add(1)
		go func(files <-chan string) {
			defer wg.Done()
			for path := range files {
				site.buildFile(path)
			}
		}(files)
	}
	return &wg, files
}

func (site *Site) buildFile(path string) error {
	subpath, _ := filepath.Rel(site.Config.SrcDir, path)
	targetPath := filepath.Join(site.Config.TargetDir, subpath)

	var contentReader io.Reader
	var err error
	templ, found := site.templates[path]
	if !found {
		// if no template found at location, treat the file as static write its contents to target
		if site.Config.LinkStatic {
			// dev optimization: link static files instead of copying them
			abs, _ := filepath.Abs(path)
			err = os.Symlink(abs, targetPath)
			return checkFileError(err)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return checkFileError(err)
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
	contentReader, err = site.injectLiveReload(targetExt, contentReader)
	if err != nil {
		return err
	}
	contentReader = site.minify(targetExt, contentReader)

	// write the file contents over to target
	return writeToFile(targetPath, contentReader)
}

func (site *Site) render(templ *templates.Template) ([]byte, error) {
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

func checkFileError(err error) error {
	// When walking the source dir it can happen that a file is present when walking starts
	// but missing or inaccessible when trying to open it (this is particularly frequent with
	// backup files from emacs and when running the dev server). We don't want to halt the build
	// process in that situation, just inform and continue.
	if os.IsNotExist(err) {
		// don't abort on missing files, usually spurious temps
		fmt.Println("skipping missing file", err)
		return nil
	}
	return err
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

	fmt.Println("wrote", targetPath)
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
	return ExtractFirstParagraph(bytes.NewReader(content))
}

// if live reload is enabled, inject the reload snippet to html files
func (site *Site) injectLiveReload(extension string, contentReader io.Reader) (io.Reader, error) {
	if !site.Config.LiveReload || extension != ".html" {
		return contentReader, nil
	}

	const JS_SNIPPET = `
const url = '%s/_events/'
const eventSource = new EventSource(url);

eventSource.onmessage = function () {
  location.reload()
};
window.onbeforeunload = function() {
  eventSource.close();
}
eventSource.onerror = function (event) {
  console.error('An error occurred:', event)
};`
	script := fmt.Sprintf(JS_SNIPPET, site.Config.SiteUrl)
	return InjectScript(contentReader, script)
}
