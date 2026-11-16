package site

import (
	"log/slog"

	fwview "github.com/arandu-io/framework/view"
	"github.com/arandu-io/hesape/geo"
	hhttp "github.com/arandu-io/hesape/http"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	"github.com/ypy-develop/ypy.dev/public"
)

// NavItem is a link of the header.
type NavItem struct {
	Label    string
	Href     string
	External bool
	// Match is the group a page declares to be in when this link is the current one.
	Match string
}

// Nav is the header, in order. The same list is the footer's first column.
func (s *Site) Nav() []NavItem {
	return []NavItem{
		{Label: "Docs", Href: "/docs", Match: "docs"},
		{Label: "Reference", Href: "/reference", Match: "reference"},
		{Label: "Examples", Href: "/docs/examples", Match: "examples"},
		{Label: "Extensions", Href: "/docs/extensions", Match: "extensions"},
		{Label: "Go SDK", Href: "/go", Match: "go"},
		{Label: "Community", Href: Community, External: true},
	}
}

// Kind says what a page is, which decides the entities of its structured data.
type Kind int

const (
	// KindPage is any page that is not one of the two below.
	KindPage Kind = iota
	// KindHome is the front page: it is where the software is described.
	KindHome
	// KindArticle is a page of the manual.
	KindArticle
)

// PageInfo is what a page says about itself, before the site turns it into a
// head: the title without the brand, the description, the language and the
// page's twin in the other one, the date it was last changed and the trail.
type PageInfo struct {
	// Path is the page's own address, "/docs/guides/events".
	Path string
	// Group is the header link the page belongs to; it also picks the cover.
	Group string
	// Title is the page's own name; the brand is added to it. Empty is the
	// front page, which has a title of its own.
	Title       string
	Description string
	// Lang is "en" or "pt-BR". Empty is English.
	Lang string
	// TwinPath and TwinLang name the same page in the other language, when it exists.
	TwinPath string
	TwinLang string
	Kind     Kind
	// Updated is the date the page says it was last changed, "2026-11-15".
	Updated string
	// Crumbs is the trail the page shows above its title, the page last.
	Crumbs []Crumb
	// NoIndex keeps the page out of the index whatever the deployment says: the
	// 404, the search results.
	NoIndex bool
}

// Chrome is what the layout draws around every page: the title, the head, the
// group the header marks as current and the facts of the site. A page's own
// data embeds it.
type Chrome struct {
	fwview.Page

	site  *Site
	Group string
	// Crumbs is the trail above the page's title; the last one is the page.
	Crumbs []Crumb

	meta geo.Metadata
}

// Chrome starts the chrome of a page that has nothing to add to its head:
// title is the page's own name, the brand is added to it.
func (s *Site) Chrome(ctx *hhttp.Context, path, group, title, description string) Chrome {
	return s.ChromeFor(ctx, PageInfo{Path: path, Group: group, Title: title, Description: description})
}

// ChromeFor starts the chrome of a page from what the page says about itself.
//
// The head is built here, once, by hesape/geo: the canonical address, the robots
// line, the language alternates and the social card. On a deployment that is
// not indexed, or for a page that asked not to be, the answer is noindex with no
// canonical, no card and no graph, and the response says so in a header too.
func (s *Site) ChromeFor(ctx *hhttp.Context, in PageInfo) Chrome {
	title := HomeTitle
	if in.Title != "" {
		title = in.Title + " · " + s.Name
	}
	meta := s.metadata(in, title)
	if meta.Canonical == "" {
		ctx.Response.Header().Set("X-Robots-Tag", "noindex, nofollow")
	}

	c := Chrome{Page: fwview.New(ctx, ""), site: s, Group: in.Group, Crumbs: in.Crumbs, meta: meta}
	c.AppName = s.Name
	c.Path = in.Path
	c.Canonical = meta.Canonical
	c.Description = in.Description
	c.Title = title
	return c
}

// metadata asks geo for the page's head. A configuration geo refuses (an
// indexed deployment whose origin is not https) stops the boot, so an error
// here is a path the site wrote wrong; the page is answered as noindex rather
// than as a failure.
func (s *Site) metadata(in PageInfo, title string) geo.Metadata {
	lang := in.Lang
	if lang == "" {
		lang = docs.English
	}
	input := geo.MetadataInput{
		Path:        in.Path,
		Title:       title,
		Description: in.Description,
		Language:    lang,
		Locale:      locale(lang),
		SiteName:    s.Name,
		Indexable:   !in.NoIndex,
	}
	if in.TwinPath != "" {
		input.OpenGraphAlternates = []string{locale(in.TwinLang)}
		input.Alternates = []geo.AlternateInput{{Tag: lang, Path: in.Path}, {Tag: in.TwinLang, Path: in.TwinPath}}
		english := in.Path
		if lang != docs.English {
			english = in.TwinPath
		}
		input.Alternates = append(input.Alternates, geo.AlternateInput{Tag: "x-default", Path: english})
	}
	meta, err := geo.MetadataFor(s.GeoConfig(), input)
	if err != nil {
		slog.Warn("site: the head of a page was refused by geo", "path", in.Path, "error", err)
		input.Indexable = false
		meta, _ = geo.MetadataFor(s.GeoConfig(), input)
		return meta
	}
	if meta.Canonical == "" {
		return meta
	}
	// The cover is asked for with a query that carries its hash, which geo's
	// path check refuses, so the image is set here.
	cover := public.Cover(in.coverKey())
	meta.Image = s.URL + cover.Path
	meta.ImageWidth, meta.ImageHeight = cover.Width, cover.Height
	meta.ImageType, meta.ImageAlt = cover.Type, cover.Alt
	meta.TwitterCard = "summary_large_image"
	return s.graph(meta, in)
}

func (in PageInfo) coverKey() string {
	switch in.Group {
	case "docs", "reference", "examples", "extensions", "go":
		return in.Group
	}
	return ""
}

func locale(lang string) string {
	if lang == docs.Portuguese {
		return "pt_BR"
	}
	return "en_US"
}

// Site is the site the page belongs to.
func (c Chrome) Site() *Site { return c.site }

// Nav is the header.
func (c Chrome) Nav() []NavItem { return c.site.Nav() }

// IsGroup says whether a header link is the one of the page being shown.
func (c Chrome) IsGroup(match string) bool { return match != "" && c.Group == match }

// YPYVersion is the version of YPY the site documents, "0.1.1".
func (c Chrome) YPYVersion() string { return c.site.YPYVersion() }

// GitHub is the organization's address.
func (c Chrome) GitHub() string { return GitHub }

// CommunityURL is where questions go.
func (c Chrome) CommunityURL() string { return Community }

// Robots is the value of the robots meta tag.
func (c Chrome) Robots() string { return c.meta.Robots }

// Head is the page's head as geo built it: canonical address, alternates, robots
// line and social card.
func (c Chrome) Head() geo.Metadata { return c.meta }

// Language is the language the page is written in, for <html lang>.
func (c Chrome) Language() string { return c.meta.Language }
