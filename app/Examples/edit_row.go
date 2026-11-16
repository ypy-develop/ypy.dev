package examples

import (
	"net/http"
	"strconv"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type member struct {
	ID    int
	Name  string
	Role  string
	Token string
	Error string
}

var team = []member{
	{ID: 1, Name: "Marina Duarte", Role: "Designer"},
	{ID: 2, Name: "Tomás Ribeiro", Role: "Engineer"},
	{ID: 3, Name: "Helena Prado", Role: "Product manager"},
	{ID: 4, Name: "Caio Nogueira", Role: "Support lead"},
}

// editedState is the rows this visitor changed: id -> [name, role].
type editedState map[string][2]string

type editView struct {
	Token   string
	Members []member
}

func (d *Demos) member(r *http.Request, id int) (member, bool) {
	if id < 1 || id > len(team) {
		return member{}, false
	}
	m := team[id-1]
	st := editedState{}
	d.state.read(r, "edit-row", &st)
	if v, ok := st[strconv.Itoa(id)]; ok {
		m.Name, m.Role = v[0], v[1]
	}
	return m, true
}

func (d *Demos) editView(r *http.Request, tok string) editView {
	v := editView{Token: tok}
	for i := range team {
		m, _ := d.member(r, i+1)
		v.Members = append(v.Members, m)
	}
	return v
}

func (d *Demos) editRow(r *hcontrollers.Router) {
	base := Prefix + "/edit-row/{id}"

	r.Action("GET", base, func(ctx *hhttp.Context) error {
		id, _ := strconv.Atoi(ctx.Param("id"))
		m, ok := d.member(ctx.Request, id)
		if !ok {
			return fragment(ctx, http.StatusNotFound, "")
		}
		return fragment(ctx, http.StatusOK, d.render("edit-show", m))
	}).Name("examples.edit-row.show")

	r.Action("GET", base+"/edit", func(ctx *hhttp.Context) error {
		id, _ := strconv.Atoi(ctx.Param("id"))
		m, ok := d.member(ctx.Request, id)
		if !ok {
			return fragment(ctx, http.StatusNotFound, "")
		}
		m.Token = token(ctx)
		return fragment(ctx, http.StatusOK, d.render("edit-form", m))
	}).Name("examples.edit-row.edit")

	r.Action("PUT", base, func(ctx *hhttp.Context) error {
		id, _ := strconv.Atoi(ctx.Param("id"))
		m, ok := d.member(ctx.Request, id)
		if !ok {
			return fragment(ctx, http.StatusNotFound, "")
		}
		in, ok := form(ctx)
		if !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		m.Name, m.Role, m.Token = clip(first(in["name"]), 40), clip(first(in["role"]), 40), token(ctx)
		if m.Name == "" || m.Role == "" {
			m.Error = "Name and role cannot be empty."
			return fragment(ctx, http.StatusUnprocessableEntity, d.render("edit-form", m))
		}
		st := editedState{}
		d.state.read(ctx.Request, "edit-row", &st)
		st[strconv.Itoa(id)] = [2]string{m.Name, m.Role}
		if !d.state.write(ctx.Response, "edit-row", st) {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		return fragment(ctx, http.StatusOK, d.render("edit-show", m))
	}).Name("examples.edit-row.save")
}

const tplEditRow = `
{{define "edit-show"}}<tr id="member-{{.ID}}">
  <td>{{.Name}}</td>
  <td>{{.Role}}</td>
  <td><button class="btn" data-variant="outline" data-size="sm" ypy-get="/examples/demo/edit-row/{{.ID}}/edit" ypy-target="#member-{{.ID}}" ypy-swap="outerHTML">Edit</button></td>
</tr>{{end}}
{{define "edit-form"}}<tr id="member-{{.ID}}">
  <td>
    <input type="hidden" name="_token" value="{{.Token}}">
    <input class="input" name="name" value="{{.Name}}" maxlength="40" aria-label="Name"{{if .Error}} aria-invalid="true" aria-describedby="member-{{.ID}}-error"{{end}}>
  </td>
  <td>
    <input class="input" name="role" value="{{.Role}}" maxlength="40" aria-label="Role"{{if .Error}} aria-invalid="true"{{end}}>
    {{if .Error}}<p class="ex-error" id="member-{{.ID}}-error">{{.Error}}</p>{{end}}
  </td>
  <td>
    <div class="ex-actions">
      <button class="btn" data-size="sm" ypy-put="/examples/demo/edit-row/{{.ID}}" ypy-include="closest tr" ypy-target="#member-{{.ID}}" ypy-swap="outerHTML">Save</button>
      <button class="btn" data-variant="ghost" data-size="sm" ypy-get="/examples/demo/edit-row/{{.ID}}" ypy-target="#member-{{.ID}}" ypy-swap="outerHTML">Cancel</button>
    </div>
  </td>
</tr>{{end}}
{{define "edit-table"}}<table class="table">
  <thead><tr><th scope="col">Name</th><th scope="col">Role</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead>
  <tbody>
    {{range .Members}}{{template "edit-show" .}}
    {{end}}
  </tbody>
</table>{{end}}
`
