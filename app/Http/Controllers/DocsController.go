package controllers

import (
	"strings"

	hhttp "github.com/arandu-io/hesape/http"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	"github.com/ypy-develop/ypy.dev/storage/framework/views"
)

// DocsController answers every page of the manual, whatever its address: the
// address is looked up among the files of content/, so a file added there is a
// route without anything being registered.
type DocsController struct {
	Controller
}

// NewDocsController builds it.
func NewDocsController(s *site.Site) *DocsController {
	return &DocsController{Controller{site: s}}
}

// Show answers /docs, /docs/..., /go and /go/....
func (c *DocsController) Show(ctx *hhttp.Context) error {
	route := strings.TrimSuffix(ctx.Request.URL.Path, "/")

	if page, ok := c.site.Manual.Page(route); ok {
		return c.page(ctx, page)
	}
	if g, section, lang, ok := c.site.Manual.Landing(route); ok {
		return c.landing(ctx, route, g, section, lang)
	}
	// An address from before the pages were gathered into long ones: the page
	// that holds it now, for good, when there is one.
	if to, ok := c.site.Manual.Redirect(route); ok {
		return Moved(ctx, to)
	}
	return c.NotFound(ctx)
}

// groupOf is the header link a page belongs to.
func groupOf(page *docs.Page) string {
	switch {
	case page.Route == "/reference":
		return "reference"
	case docs.Own[page.Group] && page.Lang == docs.English:
		return page.Group
	}
	return page.Section.ID
}

func (c *DocsController) page(ctx *hhttp.Context, page *docs.Page) error {
	m := c.site.Manual
	prev, next := m.Neighbours(page)
	twin, _ := m.Twin(page)

	info := site.PageInfo{
		Path: page.Route, Group: groupOf(page), Title: page.Title, Description: page.Description,
		Lang: page.Lang, Kind: site.KindArticle, Updated: page.Updated, Crumbs: c.site.PageCrumbs(page),
	}
	// hreflang is a cluster in which every page names every page, itself
	// included: a twin that does not name this page back is left out.
	if twin, ok := m.Twin(page); ok {
		if back, ok := m.Twin(twin); ok && back == page {
			info.TwinPath, info.TwinLang = twin.Route, twin.Lang
		}
	}
	chrome := c.site.ChromeFor(ctx, info)

	return ctx.View("doc", views.DocData{
		Chrome:       chrome,
		Page:         page,
		Body:         page.HTML,
		Tree:         m.SectionOf(page),
		SectionTitle: page.Section.Title,
		SectionHref:  page.Section.Prefix,
		Prev:         prev,
		Next:         next,
		Twin:         twin,
	})
}

func (c *DocsController) landing(ctx *hhttp.Context, route string, g docs.Group, s docs.Section, lang string) error {
	group := s.ID
	if docs.Own[g.Key] && lang == docs.English {
		group = g.Key
	}
	chrome := c.site.ChromeFor(ctx, site.PageInfo{
		Path: route, Group: group, Title: g.Title, Description: site.LandingDescription(s, g),
		Lang: lang, Crumbs: c.site.LandingCrumbs(s, lang, g),
	})
	return ctx.View("landing", views.LandingData{
		Chrome:       chrome,
		Title:        g.Title,
		SectionTitle: s.Title,
		SectionHref:  s.Prefix,
		Pages:        g.Pages,
		Subs:         g.Subs,
		Tree:         nil,
	})
}
