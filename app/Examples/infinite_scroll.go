package examples

import (
	"net/http"
	"strconv"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"
	hcontrollers "github.com/arandu-io/hesape/routing/controllers"
)

const scrollSize = 9
const scrollPages = 3

var firstNames = []string{"Marina", "Tomás", "Helena", "Caio", "Beatriz", "Rafael", "Inês", "Joaquim", "Larissa"}
var lastNames = []string{"Duarte", "Ribeiro", "Prado", "Nogueira", "Lemos", "Quintana", "Carvalho", "Teles", "Moura"}

type scrollRow struct{ Name, Email string }

type scrollView struct {
	Rows []scrollRow
	// Next is the page the last row asks for, 0 on the last page.
	Next int
}

// scrollPage is the n-th page of 27 contacts. The same page is the same rows
// for every visitor: nothing here is state.
func scrollPage(n int) scrollView {
	v := scrollView{}
	if n < 1 || n > scrollPages {
		return v
	}
	for i := (n - 1) * scrollSize; i < n*scrollSize; i++ {
		first, last := firstNames[i%len(firstNames)], lastNames[(i/len(firstNames)+i)%len(lastNames)]
		email := strings.ToLower(first+"."+last) + "@example.com"
		v.Rows = append(v.Rows, scrollRow{first + " " + last, asciiFold(email)})
	}
	if n < scrollPages {
		v.Next = n + 1
	}
	return v
}

func asciiFold(s string) string {
	return strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ã", "a", "õ", "o", "ç", "c", "ê", "e", "ô", "o").Replace(s)
}

func (d *Demos) infiniteScroll(r *hcontrollers.Router) {
	r.Action("GET", Prefix+"/infinite-scroll", func(ctx *hhttp.Context) error {
		page, _ := strconv.Atoi(ctx.Query("page"))
		return fragment(ctx, http.StatusOK, d.render("scroll-rows", scrollPage(page)))
	}).Name("examples.infinite-scroll")
}

const tplScroll = `
{{define "scroll-rows"}}{{$n := len .Rows}}{{range $i, $r := .Rows}}{{if and (eq (inc $i) $n) $.Next}}<tr ypy-get="/examples/demo/infinite-scroll?page={{$.Next}}" ypy-trigger="revealed" ypy-swap="afterend">
  <td>{{$r.Name}}</td><td>{{$r.Email}}</td>
</tr>
{{else}}<tr><td>{{$r.Name}}</td><td>{{$r.Email}}</td></tr>
{{end}}{{end}}{{end}}
{{define "scroll-table"}}<div class="ex-scroll" role="region" aria-label="Contacts" tabindex="0">
<table class="table">
  <thead><tr><th scope="col">Name</th><th scope="col">Email</th></tr></thead>
  <tbody>
{{template "scroll-rows" .}}
  </tbody>
</table>
</div>{{end}}
`
