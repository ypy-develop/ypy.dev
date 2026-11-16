---
order: 10
title: Click to edit
description: Show a contact as text, swap in an edit form on click, and swap the text back after Save or Cancel, with no page reload and no script.
---

# Click to edit

A contact is shown as plain text. A click on **Edit** replaces it with a form, and **Save** replaces the form with the
updated text. The page does not reload and nothing outside the block changes.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<div id="contact-1">
  <p>First name: Ada</p>
  <p>Email: ada@example.com</p>
  <button ypy-get="/contacts/1/edit" ypy-target="#contact-1" ypy-swap="outerHTML">Edit</button>
</div>
```

The server answers `GET /contacts/1/edit` with a form that has the same `id`:

```html
<form id="contact-1" ypy-put="/contacts/1" ypy-target="this" ypy-swap="outerHTML">
  <label>First name <input name="first" value="Ada" required></label>
  <label>Email <input name="email" type="email" value="ada@example.com" required></label>
  <button>Save</button>
  <button type="button" ypy-get="/contacts/1" ypy-target="#contact-1" ypy-swap="outerHTML">Cancel</button>
</form>
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

var tmpl = template.Must(template.New("c").Parse(`
{{define "show"}}<div id="contact-{{.ID}}">
  <p>First name: {{.First}}</p><p>Email: {{.Email}}</p>
  <button ypy-get="/contacts/{{.ID}}/edit" ypy-target="#contact-{{.ID}}" ypy-swap="outerHTML">Edit</button>
</div>{{end}}
{{define "edit"}}<form id="contact-{{.ID}}" ypy-put="/contacts/{{.ID}}" ypy-target="this" ypy-swap="outerHTML">
  <label>First name <input name="first" value="{{.First}}" required></label>
  <label>Email <input name="email" type="email" value="{{.Email}}" required></label>
  <button>Save</button>
  <button type="button" ypy-get="/contacts/{{.ID}}" ypy-target="#contact-{{.ID}}" ypy-swap="outerHTML">Cancel</button>
</form>{{end}}`))

func render(w http.ResponseWriter, name string, data any, status int) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	ypy.FragmentStatus(w, status, buf.String())
}

func routes(mux *http.ServeMux, store *Store) {
	mux.HandleFunc("GET /contacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		c, ok := store.Get(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		render(w, "show", c, http.StatusOK)
	})
	mux.HandleFunc("GET /contacts/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		c, ok := store.Get(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		render(w, "edit", c, http.StatusOK)
	})
	mux.HandleFunc("PUT /contacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		c, ok := store.Get(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		c.First, c.Email = r.PostFormValue("first"), r.PostFormValue("email")
		if c.First == "" {
			render(w, "edit", c, http.StatusUnprocessableEntity) // same form again
			return
		}
		store.Save(c)
		render(w, "show", c, http.StatusOK)
	})
}
```

Wrap the mux with `ypy.Middleware` where you build the server, so every answer carries `Vary: YPY-Request`.

## How it works

- `ypy-get` on the button asks for the edit form. `ypy-target="#contact-1"` names the block to replace and
  `ypy-swap="outerHTML"` replaces the element itself, not only what is inside it. Without `ypy-swap` the default is
  `innerMorph`, which would keep the block and morph its children.
- The form carries `ypy-put`. A form sends its fields on `submit`, so no `ypy-trigger` is needed. `ypy-target="this"`
  points at the form itself, and `outerHTML` swaps it for the read-only block.
- The edit form has the same `id` as the block it replaces, so the next click on **Edit** finds it again.
- **Cancel** is `type="button"`, so it does not submit the form. It asks for the read-only block with `ypy-get`.
- A `PUT` body is read with `r.PostFormValue`, as for `POST`. The 422 branch returns the form with its values; YPY
  swaps it in the way the target says. See [Forms](../guides/forms.md) for the error convention.

## Variations

- Use `ypy-patch` instead of `ypy-put` if your API treats the edit as a partial update.
- To show an error next to a field instead of redrawing the whole form, use
  [`ypy-error-target`](../reference/attributes/ypy-error.md).
- Add `ypy-confirm="Discard changes?"` to **Cancel** when the form is long.

## See also

- [Edit row](edit-row.md), the same idea inside a table row
- [Inline validation](inline-validation.md)
- [Attributes](../reference/attributes/index.md)
