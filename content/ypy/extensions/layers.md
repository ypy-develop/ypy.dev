---
title: "Layers: the ypy-layers extension"
description: "Modals, drawers and popups on the native <dialog> element, themed with CSS variables and never writing a style attribute, so it runs under a strict CSP."
order: 170
---
# Layers: the `ypy-layers` extension

Modals, drawers and popups that stack on the native `<dialog>` element. The default look is one optional stylesheet
with zero specificity, so any class of yours wins over it, and CSS variables theme it. Every class is configurable.
The extension never writes a `style` attribute, so it runs under a strict CSP.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-layers.min.js"></script>

<a href="/items/new" ypy-layer="modal" ypy-layer-close-on="saved" class="text-sky-700 underline">New item</a>
```

The link still works without JavaScript. With YPY the response of `/items/new` fills a modal instead of replacing the
page; links and forms inside it stay in the modal; when the server fires the `saved` event, the modal closes.

For when to choose a layer, a native `popover` or a plain `<dialog>`, and what a swap does to open dialogs, see
[Portals: popover, dialog and layers](../guides/portals.md).

Contents: [Opening](#opening-a-layer) · [Closing and results](#closing-and-results) · [Stack, focus and ARIA](#the-stack-focus-and-aria) ·
[Navigation in a layer](#navigation-inside-a-layer) · [History](#history) · [Server](#the-server-opens-a-layer-ypy-open-layer) ·
[Style](#style-with-tailwind-bootstrap-or-plain-css) · [API](#javascript-api) · [Events](#events) · [Config](#configuration) · [Limits](#limits)

## Opening a layer

| From | How |
|---|---|
| a link | `<a href="/x" ypy-layer="modal">` |
| a form | `<form action="/x" method="post" ypy-layer="drawer">`: the fields are sent, `GET` in the query string |
| any request element | `<button ypy-get="/menu" ypy-layer="popup">` |
| code | `ypy.layers.open('/x', {mode: 'drawer'})` or `ypy.layers.open('<h2>Hi</h2>')` (a string that starts with `<` is HTML) |
| the server | the `YPY-Open-Layer` response header ([below](#the-server-opens-a-layer-ypy-open-layer)) |

The value of `ypy-layer` is the kind: `modal` (also the default for any other value), `drawer` or `popup`.
The element keeps what it has: `ypy-trigger` picks the event, `ypy-select` picks part of the response,
`ypy-confirm`, `ypy-headers`, `ypy-vals`, `ypy-indicator` and `ypy-sync` work as on any request. A click with Ctrl, Meta or
Shift goes to the browser. An element inside `ypy-ignore`, or one the nonce gate refuses, is not wired.

The layer appears when its content is in. A request that fails (no connection) opens nothing and the result is
`{accepted: false, reason: 'error'}`. An error *response* is shown in the layer like any response (use
`ypy-status:500="swap:none"` or `ypy-error-*` to treat it otherwise).

| Kind | What it is |
|---|---|
| `modal` | `showModal()`: centered, the page behind is inert, Esc and the backdrop dismiss |
| `drawer` | a modal that sticks to an edge: `data-ypy-layer-side` is `end` (default), `start`, `top` or `bottom` |
| `popup` | a non-modal dialog (`show()`), anchored below its opener when the browser supports CSS anchor positioning, otherwise in a corner. Esc and a click outside dismiss it; a second click on the opener closes it |

A popup opened over a modal is placed inside the modal's dialog, because the rest of the page is inert while a modal is
open; one opened over another popup goes to the same parent as that popup. Only `modal` and `drawer` enter the browser's
top layer (`showModal()`); a popup is opened with `show()`, so it is an ordinary fixed-position element with
`z-index: var(--ypy-layer-z, 1000)`. See [Portals](../guides/portals.md#the-top-layer).

### Options

The same options come from three places, the last one winning: `ypy.config.layers`, attributes on the opener
(`ypy-layer-<name>`, in the form `close-on-url`), and the object passed to `ypy.layers.open()`.

| Option | Attribute | Meaning |
|---|---|---|
| `mode` | `ypy-layer` | `modal`, `drawer`, `popup` |
| `closeOn` | `ypy-layer-close-on` | event names that accept the layer |
| `closeOnUrl` | `ypy-layer-close-on-url` | path patterns (`*` is a wildcard) that accept it |
| `closeOnFragment` | `ypy-layer-close-on-fragment` | selector: a response that holds a match accepts it |
| `outside` | `ypy-layer-outside` | `false`: a click outside does not dismiss. Default `true` |
| `escape` | `ypy-layer-escape` | `false`: Esc does not dismiss. Default `true` |
| `history` | `ypy-layer-history` | show the layer URL in the address bar |
| `side` | `ypy-layer-side` | side of a drawer |
| `label` | `ypy-layer-label` | `aria-label` |
| `class` | `ypy-layer-class` | classes of the dialog |
| `url`, `html`, `method`, `values`, `select`, `source`, `event` | | code only: what to load and how (`source` is the element the request starts from and the one that gets the focus back) |

The `ypy-layer-*` option attributes are inherited when written as `ypy-layer-close-on:inherited="saved"` on an
ancestor of the opener, like any inherited YPY attribute. `ypy-layer`, `ypy-layer-accept`, `ypy-layer-dismiss` and
`ypy-layer-value` are read on their own element only.

## Closing and results

`ypy.layers.open()` returns a promise that settles when the layer closes; it never rejects.

```js
let result = await ypy.layers.open('/items/new', {closeOn: 'saved'});
if (result.accepted) console.log('saved', result.value);   // {accepted: true, value}
else console.log('left without saving', result.reason);    // {accepted: false, reason}
```

`promise.layer` is the handle (`element`, `mode`, `accept(value)`, `dismiss(reason)`, `closed`, `result`).

A layer accepts (a result with a value) or is dismissed (a reason). It closes by:

- an event: `ypy-layer-close-on="saved, updated"`. The first event of that name from inside the layer, or from the
  page for the top layer, accepts it. The value is the event `detail`: a response header
  `YPY-Trigger: {"saved":{"id":7}}` from a form in the layer closes it with `{id: 7}`.
- a URL: `ypy-layer-close-on-url="/items/*"`. When a request of the layer ends at a matching path (after redirects),
  the page is not swapped in: the layer is accepted with the path as its value. The classic "save, then redirect to
  the item" needs no extra response.
- a fragment: `ypy-layer-close-on-fragment="#saved"`. A response that holds a matching element is not swapped in; the
  layer is accepted with that element's `ypy-layer-value` (JSON or text).
- buttons: `<button ypy-layer-accept='{"ok":true}'>` accepts, `<button ypy-layer-dismiss="cancelled">` dismisses.
- the user: Esc, a click outside, a native `<form method="dialog">` (reason: its button `value`, or `closed`).
- code: `layer.accept(v)`, `layer.dismiss(r)`, `ypy.layers.accept(v)`, `ypy.layers.dismiss(r)`, `ypy.layers.closeAll()`.

Reasons the extension gives: `escape`, `outside`, `toggle` (a popup closed by its opener), `parent` (a layer below closed),
`navigate` (a popup whose layer navigated), `history`, `error` (the request failed, or the browser refused to show the
dialog), `dropped` (the request was dropped by `ypy-sync`), `replaced` (the first response asked for another layer with
`YPY-Open-Layer`), `cancelled` (a `ypy:layer:before:open` listener cancelled it; the promise resolves at once),
`closed` (the dialog was closed natively with no `returnValue`) and `dismissed` (the default).

A listener of `ypy:layer:before:close` can cancel the closing, for example to ask about unsaved changes. Then
`accept()` and `dismiss()` return `false` and nothing changes.

## The stack, focus and ARIA

Layers stack: each `open()` goes on top, `ypy.layers.stack` lists them (bottom first), `data-ypy-layer-depth` numbers
them. Closing a layer first dismisses the layers above it (reason `parent`); if one of them refuses, the lower layer stays.
Esc closes only the top one. A nested layer may be a modal, a drawer or a popup.

- Focus is trapped in a modal by the native dialog: the page behind cannot take it. On open, the focus goes to the
  `autofocus` control, else the first control, else the layer itself. When a swap removes the focused control the
  focus comes back into the layer. On close it returns to the opener, or to the element that was focused (which, for a
  nested layer, is in the layer below).
- ARIA: the dialog role is the element's own; a modal and a drawer carry `aria-modal="true"` (a popup does not).
  The layer is named by `ypy-layer-label` (`aria-label`), else by its first heading (`h1`, `h2`, `h3` or
  `role="heading"`, through `aria-labelledby`, kept in step when the layer navigates). `aria-busy` is set on the
  dialog while it loads, as on any target.
- Scroll lock: while a modal or a drawer is open `<html>` carries `data-ypy-layer-lock`, which the default style
  turns into `overflow:hidden`. `config.layers.scrollLock = false` leaves it out.

## Navigation inside a layer

Links and forms inside a layer navigate the layer: the response replaces the layer's content, the page does not
change. A plain link or form is taken over when it:

- is a left click without modifier keys on an `<a href>` (or a form submit) of the same origin;
- has no `target` other than `_self`, no `download`, no `href="#..."`;
- has no request attributes of its own (`ypy-get`, `ypy-post`...) and no `ypy-boost="false"`; and is not itself an opener.

Anything else is left to the browser or to YPY as usual. A link that writes `ypy-target` or `ypy-swap` keeps them.
The default is `innerHTML` into the layer's dialog, not a morph: the dialog is never replaced, its content is. A popup that
lives inside the dialog is dismissed first (reason `navigate`).
Forms are validated before they are sent. A `422` answer to a form in a layer goes into the form, as everywhere.
Every request that fills a layer or starts inside one carries the header `YPY-Layer: modal|drawer|popup`
(add `Vary: YPY-Layer` if you cache the answers), so the server can leave out the page chrome.

## History

`history: true` (or `ypy-layer-history`) shows the layer URL in the address bar while the layer is open. The layer pushes one entry;
navigating inside the layer replaces it; closing the layer goes back. The Back button closes the top layer that
pushed an entry (reason `history`) and the page behind is not requested again. Without the option the address
bar never changes, and a `YPY-Push-Url` header in a layer response is ignored. It needs `ypy.config.history` on, and a URL
(not just `html`).

## The server opens a layer: `YPY-Open-Layer`

A response with the header `YPY-Open-Layer` is not swapped: it opens a layer. The value is JSON, or just a URL.

```http
HTTP/1.1 200 OK
YPY-Open-Layer: {"url":"/items/new","mode":"drawer","closeOn":"saved","label":"New item"}
```

Keys: `url` or `html`, `mode`, `closeOn`, `closeOnUrl`, `closeOnFragment`, `history`, `side`, `label`, `outside`,
`escape`, `class`. Others are ignored. Write non-ASCII characters as `\u` escapes (headers are Latin-1). The request that
got the header may come from inside a layer: the new layer goes on top. A URL is fetched with the same rules as any
request (same-origin by default). With the Go SDK: `ypy.OpenLayer(w, url, ypy.LayerOptions{...})` and
`ypy.OpenLayerHTML(w, markup, ...)`, see [Layers with the Go SDK](../server/layers.md). A response that carries
`YPY-Open-Layer` is not swapped whatever its status (`204` is fine); when a layer's own first request answers it, that
layer is dismissed with reason `replaced` and the new one opens.

## Style with Tailwind, Bootstrap or plain CSS

The default stylesheet is added once, as a constructable stylesheet, with `:where()` so any class beats it. Turn it off with
`ypy.config.layers = {style: false}`, or theme it with variables on `:root` or on one dialog:

| Variable | Default |
|---|---|
| `--ypy-layer-bg`, `--ypy-layer-color` | `Canvas`, `CanvasText` |
| `--ypy-layer-border`, `--ypy-layer-radius`, `--ypy-layer-shadow` | thin grey border, `.5rem`, soft shadow |
| `--ypy-layer-padding`, `--ypy-layer-width`, `--ypy-layer-max-height` | `1rem`, `min(32rem, 100% - 2rem)`, viewport height minus `2rem` |
| `--ypy-layer-backdrop` | `#0006` |
| `--ypy-layer-drawer-width`, `--ypy-layer-drawer-height` | `min(24rem, 100%)`, `50dvh` |
| `--ypy-layer-popup-top`, `--ypy-layer-popup-right`, `--ypy-layer-popup-width`, `--ypy-layer-z` | corner placement when not anchored |

The hooks for your own CSS are on the dialog: `data-ypy-layer-kind="modal|drawer|popup"`, `data-ypy-layer-side`,
`data-ypy-layer-depth`, `data-ypy-layer-anchored`; and on the opener of a popup the class `ypy-layer-opener`.

Tailwind: put the utilities in the configuration, per kind, or per layer with `ypy-layer-class`. Make Tailwind scan
the file that holds them (`@source`).

```html
<script>
  ypy.config.layers = {classes: {
    modal:  'rounded-xl p-0 w-full max-w-lg shadow-2xl backdrop:bg-slate-900/60',
    drawer: 'p-0 h-dvh w-96 max-w-full backdrop:bg-slate-900/60',
    popup:  'rounded-lg border bg-white p-2 shadow-lg'
  }};
</script>
<a href="/cart" ypy-layer="drawer" ypy-layer-side="end" class="rounded bg-sky-600 px-3 py-2 text-white">Cart</a>
```

The server fragment brings its own spacing:

```html
<div class="space-y-4 p-6">
  <h2 class="text-lg font-semibold">Cart</h2>
  <button ypy-layer-dismiss class="rounded border px-3 py-1">Close</button>
</div>
```

Bootstrap: the layer is a `<dialog>`, not a `.modal`; reuse the look with the component classes and leave
Bootstrap's own modal JavaScript out of it.

```html
<script>
  ypy.config.layers = {style: false, classes: {
    modal: 'border-0 rounded-3 p-0 shadow', drawer: 'border-0 p-0 shadow', popup: 'card p-2 shadow'
  }};
</script>
<style>dialog[data-ypy-layer-kind]::backdrop { background: rgb(0 0 0 / .5) }</style>
<button class="btn btn-primary" ypy-get="/items/new" ypy-layer="modal">New</button>
```

```html
<div class="modal-header"><h2 class="modal-title fs-5">New item</h2>
  <button class="btn-close" ypy-layer-dismiss aria-label="Close"></button></div>
<form action="/items" method="post"><div class="modal-body">...</div>
  <div class="modal-footer"><button class="btn btn-primary">Save</button></div></form>
```

Use `ypy.onLoad('[data-bs-toggle="tooltip"]', ...)` to start Bootstrap components in a layer: its content is processed
like any swap.

## JavaScript API

`ypy.layers`:

| Member | |
|---|---|
| `open(spec, options)` | `spec` is a URL, a string of HTML (starts with `<`) or an options object. Returns a promise of `{accepted, value}` or `{accepted, reason}`, with `.layer` |
| `stack` | the open layers, bottom first |
| `top()` | the top layer |
| `accept(value)`, `dismiss(reason)` | on the top layer |
| `closeAll(reason)` | dismisses every layer, top first; `false` when one refused |

A layer handle has `element`, `mode`, `source`, `options`, `closed`, `result`, `accept(value)` and `dismiss(reason)`.
Types are in `ypy.d.ts` (`YpyLayers`, `YpyLayer`, `YpyLayerOptions`).

## Events

| Event | On | Detail | Cancelable |
|---|---|---|---|
| `ypy:layer:before:open` | the opener, or `document` | `{options, mode}` | yes: nothing opens |
| `ypy:layer:opened` | the dialog | `{layer}` | no |
| `ypy:layer:before:close` | the dialog | `{layer, accepted, value or reason}` | yes: it stays open |
| `ypy:layer:accepted` | the dialog, still in the page | `{layer, accepted: true, value}` | no |
| `ypy:layer:dismissed` | the dialog, still in the page | `{layer, accepted: false, reason}` | no |

They bubble to `document`.

## Configuration

`ypy.config.layers` (all optional): `style` (default `true`), `scrollLock` (`true`), `outside` (`true`), `escape` (`true`),
`history` (`false`), `classes` (`{modal, drawer, popup, opener}`). The extension reads it each time a layer opens, so it also
works on a locked configuration.

## Limits

- A layer is a native `<dialog>`; a browser without `showModal()` is not supported.
- A layer element is a child of `<body>` (a popup over a layer, a child of that layer). Page swaps whose target does not
  contain it leave it alone; keep targets narrower than `<body>`.
- The focus returns to the opener on close only if the opener is still in the page.
- Ids in a layer's content live in the same document as the page's ids; keep them distinct.
- A popup is placed with CSS anchor positioning (opener to popup). Without it, it sits in a corner (`--ypy-layer-popup-*`).
- A layer opens with a `GET` or the verb of its form; the extension does not send a body of its own.
- With the nonce extension (`ypy-csp`), openers need their `ypy-nonce`, as every YPY element does. Links and forms
  that a layer takes over are plain HTML and need none.
