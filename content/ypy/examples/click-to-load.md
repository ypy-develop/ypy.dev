---
order: 30
title: Click to load
description: A Load more button in the last table row fetches the next page of rows and replaces itself, until the server sends the final page.
---

# Click to load

A table shows the first page. The last row holds a **Load more** button. A click asks for the next page, and the answer
replaces that row with the new rows and, when more pages exist, a new button row at the end.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<table>
  <thead><tr><th>Name</th><th>Email</th></tr></thead>
  <tbody id="contacts">
    <tr><td>Ada</td><td>ada@example.com</td></tr>
    <tr><td>Grace</td><td>grace@example.com</td></tr>
    <tr id="more">
      <td colspan="2">
        <button ypy-get="/contacts?page=2" ypy-target="#more" ypy-swap="outerHTML">Load more</button>
      </td>
    </tr>
  </tbody>
</table>
```

## The server (Go)

```go
package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"

	ypy "github.com/ypy-develop/go"
)

var rows = template.Must(template.New("rows").Parse(`{{range .Items}}
<tr><td>{{.Name}}</td><td>{{.Email}}</td></tr>{{end}}
{{if .Next}}<tr id="more"><td colspan="2">
  <button ypy-get="/contacts?page={{.Next}}" ypy-target="#more" ypy-swap="outerHTML">Load more</button>
</td></tr>{{end}}`))

type pageData struct {
	Items []Contact
	Next  int // 0 when this is the last page
}

func contacts(store *Store) http.HandlerFunc {
	const size = 20
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		items, more := store.Page(page, size) // more is true when a later page exists
		data := pageData{Items: items}
		if more {
			data.Next = page + 1
		}
		var buf bytes.Buffer
		if err := rows.Execute(&buf, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		ypy.Fragment(w, buf.String())
	}
}
```

Register it with `mux.Handle("GET /contacts", ypy.Middleware(contacts(store)))`.

## How it works

- The button asks for the next page with `ypy-get` and names the row that holds it as the target. `outerHTML`
  replaces that row, button and all, with the rows in the answer.
- The new answer ends with its own `<tr id="more">` that points at the page after it. The last page leaves it out, and
  the button is gone.
- While the request runs the button carries the `ypy-request` class, because an element without `ypy-indicator` is its
  own indicator. Dim or disable it in CSS. See [CSS classes](../reference/css-classes.md).
- The first page and the later pages come from the same template, so what the user sees on load and after a click
  match.
- The URL carries the page number, so the same route works for a plain link: put
  `<a href="/contacts?page=2">` in the button's place for visitors without JavaScript, or give the server a full-page
  answer when `ypy.FromRequest(r).IsYPY` is false.

## Variations

- Append to a list that stays in place: target the list (`ypy-target="#contacts"`) with `ypy-swap="beforeend"` and put
  the button outside of it. The server then answers only with rows, and has to replace the button some other way, for
  example with a `<ypy-partial>` that targets the button's container.
- Load the next page without a click when the button comes into view: see [Infinite scroll](infinite-scroll.md).
- Stop a double click from asking for the same page twice with `ypy-sync="this:drop"`. See
  [`ypy-sync`](../reference/attributes/ypy-sync.md).

## See also

- [Infinite scroll](infinite-scroll.md)
- [Lazy loading](lazy-loading.md)
- [Swap styles](../reference/swap-and-morph.md)
