---
order: 170
title: Dialogs with layers
description: Open a form in a native modal dialog with ypy-layers, keep 422 errors inside it, and close it when the server fires a saved event.
---

# Dialogs with layers

A link opens a form in a modal dialog, errors stay inside it, and the dialog closes when the server says the data is saved.

## The page

The `ypy-layers` extension drives the native `<dialog>` element, so focus handling, Esc and the backdrop come with it.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-layers.min.js"></script>

<a href="/contacts/new" ypy-layer="modal" ypy-layer-close-on="saved" ypy-layer-label="New contact">New contact</a>

<ul id="contacts" ypy-get="/contacts" ypy-trigger="saved from:body" ypy-swap="innerHTML">
  <li>Helena Barros</li>
</ul>
```

Without JavaScript the link goes to `/contacts/new` and shows a whole page. With YPY the response fills a modal. When
the `saved` event reaches the page, the layer closes and the list asks the server for its new content.

## What the server answers

`GET /contacts/new` sends the form alone when the request comes from a layer (the `YPY-Layer` request header is
present), and the same form inside a whole page otherwise:

```text
GET /contacts/new
YPY-Request: true
YPY-Layer: modal

HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Vary: YPY-Request, YPY-Layer
```

```html
<form action="/contacts" method="post">
  <h2>New contact</h2>
  <label>Name <input name="name" required></label>
  <button>Save</button>
  <button type="button" ypy-layer-dismiss="cancelled">Cancel</button>
</form>
```

`POST /contacts` with an empty name answers `422` with the same fields and a message. The content goes into the form,
inside the layer:

```text
HTTP/1.1 422 Unprocessable Content
Content-Type: text/html; charset=utf-8
Vary: YPY-Request, YPY-Layer
```

```html
<h2>New contact</h2>
<label>Name <input name="name" value="" required aria-invalid="true" aria-describedby="name-error"></label>
<p id="name-error" role="alert">Enter a name.</p>
<button>Save</button>
<button type="button" ypy-layer-dismiss="cancelled">Cancel</button>
```

When the contact is saved, the server fires the event with the `YPY-Trigger` header and sends no body:

```text
HTTP/1.1 204 No Content
YPY-Trigger: {"saved": {"id": 7}}
Vary: YPY-Request, YPY-Layer
```

A request that did not come from a layer gets `303 See Other` with `Location: /contacts` instead.

| Request | Status | What happens |
|---|---|---|
| `GET /contacts/new` from the link | `200`, the form | the layer opens with the form |
| `POST /contacts`, name empty | `422`, the form fields with a message | the message appears inside the layer |
| `POST /contacts`, name valid | `204` with `YPY-Trigger: saved` | the layer closes, the list reloads |

## How it works

- `ypy-layer="modal"` turns the link into a request that fills a `<dialog>` opened with `showModal()`.
- The form inside has no YPY attributes. A plain form in a layer submits into the layer, and a `422` goes into the form.
- The `saved` event fires on the form, bubbles to the page and matches `ypy-layer-close-on="saved"`. The detail
  (`{"id": 7}`) becomes the value of the layer.
- The list uses `ypy-trigger="saved from:body"` to reload itself when that event bubbles up to the page.
- Send `Vary: YPY-Layer` as well as `Vary: YPY-Request` when a cache sits in front of `/contacts/new`.

## Variations

- `ypy-layer="drawer"` and `ypy-layer-side="start"` make a side panel. `ypy-layer="popup"` makes a small dialog next to
  the opener.
- To close after a redirect instead of an event, use `ypy-layer-close-on-url="/contacts/*"`.
- To open a layer from the server, answer with the `YPY-Open-Layer` header:
  `YPY-Open-Layer: {"url":"/contacts/7/confirm-delete","mode":"modal"}`.
- To read the result in script, from a script file: `let result = await ypy.layers.open('/contacts/new', {closeOn: 'saved'})`.

## See also

- [`ypy-layers`](../extensions/layers.md) for every option, event and ARIA detail
- [Layers on the server](../server/layers.md)
- [Portals: popover, dialog and layers](../guides/portals.md) to choose between a layer, a `popover` and a plain `<dialog>`
- [Confirm and prompt](confirm-and-prompt.md) for the browser's own dialogs
