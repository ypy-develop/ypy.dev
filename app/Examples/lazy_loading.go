package examples

import (
	"net/http"
	"time"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

// lazyDelay is how long the report takes, so the placeholder can be seen.
const lazyDelay = 900 * time.Millisecond

type lazyView struct {
	Rows []struct{ Label, Value string }
}

func (d *Demos) lazyLoading(r *hcontrollers.Router) {
	base := Prefix + "/lazy-loading"

	r.Action("GET", base+"/report", func(ctx *hhttp.Context) error {
		select {
		case <-time.After(lazyDelay):
		case <-ctx.Ctx().Done():
			return nil
		}
		v := lazyView{Rows: []struct{ Label, Value string }{
			{"Orders this week", "128"},
			{"Revenue", "$14,320.50"},
			{"Average order", "$111.88"},
			{"Refunds", "3"},
		}}
		return fragment(ctx, http.StatusOK, d.render("lazy-report", v))
	}).Name("examples.lazy-loading.report")

	// The placeholder again, so the demo can be run more than once.
	r.Action("GET", base+"/placeholder", func(ctx *hhttp.Context) error {
		return fragment(ctx, http.StatusOK, d.render("lazy-placeholder", nil))
	}).Name("examples.lazy-loading.placeholder")
}

const tplLazy = `
{{define "lazy-placeholder"}}<div id="report" class="ex-card" ypy-get="/examples/demo/lazy-loading/report" ypy-trigger="load" ypy-swap="outerHTML">
  <p class="ex-muted" role="status">Loading the weekly report…</p>
  <div class="ex-skeleton" aria-hidden="true">
    <div class="skeleton"></div><div class="skeleton"></div><div class="skeleton"></div>
  </div>
</div>{{end}}
{{define "lazy-report"}}<div id="report" class="ex-card">
  <dl class="ex-dl">
    {{range .Rows}}<dt>{{.Label}}</dt><dd>{{.Value}}</dd>
    {{end}}
  </dl>
  <div class="ex-actions">
    <button class="btn" data-variant="outline" data-size="sm" ypy-get="/examples/demo/lazy-loading/placeholder" ypy-target="#report" ypy-swap="outerHTML">Load it again</button>
  </div>
</div>{{end}}
`
