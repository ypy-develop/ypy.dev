package docs

import (
	"sort"
	"strings"
)

// Group is a heading of the sidebar and the pages under it.
type Group struct {
	// Key is the folder, "" for the files at the top of a section.
	Key   string
	Title string
	Pages []*Page
	// Subs are the folders below the group, each with its own pages.
	Subs []Sub
}

// Sub is a folder inside a group.
type Sub struct {
	Title string
	Pages []*Page
}

// groupOrder is the order the folders of a section are shown in. A folder that
// is not listed follows them, by name.
var groupOrder = []string{"", "guides", "server", "examples", "extensions", "reference", "migration", "security", "performance", "essays", "contributing"}

// groupTitles are the names of folders that are not their folder name with a capital.
var groupTitles = map[string]string{
	"":             "Start",
	"pt-BR":        "Português",
	"faq":          "FAQ",
	"api":          "API",
	"ypy-security": "Security",
}

// pageOrder puts the pages everyone reads first at the top of their folder. A
// front matter `order` wins over it; every other page follows by file name.
var pageOrder = map[string]int{
	"getting-started": 1,
	"core-concepts":   2,
	"guide":           1,
}

func (p *Page) sortKey() (int, string) {
	if p.Order != 0 {
		return p.Order, p.Slug
	}
	if o, ok := pageOrder[p.Slug]; ok {
		return o, p.Slug
	}
	return 1000, p.Slug
}

func sortPages(pages []*Page) {
	sort.SliceStable(pages, func(i, j int) bool {
		if pages[i].Index != pages[j].Index {
			return pages[i].Index
		}
		oi, ni := pages[i].sortKey()
		oj, nj := pages[j].sortKey()
		if oi != oj {
			return oi < oj
		}
		return ni < nj
	})
}

// arrange builds the sidebar of every section and language, and the reading
// order that "previous" and "next" follow.
func (m *Manual) arrange() {
	for _, s := range Sections {
		for _, lang := range []string{English, Portuguese} {
			key := s.ID + "/" + lang

			byGroup := map[string][]*Page{}
			for _, p := range m.pages {
				if p.Section.ID == s.ID && p.Lang == lang {
					byGroup[p.Group] = append(byGroup[p.Group], p)
				}
			}
			if len(byGroup) == 0 {
				continue
			}

			keys := make([]string, 0, len(byGroup))
			for k := range byGroup {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				ri, rj := rank(keys[i]), rank(keys[j])
				if ri != rj {
					return ri < rj
				}
				return keys[i] < keys[j]
			})

			var tree []Group
			var flat []*Page
			for _, k := range keys {
				g := Group{Key: k, Title: groupTitle(k, lang)}
				direct := []*Page{}
				subs := map[string][]*Page{}
				for _, p := range byGroup[k] {
					if p.Sub == "" {
						direct = append(direct, p)
					} else {
						subs[p.Sub] = append(subs[p.Sub], p)
					}
				}
				sortPages(direct)
				g.Pages = direct
				flat = append(flat, direct...)

				names := make([]string, 0, len(subs))
				for n := range subs {
					names = append(names, n)
				}
				sort.Strings(names)
				for _, n := range names {
					sortPages(subs[n])
					g.Subs = append(g.Subs, Sub{Title: humanize(strings.ReplaceAll(n, "/", " / ")), Pages: subs[n]})
					flat = append(flat, subs[n]...)
				}
				tree = append(tree, g)
			}
			m.trees[key] = tree
			m.flat[key] = flat
		}
	}
}

func rank(key string) int {
	for i, k := range groupOrder {
		if k == key {
			return i
		}
	}
	return len(groupOrder)
}

func groupTitle(key, lang string) string {
	if t, ok := groupTitles[key]; ok {
		if key == "" && lang == Portuguese {
			return "Início"
		}
		return t
	}
	return humanize(key)
}

// Landing is a folder that has no index page of its own: the sidebar group a
// route names, so the address of a folder lists what is in it instead of
// answering 404.
func (m *Manual) Landing(route string) (Group, Section, string, bool) {
	route = strings.TrimSuffix(route, "/")
	for _, s := range Sections {
		for _, lang := range []string{English, Portuguese} {
			for _, g := range m.trees[s.ID+"/"+lang] {
				if g.Key == "" {
					continue
				}
				prefix := s.Prefix
				if lang == Portuguese {
					prefix += "/" + Portuguese
				}
				if prefix+"/"+g.Key == route {
					return g, s, lang, true
				}
			}
		}
	}
	return Group{}, Section{}, "", false
}

// GroupTitle is the name the sidebar gives a folder, in a language: the page's
// trail and the structured data of the page use the same word.
func GroupTitle(key, lang string) string { return groupTitle(key, lang) }
