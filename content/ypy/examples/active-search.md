---
order: 90
title: Active search
description: Search a table as the user types: a 300 ms pause, a loading indicator, and no chance for a slow old answer to overwrite a newer one.
---

# Active search

A search box that asks the server after each pause in typing and puts the matching rows under it. The page never
reloads, and a slow answer to an old query cannot overwrite a newer one.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<h2>Contacts <span id="searching" class="ypy-indicator">Searching...</span></h2>

<input type="search" name="q" placeholder="Search by name or email" autocomplete="off"
       ypy-get="/contacts"
       ypy-trigger="input changed delay:300ms, search"
       ypy-target="#results"
       ypy-indicator="#searching"
       ypy-sync="replace">

<table>
  <thead><tr><th>Name</th><th>Email</th></tr></thead>
  <tbody id="results">
    <!-- the server renders the first rows with the page -->
  </tbody>
</table>
```

## The server (Go)

```go
package contacts

import (
	"html/template"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

type Contact struct{ Name, Email string }

type Handler struct{ All []Contact }

var rows = template.Must(template.New("rows").Parse(
	`{{range .}}<tr><td>{{.Name}}</td><td>{{.Email}}</td></tr>
{{else}}<tr><td colspan="2">No contact matches.</td></tr>{{end}}`))

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	var found []Contact
	for _, c := range h.All {
		if strings.Contains(strings.ToLower(c.Name+" "+c.Email), q) {
			found = append(found, c)
		}
	}
	var b strings.Builder
	if err := rows.Execute(&b, found); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ypy.Fragment(w, b.String())
}

func Routes(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /contacts", h.Search)
	return ypy.Middleware(mux)
}
```

`html/template` escapes every name and email. `ypy.Middleware` adds `Vary: YPY-Request`, so a cache never mixes
the rows with a full page.

## How it works

- `ypy-get="/contacts"` on the input sends `GET /contacts?q=<what is typed>`. An input sends its own value under its
  `name`.
- `ypy-trigger="input changed delay:300ms, search"` has two triggers. `input` fires on each keystroke, `changed`
  skips it when the value is the same as the last request, and `delay:300ms` waits for a pause in typing. `search`
  is the browser's own event for a search field (Enter, and the clear button in browsers that have one).
- `ypy-target="#results"` puts the answer in the table body. The default swap is `innerMorph`, which keeps rows that
  did not change.
- `ypy-indicator="#searching"` puts the `ypy-request` class on `#searching` while the request runs. The
  `ypy-indicator` class on the element is hidden until then (see [styling](../index.md#css-transitions)).
- `ypy-sync="replace"` aborts the request in flight when a new one starts. See
  [`ypy-sync`](../index.md#synchronization) for the other strategies; `latest` lets old requests finish and
  drops their answers.

## Variations

- Works without scripts. Wrap the input in `<form action="/contacts">` and, in the handler, answer a full page
  when `ypy.FromRequest(r).IsYPY` is false. Put `ypy-get` and the triggers on the form.
- Keep the address in step. Add `ypy-push-url="true"` to the input to give each search its own URL, and
  render the page from `q` on the server so the address can be opened again.
- Show a count. Answer with a `<ypy-partial id="count">` next to the rows (see [several places](updating-other-content.md)).

## See also

- [Triggers](../index.md#requests) and [`ypy-sync`](../index.md#synchronization)
- [Inline validation](inline-validation.md), which uses the same delay on a field
