package examples

import (
	"net/http"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type contact struct{ Name, Role, City string }

var directory = []contact{
	{"Marina Duarte", "Designer", "Lisbon"},
	{"Tomás Ribeiro", "Engineer", "Porto"},
	{"Helena Prado", "Product manager", "São Paulo"},
	{"Caio Nogueira", "Support lead", "Recife"},
	{"Beatriz Lemos", "Engineer", "Curitiba"},
	{"Rafael Quintana", "Finance analyst", "Montevideo"},
	{"Inês Carvalho", "Designer", "Coimbra"},
	{"Joaquim Teles", "Sales manager", "Porto Alegre"},
	{"Larissa Moura", "Engineer", "Belo Horizonte"},
	{"Henrique Salgado", "Legal counsel", "Lisbon"},
	{"Camila Ferraz", "Data analyst", "Florianópolis"},
	{"Otávio Brandão", "Support agent", "Salvador"},
}

type searchView struct {
	Query   string
	Results []contact
}

// searchPeople answers the people whose name, role or city holds the query,
// ignoring case. An empty query answers everyone.
func searchPeople(q string) []contact {
	q = strings.ToLower(strings.TrimSpace(q))
	var out []contact
	for _, c := range directory {
		if q == "" || strings.Contains(strings.ToLower(c.Name+" "+c.Role+" "+c.City), q) {
			out = append(out, c)
		}
	}
	return out
}

func (d *Demos) activeSearch(r *hcontrollers.Router) {
	r.Action("GET", Prefix+"/active-search", func(ctx *hhttp.Context) error {
		q := clip(ctx.Query("q"), 40)
		return fragment(ctx, http.StatusOK, d.render("search-results", searchView{Query: q, Results: searchPeople(q)}))
	}).Name("examples.active-search")
}

const tplSearch = `
{{define "search-results"}}{{if .Results}}{{range .Results}}<tr><td>{{.Name}}</td><td>{{.Role}}</td><td>{{.City}}</td></tr>
{{end}}{{else}}<tr><td colspan="3" class="ex-muted">No one matches “{{.Query}}”. Try a name, a role or a city.</td></tr>{{end}}{{end}}
{{define "search"}}<div class="ex-card">
  <label class="label" for="search-q">Search the team</label>
  <input class="input" id="search-q" type="search" name="q" placeholder="Name, role or city" autocomplete="off"
    ypy-get="/examples/demo/active-search" ypy-trigger="input changed delay:300ms, search" ypy-target="#search-results" ypy-indicator="#search-results" ypy-sync="replace">
</div>
<table class="table">
  <thead><tr><th scope="col">Name</th><th scope="col">Role</th><th scope="col">City</th></tr></thead>
  <tbody id="search-results" aria-live="polite">
{{template "search-results" .}}
  </tbody>
</table>{{end}}
`
