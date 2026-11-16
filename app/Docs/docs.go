// Package docs turns the Markdown files of content/ into the pages of the site.
//
// A page's address comes from its path and nothing else:
//
//	content/ypy/guides/events.md    -> /docs/guides/events
//	content/ypy/index.md            -> /docs
//	content/go/guide.md             -> /go/guide
//	content/go/README.md            -> /go
//
// The sidebar is the folder tree, the page contents are the headings of the
// rendered body, and "previous" and "next" follow the sidebar. Nothing is listed
// by hand: a file added to content/ is a page, and a file removed is not.
//
// The Markdown is rendered by github.com/hyz-is/arandu-markdown, which writes
// only a closed allowlist of HTML. This package adds what the module leaves to
// its caller: it points the links between files at their pages, colours the
// code on the server with classes (the site's policy forbids inline style), and
// gives every code block a copy button.
package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	markdown "github.com/hyz-is/arandu-markdown"
)

// Section is one of the two documentations the site carries.
type Section struct {
	// ID is "docs" or "go".
	ID string
	// Dir is the folder under content/.
	Dir string
	// Prefix is the address the section lives under.
	Prefix string
	// Title is its name.
	Title string
}

// Sections are the documentations of the site, in the order the header shows them.
var Sections = []Section{
	{ID: "docs", Dir: "ypy", Prefix: "/docs", Title: "YPY"},
	{ID: "go", Dir: "go", Prefix: "/go", Title: "Go SDK"},
}

// Languages the content is written in. English is the root.
const (
	English    = "en"
	Portuguese = "pt-BR"
)

// Manual is every page, loaded once at start.
type Manual struct {
	pages   []*Page
	byRoute map[string]*Page
	byFile  map[string]*Page
	trees   map[string][]Group // by section and language
	flat    map[string][]*Page // reading order, by section and language
	broken  []BrokenLink
	// relocated are links that pointed at a path where there is no file now, and
	// that were followed to the file of the same name elsewhere.
	relocated []BrokenLink
}

// BrokenLink is a link between two files that points at a file that is not a page.
type BrokenLink struct {
	From   string
	Target string
}

// Load reads every .md file under root, which is the content/ folder.
func Load(root fs.FS, md *markdown.Module) (*Manual, error) {
	m := &Manual{
		byRoute: map[string]*Page{},
		byFile:  map[string]*Page{},
		trees:   map[string][]Group{},
		flat:    map[string][]*Page{},
	}

	for _, s := range Sections {
		err := fs.WalkDir(root, s.Dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				// A documentation that has no folder yet is not an error: the site
				// serves the other one.
				if p == s.Dir && errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if d.IsDir() || !strings.HasSuffix(p, ".md") {
				return nil
			}
			raw, err := fs.ReadFile(root, p)
			if err != nil {
				return err
			}
			page := newPage(s, p, string(raw))
			if other, taken := m.byRoute[page.Route]; taken {
				return fmt.Errorf("docs: %s and %s are both %s", other.File, page.File, page.Route)
			}
			m.pages = append(m.pages, page)
			m.byRoute[page.Route] = page
			m.byFile[page.File] = page
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(m.pages) == 0 {
		return nil, fmt.Errorf("docs: no .md file under content/ypy or content/go")
	}

	for _, p := range m.pages {
		p.render(m, md)
	}

	sort.Slice(m.pages, func(i, j int) bool { return m.pages[i].Route < m.pages[j].Route })
	m.arrange()
	return m, nil
}

// Pages is every page, by address.
func (m *Manual) Pages() []*Page { return m.pages }

// Page is the page at an address, without a trailing slash.
func (m *Manual) Page(route string) (*Page, bool) {
	p, ok := m.byRoute[strings.TrimSuffix(route, "/")]
	return p, ok
}

// ByFile is the page a file became, named as under content/: "ypy/guides/events.md".
func (m *Manual) ByFile(file string) (*Page, bool) {
	p, ok := m.byFile[file]
	return p, ok
}

// Relocated lists the links that were followed to a file that moved: the file
// that holds each one needs the new path written in it.
func (m *Manual) Relocated() []BrokenLink { return m.relocated }

// Broken lists the links between files that point nowhere.
func (m *Manual) Broken() []BrokenLink { return m.broken }

// Tree is the sidebar of a section in a language.
func (m *Manual) Tree(section, lang string) []Group { return m.trees[section+"/"+lang] }

// Neighbours are the pages before and after one, in the order of the sidebar.
func (m *Manual) Neighbours(p *Page) (prev, next *Page) {
	order := m.flat[p.Section.ID+"/"+p.Lang]
	for i, q := range order {
		if q != p {
			continue
		}
		if i > 0 {
			prev = order[i-1]
		}
		if i+1 < len(order) {
			next = order[i+1]
		}
	}
	return prev, next
}

// Twin is the same page in the other language, when it exists.
func (m *Manual) Twin(p *Page) (*Page, bool) {
	parts := strings.Split(strings.TrimPrefix(p.File, p.Section.Dir), "/")
	parts = parts[1:] // the leading empty element
	var twin []string
	if p.Lang == Portuguese {
		twin = append(twin, parts[1:]...)
	} else {
		twin = append(twin, Portuguese)
		twin = append(twin, parts...)
	}
	file := path.Join(append([]string{p.Section.Dir}, twin...)...)
	q, ok := m.byFile[file]
	if !ok && p.Lang == English && len(parts) == 1 {
		// The Portuguese home of a section is its index under pt-BR.
		for _, name := range []string{"index.md", "README.md"} {
			if q, ok = m.byFile[path.Join(p.Section.Dir, Portuguese, name)]; ok {
				break
			}
		}
	}
	if !ok && p.Lang == Portuguese && len(parts) == 2 {
		for _, name := range []string{"index.md", "README.md"} {
			if q, ok = m.byFile[path.Join(p.Section.Dir, name)]; ok {
				break
			}
		}
	}
	return q, ok
}

// Page is one Markdown file, rendered.
type Page struct {
	Section Section
	// File is the path under content/.
	File string
	// Route is the address of the page.
	Route string
	// Lang is "en" or "pt-BR".
	Lang string
	// Group is the first folder under the section, "" for a file at its top.
	Group string
	// Sub is the folders below Group, "" when there are none.
	Sub  string
	Slug string
	// Index is true for index.md and README.md.
	Index bool

	Title       string
	Description string
	Updated     string
	Meta        map[string]string
	Order       int

	// HTML is the body, without its first heading, which is the Title.
	HTML     string
	Headings []markdown.Heading
	Text     string
	Words    int
	Minutes  int

	body string
}

// Contents are the headings a page lists on its side: sections and their
// subsections, which is what the Markdown module returns.
func (p *Page) Contents() []markdown.Heading { return p.Headings }

func newPage(s Section, file, raw string) *Page {
	rel := strings.TrimPrefix(file, s.Dir+"/")
	name := strings.TrimSuffix(path.Base(rel), ".md")
	dir := path.Dir(rel)
	if dir == "." {
		dir = ""
	}

	p := &Page{Section: s, File: file, Lang: English, Slug: name}
	segments := strings.Split(dir, "/")
	if segments[0] == Portuguese {
		p.Lang = Portuguese
		segments = segments[1:]
	}
	dirPath := strings.Join(segments, "/")
	if dirPath != "" {
		p.Group = segments[0]
		p.Sub = strings.Join(segments[1:], "/")
	}

	p.Index = strings.EqualFold(name, "index") || strings.EqualFold(name, "README")
	route := s.Prefix
	if p.Lang == Portuguese {
		route += "/" + Portuguese
	}
	if dirPath != "" {
		route += "/" + dirPath
	}
	if !p.Index {
		route += "/" + name
	}
	p.Route = route

	meta, body := frontMatter(raw)
	p.Meta = meta
	p.Updated = meta["updated"]
	if n, err := strconv.Atoi(meta["order"]); err == nil {
		p.Order = n
	}

	title, body := splitTitle(body)
	if t := meta["title"]; t != "" {
		title = t
	}
	if title == "" {
		title = humanize(name)
	}
	p.Title = title
	p.body = body
	p.Description = meta["description"]
	if p.Description == "" {
		p.Description = firstParagraph(body)
	}
	return p
}

// frontMatter splits the leading --- block into keys and values. It reads
// "key: value" lines only; a list or a nested value is not needed by the site
// and is left out.
func frontMatter(raw string) (map[string]string, string) {
	meta := map[string]string{}
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(raw, "---\n") {
		return meta, raw
	}
	end := strings.Index(raw[4:], "\n---")
	if end < 0 {
		return meta, raw
	}
	block := raw[4 : 4+end]
	rest := raw[4+end+4:]
	rest = strings.TrimPrefix(rest, "\n")

	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		meta[strings.TrimSpace(key)] = value
	}
	return meta, rest
}

// splitTitle takes the first "# " heading out of a body and returns its text.
// The page draws its own <h1>, so the body must not draw a second one.
func splitTitle(body string) (string, string) {
	lines := strings.Split(body, "\n")
	fenced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
		}
		if fenced {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			title := plain(strings.TrimSpace(line[2:]))
			rest := append(append([]string{}, lines[:i]...), lines[i+1:]...)
			return title, strings.Join(rest, "\n")
		}
		break
	}
	return "", body
}

// plain is a line of Markdown as text: no code ticks, no emphasis, a link as its words.
func plain(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	for {
		open := strings.Index(s, "](")
		if open < 0 {
			break
		}
		close := strings.Index(s[open:], ")")
		start := strings.LastIndex(s[:open], "[")
		if close < 0 || start < 0 {
			break
		}
		s = s[:start] + s[start+1:open] + s[open+close+1:]
	}
	return strings.TrimSpace(s)
}

// firstParagraph is the first run of ordinary text of a body, for a description.
func firstParagraph(body string) string {
	var words []string
	fenced := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if trimmed == "" {
			if len(words) > 0 {
				break
			}
			continue
		}
		if len(words) == 0 && skippable(trimmed) {
			continue
		}
		words = append(words, trimmed)
	}
	text := plain(strings.Join(words, " "))
	if len(text) > 200 {
		cut := strings.LastIndex(text[:200], " ")
		if cut < 60 {
			cut = 200
		}
		text = text[:cut] + "..."
	}
	return text
}

// humanize makes a title of a file name: "getting-started" is "Getting started".
func humanize(name string) string {
	name = strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " ")
	if name == "" {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// skippable is a line that is not a sentence: a heading, a quote, a table row,
// a list item, a rule or markup.
func skippable(line string) bool {
	switch line[0] {
	case '#', '>', '|', '-', '*', '<', '!', '=':
		return true
	}
	if line[0] >= '0' && line[0] <= '9' {
		return true
	}
	return strings.HasPrefix(line, "Contents:") || strings.HasPrefix(line, "Portuguese:")
}

// Mentions counts the times a text appears in the Markdown source of the page.
func (p *Page) Mentions(needle string) int { return strings.Count(p.body, needle) }
