// Package controllers answers the requests of the site. None of them writes:
// every route is a GET, the pages are built from the manual and the reference
// loaded at start, and the only thing a request changes is nothing.
package controllers

import (
	"net/http"

	hhttp "github.com/arandu-io/hesape/http"

	site "github.com/ypy-develop/ypy.dev/app/Site"
	errorsview "github.com/ypy-develop/ypy.dev/storage/framework/views/errors"
)

// Controller is what every controller shares: the site, and the not-found page.
type Controller struct {
	site *site.Site
}

// NotFound draws the site's own 404, with the header and the footer, instead of
// the router's plain-text one.
func (c Controller) NotFound(ctx *hhttp.Context) error {
	chrome := c.site.ChromeFor(ctx, site.PageInfo{Path: ctx.Request.URL.Path, Title: "Page not found", NoIndex: true})
	return ctx.ViewStatus(http.StatusNotFound, "errors.404", errorsview.ErrorData{
		Chrome:  chrome,
		Code:    "404",
		Heading: "There is no page here",
		Body:    "The address may have moved or never existed. The documentation index lists every page, and the search finds names.",
	})
}
