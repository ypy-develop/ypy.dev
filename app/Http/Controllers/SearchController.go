package controllers

import (
	"net/http"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	"github.com/ypy-develop/ypy.dev/storage/framework/views"
)

// SearchController answers the search page, and serves the index the page script searches.
type SearchController struct {
	Controller
}

// NewSearchController builds it.
func NewSearchController(s *site.Site) *SearchController {
	return &SearchController{Controller{site: s}}
}

// Page is /search?q=, which works with no script.
func (c *SearchController) Page(ctx *hhttp.Context) error {
	query := strings.TrimSpace(ctx.Query("q"))
	chrome := c.site.ChromeFor(ctx, site.PageInfo{Path: "/search", Group: "search", Title: "Search", NoIndex: true})
	return ctx.View("search", views.SearchData{
		Chrome:  chrome,
		Query:   query,
		Results: c.site.Index.Search(query, 40),
	})
}

// searchIndex is the index as the one resource the endpoint answers with.
type searchIndex struct{ entries docs.Index }

func (r searchIndex) ToArray() map[string]any { return map[string]any{"entries": r.entries} }
func (searchIndex) With() map[string]any      { return nil }

// Index is /search-index.json: every page and every reference name, small
// enough to fetch once and search in the browser.
func (c *SearchController) Index(ctx *hhttp.Context) error {
	ctx.Response.Header().Set("Cache-Control", "public, max-age=300")
	return ctx.JSON(http.StatusOK, searchIndex{c.site.Index})
}
