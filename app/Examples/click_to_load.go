package examples

import (
	"net/http"
	"strconv"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

type order struct {
	Number   string
	Customer string
	Total    string
	Status   string
}

// orders is the whole list; the demo shows it five at a time.
var orders = []order{
	{"#2041", "Harbor Coffee Roasters", "$1,280.00", "Paid"},
	{"#2040", "Linden Street Books", "$342.50", "Paid"},
	{"#2039", "Oak & Anvil Woodworks", "$2,975.00", "Pending"},
	{"#2038", "Blue Heron Dental", "$610.00", "Paid"},
	{"#2037", "Tidewater Print Shop", "$89.90", "Refunded"},
	{"#2036", "Meridian Cycles", "$1,115.25", "Paid"},
	{"#2035", "Fernhill Veterinary", "$480.00", "Pending"},
	{"#2034", "Copperleaf Catering", "$3,420.00", "Paid"},
	{"#2033", "Northgate Pharmacy", "$256.75", "Paid"},
	{"#2032", "Willow & Reed Florists", "$134.00", "Refunded"},
	{"#2031", "Summit Physio", "$720.00", "Paid"},
	{"#2030", "Kestrel Bike Courier", "$95.00", "Pending"},
	{"#2029", "Lakeside Bakery", "$1,008.40", "Paid"},
	{"#2028", "Ironbridge Tools", "$4,210.00", "Paid"},
}

const loadSize = 5

type loadView struct {
	Rows []order
	// Next is the page the button asks for, 0 when there is nothing more.
	Next int
	// Shown and Total feed the count under the button.
	Shown, Total int
}

func loadPage(n int) loadView {
	if n < 1 {
		n = 1
	}
	start := (n - 1) * loadSize
	if start >= len(orders) {
		start = len(orders)
	}
	end := min(start+loadSize, len(orders))
	v := loadView{Rows: orders[start:end], Shown: end, Total: len(orders)}
	if end < len(orders) {
		v.Next = n + 1
	}
	return v
}

func (d *Demos) clickToLoad(r *hcontrollers.Router) {
	r.Action("GET", Prefix+"/click-to-load", func(ctx *hhttp.Context) error {
		page, _ := strconv.Atoi(ctx.Query("page"))
		return fragment(ctx, http.StatusOK, d.render("load-rows", loadPage(page)))
	}).Name("examples.click-to-load")
}

const tplClickToLoad = `
{{define "load-rows"}}{{range .Rows}}<tr>
  <td>{{.Number}}</td><td>{{.Customer}}</td><td>{{.Total}}</td><td>{{.Status}}</td>
</tr>
{{end}}{{if .Next}}<tr id="more-orders">
  <td colspan="4"><button class="btn" data-variant="outline" data-size="sm" ypy-get="/examples/demo/click-to-load?page={{.Next}}" ypy-target="#more-orders" ypy-swap="outerHTML">Load more orders</button> <span class="ex-muted">{{.Shown}} of {{.Total}}</span></td>
</tr>{{else}}<tr id="more-orders"><td colspan="4"><span class="ex-muted">All {{.Total}} orders are shown.</span></td></tr>{{end}}{{end}}
{{define "load-table"}}<table class="table">
  <thead><tr><th scope="col">Order</th><th scope="col">Customer</th><th scope="col">Total</th><th scope="col">Status</th></tr></thead>
  <tbody>
{{template "load-rows" .}}
  </tbody>
</table>{{end}}
`
