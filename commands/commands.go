package commands

import (
	"fmt"
	"github.com/facundoolano/blorg/templates"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const SRC_DIR = "src"
const TARGET_DIR = "target"
const LAYOUT_DIR = "layouts"
const FILE_RW_MODE = 0777

func Init() error {
	// get working directory
	// default to .
	// if not exist, create directory
	// copy over default files
	fmt.Println("not implemented yet")
	return nil
}

// Read the files in src/ render them and copy the result to target/
// FIXME pass src and target by arg
func Build() error {
	_, err := os.ReadDir(SRC_DIR)
	if os.IsNotExist(err) {
		return fmt.Errorf("missing %s directory", SRC_DIR)
	} else if err != nil {
		return fmt.Errorf("couldn't read %s", SRC_DIR)
	}

	site := Site{
		layouts: make(map[string]templates.Template),
	}

	// FIXME these sound like they should be site methods too
	PHASES := []func(*Site) error{
		loadConfig,
		loadLayouts,
		loadTemplates,
		writeTarget,
	}
	for _, phaseFun := range PHASES {
		if err := phaseFun(&site); err != nil {
			return err
		}
	}

	return err
}

func loadConfig(site *Site) error {
	// context["config"]
	return nil
}

func loadLayouts(site *Site) error {
	files, err := os.ReadDir(LAYOUT_DIR)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() {
			filename := entry.Name()
			path := filepath.Join(LAYOUT_DIR, filename)
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

func loadTemplates(site *Site) error {
	return filepath.WalkDir(SRC_DIR, func(path string, entry fs.DirEntry, err error) error {
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
		}
		return nil
	})
}

func writeTarget(site *Site) error {
	// clear previous target contents
	os.RemoveAll(TARGET_DIR)
	os.Mkdir(TARGET_DIR, FILE_RW_MODE)

	// walk the source directory, creating directories and files at the target dir
	templIndex := site.templateIndex()
	return filepath.WalkDir(SRC_DIR, func(path string, entry fs.DirEntry, err error) error {
		subpath, _ := filepath.Rel(SRC_DIR, path)
		targetPath := filepath.Join(TARGET_DIR, subpath)

		if entry.IsDir() {
			os.MkdirAll(targetPath, FILE_RW_MODE)
		} else {

			if templ, ok := templIndex[path]; ok {
				// if a template was found at source, render it
				content, err := site.render(templ)
				if err != nil {
					return err
				}

				// write the file contents over to target at the same location
				targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + templ.Ext()
				fmt.Println("writing ", targetPath)
				return os.WriteFile(targetPath, []byte(content), FILE_RW_MODE)
			} else {
				// if a non template was found, copy file as is
				fmt.Println("writing ", targetPath)
				return copyFile(path, targetPath)
			}
		}

		return nil
	})
}

func copyFile(source string, target string) error {
	// does this need to be so verbose?
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	targetFile, _ := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, srcFile)
	if err != nil {
		return err
	}

	return targetFile.Sync()
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

func Serve() error {
	// build
	// serve target with file server
	// (later watch and live reload)
	fmt.Println("not implemented yet")
	return nil
}
