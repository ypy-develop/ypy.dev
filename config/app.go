// Package config is what this application reads from the environment beyond the
// framework's own settings: where the Markdown lives and whether the site may
// be indexed.
package config

import (
	"os"
	"strings"

	"github.com/arandu-io/framework/foundation/bootstrap"
)

// Config is the framework's settings and the site's own.
type Config struct {
	bootstrap.Settings

	Site Site
}

// Site is what the documentation site itself is configured with.
type Site struct {
	// ContentDir is the folder of Markdown files. Empty means the content/
	// folder next to go.mod.
	ContentDir string
	// Indexing lets crawlers index the site. It is on only in production, so a
	// staging deployment is never found by a search engine.
	Indexing bool
}

// Load reads the environment.
func Load() (Config, error) {
	settings, err := bootstrap.LoadSettings()
	if err != nil {
		return Config{}, err
	}
	return From(settings), nil
}

// From builds the configuration over settings that were read already.
func From(settings bootstrap.Settings) Config {
	indexing := settings.App.Env.IsProduction()
	if v, ok := os.LookupEnv("SITE_INDEXING"); ok && v != "" {
		indexing = strings.EqualFold(v, "true") || v == "1"
	}
	return Config{
		Settings: settings,
		Site:     Site{ContentDir: os.Getenv("CONTENT_DIR"), Indexing: indexing},
	}
}
