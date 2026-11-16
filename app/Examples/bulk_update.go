package examples

import (
	"net/http"
	"strconv"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type person struct {
	ID     int
	Name   string
	Email  string
	Active bool
}

// teamBase is the table every visitor starts from.
var teamBase = []person{
	{1, "Marina Duarte", "marina.duarte@example.com", true},
	{2, "Tomás Ribeiro", "tomas.ribeiro@example.com", true},
	{3, "Helena Prado", "helena.prado@example.com", false},
	{4, "Caio Nogueira", "caio.nogueira@example.com", true},
	{5, "Beatriz Lemos", "beatriz.lemos@example.com", false},
}

// bulkView is what the bulk-update template draws.
type bulkView struct {
	Token   string
	People  []person
	Message string
	Error   string
}

// bulkState keeps the ids a visitor changed from the base: id -> active.
type bulkState map[string]bool

func (d *Demos) bulk(r *http.Request, token, message string) bulkView {
	st := bulkState{}
	d.state.read(r, "bulk-update", &st)
	return bulkFrom(st, token, message)
}

func bulkFrom(st bulkState, token, message string) bulkView {
	people := make([]person, len(teamBase))
	copy(people, teamBase)
	for i := range people {
		if v, ok := st[strconv.Itoa(people[i].ID)]; ok {
			people[i].Active = v
		}
	}
	return bulkView{Token: token, People: people, Message: message}
}

func (d *Demos) bulkUpdate(r *hcontrollers.Router) {
	r.Action("PUT", Prefix+"/bulk-update", func(ctx *hhttp.Context) error {
		in, ok := form(ctx)
		if !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		st := bulkState{}
		d.state.read(ctx.Request, "bulk-update", &st)

		ids := map[string]bool{}
		for _, id := range in["ids"] {
			for _, p := range teamBase {
				if id == strconv.Itoa(p.ID) {
					ids[id] = true
				}
			}
		}
		status := first(in["status"])
		if len(ids) == 0 || (status != "active" && status != "inactive") {
			v := bulkFrom(st, token(ctx), "")
			v.Error = "Tick at least one person first."
			return fragment(ctx, http.StatusUnprocessableEntity, d.render("bulk", v))
		}
		for id := range ids {
			st[id] = status == "active"
		}
		if !d.state.write(ctx.Response, "bulk-update", st) {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		word := "activated"
		if status == "inactive" {
			word = "deactivated"
		}
		msg := strconv.Itoa(len(ids)) + " people " + word + "."
		if len(ids) == 1 {
			msg = "1 person " + word + "."
		}
		return fragment(ctx, http.StatusOK, d.render("bulk", bulkFrom(st, token(ctx), msg)))
	}).Name("examples.bulk-update")
}

const tplBulk = `
{{define "bulk"}}<form id="team" ypy-put="/examples/demo/bulk-update" ypy-target="this" ypy-swap="outerHTML">
  <input type="hidden" name="_token" value="{{.Token}}">
  <table class="table">
    <thead>
      <tr><th scope="col"><span class="sr-only">Select</span></th><th scope="col">Name</th><th scope="col">Email</th><th scope="col">Status</th></tr>
    </thead>
    <tbody>
      {{range .People}}<tr>
        <td><input class="input" type="checkbox" name="ids" value="{{.ID}}" aria-label="Select {{.Name}}"></td>
        <td>{{.Name}}</td>
        <td>{{.Email}}</td>
        <td>{{if .Active}}<span class="badge">Active</span>{{else}}<span class="badge" data-variant="outline">Inactive</span>{{end}}</td>
      </tr>{{end}}
    </tbody>
  </table>
  <div class="ex-actions">
    <button class="btn" data-size="sm" name="status" value="active">Activate</button>
    <button class="btn" data-variant="outline" data-size="sm" name="status" value="inactive">Deactivate</button>
  </div>
  {{if .Error}}<p class="ex-error" role="alert">{{.Error}}</p>{{end}}
  {{if .Message}}<p class="ex-ok" role="status">{{.Message}}</p>{{end}}
</form>{{end}}
`
