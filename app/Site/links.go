package site

import (
	"strings"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
)

// DocFor is the page of the manual that explains a name of the reference, or ""
// when there is none yet. The reference tables link to it from every row.
func (s *Site) DocFor(it reference.Item) string { return s.docLinks[string(it.Kind)+"/"+it.Name] }

// linkReference finds, for every row of the tables, the page that explains it.
//
// A row can say so itself (`doc`). Failing that, the page is found by name: a
// page called after the name, then the page of the extension, then the English
// page that mentions the name between backticks the most -- the reference folder
// before the guides. A row nothing mentions has no link, and the table says so
// rather than pointing somewhere vague.
func (s *Site) linkReference() {
	s.docLinks = map[string]string{}

	var pages []*docs.Page
	for _, p := range s.Manual.Pages() {
		if p.Lang == docs.English && p.Section.ID == "docs" {
			pages = append(pages, p)
		}
	}

	byStem := map[string]*docs.Page{}
	s.extGuides = map[string]string{}
	for _, p := range pages {
		if p.Index {
			continue
		}
		if p.Group == "extensions" {
			s.extGuides["ypy-"+p.Slug] = p.Route
			continue
		}
		if _, taken := byStem[p.Slug]; !taken || p.Group == "reference" {
			byStem[p.Slug] = p
		}
	}

	for _, k := range reference.Kinds {
		for _, it := range s.Reference.Of(k) {
			key := string(k) + "/" + it.Name
			if it.Doc != "" {
				if _, ok := s.Manual.Page(strings.SplitN(it.Doc, "#", 2)[0]); ok {
					s.docLinks[key] = it.Doc
					continue
				}
			}
			if p := s.bestPage(pages, byStem, it); p != nil {
				s.docLinks[key] = p.Route
			}
		}
	}
}

func (s *Site) bestPage(pages []*docs.Page, byStem map[string]*docs.Page, it reference.Item) *docs.Page {
	stem := strings.TrimPrefix(it.Name, "ypy-")
	if p, ok := byStem[it.Name]; ok {
		return p
	}
	if it.Kind == reference.Attributes || it.Kind == reference.Elements {
		if p, ok := byStem[stem]; ok && p.Group == "reference" {
			return p
		}
	}
	if it.Extension != "" {
		if route, ok := s.extGuides[it.Extension]; ok {
			p, _ := s.Manual.Page(route)
			return p
		}
	}

	needle := "`" + it.Name
	var best *docs.Page
	bestScore := 0
	for _, p := range pages {
		n := p.Mentions(needle)
		if n == 0 {
			continue
		}
		score := n
		if p.Group == "reference" {
			score += 1000
		}
		if p.Group == "guides" {
			score += 500
		}
		if score > bestScore {
			best, bestScore = p, score
		}
	}
	return best
}

// ExtensionGuide is the page of an extension's own guide, "" when it has none.
func (s *Site) ExtensionGuide(it reference.Item) string { return s.extGuides[it.Name] }
