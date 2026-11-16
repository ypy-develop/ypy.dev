---
title: "Reorder a list by hand: the ypy-sort extension"
description: "Lets a person reorder a list by dragging or with the keyboard, and sends the new order of the item ids to the server in one request."
order: 380
---
# Reorder a list by hand: the `ypy-sort` extension

`ypy-sort` lets a person put the items of a list in a new order, by dragging them or with the keyboard, and sends the
new order to the server. It is for manual order: a task list, a playlist, a ranking the user arranges. It never sorts by a field or a
column, and it never compares the content of the items.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-sort.js"></script>

<ul ypy-sort ypy-sort-url="/tasks/order" class="space-y-2">
  <li id="task-1" class="rounded border p-2">Write</li>
  <li id="task-2" class="rounded border p-2">Review</li>
  <li id="task-3" class="rounded border p-2">Ship</li>
</ul>
```

Drag "Ship" above "Write" and the page sends one `POST /tasks/order` with the field `id` three times: `task-3`,
`task-1`, `task-2`. If the server refuses or does not answer, the list goes back to the old order.

If you set `ypy.config.extensions`, list `ypy-sort` in it, or the extension is refused. It is listed with the other
official extensions in [Extensions](index.md) and it is part of `ypy.max`. It uses the browser's native drag and
drop, so it needs no style and no extra script.

## Attributes

| Attribute | On | Meaning |
|---|---|---|
| `ypy-sort` | the list container | turns on reordering for its children that have an `id`. `ypy-sort="false"` leaves the container to the page |
| `ypy-sort-url` | the same container | the URL that gets a `POST` when an item is put in a new place. Without it the list only reorders and fires `ypy:sort:change` |

Neither attribute is inherited. A container inside an `ypy-ignore` element is not touched.

The items are the element children of the container that have an `id`. Children without an `id` (a heading, a
"drop here" row) stay where they are and cannot be dragged. Each item gets `draggable="true"` and, if it has no
`tabindex`, `tabindex="0"`; an attribute you set yourself is kept. Items that a later swap adds are tagged as well.

## Dragging

While an item is dragged it has `data-ypy-state="dragging"`, and the list moves as the pointer passes over the other
items (a row of items is crossed by the horizontal position, a column by the vertical one). When the item is dropped
somewhere new, the extension fires `ypy:sort:change` and then sends the request. If the drop does not happen (the user
presses Escape or drops outside the list), everything goes back and `ypy:sort:revert` fires with the reason `cancelled`.

Style the states with the attribute:

```css
[data-ypy-state~="dragging"], [data-ypy-state~="grabbed"] { opacity: .6; outline: 2px solid #6366f1; }
```

## Keyboard

On a focused item:

| Key | Effect |
|---|---|
| `Space` | picks the item up (`data-ypy-state="grabbed"`); pressed again it puts the item down and sends the order |
| `ArrowUp`, `ArrowLeft` | moves the picked-up item one place before; with nothing picked up, moves the focus to the previous item |
| `ArrowDown`, `ArrowRight` | the same, one place after |
| `Escape` | with an item picked up, puts everything back and sends nothing |

Nothing is sent while the item is moved, only when it is put down. A key pressed inside a field in an item (an `<input>`
in the `<li>`) is left to the field. Keys with `Alt`, `Ctrl` or `Meta` are ignored.

### Spoken messages

The extension has no text of its own. To announce what happens to screen readers, give it messages: it writes them in a
polite live region (`role="status"`, created on first use, hidden with an adopted style sheet, so it works under a strict
CSP). `{position}` and `{total}` are replaced. A message that is not set is not said.

```html
<meta name="ypy-config" content='{"sort":{"messages":{
  "grab": "Picked up. Place {position} of {total}. Use the arrow keys, Space to drop.",
  "move": "Place {position} of {total}",
  "drop": "Dropped at place {position} of {total}",
  "cancel": "Cancelled",
  "revert": "The new order was not saved"}}}'>
```

## The request

When the order changed, the extension sends `POST` to `ypy-sort-url`. The body has the field `id` once per item, in the
new order, plus whatever the container's `ypy-vals` and an enclosing form carry:

```
id=task-3&id=task-1&id=task-2&list=todo
```

It is an ordinary YPY request: indicators, `ypy-confirm`, `ypy-headers`, the [CSRF](csrf.md) header of `ypy-csrf` and the request events
all apply. The response is not swapped (the swap is `none`), but out-of-band elements and partials in it are, so the
server can answer with an updated counter.

| Answer | Result |
|---|---|
| `2xx` (a `204` is fine) | the new order stays |
| `400` or more | the list goes back, `ypy:sort:revert` fires with `reason: "failed"` |
| no answer (the connection fails) | the same |

On a revert, the focused item keeps its focus.

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:sort:change` | the container | `ids` (the new order), `item` (the element that moved). Cancelable: cancel it to put the list back (the revert has `reason: "cancelled"`) and send nothing |
| `ypy:sort:revert` | the container | `ids` (the order after it went back), `reason` (`cancelled` or `failed`) |

`ypy:sort:change` does not fire when the order is the same as before.

```js
document.addEventListener('ypy:sort:change', evt => {
  if (evt.detail.ids.length > 50) evt.preventDefault();   // too long to reorder
});
```

## Configuration

| Key | Meaning |
|---|---|
| `sort.messages.grab` | said when an item is picked up |
| `sort.messages.move` | said when a picked-up item moves |
| `sort.messages.drop` | said when an item is put down |
| `sort.messages.cancel` | said when a pick-up is cancelled with Escape |
| `sort.messages.revert` | said whenever the list is put back: a refusal, a failed connection, a cancelled change or a cancelled pick-up |

All are strings; none is set by default.

## The Go handler

The server gets the ids in the new order. Check that they are the ones you expect, save, and answer `204`. Any status of
400 or more puts the list back in the browser.

```go
func reorder(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ids := r.PostForm["id"] // []string{"task-3", "task-1", "task-2"}
	if err := tasks.Reorder(r.Context(), ids); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // the list goes back
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

The values are the `id` attributes of the elements, `task-3` in the example: strip the prefix in the handler, or choose
ids you can read as they are. Never trust the list: check that every id belongs to the current user.

## Pitfalls

- This is not a sort by value. To sort a table by a column, ask the server for the rows in that order.
- Only children with an `id` move. The ids must be unique in the page, and are what the server gets.
- A swap that replaces the container is fine. The new container is set up when it is processed, and items a swap
  adds to the old one are tagged too.
- Do not swap the answer. The request uses `swap:none`; to refresh the list send the new HTML in an out-of-band
  element.
- A second drag while the first request is in flight starts a second request. Make the handler idempotent: it
  receives the whole order, not a move.
- A list inside a form sends the form's fields too. Put the list outside the form if you do not want them.
