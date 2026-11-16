---
order: 40
title: Delete row
description: Confirm in the browser, send DELETE, and remove the table row when the server answers. Covers why the handler returns 200 and not 204.
---

# Delete row

Each row of a table has a **Delete** button. A click asks the user to confirm, sends a `DELETE` request, and removes the
row when the server says it is done.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<table>
  <thead><tr><th>Name</th><th>Email</th><th></th></tr></thead>
  <tbody>
    <tr>
      <td>Ada</td><td>ada@example.com</td>
      <td>
        <button ypy-delete="/contacts/1" ypy-confirm="Delete Ada?"
                ypy-target="closest tr" ypy-swap="delete">Delete</button>
      </td>
    </tr>
    <tr>
      <td>Grace</td><td>grace@example.com</td>
      <td>
        <button ypy-delete="/contacts/2" ypy-confirm="Delete Grace?"
                ypy-target="closest tr" ypy-swap="delete">Delete</button>
      </td>
    </tr>
  </tbody>
</table>
```

## The server (Go)

```go
package main

import (
	"net/http"

	ypy "github.com/ypy-develop/go"
)

func deleteContact(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !store.Delete(r.PathValue("id")) {
			http.NotFound(w, r)
			return
		}
		ypy.Fragment(w, "") // 200 with an empty body
	}
}
```

Register it with `mux.Handle("DELETE /contacts/{id}", ypy.Middleware(deleteContact(store)))`.

## How it works

- `ypy-delete` sends `DELETE /contacts/1` when the button is clicked. A button's default trigger is `click`.
- `ypy-confirm` shows the browser's confirm dialog first. If the user cancels, no request is made.
- `ypy-target="closest tr"` picks the table row that holds the button. `closest` and the other extended selectors are
  listed under [`ypy-target`](../reference/attributes/index.md).
- `ypy-swap="delete"` removes the target whatever the response holds, so the answer can be an empty `200`. The server
  answers `200`, not `204`, because by default a `204` is not swapped at all (`config.noSwap`), and the row would
  stay.
- If the server answers with an error status, the row stays. The failure goes where
  [the error convention](../reference/attributes/ypy-error.md) says.

## Variations

- Fade the row out before it is removed with a swap delay: `ypy-swap="delete swap:300ms"` and a CSS rule on the
  `ypy-swapping` class that sets `opacity: 0` with a transition. See [Animations](animations.md).
- Replace the browser dialog with your own with the `ypy:confirm` event, or with a [layer](dialogs-layers.md).
- To confirm with a typed answer, see [Confirm and prompt](confirm-and-prompt.md).
- Keep the user's place and give feedback: add a `<ypy-partial id="toast">` to the answer, as in
  [Bulk update](bulk-update.md).

## See also

- [Edit row](edit-row.md)
- [`ypy-confirm` and the `ypy:confirm` event](../reference/events/list.md)
- [Security checklist](../guides/security-checklist.md), because a `DELETE` is a state change and needs a CSRF check
