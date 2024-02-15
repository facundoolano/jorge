package commands

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/facundoolano/blorg/site"
	"github.com/fsnotify/fsnotify"
)

// Generate and serve the site, rebuilding when the source files change.
func Serve() error {

	if err := rebuild(); err != nil {
		return err
	}

	// watch for changes in src and layouts, and trigger a rebuild
	watcher, err := setupWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// serve the target dir with a file server
	fs := http.FileServer(HTMLDir{http.Dir("target/")})
	http.Handle("/", http.StripPrefix("/", fs))
	fmt.Println("server listening at http://localhost:4001/")
	http.ListenAndServe(":4001", nil)

	return nil
}

func rebuild() error {
	site, err := site.Load(SRC_DIR, LAYOUTS_DIR)
	if err != nil {
		return err
	}

	if err := site.Build(SRC_DIR, TARGET_DIR, false, true); err != nil {
		return err
	}

	return nil
}

// Tweaks the http file system to construct a server that hides the .html suffix from requests.
// Based on https://stackoverflow.com/a/57281956/993769
type HTMLDir struct {
	d http.Dir
}

func (d HTMLDir) Open(name string) (http.File, error) {
	// Try name as supplied
	f, err := d.d.Open(name)
	if os.IsNotExist(err) {
		// Not found, try with .html
		if f, err := d.d.Open(name + ".html"); err == nil {
			return f, nil
		}
	}
	return f, err
}

func setupWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// chmod events are noisy, ignore them
				if event.Has(fsnotify.Chmod) {
					fmt.Printf("\nFile %s changed, triggering rebuild.\n", event.Name)

					// since new nested directories could be triggering this change, and we need to watch those too
					// and since re-watching files is a noop, I just re-add the entire src everytime there's a change
					if err := addAll(watcher); err != nil {
						fmt.Println("error:", err)
						return
					}

					if err := rebuild(); err != nil {
						fmt.Println("error:", err)
						return
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	err = addAll(watcher)

	return watcher, err
}

// Add the layouts and all source directories to the given watcher
func addAll(watcher *fsnotify.Watcher) error {
	err := watcher.Add(LAYOUTS_DIR)
	// fsnotify watches all files within a dir, but non recursively
	// this walks through the src dir and adds watches for each found directory
	filepath.WalkDir(SRC_DIR, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
	return err
}
