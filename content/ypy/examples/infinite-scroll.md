---
order: 80
title: Infinite scroll
description: Fetch the next page of rows when the last row scrolls into view. Each answer carries its own next trigger, so the server decides when to stop.
---

# Infinite scroll

The last row of a table asks for the next page as soon as it comes into view. The answer replaces that row with more
rows, and a new last row that asks for the page after it. When no pages are left, the answer ends without one.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<table>
  <thead><tr><th>Name</th><th>Email</th></tr></thead>
  <tbody>
    <tr><td>Ada</td><td>ada@example.com</td></tr>
    <tr><td>Grace</td><td>grace@example.com</td></tr>
    <tr ypy-get="/contacts?page=2" ypy-trigger="revealed" ypy-swap="outerHTML">
      <td colspan="2">Loading more...</td>
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
{{if .Next}}<tr ypy-get="/contacts?page={{.Next}}" ypy-trigger="revealed" ypy-swap="outerHTML">
  <td colspan="2">Loading more...</td>
</tr>{{end}}`))

func contacts(store *Store) http.HandlerFunc {
	const size = 20
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		items, more := store.Page(page, size)
		data := struct {
			Items []Contact
			Next  int
		}{Items: items}
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

- `ypy-trigger="revealed"` fires once, when the row scrolls into the viewport, and the observer then disconnects.
- `ypy-swap="outerHTML"` replaces the row with the answer. The answer holds the next rows and, if more pages exist, a
  new sentinel row with the same attributes pointing at the next page.
- Because the sentinel is replaced, it can fire only once. A fast scroll cannot ask for the same page twice.
- The server decides when to stop: the last page has no sentinel row, so no request follows.

## Variations

- Start loading before the user reaches the end: use `ypy-trigger="intersect once"` on a row some lines before the
  last one, or `intersect root:#scroller threshold:0` when the rows scroll inside a box with that `id`.
  See [`ypy-trigger`](../reference/attributes/index.md).
- Keep a list in place and add to its end: target the list with `ypy-swap="beforeend"` and let the sentinel live outside
  of it. The sentinel then needs a `<ypy-partial>` in the answer to replace itself.
- Visitors without JavaScript see only the first page. Add a normal link to the next page inside the sentinel row,
  so the content is reachable: `<a href="/contacts?page=2">Next page</a>`.
- The [`ypy-list`](../extensions/list.md) extension keeps a growing list bounded, with `limit:<n>` on the swap, so a long
  session does not hold thousands of rows.

## See also

- [Click to load](click-to-load.md), the same pattern with a button
- [Lazy loading](lazy-loading.md)
- [`ypy-list`](../extensions/list.md)
