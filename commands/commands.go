package commands

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/facundoolano/blorg/site"
	"github.com/niklasfasching/go-org/org"
)

const SRC_DIR = "src"
const TARGET_DIR = "target"
const LAYOUTS_DIR = "layouts"
const FILE_RW_MODE = 0777

func Init() error {
	// get working directory
	// default to .
	// if not exist, create directory
	// copy over default files
	fmt.Println("not implemented yet")
	return nil
}

func New() error {
	// prompt for title
	// slugify
	// fail if file already exist
	// create a new .org file with the slug
	// add front matter and org options
	fmt.Println("not implemented yet")
	return nil
}

// Read the files in src/ render them and copy the result to target/
// TODO add root dir override support
func Build() error {
	site, err := site.Load(SRC_DIR, LAYOUTS_DIR)
	if err != nil {
		return err
	}

	return buildTarget(site, true, false)
}

// TODO consider moving to site
// TODO consider making minify and reload site.config values
func buildTarget(site *site.Site, minify bool, htmlReload bool) error {
	// clear previous target contents
	os.RemoveAll(TARGET_DIR)
	os.Mkdir(TARGET_DIR, FILE_RW_MODE)

	// walk the source directory, creating directories and files at the target dir
	return filepath.WalkDir(SRC_DIR, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		subpath, _ := filepath.Rel(SRC_DIR, path)
		targetPath := filepath.Join(TARGET_DIR, subpath)

		// if it's a directory, just create the same at the target
		if entry.IsDir() {
			return os.MkdirAll(targetPath, FILE_RW_MODE)
		}

		contentReader, found, err := site.RenderTemplate(path)
		if err != nil {
			return err
		} else if !found {
			// if no template found at location, treat the file as static
			// write its contents to target
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			contentReader = srcFile
		}

		// if it's org or markdown, export to html, updating the target extension accordingly
		switch filepath.Ext(targetPath) {
		case ".org":
			{
				doc := org.New().Parse(contentReader, path)
				content, err := doc.Write(org.NewHTMLWriter())
				if err != nil {
					return err
				}
				contentReader = strings.NewReader(content)
				targetPath = strings.TrimSuffix(targetPath, ".org") + ".html"
			}
		case ".md":
			{
				// TODO parse markdown
				targetPath = strings.TrimSuffix(targetPath, ".md") + ".html"
			}
		}

		// if live reload is enabled, inject the reload snippet to html files
		ext := filepath.Ext(targetPath)
		if htmlReload && ext == ".html" {
			// TODO inject live reload snippet
		}

		// if enabled, minify web files
		if minify && (ext == ".html" || ext == ".css" || ext == ".js") {
			// TODO minify output
		}

		// write the file contents over to target
		fmt.Println("writing", targetPath)
		return writeToFile(targetPath, contentReader)
	})
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
