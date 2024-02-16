package config

import (
	"fmt"
	"maps"
	"path/filepath"
)

// The properties that are depended upon in the source code are declared explicitly in the config struct.
// The constructors will set default values for most.
// Depending on the command, different defaults will be used (serve is assumed to be a "dev" environment
// while build is assumed to be prod)
// Some defaults could be overridden by cli flags (eg disable live reload on serve).
// The user can override some of those via config yaml.
// The non declared values found in config yaml will just be passed as site.config values

type Config struct {
	RootDir     string
	SrcDir      string
	TargetDir   string
	LayoutsDir  string
	IncludesDir string
	DataDir     string

	SiteUrl    string
	SlugFormat string

	Minify     bool
	LiveReload bool
	LinkStatic bool

	ServerHost string
	ServerPort int

	pageDefaults map[string]interface{}

	// the user provided overrides, as found in config.yml
	// these will passed as found as template context
	overrides map[string]interface{}
}

func Load(rootDir string) (*Config, error) {
	// FIXME change defaults based on command mode

	config := &Config{
		RootDir:      rootDir,
		SrcDir:       filepath.Join(rootDir, "src"),
		TargetDir:    filepath.Join(rootDir, "target"),
		LayoutsDir:   filepath.Join(rootDir, "layouts"),
		IncludesDir:  filepath.Join(rootDir, "includes"),
		DataDir:      filepath.Join(rootDir, "data"),
		SlugFormat:   ":title",
		Minify:       true,
		LiveReload:   false,
		LinkStatic:   false,
		pageDefaults: map[string]interface{}{},
	}

	// TODO load overrides from config.yml
	// TODO set siteUrl from overrides["url"]
	// TODO set slugFormat from overrides["slug"]

	return config, nil
}

func LoadDevServer(rootDir string) (*Config, error) {
	config, err := Load(rootDir)
	if err != nil {
		return nil, err
	}

	// setup serve command specific overrides (these could be eventually tweaked with flags)
	config.ServerHost = "localhost"
	config.ServerPort = 4001
	config.Minify = false
	config.LiveReload = true
	config.LinkStatic = true
	config.SiteUrl = fmt.Sprintf("http://%s:%d", config.ServerHost, config.ServerPort)

	return config, nil
}

func (config Config) AsContext() map[string]interface{} {
	context := map[string]interface{}{
		"url": config.SiteUrl,
	}
	maps.Copy(context, config.overrides)
	return context
}
