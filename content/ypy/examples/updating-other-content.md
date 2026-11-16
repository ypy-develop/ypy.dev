---
order: 210
title: Updating other content
description: Update the cart, the item count and a message from one response, with partials, a retarget header, or an event the server fires.
---

# Updating other content

"Add to cart" changes the cart list, the item count in the header and shows a message. There are three ways to reach
the parts that are not the request's own target. Pick the first one unless you have a reason not to.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<header>Cart (<span id="cart-count">0</span>)</header>
<div id="toast"></div>

<ul id="cart"></ul>

<form ypy-post="/cart/items" ypy-target="#cart" ypy-swap="beforeend">
  <input type="hidden" name="sku" value="A-100">
  <button>Add to cart</button>
</form>
```

The form's own target is `#cart`, where the new row goes. The count and the message live elsewhere.

## 1. Partials: the server names the targets

```go
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	item, count := h.cart.Add(r.Context(), r.PostFormValue("sku"))

	ypy.Fragment(w, `<li>`+html.EscapeString(item.Name)+`</li>`) // goes to #cart
	ypy.Partial(w, ypy.PartialSpec{Target: "#cart-count", Swap: "innerHTML"}, strconv.Itoa(count))
	ypy.Partial(w, ypy.PartialSpec{Target: "#toast", Swap: "innerHTML"}, `<p role="status">Added to the cart</p>`)
}
```

The response is the `<li>` followed by two `<ypy-partial>` elements. YPY swaps the main content first, then each partial
into its own target. Each piece carries its destination, so the page needs no extra attributes. A partial whose target is
not on the page fires `ypy:target:missing` and is skipped.

## 2. Change the target from the server

When the whole answer belongs somewhere else, say where in a header. This is how a form shows a message when the
cart is full, whatever target the form has:

```go
if full {
	ypy.Retarget(w, "#toast")
	ypy.Reswap(w, "innerHTML")
	ypy.Fragment(w, `<p role="alert">The cart is full.</p>`)
	return
}
```

`ypy.Reselect(w, "#message")` keeps only part of the response.

## 3. An event the server fires

Other parts of the page can refresh themselves when something happens. The server only names the event:

```go
ypy.Trigger(w, "cart-changed", map[string]any{"count": count})
```

```html
<span id="cart-count" ypy-get="/cart/count" ypy-trigger="cart-changed from:body" ypy-swap="innerHTML">0</span>
```

The event fires on the element that sent the request and bubbles to `<body>`. The count asks the server for its new
value. This costs one more request, but the server code for the action does not need to know who is on the page.

## How it works

- Partials, `YPY-Retarget` and `YPY-Trigger` are all applied after the response arrives. The order of a swap is: main
  content, then `ypy-select-oob` pieces, then `ypy-swap-oob` pieces, then partials.
- `ypy-swap-oob="true"` on an element of the response swaps it into the element with the same `id`. It exists for
  responses written before partials. In new code, a partial says the same thing with an explicit target.
- If several pieces must change together or not at all, see [atomic swaps](../reference/atomic.md).

## Variations

- `ypy-select-oob="#cart-count,#toast"` on the form picks those elements out of the response and swaps each into the
  element with the same id.
- A partial can use any swap style, for example `Swap: "beforeend"` to add a message to a list of messages.

## See also

- [Multiple targets in one response](../guides/multiple-targets.md)
- [`<ypy-partial>`](../reference/tags/ypy-partial.md)
- [Response headers](../reference/headers.md)
- [Atomic swaps](../reference/atomic.md)
