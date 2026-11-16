// Package providers holds this application's service provider: a
// foundation.Module listed in bootstrap/app.go, built by hand over what the
// bootstrap constructed. There is no container; every dependency is passed in.
package providers

import (
	"context"
	"log/slog"
	"time"

	"github.com/arandu-io/framework/foundation"
	hfoundation "github.com/arandu-io/hesape/foundation"
	"github.com/arandu-io/hesape/geo"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
	markdown "github.com/hyz-is/arandu-markdown"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
	controllers "github.com/ypy-develop/ypy.dev/app/Http/Controllers"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	stars "github.com/ypy-develop/ypy.dev/app/Stars"
	"github.com/ypy-develop/ypy.dev/routes"
)

// StartupRefreshTimeout bounds the star count fetched before the server takes
// traffic. An unavailable API costs the start at most this long.
const StartupRefreshTimeout = 5 * time.Second

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
	// GitHub is the client the header's star count is read through.
	GitHub clients.GitHub

	web routes.Deps
}

// The interfaces this provider implements, asserted at compile time so a typo
// in a method name fails the build instead of silently doing nothing.
var (
	_ foundation.Module      = (*AppServiceProvider)(nil)
	_ hfoundation.Background = (*AppServiceProvider)(nil)
)

// NewAppServiceProvider builds the controllers over the site. github is the
// client the star count is read through: bootstrap configures it, a test hands
// over a fake.
func NewAppServiceProvider(s *site.Site, md *markdown.Module, github clients.GitHub) (*AppServiceProvider, error) {
	docs := controllers.NewDocsController(s)
	return &AppServiceProvider{
		Site:     s,
		Markdown: md,
		GitHub:   github,
		Geo:      geo.NewModule(s.GeoConfig(), s.Catalog()),
		web: routes.Deps{
			Home:      controllers.NewHomeController(s),
			Docs:      docs,
			Reference: controllers.NewReferenceController(s, docs),
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

// Console answers what routes/console.go's schedule reaches.
func (p *AppServiceProvider) Console() routes.ConsoleDeps {
	return routes.ConsoleDeps{GitHub: p.GitHub}
}

// Start primes the star count before the server takes traffic, in the process
// that serves and in no other: every command boots, and only serve starts.
//
// The refresh is best effort. A failure is logged and the process still starts
// with no count, and the bounded context keeps an unavailable API out of the
// startup path; the scheduled task tries again within the hour.
func (p *AppServiceProvider) Start(ctx context.Context) error {
	refresh, cancel := context.WithTimeout(ctx, StartupRefreshTimeout)
	defer cancel()
	if err := stars.Refresh(refresh, p.GitHub); err != nil {
		slog.Default().Warn("the star count was not fetched at start", "error", err)
	}
	return nil
}
