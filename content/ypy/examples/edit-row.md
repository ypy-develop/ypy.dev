---
order: 50
title: Edit row
description: Swap a table row for a row of inputs, save it with ypy-include because a row cannot hold a form, or cancel without touching the server.
---

# Edit row

A click on **Edit** swaps one `<tr>` for a row of inputs. **Save** sends the inputs of that row and swaps the plain row
back. **Cancel** swaps it back without sending anything.

A table row cannot hold a `<form>`, so the **Save** button names the row with `ypy-include` and YPY collects the inputs
inside it.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<table>
  <thead><tr><th>Name</th><th>Email</th><th></th></tr></thead>
  <tbody>
    <tr id="contact-1">
      <td>Ada</td><td>ada@example.com</td>
      <td><button ypy-get="/contacts/1/edit" ypy-target="closest tr" ypy-swap="outerHTML">Edit</button></td>
    </tr>
  </tbody>
</table>
```

The answer to `GET /contacts/1/edit`:

```html
<tr id="contact-1">
  <td><input name="name" value="Ada" aria-label="Name" required></td>
  <td><input name="email" type="email" value="ada@example.com" aria-label="Email" required></td>
  <td>
    <button ypy-put="/contacts/1" ypy-include="closest tr" ypy-target="closest tr" ypy-swap="outerHTML">Save</button>
    <button ypy-get="/contacts/1" ypy-target="closest tr" ypy-swap="outerHTML">Cancel</button>
  </td>
</tr>
```

## The server (Go)

```go
package main

import (
	"bytes"
	"html/template"
	"net/http"

	ypy "github.com/ypy-develop/go"
)

var rowTmpl = template.Must(template.New("r").Parse(`
{{define "show"}}<tr id="contact-{{.ID}}">
  <td>{{.Name}}</td><td>{{.Email}}</td>
  <td><button ypy-get="/contacts/{{.ID}}/edit" ypy-target="closest tr" ypy-swap="outerHTML">Edit</button></td>
</tr>{{end}}
{{define "edit"}}<tr id="contact-{{.ID}}">
  <td><input name="name" value="{{.Name}}" aria-label="Name" required></td>
  <td><input name="email" type="email" value="{{.Email}}" aria-label="Email" required></td>
  <td>
    <button ypy-put="/contacts/{{.ID}}" ypy-include="closest tr" ypy-target="closest tr" ypy-swap="outerHTML">Save</button>
    <button ypy-get="/contacts/{{.ID}}" ypy-target="closest tr" ypy-swap="outerHTML">Cancel</button>
  </td>
</tr>{{end}}`))

func row(w http.ResponseWriter, status int, name string, c Contact) {
	var buf bytes.Buffer
	if err := rowTmpl.ExecuteTemplate(&buf, name, c); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	ypy.FragmentStatus(w, status, buf.String())
}

func routes(mux *http.ServeMux, store *Store) {
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, Contact)) {
		mux.Handle(pattern, ypy.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := store.Get(r.PathValue("id"))
			if !ok {
				http.NotFound(w, r)
				return
			}
			fn(w, r, c)
		})))
	}
	handle("GET /contacts/{id}", func(w http.ResponseWriter, _ *http.Request, c Contact) { row(w, http.StatusOK, "show", c) })
	handle("GET /contacts/{id}/edit", func(w http.ResponseWriter, _ *http.Request, c Contact) { row(w, http.StatusOK, "edit", c) })
	handle("PUT /contacts/{id}", func(w http.ResponseWriter, r *http.Request, c Contact) {
		c.Name, c.Email = r.PostFormValue("name"), r.PostFormValue("email")
		if c.Name == "" {
			row(w, http.StatusUnprocessableEntity, "edit", c) // the edit row again
			return
		}
		store.Save(c)
		row(w, http.StatusOK, "show", c)
	})
}
```

## How it works

- `ypy-target="closest tr"` finds the row that holds the clicked button, and `outerHTML` replaces that row. The
  answer is one `<tr>` with the same `id`.
- On a button, a `PUT` carries only the button's own value. `ypy-include="closest tr"` adds every named input inside
  the row, so `name` and `email` reach the server as a normal form body.
- The server reads the fields with `r.PostFormValue`, saves them, and answers with the plain row.
- **Cancel** is a `GET` of the plain row. It sends nothing and changes nothing on the server.
- A row is not a form, so the browser does not check `required` or `type="email"` before the request. The server
  checks the values and answers `422` with the edit row when one is not valid.

## Variations

- Mark the invalid input in the `422` answer with `aria-invalid="true"` and a message. The `ypy-error-target`
  attributes decide where the answer goes, see [`ypy-error-*`](../reference/attributes/ypy-error.md).
- Ask before throwing the edit away: `ypy-confirm="Discard changes?"` on **Cancel**.
- For a form outside a table, see [Click to edit](click-to-edit.md).

## See also

- [Delete row](delete-row.md)
- [Bulk update](bulk-update.md)
- [Attribute inheritance](../reference/inheritance.md), if you want to write `ypy-target` once on the `<tbody>`
