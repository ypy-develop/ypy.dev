---
title: "Layers: the ypy-layers extension"
description: "Modals, drawers and popups on the native <dialog> element, themed with CSS variables and never writing a style attribute, so it runs under a strict CSP."
order: 170
---
# Layers: the `ypy-layers` extension

Modals, drawers and popups that stack on the native `<dialog>` element. The default look is one optional stylesheet with
zero specificity, so any class of yours wins over it, and CSS variables theme it. The extension never writes a `style`
attribute, so it runs under a strict CSP.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-layers.min.js"></script>

<a href="/items/new" ypy-layer="modal" ypy-layer-close-on="saved">New item</a>
```

The link still works without JavaScript. With YPY the response of `/items/new` fills a modal instead of replacing the
page, links and forms inside it stay in the modal, and when the server fires the `saved` event, the modal closes.

## Opening a layer

| From | How |
|---|---|
| a link | `<a href="/x" ypy-layer="modal">` |
| a form | `<form action="/x" method="post" ypy-layer="drawer">`; a GET puts the fields in the query string |
| any request element | `<button ypy-get="/menu" ypy-layer="popup">` |
| code | `ypy.layers.open('/x', {mode: 'drawer'})` or `ypy.layers.open('<h2>Hi</h2>')` (a string that starts with `<` is HTML) |
| the server | the `YPY-Open-Layer` response header |

The value of `ypy-layer` is the kind: `modal` (also the default for any other value), `drawer` or `popup`.

| Kind | What it is |
|---|---|
| `modal` | `showModal()`: centered, the page behind is inert, Esc and the backdrop dismiss |
| `drawer` | a modal that sticks to an edge: `data-ypy-layer-side` is `end` (default), `start`, `top` or `bottom` |
| `popup` | a non-modal dialog (`show()`), anchored below its opener when the browser supports CSS anchor positioning, otherwise in a corner. Esc and a click outside dismiss it |

The opener keeps what it has: `ypy-trigger`, `ypy-select`, `ypy-confirm`, `ypy-headers`, `ypy-vals`, `ypy-indicator` and
`ypy-sync` work as on any request. A click with Ctrl, Meta or Shift goes to the browser. A request that fails opens nothing
and the result is `{accepted: false, reason: 'error'}`. An error response is shown in the layer like any response.

### Options

The same options come from `ypy.config.layers`, from attributes on the opener (`ypy-layer-<name>`, as in
`ypy-layer-close-on-url`) and from the object passed to `ypy.layers.open()`. The last one wins.

| Option | Attribute | Meaning |
|---|---|---|
| `mode` | `ypy-layer` | `modal`, `drawer`, `popup` |
| `closeOn` | `ypy-layer-close-on` | event names that accept the layer |
| `closeOnUrl` | `ypy-layer-close-on-url` | path patterns (`*` is a wildcard) that accept it |
| `closeOnFragment` | `ypy-layer-close-on-fragment` | selector: a response that holds a match accepts it |
| `outside`, `escape` | `ypy-layer-outside`, `ypy-layer-escape` | `false` stops a click outside, or Esc, from dismissing. Default `true` |
| `history` | `ypy-layer-history` | show the layer URL in the address bar |
| `side` | `ypy-layer-side` | side of a drawer |
| `label` | `ypy-layer-label` | `aria-label` |
| `class` | `ypy-layer-class` | classes of the dialog |
| `url`, `html`, `method`, `values`, `select`, `source`, `event` | | code only: what to load and how |

The `ypy-layer-*` option attributes can be inherited, as in `ypy-layer-close-on:inherited="saved"` on an ancestor.

## Closing and results

`ypy.layers.open()` returns a promise that settles when the layer closes and never rejects.

```js
let result = await ypy.layers.open('/items/new', {closeOn: 'saved'});
if (result.accepted) console.log('saved', result.value);
else console.log('left without saving', result.reason);
```

A layer is accepted, with a value, or dismissed, with a reason. It closes by:

- an event: `ypy-layer-close-on="saved, updated"`. The first event of that name from inside the layer accepts it, and the
  value is the event `detail`. A response header `YPY-Trigger: {"saved":{"id":7}}` from a form in the layer closes it with
  `{id: 7}`.
- a URL: with `ypy-layer-close-on-url="/items/*"`, a request of the layer that ends at a matching path is not swapped in
  and the layer is accepted with the path. "Save, then redirect to the item" needs nothing else.
- a fragment: with `ypy-layer-close-on-fragment="#saved"`, a response that holds a matching element is not swapped in and
  the layer is accepted with that element's `ypy-layer-value`.
- buttons: `<button ypy-layer-accept='{"ok":true}'>` accepts and `<button ypy-layer-dismiss="cancelled">` dismisses.
- the user: Esc, a click outside, or a native `<form method="dialog">`.
- code: `layer.accept(v)`, `layer.dismiss(r)`, `ypy.layers.accept(v)`, `ypy.layers.dismiss(r)`, `ypy.layers.closeAll()`.

The reasons the extension gives are `escape`, `outside`, `toggle`, `parent`, `navigate`, `history`, `error`, `dropped`,
`replaced`, `cancelled`, `closed` and `dismissed` (the default). A listener of `ypy:layer:before:close` can cancel the
closing, for example to ask about unsaved changes.

## The stack, focus and ARIA

Layers stack. `ypy.layers.stack` lists them bottom first and `data-ypy-layer-depth` numbers them. Closing a layer first
dismisses the ones above it (reason `parent`), and Esc closes only the top one.

- On open, focus goes to the `autofocus` control, else the first control, else the layer. On close it returns to the
  opener, if it is still in the page.
- A modal and a drawer carry `aria-modal="true"`. The layer is named by `ypy-layer-label`, else by its first heading.
  `aria-busy` is set while it loads.
- While a modal or a drawer is open, `<html>` carries `data-ypy-layer-lock`, which the default style turns into
  `overflow:hidden`. `config.layers.scrollLock = false` leaves it out.

## Navigation inside a layer

Links and forms inside a layer navigate the layer: the response replaces its content with `innerHTML` and the page does
not change. A plain link or form is taken over when it is a left click without modifier keys, on the same origin, with no
`target` other than `_self`, no `download`, no `href="#..."`, no request attributes of its own and no `ypy-boost="false"`.
Forms are validated first, and a `422` goes into the form as everywhere.

Every request that fills a layer or starts inside one carries `YPY-Layer: modal|drawer|popup`, so the server can leave out
the page chrome. Add `Vary: YPY-Layer` if you cache the answers.

## History

`history: true` (or `ypy-layer-history`) shows the layer URL in the address bar while the layer is open. Navigating inside
the layer replaces the entry, and closing goes back. The Back button closes the top layer that pushed an entry. Without the
option the address bar never changes. It needs `ypy.config.history` on and a URL, not just `html`.

## The server opens a layer

A response with the header `YPY-Open-Layer` is not swapped, whatever its status: it opens a layer. The value is JSON, or
just a URL.

```http
HTTP/1.1 200 OK
YPY-Open-Layer: {"url":"/items/new","mode":"drawer","closeOn":"saved","label":"New item"}
```

Keys: `url` or `html`, `mode`, `closeOn`, `closeOnUrl`, `closeOnFragment`, `history`, `side`, `label`, `outside`, `escape`
and `class`. Write non-ASCII characters as `\u` escapes, because headers are Latin-1. When the first request of a layer
answers with the header, that layer is dismissed with reason `replaced` and the new one opens.

## Style

The default stylesheet is added once, as a constructable stylesheet, with `:where()` so any class beats it. Turn it off
with `ypy.config.layers = {style: false}`, or theme it with variables on `:root` or on one dialog.

| Variable | Default |
|---|---|
| `--ypy-layer-bg`, `--ypy-layer-color` | `Canvas`, `CanvasText` |
| `--ypy-layer-border`, `--ypy-layer-radius`, `--ypy-layer-shadow` | thin grey border, `.5rem`, soft shadow |
| `--ypy-layer-padding`, `--ypy-layer-width`, `--ypy-layer-max-height` | `1rem`, `min(32rem, 100% - 2rem)`, viewport height minus `2rem` |
| `--ypy-layer-backdrop` | `#0006` |
| `--ypy-layer-drawer-width`, `--ypy-layer-drawer-height` | `min(24rem, 100%)`, `50dvh` |
| `--ypy-layer-popup-top`, `--ypy-layer-popup-right`, `--ypy-layer-popup-width`, `--ypy-layer-z` | corner placement when not anchored |

The hooks for your own CSS are on the dialog: `data-ypy-layer-kind`, `data-ypy-layer-side`, `data-ypy-layer-depth` and
`data-ypy-layer-anchored`, and the class `ypy-layer-opener` on the opener of a popup. To use the classes of a CSS
framework, set them per kind:

```js
ypy.config.layers = {style: false, classes: {
  modal:  'rounded-xl p-0 w-full max-w-lg shadow-2xl',
  drawer: 'p-0 h-dvh w-96 max-w-full',
  popup:  'rounded-lg border p-2 shadow-lg'
}};
```

## JavaScript API and events

| Member | |
|---|---|
| `ypy.layers.open(spec, options)` | `spec` is a URL, a string of HTML or an options object. Returns a promise of `{accepted, value}` or `{accepted, reason}`, with `.layer` |
| `ypy.layers.stack`, `top()` | the open layers, bottom first, and the top one |
| `ypy.layers.accept(value)`, `dismiss(reason)` | on the top layer |
| `ypy.layers.closeAll(reason)` | dismisses every layer, top first; `false` when one refused |

A layer handle has `element`, `mode`, `source`, `options`, `closed`, `result`, `accept(value)` and `dismiss(reason)`.

| Event | On | Cancelable |
|---|---|---|
| `ypy:layer:before:open` | the opener, or `document` | yes: nothing opens |
| `ypy:layer:opened` | the dialog | no |
| `ypy:layer:before:close` | the dialog | yes: it stays open |
| `ypy:layer:accepted`, `ypy:layer:dismissed` | the dialog, still in the page | no |

`ypy.config.layers` also takes `scrollLock`, `outside`, `escape`, `history` and `classes`, all optional.

## Limits

- A layer is a native `<dialog>`. A browser without `showModal()` is not supported.
- A layer is a child of `<body>`, so keep swap targets narrower than `<body>`.
- Ids in a layer live in the same document as the page's ids, so keep them distinct.
- A layer opens with a GET or the verb of its form, and the extension sends no body of its own.
- With [`ypy-csp`](csp.md), openers need their nonce like every YPY element.
