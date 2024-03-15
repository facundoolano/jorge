package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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

	SiteUrl        string
	PostFormat     string
	Lang           string
	HighlightTheme string

	Minify           bool
	MinifyExclusions []string
	LiveReload       bool
	LinkStatic       bool
	IncludeDrafts    bool

	ServerHost string
	ServerPort int

	pageDefaults map[string]interface{}

	// the user provided overrides, as found in config.yml
	// these will passed as found as template context
	overrides map[string]interface{}
}

func Load(rootDir string) (*Config, error) {
	// TODO allow to disable minify

	config := &Config{
		RootDir:          rootDir,
		SrcDir:           filepath.Join(rootDir, "src"),
		TargetDir:        filepath.Join(rootDir, "target"),
		LayoutsDir:       filepath.Join(rootDir, "layouts"),
		IncludesDir:      filepath.Join(rootDir, "includes"),
		DataDir:          filepath.Join(rootDir, "data"),
		PostFormat:       "blog/:title.org",
		Lang:             "en",
		HighlightTheme:   "github",
		Minify:           true,
		MinifyExclusions: make([]string, 0),
		LiveReload:       false,
		LinkStatic:       false,
		IncludeDrafts:    false,
		pageDefaults:     map[string]interface{}{},
	}

	// load overrides from config.yml
	configPath := filepath.Join(rootDir, "config.yml")
	yamlContent, err := os.ReadFile(configPath)

	if errors.Is(err, os.ErrNotExist) {
		// config file is not mandatory
		return config, nil
	} else if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlContent, &config.overrides)
	if err != nil {
		return nil, err
	}

	// set user-provided overrides of declared config keys
	// FIXME less copypasty way of declaring config overrides
	if url, found := config.overrides["url"]; found {
		config.SiteUrl = url.(string)
	}
	if format, found := config.overrides["post_format"]; found {
		config.PostFormat = format.(string)
	}
	if lang, found := config.overrides["lang"]; found {
		config.Lang = lang.(string)
	}
	if theme, found := config.overrides["highlight_theme"]; found {
		config.HighlightTheme = theme.(string)
	}
	if exclusions, found := config.overrides["minify_exclusions"]; found {
		for _, exclusion := range exclusions.([]interface{}) {
			config.MinifyExclusions = append(config.MinifyExclusions, exclusion.(string))
		}
	}

	return config, nil
}

func LoadDev(rootDir string, host string, port int, reload bool) (*Config, error) {
	// TODO revisit is this Load vs LoadDevServer is the best way to handle both modes
	// TODO some of the options need to be overridable: host, port, live reload at least

	config, err := Load(rootDir)
	if err != nil {
		return nil, err
	}

	// setup serve command specific overrides (these could be eventually tweaked with flags)
	config.ServerHost = host
	config.ServerPort = port
	config.LiveReload = reload
	config.Minify = false
	config.LinkStatic = true
	config.IncludeDrafts = true
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
