package controllers

import (
	"fmt"
	"net/http"

	hhttp "github.com/arandu-io/hesape/http"

	reference "github.com/ypy-develop/ypy.dev/app/Reference"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	"github.com/ypy-develop/ypy.dev/storage/framework/views"
)

// HomeController answers the front page, the 404 for anything unclaimed, and
// the one endpoint the front page's example calls.
type HomeController struct {
	Controller
}

// NewHomeController builds it.
func NewHomeController(s *site.Site) *HomeController {
	return &HomeController{Controller{site: s}}
}

// Index is the front page.
func (c *HomeController) Index(ctx *hhttp.Context) error {
	s := c.site
	v := s.YPYVersion()

	chrome := s.ChromeFor(ctx, site.PageInfo{Path: "/", Group: "home", Description: site.HomeDescription, Kind: site.KindHome})

	return ctx.View("home", views.HomeData{
		Chrome:      chrome,
		Version:     v,
		JsDelivrTag: fmt.Sprintf(`<script src="`+site.CDNJsDelivr+`"></script>`, v),
		YpyTag:      fmt.Sprintf(`<script src="`+site.CDNYpy+`"></script>`, v),
		NpmInstall:  "npm install @ypy-dev/ypy@" + v,
		GoGet:       "go get github.com/ypy-develop/go@" + site.GoSDKVersion,
		Example: `<button ypy-get="/hello" ypy-target="#out">Say hello</button>
<p id="out"></p>`,
		Attributes: len(s.Reference.Of(reference.Attributes)),
		Events:     len(s.Reference.Of(reference.Events)),
		Extensions: len(s.Reference.Extensions),
		Cards: []views.Card{
			{Title: "Docs", Body: "Start here: one long page, from the first attribute to security and testing.", Href: "/docs"},
			{Title: "Reference", Body: "Every attribute, event, header and configuration key, generated from the library.", Href: "/reference"},
			{Title: "Examples", Body: "Click to edit, active search, bulk update, delete a row: complete and short.", Href: "/examples"},
			{Title: "Extensions", Body: "Streaming, WebSocket, history cache, optimistic updates and more, one file each.", Href: "/extensions"},
			{Title: "Essays", Body: "Why YPY is built the way it is: size, security, performance and the choices behind them.", Href: "/essays"},
			{Title: "Community", Body: "Questions, ideas and show-and-tell, in the GitHub discussions.", Href: site.Community, External: true},
			{Title: "GitHub", Body: "The code, the issues and the releases.", Href: site.GitHub, External: true},
		},
	})
}

// Hello answers the example on the front page. A YPY request gets the fragment
// the button asked for; anyone else, who followed the address by hand, gets the
// front page, because a bare fragment is not a page.
func (c *HomeController) Hello(ctx *hhttp.Context) error {
	if !ctx.IsYPY() {
		return ctx.RedirectRoute("home")
	}
	ctx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.Response.Header().Set("Vary", "YPY-Request")
	ctx.Response.WriteHeader(http.StatusOK)
	_, err := ctx.Response.Write([]byte("Hello from the server. This sentence came back as HTML and replaced the text above, and the page did not reload."))
	return err
}

// Missing is every address no route claims.
func (c *HomeController) Missing(ctx *hhttp.Context) error { return c.NotFound(ctx) }
