package examples

import (
	"net/http"
	"slices"
	"strconv"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type invoice struct {
	ID       int
	Customer string
	Amount   string
}

var invoices = []invoice{
	{1, "Harbor Coffee Roasters", "$1,280.00"},
	{2, "Linden Street Books", "$342.50"},
	{3, "Oak & Anvil Woodworks", "$2,975.00"},
	{4, "Blue Heron Dental", "$610.00"},
}

// deletedState is the list of ids this visitor deleted.
type deletedState []int

type deleteView struct {
	Token    string
	Invoices []invoice
}

func (d *Demos) deleteView(r *http.Request, tok string) deleteView {
	var gone deletedState
	d.state.read(r, "delete-row", &gone)
	v := deleteView{Token: tok}
	for _, in := range invoices {
		if !slices.Contains(gone, in.ID) {
			v.Invoices = append(v.Invoices, in)
		}
	}
	return v
}

func (d *Demos) deleteRow(r *hcontrollers.Router) {
	base := Prefix + "/delete-row"

	// The row is deleted with DELETE. A DELETE has no body, so the CSRF token
	// travels in the X-CSRF-Token header the form names with ypy-headers.
	r.Action("DELETE", base+"/{id}", func(ctx *hhttp.Context) error {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id < 1 || id > len(invoices) {
			return fragment(ctx, http.StatusNotFound, "")
		}
		var gone deletedState
		d.state.read(ctx.Request, "delete-row", &gone)
		if !slices.Contains(gone, id) {
			gone = append(gone, id)
		}
		if !d.state.write(ctx.Response, "delete-row", gone) {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		// 200 with an empty body: the row is gone and there is nothing to draw.
		return fragment(ctx, http.StatusOK, "")
	}).Name("examples.delete-row.delete")

	r.Action("POST", base+"/restore", func(ctx *hhttp.Context) error {
		if _, ok := form(ctx); !ok {
			return fragment(ctx, http.StatusRequestEntityTooLarge, "")
		}
		d.state.write(ctx.Response, "delete-row", deletedState{})
		return fragment(ctx, http.StatusOK, d.render("delete-table", deleteView{Token: token(ctx), Invoices: invoices}))
	}).Name("examples.delete-row.restore")
}

const tplDeleteRow = `
{{define "delete-table"}}<div id="invoices">
  <table class="table">
    <thead><tr><th scope="col">Customer</th><th scope="col">Amount</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead>
    <tbody>
      {{range .Invoices}}<tr>
        <td>{{.Customer}}</td>
        <td>{{.Amount}}</td>
        <td>
          <form ypy-delete="/examples/demo/delete-row/{{.ID}}" ypy-headers='{"X-CSRF-Token": "{{$.Token}}"}' ypy-confirm="Delete the invoice for {{.Customer}}?" ypy-target="closest tr" ypy-swap="delete">
            <button class="btn" data-variant="destructive" data-size="sm">Delete</button>
          </form>
        </td>
      </tr>{{end}}
    </tbody>
  </table>
  <form ypy-post="/examples/demo/delete-row/restore" ypy-target="#invoices" ypy-swap="outerHTML">
    <input type="hidden" name="_token" value="{{.Token}}">
    <button class="btn" data-variant="ghost" data-size="sm">Restore the rows</button>
  </form>
</div>{{end}}
`
