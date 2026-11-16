package site

import (
	"fmt"
	"strings"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
)

// PageCrumbs is the trail a page of the manual shows above its title: the
// section, the folder it is in and the page itself, last. The page's header and
// its BreadcrumbList are both drawn from this one list, so the data never says
// more than the page does.
func (s *Site) PageCrumbs(p *docs.Page) []Crumb {
	prefix := p.Section.Prefix
	start := prefix
	if p.Lang == docs.Portuguese {
		prefix += "/" + docs.Portuguese
		if _, ok := s.Manual.Page(prefix); ok {
			start = prefix
		}
	}
	crumbs := []Crumb{{Name: p.Section.Title, Href: start}}
	if p.Group != "" {
		crumbs = append(crumbs, Crumb{Name: docs.GroupTitle(p.Group, p.Lang), Href: docs.GroupRoute(p.Section, p.Lang, p.Group)})
	}
	if last := crumbs[len(crumbs)-1]; last.Href == p.Route {
		// The page is the folder's own page, or the section's: it is already the
		// last step.
		crumbs[len(crumbs)-1].Name = p.Title
		return crumbs
	}
	return append(crumbs, Crumb{Name: p.Title, Href: p.Route})
}

// LandingCrumbs is the trail of a folder's page.
func (s *Site) LandingCrumbs(sec docs.Section, lang string, g docs.Group) []Crumb {
	start := sec.Prefix
	if lang == docs.Portuguese {
		if _, ok := s.Manual.Page(sec.Prefix + "/" + docs.Portuguese); ok {
			start = sec.Prefix + "/" + docs.Portuguese
		}
	}
	return []Crumb{{Name: sec.Title, Href: start}, {Name: g.Title, Href: LandingRoute(sec, lang, g)}}
}

// LlmsText is /llms.txt: what YPY is, then the English pages of the manual under
// the heading of the folder they are in, so a model can pick a section. The
// Portuguese pages and the full text come last, under Optional.
func (s *Site) LlmsText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n> %s Version %s, MIT licence.\n\n", s.Name, HomeDescription, s.YPYVersion())
	fmt.Fprintf(&b, "Source code: %s\n", Repository)

	if !s.longReference {
		fmt.Fprintf(&b, "\n## Reference\n\n")
		for _, k := range reference.Kinds {
			fmt.Fprintf(&b, "- [%s](%s): %d names\n", reference.Titles[k], s.AbsoluteURL("/reference/"+string(k)), len(s.Reference.Of(k)))
		}
		fmt.Fprintf(&b, "- [Extensions](%s): %d official extensions\n", s.AbsoluteURL("/reference/extensions"), len(s.Reference.Extensions))
	}

	for _, sec := range docs.Sections {
		for _, g := range s.Manual.Tree(sec.ID, docs.English) {
			heading := g.Title
			switch {
			case sec.ID == "docs" && g.Key == "":
				heading = "Docs"
			case sec.ID == "go" && g.Key == "":
				heading = "Go SDK"
			case sec.ID == "go":
				heading = "Go SDK: " + g.Title
			}
			fmt.Fprintf(&b, "\n## %s\n\n", heading)
			s.llmsPages(&b, g.Pages)
			for _, sub := range g.Subs {
				s.llmsPages(&b, sub.Pages)
			}
		}
	}

	fmt.Fprintf(&b, "\n## Optional\n\n- [YPY documentation, in full](%s): every page below as plain text\n", s.AbsoluteURL("/llms-full.txt"))
	for _, sec := range docs.Sections {
		for _, g := range s.Manual.Tree(sec.ID, docs.Portuguese) {
			s.llmsPages(&b, g.Pages)
			for _, sub := range g.Subs {
				s.llmsPages(&b, sub.Pages)
			}
		}
	}
	return b.String()
}

func (s *Site) llmsPages(b *strings.Builder, pages []*docs.Page) {
	for _, p := range pages {
		line := strings.Join(strings.Fields(p.Description), " ")
		if line != "" {
			line = ": " + line
		}
		title := strings.NewReplacer("[", "\\[", "]", "\\]").Replace(p.Title)
		fmt.Fprintf(b, "- [%s](%s)%s\n", title, s.AbsoluteURL(p.Route), line)
	}
}
