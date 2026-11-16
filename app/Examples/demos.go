// Package examples answers the live demos of the Examples section.
//
// A demo is the markup the page teaches, running against real routes of this
// site under /examples/demo/. The routes answer HTML fragments, never pages, so
// they are kept out of the catalogue and carry noindex.
//
// Nothing a visitor does is shared. A demo that has state keeps it in a small
// signed cookie of that visitor, capped in size; there is no database and no
// global variable that two visitors could both write.
package examples

import (
	"bytes"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

// Prefix is where every demo route lives.
const Prefix = "/examples/demo"

// Demos is the set of demos: their templates and the store of visitor state.
type Demos struct {
	tpl   *template.Template
	state store
}

// New builds the demos. key signs the visitor-state cookies; secure marks them
// Secure when the site is served over https.
func New(key []byte, secure bool) *Demos {
	return &Demos{
		tpl:   template.Must(template.New("demos").Funcs(template.FuncMap{"inc": func(i int) int { return i + 1 }}).Parse(templates)),
		state: newStore(key, secure),
	}
}

// Register adds the demo routes: one group, /examples/demo/<example>/...
func (d *Demos) Register(r *hcontrollers.Router) {
	d.clickToEdit(r)
	d.bulkUpdate(r)
	d.clickToLoad(r)
	d.deleteRow(r)
	d.editRow(r)
	d.lazyLoading(r)
	d.inlineValidation(r)
	d.activeSearch(r)
}

// ---- the marker a page uses to say where a demo goes ----------------------

var markerLine = regexp.MustCompile(`(?m)^[ \t]*<!--\s*demo:([a-z0-9-]+)\s*-->[ \t]*$`)
var markerPara = regexp.MustCompile(`<p>YPYDEMOMARKER-([a-z0-9-]+)</p>`)
var markerText = regexp.MustCompile(`YPYDEMOMARKER-[a-z0-9-]+`)

// Prepare turns each `<!-- demo:name -->` line of a page into a paragraph the
// Markdown renderer lets through, so Embed can find it in the HTML.
func Prepare(body string) string {
	return markerLine.ReplaceAllString(body, "\nYPYDEMOMARKER-$1\n")
}

// Unmark takes the marker out of the plain text kept for search.
func Unmark(text string) string { return strings.TrimSpace(markerText.ReplaceAllString(text, "")) }

// Embed replaces each marker paragraph of a rendered page with the demo's HTML
// in the state of this visitor. token is asked for only when the page has a demo.
func (d *Demos) Embed(page string, r *http.Request, token func() string) string {
	if !strings.Contains(page, "YPYDEMOMARKER-") {
		return page
	}
	tok := token()
	return markerPara.ReplaceAllStringFunc(page, func(m string) string {
		name := markerPara.FindStringSubmatch(m)[1]
		inner, ok := d.initial(name, r, tok)
		if !ok {
			return ""
		}
		return `<section class="ex-demo" aria-label="Live demo">` +
			`<p class="ex-demo-label">Live demo</p>` +
			`<div class="ex-demo-stage">` + inner + `</div>` +
			`<p class="ex-demo-note">Real requests to this site. What you change is kept in a cookie of your own for an hour and is invisible to every other visitor.</p>` +
			`</section>`
	})
}

// initial is the first HTML of a demo, as the page shows it.
func (d *Demos) initial(name string, r *http.Request, token string) (string, bool) {
	switch name {
	case "click-to-edit":
		return d.render("contact-show", d.contact(r)), true
	case "bulk-update":
		return d.render("bulk", d.bulk(r, token, "")), true
	case "click-to-load":
		return d.render("load-table", loadPage(1)), true
	case "delete-row":
		return d.render("delete-table", d.deleteView(r, token)), true
	case "edit-row":
		return d.render("edit-table", d.editView(r, token)), true
	case "lazy-loading":
		return d.render("lazy-placeholder", nil), true
	case "inline-validation":
		return d.render("signup", signup{Token: token}), true
	case "active-search":
		return d.render("search", searchView{Results: searchPeople("")}), true
	}
	return "", false
}

// ---- helpers every demo shares -------------------------------------------

func (d *Demos) render(name string, data any) string {
	var b bytes.Buffer
	if err := d.tpl.ExecuteTemplate(&b, name, data); err != nil {
		return "<p>The demo could not be drawn.</p>"
	}
	return b.String()
}

// fragment writes an HTML fragment: the status, Vary: YPY-Request so a cache
// never serves it as a page, and noindex so a crawler that finds the address
// does not keep it.
func fragment(ctx *hhttp.Context, status int, html string) error {
	h := ctx.Response.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Vary", "YPY-Request")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Robots-Tag", "noindex, nofollow")
	ctx.Response.WriteHeader(status)
	_, err := ctx.Response.Write([]byte(html))
	return err
}

// token is the CSRF token the framework issued for this request, for the forms
// an answer draws.
func token(ctx *hhttp.Context) string {
	t, _ := hhttp.CSRFTokenFrom(ctx.Ctx())
	return t
}

// form reads a small form body: anything above 8 KiB is refused by the reader.
func form(ctx *hhttp.Context) (map[string][]string, bool) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Response, ctx.Request.Body, 8<<10)
	if err := ctx.Request.ParseForm(); err != nil {
		return nil, false
	}
	return ctx.Request.Form, true
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
