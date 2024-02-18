package commands

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/facundoolano/jorge/config"
	"github.com/facundoolano/jorge/site"
	"github.com/fsnotify/fsnotify"
)

// Generate and serve the site, rebuilding when the source files change.
func Serve(rootDir string) error {
	config, err := config.LoadDevServer(rootDir)
	if err != nil {
		return err
	}

	if err := rebuild(config); err != nil {
		return err
	}

	// watch for changes in src and layouts, and trigger a rebuild
	watcher, err := setupWatcher(config)
	if err != nil {
		return err
	}
	defer watcher.Close()

	// serve the target dir with a file server
	fs := http.FileServer(HTMLDir{http.Dir(config.TargetDir)})
	http.Handle("/", http.StripPrefix("/", fs))

	addr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	fmt.Printf("server listening at http://%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func rebuild(config *config.Config) error {

	site, err := site.Load(*config)
	if err != nil {
		return err
	}

	if err := site.Build(); err != nil {
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

func setupWatcher(config *config.Config) (*fsnotify.Watcher, error) {
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

				// chmod events are noisy, ignore them. also skip create events
				// which we assume meaningless until the write that comes next
				if event.Has(fsnotify.Chmod) || event.Has(fsnotify.Create) {
					continue
				}

				fmt.Printf("\nFile %s changed, triggering rebuild.\n", event.Name)

				// since new nested directories could be triggering this change, and we need to watch those too
				// and since re-watching files is a noop, I just re-add the entire src everytime there's a change
				if err := addAll(watcher, config); err != nil {
					fmt.Println("couldn't add watchers:", err)
					continue
				}

				if err := rebuild(config); err != nil {
					fmt.Println("build error:", err)
					continue
				}

				fmt.Println("done\nserver listening at", config.SiteUrl)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	err = addAll(watcher, config)

	return watcher, err
}

// Add the layouts and all source directories to the given watcher
func addAll(watcher *fsnotify.Watcher, config *config.Config) error {
	err := watcher.Add(config.LayoutsDir)
	err = watcher.Add(config.DataDir)
	err = watcher.Add(config.IncludesDir)
	// fsnotify watches all files within a dir, but non recursively
	// this walks through the src dir and adds watches for each found directory
	filepath.WalkDir(config.SrcDir, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
	return err
}
