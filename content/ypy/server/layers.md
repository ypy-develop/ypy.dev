---
order: 70
title: "Layers with the Go SDK"
description: "Answer layer requests, open a layer from a response and close one with an event."
---

# Layers with the Go SDK

How a Go server drives the [`ypy-layers`](../extensions/layers.md) extension: it answers the fragment a layer shows,
tells a layer request from a page request, opens a layer from a response, and closes one with an event.

## 1. The fragment of a layer

The browser sends `YPY-Layer: modal|drawer|popup` with every request that fills a layer or starts inside one. Answer
the content only, without the page chrome, and vary on the header:

```go
func newItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "YPY-Layer")
	if ypy.FromRequest(r).InLayer() { // YPY-Layer is set
		_ = ypy.Fragment(w, itemForm(Item{}, nil)) // the form only, in the layer
		return
	}
	renderPage(w, itemForm(Item{}, nil)) // the same form in a whole page: no JavaScript, same URL
}
```

```html
<a href="/items/new" ypy-layer="modal" ypy-layer-close-on="saved">New item</a>
```

## 2. Save: 422 stays in the layer, success closes it

A form in a layer submits into the layer. Answer `422` with the form to show the errors (the layer stays open, the form is
re-rendered). On success, fire the event the opener listens to; the layer accepts with the event detail as its value:

```go
func createItem(w http.ResponseWriter, r *http.Request) {
	item, errs := parseItem(r)
	if errs != nil {
		_ = ypy.ValidationError(w, itemForm(item, errs)) // 422
		return
	}
	saved := store.Create(item)
	if ypy.FromRequest(r).InLayer() {
		_ = ypy.Trigger(w, "saved", map[string]any{"id": saved.ID}) // closes the layer
		ypy.NoContent(w)
		return
	}
	ypy.SeeOther(w, "/items/"+saved.ID) // not in a layer: a normal redirect
}
```

Without the `saved` event, `ypy-layer-close-on-url="/items/*"` does the same from a redirect: the layer follows the
`303` to `/items/7`, sees the matching path and is accepted with `/items/7` as its value, without swapping that page in.

## 3. Open a layer from a response: `YPY-Open-Layer`

`ypy.OpenLayer` sets the header. It only sets the header: the status and the body are yours (usually `ypy.NoContent`).
Options go in a `ypy.LayerOptions`; what you leave out keeps the client's configuration.

```go
func askDelete(w http.ResponseWriter, r *http.Request) {
	err := ypy.OpenLayer(w, "/items/7/confirm-delete", ypy.LayerOptions{
		Mode:    ypy.LayerModal,
		CloseOn: []string{"deleted"},
		Label:   "Delete item",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ypy.NoContent(w)
}
```

With no options and a plain ASCII URL the header is the bare URL (`YPY-Open-Layer: /items/7/confirm-delete`); otherwise it
is a JSON object, always ASCII (non-ASCII characters are `\u` escapes, as headers are Latin-1). `ypy.OpenLayerHTML(w, markup)`
opens a layer with markup you already have, in the header: keep it small and escape user data.

| `LayerOptions` field | Header key | Notes |
|---|---|---|
| `Mode` | `mode` | `ypy.LayerModal`, `ypy.LayerDrawer`, `ypy.LayerPopup` |
| `Side` | `side` | `ypy.SideEnd` (default), `SideStart`, `SideTop`, `SideBottom`; drawers only |
| `CloseOn` | `closeOn` | event names that accept the layer |
| `CloseOnURL` | `closeOnUrl` | path patterns, `*` is a wildcard |
| `CloseOnFragment` | `closeOnFragment` | a selector |
| `History`, `Outside`, `Escape` | `history`, `outside`, `escape` | `*bool`: nil keeps the client's setting |
| `Label`, `Class` | `label`, `class` | `aria-label` and classes of the dialog |

`OpenLayer` returns an error, and sets nothing, for an empty URL, a control character, a mode or a side the client does
not know, or an event name with a separator. A layer opened this way from a request made inside another layer goes on
top of it.

## 4. Test it without a browser

```go
res := ypytest.Do(handler, ypytest.NewRequest("GET", "/items/new", ypytest.Layer(ypy.LayerModal)))
res.AssertVary(t, ypy.HeaderLayer)

res = ypytest.Do(handler, ypytest.NewRequest("POST", "/items/7/delete"))
res.AssertOpenLayer(t, "/items/7/confirm-delete")
if o := res.OpenLayer(t); o.Mode != ypy.LayerModal {
	t.Errorf("mode = %q", o.Mode)
}
```

`ypytest.Layer(mode)` sets `YPY-Layer` on the request, `OpenLayer(t)` decodes the header the way the client reads it, and
`AssertTrigger(t, ypy.HeaderTrigger, "saved", ...)` checks the event of section 2.

See also [Portals: popover, dialog and layers](../guides/portals.md) for when to pick a layer over a native popover or dialog.
