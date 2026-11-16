package examples

import (
	"net/http"
	"time"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

// jobLength is how long the pretend export runs.
const jobLength = 6 * time.Second

type jobState struct {
	Start int64 `json:"s"`
}

type jobView struct {
	Token   string
	Percent int
	Done    bool
	Idle    bool
}

// jobAt is the progress of a job that began at start, as of now. It is a
// function of two clocks, so the server keeps no job: a signed cookie carries
// the start, and the percentage is worked out on every poll.
func jobAt(start, now time.Time) jobView {
	pct := int(now.Sub(start) * 100 / jobLength)
	pct = pct / 10 * 10
	if pct >= 100 {
		return jobView{Percent: 100, Done: true}
	}
	return jobView{Percent: max(pct, 0)}
}

func (d *Demos) progressBar(r *hcontrollers.Router) {
	base := Prefix + "/progress-bar"

	r.Action("POST", base+"/start", func(ctx *hhttp.Context) error {
		if _, ok := form(ctx); !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		d.state.write(ctx.Response, "progress-bar", jobState{Start: time.Now().UnixMilli()})
		return fragment(ctx, http.StatusOK, d.render("job", jobView{Token: token(ctx)}))
	}).Name("examples.progress-bar.start")

	r.Action("GET", base+"/progress", func(ctx *hhttp.Context) error {
		var st jobState
		if !d.state.read(ctx.Request, "progress-bar", &st) {
			return fragment(ctx, http.StatusOK, d.render("job-idle", jobView{Token: token(ctx)}))
		}
		v := jobAt(time.UnixMilli(st.Start), time.Now())
		v.Token = token(ctx)
		return fragment(ctx, http.StatusOK, d.render("job", v))
	}).Name("examples.progress-bar.progress")
}

const tplProgress = `
{{define "job"}}{{if .Done}}<div id="job" class="ex-card" role="status">
  <p class="ex-ok">The export is ready: 1,284 rows.</p>
  <div class="ex-actions"><button class="btn" data-variant="outline" data-size="sm" ypy-post="/examples/demo/progress-bar/start" ypy-headers='{"X-CSRF-Token": "{{.Token}}"}' ypy-target="#job" ypy-swap="outerHTML">Start again</button></div>
</div>{{else}}<div id="job" class="ex-card" ypy-get="/examples/demo/progress-bar/progress" ypy-trigger="load delay:600ms" ypy-swap="outerHTML" role="status">
  <progress class="ex-progress" value="{{.Percent}}" max="100" aria-label="Export progress"></progress>
  <p class="ex-muted">{{.Percent}}%</p>
</div>{{end}}{{end}}
{{define "job-idle"}}<div id="job" class="ex-card">
  <button class="btn" data-size="sm" ypy-post="/examples/demo/progress-bar/start" ypy-headers='{"X-CSRF-Token": "{{.Token}}"}' ypy-target="#job" ypy-swap="outerHTML">Start the export</button>
</div>{{end}}
`
