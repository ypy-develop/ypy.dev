---
title: "Portals: popover, dialog and layers"
description: "Pick between native popover, native dialog and ypy-layers for menus, confirmations, modal forms and drawers, and what a swap or morph does to an open one."
order: 60
---
# Portals: popover, dialog and layers

Menus, tooltips, confirmations, forms in a modal and side panels all draw above the page. The browser has three native
pieces for that: the `popover` attribute, the `<dialog>` element and the top layer they live in. YPY adds one
extension, [`ypy-layers`](../extensions/layers.md), which drives `<dialog>` from the server. This guide says which one
to use, what the top layer changes, and what a swap or a morph does to an element that is open.

## Which one to use

| You need | Use | Why |
|---|---|---|
| a menu, a tooltip, a small card next to a button, no server round trip | native `popover` | no JavaScript; light dismiss and Esc come from the browser |
| the same, with content that the server renders | native `popover` + `ypy-get` on the button | YPY fills the popover; the browser opens it |
| a confirmation or a message whose content is already in the page | native `<dialog>` | `showModal()` and `<form method="dialog">` are enough |
| a form or a page that the server renders in a modal, with validation, redirects and a result | `ypy-layer="modal"` | links and forms stay in the dialog; the server closes it with an event |
| a side panel (cart, filters, details) | `ypy-layer="drawer"` | a modal glued to an edge |
| a small menu that the server renders, over a modal, with a stack | `ypy-layer="popup"` | non-modal dialog, placed inside the modal below it |
| a modal that the server decides to open | `YPY-Open-Layer` response header | a delete button can ask "are you sure?" from the handler |

Rule of thumb: if the content is static or the page already holds it, stay native. If the content is a page of the
server with its own links and forms, use a layer.

## The top layer

The top layer is a part of the browser's rendering that sits above the whole page. An element in it is drawn over
everything else regardless of `z-index`, and the `overflow` or `transform` of its ancestors does not clip it.

- `popover` elements and `<dialog>` elements opened with `showModal()` enter the top layer. Later ones draw over earlier ones.
- `showModal()` also makes the rest of the document inert (it cannot be clicked or focused), keeps the focus in the
  dialog and turns Esc into a `cancel` event. A popover with `popover="auto"` closes on Esc and on a click
  outside (light dismiss); `popover="manual"` does neither.
- A `<dialog>` opened with `show()` is not in the top layer and is not modal. It is an ordinary element, so `z-index`
  and clipping apply.

What this means for YPY:

- A layer of kind `modal` or `drawer` uses `showModal()`: top layer, inert page, native focus trap
  (`ypy-layers.js`, function `reveal`).
- A layer of kind `popup` uses `show()`, so it is not in the top layer. The default style gives it `position: fixed`
  and `z-index: var(--ypy-layer-z, 1000)`. Change `--ypy-layer-z` if something on your page is above 1000.
- Because the page behind a modal is inert, a popup opened while a modal is on top is placed inside that
  modal's dialog. A popup over another popup goes to the same parent as that popup.
- `ypy-inert` (when on) makes the target of a request `inert` while it runs. Do not make a region that holds the open
  dialog the target of a slow request if people must keep using the dialog: see
  [`ypy-inert`](../reference/attributes/ypy-inert.md).

## Native popover with server content

The browser opens and closes the popover, and YPY fills it. Give the popover an `id`, because the morph recognizes it by that.

```html
<button popovertarget="user-menu" ypy-get="/menu" ypy-target="#user-menu"
        class="rounded border px-3 py-1.5">Account</button>

<div id="user-menu" popover class="rounded-lg border bg-white p-2 shadow-lg">Loading...</div>
```

The server answers a fragment for the popover:

```http
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Vary: YPY-Request

<a href="/profile" class="block px-3 py-1.5">Profile</a>
<form action="/logout" method="post"><button class="block px-3 py-1.5">Sign out</button></form>
```

```go
func menu(w http.ResponseWriter, r *http.Request) {
	_ = ypy.Fragment(w, `<a href="/profile" class="block px-3 py-1.5">Profile</a>
<form action="/logout" method="post"><button class="block px-3 py-1.5">Sign out</button></form>`)
}
```

YPY has no code or test of its own for `popover`, because it is the browser's attribute. The default swap is a morph,
so the popover node stays in the page and only its content changes.

## Native dialog

```html
<dialog id="confirm" class="rounded-xl p-6 backdrop:bg-slate-900/60">
  <form method="dialog" class="space-y-4">
    <p>Delete this item?</p>
    <button value="cancel" class="rounded border px-3 py-1">Cancel</button>
    <button value="ok" class="rounded bg-red-600 px-3 py-1 text-white">Delete</button>
  </form>
</dialog>
<button data-open-dialog="confirm" class="rounded border px-3 py-1.5">Delete</button>
```

`showModal()` is a method, so something has to call it. With a strict CSP, put it in a script file, not in an `onclick`:

```js
document.addEventListener('click', e => {
  let b = e.target.closest('[data-open-dialog]');
  if (b) document.getElementById(b.dataset.openDialog).showModal();
});
```

A form with `method="dialog"` (or a button with `formmethod="dialog"`) is left to the browser: YPY makes no request for
it and does not boost it. The button `value` becomes the dialog's `returnValue`.

## A layer, end to end

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-layers.min.js"></script>

<a href="/items/new" ypy-layer="modal" ypy-layer-close-on="saved"
   class="rounded bg-sky-600 px-3 py-2 text-white">New item</a>
<ul id="items" ypy-get="/items" ypy-trigger="saved from:document">...</ul>
```

`/items/new` answers with the form only. The browser sends `YPY-Layer: modal`, and the handler leaves out the page chrome
and, on success, fires the event the opener listens to:

```go
func newItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "YPY-Layer")
	q := ypy.FromRequest(r)
	if r.Method == http.MethodGet {
		if q.InLayer() {
			_ = ypy.Fragment(w, itemForm(Item{}, nil))
			return
		}
		renderPage(w, itemForm(Item{}, nil)) // no JavaScript: the same form in a page
		return
	}
	item, errs := parseItem(r)
	if errs != nil {
		_ = ypy.ValidationError(w, itemForm(item, errs)) // 422: stays in the layer
		return
	}
	saved := store.Create(item)
	if !q.InLayer() {
		ypy.SeeOther(w, "/items/"+saved.ID)
		return
	}
	_ = ypy.Trigger(w, "saved", map[string]any{"id": saved.ID}) // accepts the layer, value {id}
	ypy.NoContent(w)
}
```

The layer closes, `ypy.layers.open()` resolves with `{accepted: true, value: {id: ...}}`, and the list refreshes
because `#items` listens to `saved`. More on this handler and on `YPY-Open-Layer`:
[Layers with the Go SDK](../server/layers.md).

## What a swap and a morph do to them

A dialog or popover inside a region that the page swaps keeps or loses its state depending on the swap.

| Situation | Result |
|---|---|
| Morph (the default swap) where the response has the same `<dialog>` | the same DOM node stays; its attributes and content follow the response, **except `open`**; an open dialog stays open and a modal stays modal (tests `morph-default.js`, `morph-state.js`) |
| The response draws `<dialog open>` but the user closed it | it stays closed: the page's state wins. `ypy-morph-force` on the dialog in the response lets the response set `open` |
| Raw swap (`innerHTML`, `outerHTML`) of a region that holds a dialog or popover | the old node is removed and the response's node is a new element. Its open state is gone; an open modal does not come back as a modal |
| The response moves the element to another place in the same target | the morph matches elements by `id` and moves the existing node (with `moveBefore` where the browser has it, else `insertBefore`); give dialogs and popovers an `id` |
| `ypy-preserve` on the element | the node leaves the page for the swap and is put back in place of the element with the same `id`. It needs the `id` on the page and in the response; see [`ypy-preserve`](../reference/attributes/ypy-preserve.md) |
| Content swapped **into** a layer (links and forms in it) | `innerHTML` into the layer's dialog; the dialog itself is not replaced |
| A swap into another target of the page while a layer is open | the layer is a child of `<body>` (or of the modal below it, for a popup); a swap that does not contain that parent does not touch it |

Four details are worth knowing:

- The `open` attribute of `<dialog>` and `<details>` mirrors something the user changes, so the morph never copies it
  from the response onto an element that is already on the page (`__copyAttributes` in `ypy.js`). Closing or opening a
  dialog from the server goes through a layer (`YPY-Open-Layer`, an event, a button), not through the attribute.
- A popup of a layer that lives inside a layer's dialog is dismissed with reason `navigate` when that layer
  navigates, because the swap would remove it.
- The focus goes back to the opener when a layer closes, if the opener is still in the page. If a swap removed it, the
  focus is not restored.
- A layer's own content is not morphed on navigation: `ypy.layers` swaps with `innerHTML`, so a field typed in a form is
  replaced by the next response. A `422` answer goes into the form, as everywhere, and keeps the typed values.

## Pitfalls

- A server-rendered `open`: do not render `open` to show a dialog on a page that YPY swaps. The morph ignores it for
  an element already there, and a dialog drawn `open` is not modal. Open it with `showModal()` or a layer.
- No `id`: a dialog or popover without an `id` is matched by position and tag. If the response reorders siblings, the
  wrong node can receive the content, so give each one an `id`.
- Swapping `body`: a layer is a child of `<body>`, so a request whose target is `body` swaps the parent of every layer,
  and no test covers what that does to them. Keep targets narrower, or close the layers first
  (`ypy.layers.closeAll()`).
- Popup z-index: a `popup` layer is not in the top layer. If it hides behind a sticky header, raise `--ypy-layer-z`, or
  use a native `popover` for that menu.
- Popover inside a modal: while a modal is open the rest of the page is inert, so a native popover that must work over
  a modal has to sit inside the modal's content.
- Lost opener: if the page that held the opener is swapped while its layer is open, the focus cannot return to it.
  Keep the opener outside the swapped region, or set the focus yourself in `ypy:layer:dismissed`.
- Full page in a layer: a handler that ignores `YPY-Layer` sends the whole page into the dialog. Check `q.InLayer()`
  and send `Vary: YPY-Layer`.
- Inline handlers: under a strict CSP use `<form method="dialog">`, layer attributes (`ypy-layer-dismiss`,
  `ypy-layer-accept`) or a script file. The layer writes no `style` attribute, so it needs no CSP relaxation.
- Repeated ids: the content of a layer is in the same document as the page, so an `id` that exists in both is two
  elements with one id. Prefix the ids of layer content.

## See also

[`ypy-layers` reference](../extensions/layers.md) · [Layers with the Go SDK](../server/layers.md) ·
[Swap styles and morph](../reference/swap-and-morph.md) · [`ypy-preserve`](../reference/attributes/ypy-preserve.md) ·
[Accessibility](accessibility.md)
