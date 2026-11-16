// Package routes is where this application declares what it answers: the
// browser-facing routes in web.go and the console in console.go. There is no
// api.go; the site has no JSON API, only the search index it serves to itself.
package routes

import (
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"

	examples "github.com/ypy-develop/ypy.dev/app/Examples"
	controllers "github.com/ypy-develop/ypy.dev/app/Http/Controllers"
	"github.com/ypy-develop/ypy.dev/assets"
	"github.com/ypy-develop/ypy.dev/public"
)

// Deps carries the controllers the routes dispatch to, so reading
// bootstrap/app.go tells what every route was given.
type Deps struct {
	Home      *controllers.HomeController
	Demos     *examples.Demos
	Docs      *controllers.DocsController
	Reference *controllers.ReferenceController
	Search    *controllers.SearchController
	Machine   *controllers.MachineController
}

// Web registers the routes. Every one is a GET: the site changes nothing, so
// there is no form, no session to guard and no Policy to ask.
//
// The documentation is one handler per section, not one route per file. Which
// addresses exist is the content/ folder's answer; the handler looks the path
// up and draws the 404 itself when it finds nothing.
func Web(r *hcontrollers.Router, d Deps) {
	r.Action("GET", "/{$}", d.Home.Index).Name("home")
	r.Action("GET", "/demo/hello", d.Home.Hello).Name("demo.hello")

	r.Action("GET", "/docs", d.Docs.Show).Name("docs")
	r.Action("GET", "/docs/{path...}", d.Docs.Show).Name("docs.page")
	r.Action("GET", "/go", d.Docs.Show).Name("go")
	r.Action("GET", "/go/{path...}", d.Docs.Show).Name("go.page")

	r.Action("GET", "/reference", d.Reference.Index).Name("reference")
	r.Action("GET", "/reference/{kind}", d.Reference.Table).Name("reference.table")

	d.Demos.Register(r) // /examples/demo/..., the live demos of the Examples section

	r.Action("GET", "/search", d.Search.Page).Name("search")
	r.Action("GET", "/search-index.json", d.Search.Index).Name("search.index")

	// robots.txt, sitemap.xml and llms-full.txt are hesape/geo's, registered by
	// the module listed in bootstrap/app.go; llms.txt is the site's own.
	r.Action("GET", "/llms.txt", d.Machine.Llms).Name("llms")

	public.Routes(r, assets.Icon())

	// Anything no route above claims is the site's own 404, not the router's
	// plain text. "/{$}" and every named route are more specific, so each real
	// address still wins its handler.
	r.Action("GET", "/{missing...}", d.Home.Missing).Name("missing")
}
