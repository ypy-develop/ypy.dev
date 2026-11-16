// Package bootstrap configures the application, and nothing else.
//
// It is the one place the application is composed: the drivers the framework
// builds from config/, the Markdown module, the site, the provider, the routes.
// Read it top to bottom and you know what this application is. Nothing is
// resolved from a container and nothing is discovered: every line is a typed call.
package bootstrap

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arandu-io/framework/foundation"
	fwview "github.com/arandu-io/framework/view"
	"github.com/arandu-io/hesape/http/client"
	"github.com/arandu-io/hesape/routing/controllers"
	markdown "github.com/hyz-is/arandu-markdown"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
	examples "github.com/ypy-develop/ypy.dev/app/Examples"
	providers "github.com/ypy-develop/ypy.dev/app/Providers"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	appconfig "github.com/ypy-develop/ypy.dev/config"
	"github.com/ypy-develop/ypy.dev/routes"

	// The compiled stylesheet and the browser files, embedded. Without this
	// import the browser gets the framework's default stylesheet and no script.
	_ "github.com/ypy-develop/ypy.dev/assets"

	// The engine this binary speaks. The site stores nothing, but the framework
	// opens a database handle at boot; it is an in-memory SQLite, which needs no
	// file, no server and no cgo. A setting that names another engine stops the
	// boot and prints the line to add here.
	_ "github.com/arandu-io/hesape/database/connectors/sqlite"

	// Importing the views is what registers them: every generated view calls
	// view.Register from init(). `aru view:build` writes the packages under
	// storage/framework/views/, and a directory a controller does not import
	// needs a line here.
	_ "github.com/ypy-develop/ypy.dev/storage/framework/views"
	_ "github.com/ypy-develop/ypy.dev/storage/framework/views/layouts"
)

// App loads the configuration, builds the drivers and the site over them, and
// configures the application. version and commit are what the build stamped.
func App(version, commit string) (*foundation.Application, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, err
	}

	drivers, err := foundation.NewDrivers(cfg.Settings)
	if err != nil {
		return nil, err
	}

	// GitHub, which the header's star count is read from. Anonymous, against
	// the public REST API, and bounded at ten seconds: nothing waits for the
	// answer, because a scheduled task fetches it and every page reads the last
	// one from memory.
	github := clients.NewGitHubClient(clients.GitHubConfig{
		BaseURL: "https://api.github.com",
		Timeout: 10 * time.Second,
	}, client.NewFactory(nil))

	app, err := NewProvider(cfg, github)
	if err != nil {
		_ = drivers.Close()
		return nil, err
	}

	return Configure(drivers, app).WithVersion(version, commit).Create()
}

// NewProvider builds the site and the provider over it. A test calls it with a
// configuration of its own and a GitHub client that never leaves the process.
func NewProvider(cfg appconfig.Config, github clients.GitHub) (*providers.AppServiceProvider, error) {
	// The ids the layout and the page chrome already use, so no heading takes one.
	md, err := markdown.New(markdown.Config{
		Reserved: []string{"main-content", "site-search-input", "site-search-results", "search-page-input", "demo-out"},
	})
	if err != nil {
		return nil, fmt.Errorf("markdown: %w", err)
	}

	dir := cfg.Site.ContentDir
	if dir == "" {
		dir, err = site.ContentDir()
		if err != nil {
			return nil, err
		}
	}

	s, err := site.New(site.Config{
		Name:     cfg.App.Name,
		URL:      cfg.App.URL,
		Indexing: cfg.Site.Indexing,
		Content:  os.DirFS(dir),
	}, md)
	if err != nil {
		return nil, err
	}
	// The demos of the Examples section sign their visitor-state cookies with a
	// key derived from APP_KEY, and mark them Secure when the site is https.
	demos := examples.New(cfg.Framework.App.Key, strings.HasPrefix(cfg.App.URL, "https://"))
	return providers.NewAppServiceProvider(s, md, github, demos)
}

// Configure is the application: the providers in the order they boot, the
// routes and the commands. App calls it with what it loaded; a test calls it
// with drivers of its own.
//
// The global pipeline is the framework's, composed from the configuration --
// the body limit, the trusted proxies, the security headers with a strict
// content security policy, the request budget -- and this site changes nothing
// about it.
func Configure(drivers *foundation.Drivers, app *providers.AppServiceProvider) *foundation.ApplicationBuilder {
	return foundation.Configure(drivers).
		WithProviders(
			// The view layer: the renderer ctx.View needs and the embedded
			// assets. Without it every page errors and every stylesheet 404s.
			fwview.NewModule(),
			// The site's own controllers.
			app,
			// The Markdown module renders the pages. It registers no route, no
			// table and no loop: a body reaches a page through a controller.
			app.Markdown,
			// robots.txt, sitemap.xml and llms-full.txt, from the same list of
			// pages the site is made of. The framework boots it, so indexing
			// without an https APP_URL stops the start instead of publishing
			// addresses nobody can fetch.
			foundation.Component(app.Geo),
		).
		WithRouting(func(r *controllers.Router) {
			routes.Web(r, app.Web())
		}).
		WithCommands(routes.Commands()...).
		WithSchedule(routes.Schedule(app.Console())...)
}
