---
order: 40
title: Template fragments
description: Define a row, a form or a list once as a named template block, then render it inside the full page and alone as the answer to a YPY request.
---

# Template fragments

A YPY response is often a piece of a page: a row, a form, a list of results. If you write that piece twice, once in
the page template and once for the response, the two copies drift apart. A fragment is a named part of a template. The
page uses it, and a handler can also render it alone.

## The idea

Define the piece once, with a name. The page includes it, and the handler for a YPY request renders only that name.

```html
{{define "page"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Items</title>
  <script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
</head>
<body>
  <h1>Items</h1>
  <form method="post" action="/items" ypy-post="/items" ypy-target="#list" ypy-swap="beforeend">
    <input name="title" aria-label="New item"><button>Add</button>
  </form>
  <ul id="list">{{range .}}{{template "item" .}}{{end}}</ul>
</body>
</html>{{end}}

{{define "item"}}<li id="item-{{.ID}}">{{.Title}}</li>{{end}}
```

The page draws every `item`. The response to the form draws one.

## The handler

With Go's `html/template`, `ExecuteTemplate` picks the block by name. `ypy.Fragment` writes the result as an HTML
answer with the right headers.

```go
package items

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

type Item struct {
	ID    int
	Title string
}

type Items struct {
	tmpl  *template.Template // parsed once, with "page" and "item"
	store *Store
}

func (h *Items) render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	ypy.FragmentStatus(w, status, buf.String())
}

// List answers the page for a browser and the page too for a history restore.
func (h *Items) List(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "page", h.store.All())
}

// Create saves the item and answers only the new row to a YPY request.
func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		http.Error(w, "Write a title.", http.StatusUnprocessableEntity)
		return
	}
	item := h.store.Add(title)
	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/items") // no script: post, redirect, get
		return
	}
	h.render(w, http.StatusOK, "item", item)
}
```

`ypy.FromRequest(r).IsYPY` tells the two cases apart. Wrap the handlers with `ypy.Middleware` so each answer has
`Vary: YPY-Request`. The `Store` is yours.

## Several pieces in one answer

A response can fill more than one place. Render each piece from its block and wrap the extra ones with `ypy.Partial`:

```go
var buf bytes.Buffer
h.tmpl.ExecuteTemplate(&buf, "item", item)
ypy.Fragment(w, buf.String()) // goes to the target of the request

buf.Reset()
h.tmpl.ExecuteTemplate(&buf, "count", h.store.Len())
ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, buf.String())
```

See [Multiple targets](../guides/multiple-targets.md) and [`<ypy-partial>`](../reference/tags/ypy-partial.md).

## Four habits that help

Give a swapped element an `id`. A morph and a partial match on it, and `ypy-target` can point to it.

Keep the fragment self-contained, so it takes its data from its arguments and not from variables of the page.

Leave the page shell out of fragments. If a piece must also work as a full page, the browser path renders `page` and
the YPY path renders the block.

Other template engines have the same idea under other names: Jinja has `{% block %}` and macros, Rails has partials,
Laravel has Blade components and `@include`, and templ has components. The [recipes](../server/recipes/go.md) show
each stack.

## See also

[Progressive enhancement](progressive-enhancement.md) · [Forms](../guides/forms.md) ·
[Examples](../examples/index.md) · [Go recipe](../server/recipes/go.md)
