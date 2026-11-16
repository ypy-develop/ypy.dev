---
order: 210
title: Updating other content
description: Update the cart, the item count and a message from one response, with partials, a retarget header, or an event that the server fires.
---

# Updating other content

"Add to cart" changes the cart list, the item count in the header and shows a message. There are three ways to reach the
parts that are not the request's own target. Pick the first one unless you have a reason not to.

## The page

```html
<header>Cart (<span id="cart-count">0</span>)</header>
<div id="toast"></div>

<ul id="cart"></ul>

<form ypy-post="/cart/items" ypy-target="#cart" ypy-swap="beforeend">
  <input type="hidden" name="sku" value="A-100">
  <button>Add to cart</button>
</form>
```

The form's own target is `#cart`, where the new row goes. The count and the message live elsewhere.

## What the server answers

### 1. Partials: the answer names the targets

`POST /cart/items` answers `200` with the new row, followed by one `<ypy-partial>` per extra piece:

```html
<li>Ceramic mug, $14.50</li>

<ypy-partial ypy-target="#cart-count" ypy-swap="innerHTML">3</ypy-partial>
<ypy-partial ypy-target="#toast" ypy-swap="innerHTML"><p role="status">Added to the cart</p></ypy-partial>
```

YPY swaps the main content first, then each partial into its own target. Each piece carries its destination, so the page
needs no extra attributes. A partial whose target is not on the page fires `ypy:target:missing` and is skipped.

### 2. Change the target from the server

When the whole answer belongs somewhere else, say where in a header. When the cart is full, the server answers with a
message, whatever target the form has:

```http
HTTP/1.1 200 OK
YPY-Retarget: #toast
YPY-Reswap: innerHTML
Vary: YPY-Request

<p role="alert">The cart is full.</p>
```

`YPY-Reselect: #message` keeps only the part of the response that matches the selector.

### 3. An event the server fires

Other parts of the page can refresh themselves when something happens. The server only names the event, in a header:

```http
HTTP/1.1 200 OK
YPY-Trigger: {"cart-changed": {"count": 3}}
Vary: YPY-Request

<li>Ceramic mug, $14.50</li>
```

```html
<span id="cart-count" ypy-get="/cart/count" ypy-trigger="cart-changed from:body" ypy-swap="innerHTML">0</span>
```

The event fires on the element that sent the request and bubbles to `<body>`. The count asks the server for its new
value. That costs one more request, but the server code for the action does not need to know who is on the page.

## How it works

- Partials, `YPY-Retarget` and `YPY-Trigger` are all applied after the response arrives.
- `ypy-swap-oob="true"` on an element of the response swaps it into the element with the same `id`. It exists for
  responses written before partials. In new code, a partial says the same thing with an explicit target.
- A partial with no `ypy-swap` merges as `innerMorph`, so it keeps what the visitor is doing inside the target.
- If several pieces must change together or not at all, the server sends `YPY-Group: atomic`. See
  [out-of-band swaps](../index.md#out-of-band-swaps).

## Variations

- `ypy-select-oob="#cart-count,#toast"` on the form picks those elements out of the response and swaps each into the
  element with the same id.
- A partial can use any swap style, for example `ypy-swap="beforeend"` to add a message to a list of messages.

## See also

- [Out-of-band swaps](../index.md#out-of-band-swaps)
- [Response headers](../reference.md#response-headers)
- [Reset the form after a save](reset-user-input.md)
