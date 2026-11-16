---
order: 200
title: Drag and drop sorting
description: Reorder a task list by dragging or with Space and the arrow keys. One POST saves the whole order, and a failed save puts the list back.
---

# Drag and drop sorting

A task list the user arranges by hand. The `ypy-sort` extension uses the browser's drag and drop and a keyboard mode,
and sends the new order to the server in one `POST`. If the server refuses, the list goes back to its old order.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-sort.min.js"></script>

<ul id="tasks" ypy-sort ypy-sort-url="/tasks/order" ypy-vals='{"list":"todo"}'>
  <li id="task-1">Write the report</li>
  <li id="task-2">Review the draft</li>
  <li id="task-3">Send it</li>
</ul>
```

Only children with an `id` can move. Each gets `draggable="true"` and `tabindex="0"`. Drag "Send it" above the first
item and the page sends:

```
POST /tasks/order
id=task-3&id=task-1&id=task-2&list=todo
```

## The server (Go)

```go
func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ids := r.PostForm["id"] // ["task-3", "task-1", "task-2"]
	numbers := make([]int, 0, len(ids))
	for _, id := range ids {
		n, err := strconv.Atoi(strings.TrimPrefix(id, "task-"))
		if err != nil {
			http.Error(w, "unknown task", http.StatusUnprocessableEntity)
			return
		}
		numbers = append(numbers, n)
	}
	// Check that every id is a task of this user, then save the whole order.
	if err := h.store.Reorder(r.Context(), numbers); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // the list goes back
		return
	}
	ypy.NoContent(w)
}
```

Any status of 400 or more puts the list back in the browser. The handler receives the whole order, not a move, so
sending it twice is harmless.

## Styles and spoken messages

While an item is dragged it has `data-ypy-state="dragging"`; a picked-up item in keyboard mode has
`data-ypy-state="grabbed"`.

```css
[data-ypy-state~="dragging"], [data-ypy-state~="grabbed"] { opacity: .6; outline: 2px solid #6366f1; }
```

Give screen readers something to announce with the `sort.messages` settings:

```html
<meta name="ypy-config" content='{"sort":{"messages":{
  "grab": "Picked up. Place {position} of {total}. Use the arrow keys, Space to drop.",
  "move": "Place {position} of {total}",
  "drop": "Dropped at place {position} of {total}",
  "revert": "The new order was not saved"}}}'>
```

## How it works

- On a focused item, `Space` picks it up, the arrow keys move it, `Space` drops it and sends the order, `Escape` puts it back.
- The request is an ordinary YPY request, so `ypy-vals`, `ypy-headers`, `ypy-confirm` and the CSRF header apply. Its
  answer is not swapped, but a `<ypy-partial>` in it is, so the server can answer with an updated counter.
- `ypy:sort:change` fires first and can be cancelled. `ypy:sort:revert` fires when the list goes back, with the reason
  `cancelled` or `failed`.

```js
document.addEventListener('ypy:sort:change', evt => {
  if (evt.detail.ids.length > 50) evt.preventDefault(); // too long to reorder
});
```

## Variations

- Leave out `ypy-sort-url` to reorder on the page only and listen to `ypy:sort:change` yourself.
- To sort by a column, this is the wrong tool: ask the server for the rows in the order you want.

## See also

- [`ypy-sort`](../extensions/sort.md)
- [Keyed and bounded lists](../extensions/list.md)
- [Accessibility](../guides/accessibility.md)
