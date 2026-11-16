package examples

import (
	"net/http"
	"strconv"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type tab struct {
	ID          int
	Label, Body string
}

var tabList = []tab{
	{1, "Profile", "Marina Duarte, marina.duarte@example.com. Member since March 2024."},
	{2, "Billing", "Plan: Team, 12 seats. Next invoice on 1 December, $240.00."},
	{3, "Team", "4 members: Marina Duarte, Tomás Ribeiro, Helena Prado and Caio Nogueira."},
}

type tabsView struct {
	Selected int
	Tabs     []tab
	Body     string
	// OOB marks the strip as an out-of-band part of an answer.
	OOB bool
}

func tabsFor(n int, oob bool) tabsView {
	return tabsView{Selected: n, Tabs: tabList, Body: tabList[n-1].Body, OOB: oob}
}

func (d *Demos) tabs(r *hcontrollers.Router) {
	r.Action("GET", Prefix+"/tabs/{n}", func(ctx *hhttp.Context) error {
		n, err := strconv.Atoi(ctx.Param("n"))
		if err != nil || n < 1 || n > len(tabList) {
			return fragment(ctx, http.StatusNotFound, "")
		}
		v := tabsFor(n, true)
		return fragment(ctx, http.StatusOK, d.render("tab-panel", v)+"\n"+d.render("tab-strip", v))
	}).Name("examples.tabs")
}

const tplTabs = `
{{define "tab-strip"}}<div role="tablist" id="tab-strip" data-variant="line"{{if .OOB}} ypy-swap-oob="true"{{end}} ypy-target:inherited="#tab-panel" ypy-swap:inherited="outerHTML">
  {{range .Tabs}}<button role="tab" id="tab-{{.ID}}" aria-selected="{{if eq .ID $.Selected}}true{{else}}false{{end}}" aria-controls="tab-panel" ypy-get="/examples/demo/tabs/{{.ID}}">{{.Label}}</button>
  {{end}}
</div>{{end}}
{{define "tab-panel"}}<div role="tabpanel" id="tab-panel" aria-labelledby="tab-{{.Selected}}">
  <p>{{.Body}}</p>
</div>{{end}}
{{define "tabs"}}<div class="tabs">
{{template "tab-strip" .}}
{{template "tab-panel" .}}
</div>{{end}}
`
