---
order: 20
title: Bulk update
description: Update many table rows with one form of checkboxes, then swap the table and fill a status message from the same server response.
---

# Bulk update

A table of users has a checkbox on each row. One button sends the whole form, the server updates every row, and the
answer replaces the table and fills a message area above it.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<div id="toast" role="status" aria-live="polite"></div>

<form ypy-post="/users/active" ypy-target="#users" ypy-swap="outerHTML">
  <table id="users">
    <thead><tr><th>Name</th><th>Email</th><th>Active</th></tr></thead>
    <tbody>
      <tr><td>Ada</td><td>ada@example.com</td>
          <td><input type="checkbox" name="active" value="1" checked></td></tr>
      <tr><td>Grace</td><td>grace@example.com</td>
          <td><input type="checkbox" name="active" value="2"></td></tr>
    </tbody>
  </table>
  <button>Save</button>
</form>
```

Each checkbox is named `active` and its value is the user's id. The browser sends only the checked ones.

## The server (Go)

```go
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	ypy "github.com/ypy-develop/go"
)

var table = template.Must(template.New("t").Parse(`<table id="users">
  <thead><tr><th>Name</th><th>Email</th><th>Active</th></tr></thead>
  <tbody>{{range .}}
    <tr><td>{{.Name}}</td><td>{{.Email}}</td>
        <td><input type="checkbox" name="active" value="{{.ID}}"{{if .Active}} checked{{end}}></td></tr>
  {{end}}</tbody>
</table>`))

func setActive(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		ids := r.PostForm["active"] // only the checked boxes are in the form
		changed := store.SetActive(ids)

		var buf bytes.Buffer
		if err := table.Execute(&buf, store.All()); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		ypy.Fragment(w, buf.String()) // goes to #users
		ypy.Partial(w, ypy.PartialSpec{Target: "#toast", Swap: "innerHTML"},
			fmt.Sprintf("<p>%d users updated.</p>", changed))
	}
}
```

Register it with `mux.HandleFunc("POST /users/active", setActive(store))`.

## How it works

- The `ypy-post` is on the **form**, so the form is the element that makes the request and its fields are the
  parameters. The default trigger of a form is `submit`.
- `ypy-target="#users"` with `ypy-swap="outerHTML"` replaces the whole table, so the server answers with a complete
  `<table id="users">`. Replacing a table, and not its `<tbody>`, keeps the markup valid in every browser.
- The answer has two parts. The first (`ypy.Fragment`) is the main content and goes to the request target. The
  second (`ypy.Partial`) is a `<ypy-partial>` that the client routes to `#toast` on its own. See
  [`<ypy-partial>`](../reference/tags/ypy-partial.md) and [Multiple targets](../guides/multiple-targets.md).
- The message area has `role="status"` and `aria-live="polite"`, so a screen reader announces the change.

## Variations

- To show the message and leave the table alone, answer with `ypy.Partial` only. A response that holds only partials
  does not swap the main target with empty content.
- Send a custom event to the page with `ypy.Trigger(w, "users:updated", map[string]any{"count": changed})` and listen
  for it with `ypy-trigger="users:updated from:body"` on another element.
- A form without JavaScript posts to its `action`. Add `action="/users/active"` and answer a plain browser with
  `ypy.SeeOther(w, "/users")`, using `ypy.FromRequest(r).IsYPY` to tell the two apart.

## See also

- [Click to edit](click-to-edit.md)
- [Updating other content](updating-other-content.md)
- [Forms guide](../guides/forms.md)
