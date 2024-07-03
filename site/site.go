package site

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/markup"
	"gopkg.in/yaml.v3"
)

const FILE_RW_MODE = 0666
const DIR_RWE_MODE = 0777

type site struct {
	config  config.Config
	layouts map[string]markup.Template
	posts   []map[string]interface{}
	pages   []map[string]interface{}
	tags    map[string][]map[string]interface{}
	data    map[string]interface{}

	templateEngine *markup.Engine
	templates      map[string]*markup.Template

	minifier markup.Minifier
}

// Load the site project pointed by `config`, then walk `config.SrcDir`
// and recreate it at `config.TargetDir` by rendering template files and copying static ones.
// The previous target dir contents are deleted.
func Build(config config.Config) error {
	site, err := load(config)
	if err != nil {
		return err
	}

	return site.build()
}

// Create a new site instance by scanning the project directories
// pointed by `config`, loading layouts, templates and data files.
func load(config config.Config) (*site, error) {
	site := site{
		layouts:        make(map[string]markup.Template),
		templates:      make(map[string]*markup.Template),
		config:         config,
		tags:           make(map[string][]map[string]interface{}),
		data:           make(map[string]interface{}),
		templateEngine: markup.NewEngine(config.SiteUrl, config.IncludesDir),
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

	site.minifier = markup.LoadMinifier(config.MinifyExclusions)

	return &site, nil
}

func (site *site) loadLayouts() error {
	files, err := os.ReadDir(site.config.LayoutsDir)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(site.config.LayoutsDir, filename)
			templ, err := markup.Parse(site.templateEngine, path)
			if err != nil {
				return checkFileError(err)
			}

			layout_name := strings.TrimSuffix(filename, filepath.Ext(filename))
			site.layouts[layout_name] = *templ
		}
	}

	return nil
}

func (site *site) loadDataFiles() error {
	files, err := os.ReadDir(site.config.DataDir)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(site.config.DataDir, filename)

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

func (site *site) loadTemplates() error {
	if _, err := os.Stat(site.config.SrcDir); err != nil {
		return fmt.Errorf("missing src directory")
	}

	err := filepath.WalkDir(site.config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			templ, err := markup.Parse(site.templateEngine, path)
			// if something fails or this is not a template, skip
			if err != nil || templ == nil {
				return checkFileError(err)
			}

			// set site related (?) metadata. Not sure if this should go elsewhere
			relPath, _ := filepath.Rel(site.config.SrcDir, path)
			srcPath, _ := filepath.Rel(site.config.RootDir, path)
			relPath = strings.TrimSuffix(relPath, filepath.Ext(relPath)) + templ.TargetExt()
			templ.Metadata["src_path"] = srcPath
			templ.Metadata["path"] = relPath
			templ.Metadata["url"] = "/" + strings.TrimSuffix(strings.TrimSuffix(relPath, "index.html"), ".html")
			templ.Metadata["dir"] = "/" + filepath.Dir(relPath)

			// if drafts are disabled, exclude from posts, page and tags indexes, but not from site.templates
			// we want to explicitly exclude the template from the target, rather than treating it as a non template file
			if !templ.IsDraft() || site.config.IncludeDrafts {
				// posts are templates that can be chronologically sorted --that have a date.
				// the rest are pages.
				if templ.IsPost() {

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
			}

			site.templates[path] = templ
		}
		return nil
	})

	if err != nil {
		return err
	}

	// sort by reverse chronological order when date is present
	// otherwise by path alphabetical
	CompareTemplates := func(a map[string]interface{}, b map[string]interface{}) int {
		if bdate, ok := b["date"]; ok {
			if adate, ok := a["date"]; ok {
				return bdate.(time.Time).Compare(adate.(time.Time))
			}
		}
		return strings.Compare(a["path"].(string), b["path"].(string))
	}
	slices.SortFunc(site.posts, CompareTemplates)
	slices.SortFunc(site.pages, CompareTemplates)
	for _, posts := range site.tags {
		slices.SortFunc(posts, CompareTemplates)
	}

	// populate previous and next in template index
	site.addPrevNext(site.pages)
	site.addPrevNext(site.posts)

	return nil
}

func (site *site) addPrevNext(posts []map[string]interface{}) {
	for i, post := range posts {
		path := filepath.Join(site.config.RootDir, post["src_path"].(string))

		// only consider them part of the same collection if they share the directory
		if i > 0 && post["dir"] == posts[i-1]["dir"] {
			// make a copy of the map, without prev/next (to avoid weird recursion)
			previous := maps.Clone(posts[i-1])
			delete(previous, "previous")
			delete(previous, "next")
			site.templates[path].Metadata["previous"] = previous
		}

		if i < len(posts)-1 && post["dir"] == posts[i+1]["dir"] {
			next := maps.Clone(posts[i+1])
			delete(next, "previous")
			delete(next, "next")
			site.templates[path].Metadata["next"] = next
		}
	}
}

// Walk the `site.Config.SrcDir` directory and reproduce it at `site.Config.TargetDir`,
// rendering template files and copying static ones.
func (site *site) build() error {
	// clear previous target contents
	os.RemoveAll(site.config.TargetDir)

	wg, files := spawnBuildWorkers(site)
	defer wg.Wait()
	defer close(files)

	// walk the source directory, creating directories and files at the target dir
	return filepath.WalkDir(site.config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			// skip dot files and directories
			return nil
		}
		subpath, _ := filepath.Rel(site.config.SrcDir, path)
		targetPath := filepath.Join(site.config.TargetDir, subpath)

		// if it's a directory, just create the same at the target
		if entry.IsDir() {
			return os.MkdirAll(targetPath, DIR_RWE_MODE)
		}
		// if it's a file (either static or template) send the path to a worker to build in target
		files <- path
		return nil
	})
}

// Create a channel to send paths to build and a worker pool to handle them concurrently
func spawnBuildWorkers(site *site) (*sync.WaitGroup, chan string) {

	var wg sync.WaitGroup
	files := make(chan string, 20)

	for range runtime.NumCPU() {
		wg.Add(1)
		go func(files <-chan string) {
			defer wg.Done()
			for path := range files {
				err := site.buildFile(path)
				if err != nil {
					fmt.Printf("error in %s: %s\n", path, err)
				}
			}
		}(files)
	}
	return &wg, files
}

func (site *site) buildFile(path string) error {
	subpath, _ := filepath.Rel(site.config.SrcDir, path)
	targetPath := filepath.Join(site.config.TargetDir, subpath)

	var contentReader io.Reader
	var err error
	templ, found := site.templates[path]
	if !found {
		// if no template found at location, treat the file as static write its contents to target
		if site.config.LinkStatic {
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
		if templ.IsDraft() && !site.config.IncludeDrafts {
			fmt.Println("skipping draft", targetPath)
			return nil
		}

		content, err := site.render(templ)
		if err != nil {
			return err
		}

		targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + templ.TargetExt()
		contentReader = bytes.NewReader(content)
	}
	targetExt := filepath.Ext(targetPath)

	// arrange paths to ensure pretty uris, eg move blog/tags.html to blog/tags/index.html
	if targetExt == ".html" && filepath.Base(targetPath) != "index.html" {
		targetDir := strings.TrimSuffix(targetPath, ".html")
		targetPath = filepath.Join(targetDir, "index.html")
		err = os.MkdirAll(targetDir, DIR_RWE_MODE)
		if err != nil {
			return err
		}
	}

	// post process file acording to extension and config
	contentReader, err = markup.Smartify(targetExt, contentReader)
	if err != nil {
		return err
	}
	contentReader, err = site.injectLiveReload(targetExt, contentReader)
	if err != nil {
		return err
	}
	if site.config.Minify {
		contentReader = site.minifier.Minify(subpath, contentReader)
	}

	// write the file contents over to target
	return writeToFile(targetPath, contentReader)
}

func (site *site) render(templ *markup.Template) ([]byte, error) {
	ctx := map[string]interface{}{
		"site": map[string]interface{}{
			"config": site.config.AsContext(),
			"posts":  site.posts,
			"tags":   site.tags,
			"pages":  site.pages,
			"data":   site.data,
		},
	}

	ctx["page"] = templ.Metadata
	content, err := templ.RenderWith(ctx, site.config.HighlightTheme)
	if err != nil {
		return nil, err
	}

	// recursively render parent layouts
	layout := templ.Metadata["layout"]
	for layout != nil && err == nil {
		if layout_templ, ok := site.layouts[layout.(string)]; ok {
			ctx["layout"] = layout_templ.Metadata
			ctx["content"] = content
			content, err = layout_templ.RenderWith(ctx, site.config.HighlightTheme)
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
func getExcerpt(templ *markup.Template) string {
	if excerpt, ok := templ.Metadata["excerpt"]; ok {
		return excerpt.(string)
	}

	// if we don't expect this to render to html don't bother parsing it
	if templ.TargetExt() != ".html" {
		return ""
	}

	content, err := templ.Render()
	if err != nil {
		return ""
	}
	return markup.ExtractFirstParagraph(bytes.NewReader(content))
}

// if live reload is enabled, inject the reload snippet to html files
func (site *site) injectLiveReload(extension string, contentReader io.Reader) (io.Reader, error) {
	if !site.config.LiveReload || extension != ".html" {
		return contentReader, nil
	}

	const JS_SNIPPET = `
const url = location.origin + '/_events/'
var eventSource;
function newSSE() {
  console.log("connecting to server events");
  eventSource = new EventSource(url);
  eventSource.onmessage = function () {
    location.reload()
  };
  window.onbeforeunload = function() {
    eventSource.close();
  }
  eventSource.onerror = function (event) {
    console.error('An error occurred:', event);
    eventSource.close();
    setTimeout(newSSE, 5000)
  };
}
newSSE();`
	return markup.InjectScript(contentReader, JS_SNIPPET)
}
