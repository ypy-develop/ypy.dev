package site

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arandu-io/hesape/geo"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
)

// The sentences the front page and the discovery documents say about YPY.
const (
	// HomeTitle is the title of the front page: the brand alone would say nothing
	// in a search result.
	HomeTitle = "YPY · Hypermedia framework: one script, no build step"
	// HomeDescription is the front page's description.
	HomeDescription = "YPY is a hypermedia framework: HTML attributes ask the server for HTML and say where the answer goes. One script, no build step, no dependency."
)

// Entry is one public page the discovery documents list: the document the geo
// module serves, and where the page sits in the manual, for llms.txt.
type Entry struct {
	geo.Document
	// Section is "docs" or "go" for a page of the manual, "" for the front page and
	// the generated reference.
	Section string
	// Group is the folder under the section, "" for the pages at its top.
	Group string
	Lang  string
}

// GeoConfig is the configuration of the discovery surfaces of this deployment.
//
// Indexing and Origin come from the configuration, never from a request: a
// deployment that is not production keeps robots.txt closed, the sitemap empty
// and every page on noindex. The module serves robots.txt, sitemap.xml and
// llms-full.txt; llms.txt is the site's own (see MachineController) because it
// groups the pages by section, which the module's flat list does not.
func (s *Site) GeoConfig() geo.Config {
	return geo.Config{
		Enabled:     true,
		Indexing:    s.Indexing,
		Origin:      s.URL,
		Surfaces:    geo.Robots.AsSet() | geo.Sitemap.AsSet() | geo.LLMsFull.AsSet(),
		Language:    "en",
		IndexTitle:  s.Name,
		CorpusTitle: "YPY documentation, in full",
		CorpusIntro: "The public pages of ypy.dev as plain text: the YPY documentation, the Go SDK and the reference tables generated from the library. YPY is a hypermedia framework: " +
			"HTML attributes ask the server for HTML and say where the answer goes. One script, no build step, no dependency. Version " + s.YPYVersion() + ", MIT licence.",
		Robots: geo.RobotsPolicy{
			Allow: []string{"/"},
			// The search results are a query, the demo answers a fragment and the
			// index is a file for the page's own script: none is a page to rank.
			Disallow: []string{"/search", "/search-index.json", "/demo"},
		},
	}
}

// Catalog is the public pages, for the module. Nothing here is private: the
// site has no sign-in, and a page of the manual is public by being in content/.
func (s *Site) Catalog() geo.Catalog {
	return geo.CatalogFunc(func(context.Context) ([]geo.Document, error) {
		out := make([]geo.Document, len(s.entries))
		for i, e := range s.entries {
			out[i] = e.Document
		}
		return out, nil
	})
}

// Entries are every public page, in address order.
func (s *Site) Entries() []Entry { return s.entries }

// buildEntries lists the front page, the generated reference, every page of the
// manual in both languages and the folders that have no page of their own.
func (s *Site) buildEntries() []Entry {
	var latest time.Time
	for _, p := range s.Manual.Pages() {
		if t := parseDate(p.Updated); t.After(latest) {
			latest = t
		}
	}

	entries := []Entry{{
		Document: geo.Document{Path: "/", Title: HomeTitle, Description: HomeDescription, UpdatedAt: latest, Content: s.homeText()},
	}}

	entries = append(entries, Entry{Document: geo.Document{
		Path:        "/reference",
		Title:       "Reference",
		Description: "Every attribute, event, header and configuration key of YPY, with the official extensions.",
		Content:     s.referenceIndexText(),
	}})
	for _, k := range reference.Kinds {
		title := s.ReferenceTitle(k)
		entries = append(entries, Entry{Document: geo.Document{
			Path:        "/reference/" + string(k),
			Title:       title + " reference",
			Description: "Every YPY " + title + ", generated from the library's own name tables.",
			Content:     s.referenceText(k),
		}})
	}
	entries = append(entries, Entry{Document: geo.Document{
		Path:        "/reference/extensions",
		Title:       "Extensions reference",
		Description: "The official YPY extensions, the script of each and the names it adds.",
		Content:     s.extensionsText(),
	}})

	for _, p := range s.Manual.Pages() {
		entries = append(entries, Entry{
			Document: geo.Document{
				Path: p.Route, Title: p.Title, Description: p.Description,
				UpdatedAt: parseDate(p.Updated), Content: p.Text,
			},
			Section: p.Section.ID, Group: p.Group, Lang: p.Lang,
		})
	}

	for _, sec := range docs.Sections {
		for _, lang := range []string{docs.English, docs.Portuguese} {
			for _, g := range s.Manual.Tree(sec.ID, lang) {
				if g.Key == "" {
					continue
				}
				route := LandingRoute(sec, lang, g)
				if _, isPage := s.Manual.Page(route); isPage {
					continue
				}
				entries = append(entries, Entry{
					Document: geo.Document{Path: route, Title: g.Title, Description: LandingDescription(sec, g)},
					Section:  sec.ID, Group: g.Key, Lang: lang,
				})
			}
		}
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries
}

// LandingRoute is the address of a folder that has no page of its own.
func LandingRoute(sec docs.Section, lang string, g docs.Group) string {
	route := sec.Prefix
	if lang == docs.Portuguese {
		route += "/" + docs.Portuguese
	}
	return route + "/" + g.Key
}

// LandingDescription says what a folder's address lists, from the pages in it.
func LandingDescription(sec docs.Section, g docs.Group) string {
	var titles []string
	for _, p := range g.Pages {
		titles = append(titles, p.Title)
	}
	for _, sub := range g.Subs {
		for _, p := range sub.Pages {
			titles = append(titles, p.Title)
		}
	}
	total := len(titles)
	if total > 5 {
		titles = titles[:5]
	}
	line := fmt.Sprintf("The %s pages of the %s documentation: %s", g.Title, sec.Title, strings.Join(titles, ", "))
	if total > 5 {
		line += fmt.Sprintf(" and %d more", total-5)
	}
	return line + "."
}

// ReferenceTitle is the plural name of a kind of the reference tables.
func (s *Site) ReferenceTitle(k reference.Kind) string { return reference.Titles[k] }

func (s *Site) homeText() string {
	return HomeDescription + "\n\nInstall: npm install @ypy-dev/ypy@" + s.YPYVersion() + "\n" +
		fmt.Sprintf("The library has %d attributes, %d events and %d official extensions.",
			len(s.Reference.Of(reference.Attributes)), len(s.Reference.Of(reference.Events)), len(s.Reference.Extensions))
}

func (s *Site) referenceIndexText() string {
	var b strings.Builder
	b.WriteString("The tables are generated from the name tables that ship with YPY " + s.YPYVersion() + ".\n")
	for _, k := range reference.Kinds {
		fmt.Fprintf(&b, "%s: %d names\n", reference.Titles[k], len(s.Reference.Of(k)))
	}
	fmt.Fprintf(&b, "Extensions: %d", len(s.Reference.Extensions))
	return b.String()
}

func (s *Site) referenceText(k reference.Kind) string {
	var b strings.Builder
	for _, it := range s.Reference.Of(k) {
		line := it.Name
		if it.Extension != "" {
			line += " (" + it.Extension + ")"
		}
		if it.Description != "" {
			line += ": " + stripTicks(it.Description)
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimSpace(b.String())
}

func (s *Site) extensionsText() string {
	var b strings.Builder
	for _, e := range s.Reference.Extensions {
		fmt.Fprintf(&b, "%s (script %s): %d attributes, %d events, %d headers, %d config keys, %d classes\n",
			e.Name, e.Script, e.Attributes, e.Events, e.Headers, e.Config, e.Classes)
	}
	return strings.TrimSpace(b.String())
}

// parseDate reads the `updated` of a page's front matter, "2026-11-15". A page
// without a readable date has no date: a sitemap that guessed one would lie.
func parseDate(v string) time.Time {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(v))
	if err != nil {
		return time.Time{}
	}
	return t
}
