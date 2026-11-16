---
order: 170
title: Dialogs with layers
description: Open a form in a native modal dialog with ypy-layers, keep 422 errors inside it, and close it when the server fires a saved event.
---

# Dialogs with layers

A link opens a form in a modal dialog. The server answers the form without the page around it, shows errors inside the
dialog, and closes it with an event when the data is saved. The `ypy-layers` extension drives the native `<dialog>`
element, so focus handling, Esc and the backdrop come with it.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-layers.min.js"></script>

<a href="/contacts/new" ypy-layer="modal" ypy-layer-close-on="saved" ypy-layer-label="New contact">New contact</a>

<ul id="contacts" ypy-get="/contacts" ypy-trigger="saved from:body" ypy-swap="innerHTML">
  <li>Ada Lovelace</li>
</ul>
```

Without JavaScript the link goes to `/contacts/new` and shows a whole page. With YPY the response fills a modal. When
the `saved` event reaches the page, the layer closes and the list asks the server for its new content.

## The server (Go)

```go
package contacts

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

func fields(name, problem string) string {
	invalid := ""
	msg := ""
	if problem != "" {
		invalid = ` aria-invalid="true" aria-describedby="name-error"`
		msg = `<p id="name-error" role="alert">` + html.EscapeString(problem) + `</p>`
	}
	return `<h2>New contact</h2>
<label>Name <input name="name" value="` + html.EscapeString(name) + `" required` + invalid + `></label>` + msg + `
<button>Save</button> <button type="button" ypy-layer-dismiss="cancelled">Cancel</button>`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	form := `<form action="/contacts" method="post">` + fields("", "") + `</form>`
	if ypy.FromRequest(r).InLayer() {
		ypy.Fragment(w, form) // only the form, the layer is the page
		return
	}
	h.renderPage(w, form) // the same form in a whole page
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		// 422: the content of the form is swapped into the form, in the layer
		ypy.ValidationError(w, fields(name, "Enter a name."))
		return
	}
	c := h.store.Add(name)
	if ypy.FromRequest(r).InLayer() {
		ypy.Trigger(w, "saved", map[string]any{"id": c.ID}) // closes the layer
		ypy.NoContent(w)
		return
	}
	ypy.SeeOther(w, "/contacts")
}
```

`ypy.FromRequest(r).InLayer()` is true when the browser sent the `YPY-Layer` header. `ypy.Middleware` adds
`Vary: YPY-Request` for you; add `Vary: YPY-Layer` too if a cache sits in front of `/contacts/new`.

## How it works

- `ypy-layer="modal"` turns the link into a request that fills a `<dialog>` opened with `showModal()`.
- The form inside has no YPY attributes. A plain form in a layer submits into the layer, and a `422` goes into the form.
- `ypy.Trigger` sets the `YPY-Trigger` header. The `saved` event fires on the form, bubbles to the page and matches
  `ypy-layer-close-on="saved"`. The detail (`{id: 7}`) becomes the value of the layer.
- The list uses `ypy-trigger="saved from:body"` to reload itself when that event bubbles up to the page.

## Variations

- `ypy-layer="drawer"` and `ypy-layer-side="start"` make a side panel. `ypy-layer="popup"` makes a small dialog next to the opener.
- To close after a redirect instead of an event, use `ypy-layer-close-on-url="/contacts/*"`.
- To open a layer from the server, call `ypy.OpenLayer(w, "/contacts/7/confirm-delete", ypy.LayerOptions{Mode: ypy.LayerModal})`.
- To read the result in script: `let result = await ypy.layers.open('/contacts/new', {closeOn: 'saved'})`.

## See also

- [`ypy-layers`](../extensions/layers.md) for every option, event and ARIA detail
- [Layers with the Go SDK](../server/layers.md)
- [Portals: popover, dialog and layers](../guides/portals.md) to choose between a layer, a `popover` and a plain `<dialog>`
- [Confirm and prompt](confirm-and-prompt.md) for the browser's own dialogs
