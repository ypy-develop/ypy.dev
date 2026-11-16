package examples

import (
	"net/http"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

var models = map[string][]string{
	"audi":       {"A3", "A4", "Q5"},
	"toyota":     {"Corolla", "Yaris", "Hilux"},
	"volkswagen": {"Golf", "Polo", "T-Cross"},
}

type modelsView struct{ Models []string }

func (d *Demos) valueSelect(r *hcontrollers.Router) {
	r.Action("GET", Prefix+"/value-select", func(ctx *hhttp.Context) error {
		// An unknown make is not an error: it is the "choose a make first" select.
		return fragment(ctx, http.StatusOK, d.render("models", modelsView{Models: models[ctx.Query("make")]}))
	}).Name("examples.value-select")
}

const tplValueSelect = `
{{define "models"}}{{if .Models}}<select class="select" id="model" name="model">
  <option value="">Choose a model</option>
  {{range .Models}}<option>{{.}}</option>
  {{end}}
</select>{{else}}<select class="select" id="model" name="model" disabled>
  <option value="">Choose a make first</option>
</select>{{end}}{{end}}
{{define "value-select"}}<div class="ex-card">
  <div class="ex-field">
    <label class="label" for="make">Make</label>
    <select class="select" id="make" name="make" ypy-get="/examples/demo/value-select" ypy-target="#model" ypy-swap="outerHTML">
      <option value="">Choose a make</option>
      <option value="audi">Audi</option>
      <option value="toyota">Toyota</option>
      <option value="volkswagen">Volkswagen</option>
    </select>
  </div>
  <div class="ex-field">
    <label class="label" for="model">Model</label>
    {{template "models" .}}
  </div>
</div>{{end}}
`
