package examples

import (
	"net/http"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

// taken are the addresses that already have an account in the demo.
var taken = map[string]bool{
	"marina.duarte@example.com": true,
	"admin@example.com":         true,
	"hello@example.com":         true,
}

type signup struct {
	Token string
	Email string
	Error string
	Done  bool
}

// checkEmail is the rule of the demo; it answers "" when the address is fine.
func checkEmail(email string) string {
	switch {
	case email == "":
		return "Enter your email address."
	case !plausibleEmail(email):
		return "Use an address like name@company.com."
	case taken[strings.ToLower(email)]:
		return "This address already has an account."
	}
	return ""
}

func (d *Demos) inlineValidation(r *hcontrollers.Router) {
	base := Prefix + "/inline-validation"

	// The field asks on blur. A 422 carries the message, a 200 clears it.
	r.Action("POST", base+"/email", func(ctx *hhttp.Context) error {
		in, ok := form(ctx)
		if !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		email := clip(first(in["email"]), 80)
		if msg := checkEmail(email); msg != "" {
			status := http.StatusUnprocessableEntity
			return fragment(ctx, status, d.render("signup-msg", signup{Error: msg}))
		}
		return fragment(ctx, http.StatusOK, d.render("signup-msg", signup{Email: email}))
	}).Name("examples.inline-validation.email")

	r.Action("POST", base, func(ctx *hhttp.Context) error {
		in, ok := form(ctx)
		if !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		v := signup{Token: token(ctx), Email: clip(first(in["email"]), 80)}
		if v.Error = checkEmail(v.Email); v.Error != "" {
			return fragment(ctx, http.StatusUnprocessableEntity, d.render("signup", v))
		}
		v.Done = true
		return fragment(ctx, http.StatusOK, d.render("signup", v))
	}).Name("examples.inline-validation.submit")
}

const tplInline = `
{{define "signup-msg"}}{{if .Error}}<p class="ex-error" id="email-msg">{{.Error}}</p>{{else}}<p class="ex-ok" id="email-msg">{{.Email}} is available.</p>{{end}}{{end}}
{{define "signup"}}<form id="signup" class="ex-card" ypy-post="/examples/demo/inline-validation" ypy-target="this" ypy-swap="outerHTML" novalidate>
  <input type="hidden" name="_token" value="{{.Token}}">
  {{if .Done}}<p class="ex-ok" role="status">Account created for {{.Email}}.</p>{{end}}
  <div class="ex-field">
    <label class="label" for="signup-email">Email</label>
    <input class="input" id="signup-email" type="email" name="email" value="{{.Email}}" autocomplete="off" aria-describedby="email-msg"{{if .Error}} aria-invalid="true"{{end}}
      ypy-post="/examples/demo/inline-validation/email" ypy-trigger="change" ypy-target="#email-msg" ypy-swap="outerHTML"
      ypy-error-target="#email-msg" ypy-error-swap="outerHTML">
    {{if .Error}}<p class="ex-error" id="email-msg">{{.Error}}</p>{{else}}<p class="ex-muted" id="email-msg">Try admin@example.com, which is taken.</p>{{end}}
  </div>
  <div class="ex-actions"><button class="btn" data-size="sm">Create account</button></div>
</form>{{end}}
`
