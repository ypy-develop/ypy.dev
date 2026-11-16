// Package providers holds this application's service provider: a
// foundation.Module listed in bootstrap/app.go, built by hand over what the
// bootstrap constructed. There is no container; every dependency is passed in.
package providers

import (
	"github.com/arandu-io/framework/foundation"
	"github.com/arandu-io/hesape/geo"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
	markdown "github.com/hyz-is/arandu-markdown"

	examples "github.com/ypy-develop/ypy.dev/app/Examples"
	controllers "github.com/ypy-develop/ypy.dev/app/Http/Controllers"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	"github.com/ypy-develop/ypy.dev/routes"
)

// AppServiceProvider builds the controllers once, over the site, and hands them
// to the routes. It registers no routes itself: those are routes/web.go.
type AppServiceProvider struct {
	// Site is the manual, the reference and the search index, loaded at start.
	Site *site.Site
	// Markdown is the module that renders the pages. It is registered in the
	// bootstrap beside this provider.
	Markdown *markdown.Module
	// Geo serves robots.txt, sitemap.xml and llms-full.txt from the site's public
	// pages. Its configuration is the deployment's: indexing is off unless the
	// environment turns it on, and the origin is APP_URL.
	Geo *geo.Module

	web routes.Deps
}

var _ foundation.Module = (*AppServiceProvider)(nil)

// NewAppServiceProvider builds the controllers over the site.
func NewAppServiceProvider(s *site.Site, md *markdown.Module, demos *examples.Demos) (*AppServiceProvider, error) {
	return &AppServiceProvider{
		Site:     s,
		Markdown: md,
		Geo:      geo.NewModule(s.GeoConfig(), s.Catalog()),
		web: routes.Deps{
			Home:      controllers.NewHomeController(s),
			Demos:     demos,
			Docs:      controllers.NewDocsController(s, demos),
			Reference: controllers.NewReferenceController(s),
			Search:    controllers.NewSearchController(s),
			Machine:   controllers.NewMachineController(s),
		},
	}, nil
}

// Name is the module identifier.
func (*AppServiceProvider) Name() string { return "app" }

// Routes registers nothing: the routes are routes/web.go, handed to WithRouting.
func (*AppServiceProvider) Routes(*hcontrollers.Router) {}

// Web answers what routes/web.go dispatches to.
func (p *AppServiceProvider) Web() routes.Deps { return p.web }
