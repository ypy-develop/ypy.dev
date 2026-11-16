// Package site is what every page of ypy.dev shares: the manual, the reference
// tables, the search index, the addresses the header and the footer link to and
// the Chrome the layout is drawn from.
package site

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	markdown "github.com/hyz-is/arandu-markdown"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
)

// Facts that are the site's own and not read from anywhere: the addresses of
// the project and the release the install snippets name. The YPY version is
// read from the VERSION file YPY ships, not written here.
const (
	// GoSDKVersion is the release of github.com/ypy-develop/go the pages name.
	GoSDKVersion = "v0.1.0"
	// GitHub is the organization.
	GitHub = "https://github.com/ypy-develop"
	// Community is where questions go.
	Community = "https://github.com/orgs/ypy-develop/discussions"
	// CDN is the address of the published script, for each of its two homes.
	CDNJsDelivr = "https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v%s/dist/ypy.min.js"
	CDNYpy      = "https://cdn.ypy.dev/ypy@%s/ypy.min.js"
)

// Config is what the site is built from.
type Config struct {
	// Name is the brand, "YPY".
	Name string
	// URL is the public origin, without a trailing slash.
	URL string
	// Indexing is false on a deployment that must not be indexed.
	Indexing bool
	// Content is the folder of Markdown files, content/.
	Content fs.FS
}

// Site is built once at start and read by every request.
type Site struct {
	Name     string
	URL      string
	Indexing bool

	Manual    *docs.Manual
	Reference *reference.Catalog
	Index     docs.Index

	docLinks  map[string]string
	extGuides map[string]string
	entries   []Entry
}

// New loads the manual and the reference tables and checks that they agree.
func New(cfg Config, md *markdown.Module) (*Site, error) {
	if cfg.Content == nil {
		return nil, errors.New("site: no content folder")
	}
	base, err := url.Parse(cfg.URL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, errors.New("site: APP_URL has to be an absolute address such as https://ypy.dev")
	}

	manual, err := docs.Load(cfg.Content, md)
	if err != nil {
		return nil, err
	}
	catalog, err := reference.Load()
	if err != nil {
		return nil, err
	}

	s := &Site{
		Name: cfg.Name, URL: strings.TrimSuffix(cfg.URL, "/"), Indexing: cfg.Indexing,
		Manual: manual, Reference: catalog,
	}
	s.linkReference()
	s.Index = s.searchIndex()
	s.entries = s.buildEntries()
	return s, nil
}

// ContentDir finds the content folder: CONTENT_DIR when it is set, otherwise
// the content/ folder next to go.mod, so a test and a server started from
// another directory read the same files.
func ContentDir() (string, error) {
	if dir := os.Getenv("CONTENT_DIR"); dir != "" {
		return dir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := wd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "content")); err == nil {
			if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
				return filepath.Join(d, "content"), nil
			}
		}
		if filepath.Dir(d) == d {
			return "", errors.New("site: no content/ folder found; set CONTENT_DIR")
		}
	}
}

// YPYVersion is the version of the library the reference describes, "0.1.1".
func (s *Site) YPYVersion() string { return s.Reference.Version }

// AbsoluteURL is the full address of a path on this site.
func (s *Site) AbsoluteURL(p string) string { return s.URL + p }

// searchIndex is the manual and every name of the reference.
func (s *Site) searchIndex() docs.Index {
	idx := s.Manual.Entries()
	for _, k := range reference.Kinds {
		for _, it := range s.Reference.Of(k) {
			where := reference.Titles[k]
			if it.Extension != "" {
				where += " (" + it.Extension + ")"
			}
			idx = append(idx, docs.Entry{
				Title:       it.Name,
				URL:         "/reference/" + string(k) + "#" + it.Anchor,
				Where:       "Reference: " + where,
				Description: stripTicks(it.Description),
			})
		}
	}
	return idx
}

func stripTicks(s string) string { return strings.ReplaceAll(s, "`", "") }
