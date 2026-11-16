package examples

import (
	"net/http"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

// contactState is the one contact of the demo: what the visitor saved.
type contactState struct {
	Name  string `json:"n"`
	Email string `json:"e"`
}

// contactView is what the contact templates draw.
type contactView struct {
	Name, Email string
	Token       string
	Errors      map[string]string
}

var defaultContact = contactState{Name: "Marina Duarte", Email: "marina.duarte@example.com"}

func (d *Demos) contact(r *http.Request) contactView {
	c := defaultContact
	d.state.read(r, "click-to-edit", &c)
	return contactView{Name: c.Name, Email: c.Email}
}

func (d *Demos) clickToEdit(r *hcontrollers.Router) {
	base := Prefix + "/click-to-edit"

	r.Action("GET", base, func(ctx *hhttp.Context) error {
		return fragment(ctx, http.StatusOK, d.render("contact-show", d.contact(ctx.Request)))
	}).Name("examples.click-to-edit.show")

	r.Action("GET", base+"/edit", func(ctx *hhttp.Context) error {
		v := d.contact(ctx.Request)
		v.Token = token(ctx)
		return fragment(ctx, http.StatusOK, d.render("contact-edit", v))
	}).Name("examples.click-to-edit.edit")

	r.Action("PUT", base, func(ctx *hhttp.Context) error {
		in, ok := form(ctx)
		if !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		v := contactView{Name: clip(first(in["name"]), 60), Email: clip(first(in["email"]), 80), Token: token(ctx), Errors: map[string]string{}}
		if v.Name == "" {
			v.Errors["name"] = "Write the name of the contact."
		}
		if !plausibleEmail(v.Email) {
			v.Errors["email"] = "Use an address like name@company.com."
		}
		if len(v.Errors) > 0 {
			return fragment(ctx, http.StatusUnprocessableEntity, d.render("contact-edit", v))
		}
		if !d.state.write(ctx.Response, "click-to-edit", contactState{Name: v.Name, Email: v.Email}) {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		return fragment(ctx, http.StatusOK, d.render("contact-show", v))
	}).Name("examples.click-to-edit.save")
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// plausibleEmail is the demo's check: one @, a dot in the domain, no spaces.
func plausibleEmail(s string) bool {
	local, domain, ok := strings.Cut(s, "@")
	return ok && local != "" && !strings.ContainsAny(s, " \t") && strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".") && !strings.Contains(domain, "@")
}

const tplClickToEdit = `
{{define "contact-show"}}<div id="contact" class="ex-card">
  <dl class="ex-dl">
    <dt>Name</dt><dd>{{.Name}}</dd>
    <dt>Email</dt><dd>{{.Email}}</dd>
  </dl>
  <button class="btn" data-variant="outline" data-size="sm" ypy-get="/examples/demo/click-to-edit/edit" ypy-target="#contact" ypy-swap="outerHTML">Edit</button>
</div>{{end}}
{{define "contact-edit"}}<form id="contact" class="ex-card" ypy-put="/examples/demo/click-to-edit" ypy-target="this" ypy-swap="outerHTML">
  <input type="hidden" name="_token" value="{{.Token}}">
  <div class="ex-field">
    <label class="label" for="contact-name">Name</label>
    <input class="input" id="contact-name" name="name" value="{{.Name}}" maxlength="60" autocomplete="off"{{if .Errors.name}} aria-invalid="true" aria-describedby="contact-name-error"{{end}}>
    {{if .Errors.name}}<p class="ex-error" id="contact-name-error">{{.Errors.name}}</p>{{end}}
  </div>
  <div class="ex-field">
    <label class="label" for="contact-email">Email</label>
    <input class="input" id="contact-email" type="email" name="email" value="{{.Email}}" maxlength="80" autocomplete="off"{{if .Errors.email}} aria-invalid="true" aria-describedby="contact-email-error"{{end}}>
    {{if .Errors.email}}<p class="ex-error" id="contact-email-error">{{.Errors.email}}</p>{{end}}
  </div>
  <div class="ex-actions">
    <button class="btn" data-size="sm">Save</button>
    <button class="btn" data-variant="ghost" data-size="sm" type="button" ypy-get="/examples/demo/click-to-edit" ypy-target="#contact" ypy-swap="outerHTML">Cancel</button>
  </div>
</form>{{end}}
`
