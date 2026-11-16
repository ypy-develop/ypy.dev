---
order: 160
title: Confirm and prompt
description: Ask yes or no with ypy-confirm, or collect a reason with ypy-prompt that reaches the server in a header. A cancel sends no request.
---

# Confirm and prompt

Two small questions before a request goes out. `ypy-confirm` asks yes or no. `ypy-prompt` asks for a line of text
and sends the answer to the server in the `YPY-Prompt` header. If the user says no, or cancels, no request is made.

## The page

`ypy-prompt` is an extension, so it needs its own script.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-prompt.min.js"></script>

<ul id="items">
  <li id="item-5">
    Quarterly report
    <button ypy-delete="/items/5"
            ypy-confirm="Delete the quarterly report?"
            ypy-target="closest li"
            ypy-swap="delete">Delete</button>
  </li>
</ul>

<div id="order-12">
  <button ypy-post="/orders/12/cancel"
          ypy-prompt="Why are you cancelling?"
          ypy-target="#order-12"
          ypy-swap="innerHTML"
          ypy-error-target="#cancel-error"
          ypy-error-swap="innerHTML">Cancel order</button>
  <p id="cancel-error"></p>
</div>
```

## The server (Go)

```go
package items

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	// ... delete item r.PathValue("id") ...
	// 200 with an empty body: the default `noSwap` list holds 204, so a 204 would leave the row in place.
	ypy.Fragment(w, "")
}

func CancelOrder(w http.ResponseWriter, r *http.Request) {
	reason := strings.TrimSpace(ypy.FromRequest(r).Prompt) // already decoded
	if reason == "" {
		ypy.ValidationError(w, `<p role="alert">A reason is required.</p>`)
		return
	}
	// ... cancel order r.PathValue("id") and store the reason ...
	ypy.Fragment(w, `<p>Order cancelled: `+html.EscapeString(reason)+`</p>`)
}

func Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /items/{id}", Delete)
	mux.HandleFunc("POST /orders/{id}/cancel", CancelOrder)
	return ypy.Middleware(mux)
}
```

## How it works

- `ypy-confirm="..."` shows the browser's `confirm()` dialog with that text before the request. "Cancel" drops the
  request. `ypy-swap="delete"` removes the target whatever the answer holds, so the server's reply can be empty.
- `ypy-prompt="..."` shows `prompt()`. A cancel aborts the request. Any other answer, even an empty one, is sent as
  the `YPY-Prompt` header, URI-encoded. `ypy.FromRequest(r).Prompt` decodes it.
- The prompt answer is user input. Validate it, escape it before it goes back into HTML, and never trust that it came
  from the dialog: a request can carry the header without one.
- `ypy:prompt` fires on the element after the user answers. Cancel it to abort the request.
- Both attributes are inherited with `:inherited`: `ypy-confirm:inherited="Are you sure?"` on a table asks before every
  delete button inside it.

## Variations

- Your own dialog for `confirm`: listen to `ypy:confirm`, call `event.preventDefault()`, open a `<dialog>`, and
  call `issueRequest()` or `dropRequest()` from its buttons. The request waits until one of them runs.

  ```js
  document.addEventListener('ypy:confirm', (event) => {
    event.preventDefault();
    const { issueRequest, dropRequest } = event.detail;
    const dialog = document.querySelector('#sure');
    dialog.addEventListener('close', () => (dialog.returnValue === 'yes' ? issueRequest() : dropRequest()), { once: true });
    dialog.showModal();
  });
  ```
- A dialog for `prompt`: define `window.ypyPrompt` as a function that returns a string, or `null` to cancel. It
  must be synchronous.
- A confirmation written in code: `ypy-confirm="js:..."` runs an expression, which may return a promise.
  It needs `config.callbacks` turned on (see [security](../index.md#security)).
- A bigger decision deserves a page of its own: open it in a [layer](dialogs-layers.md) instead.

## See also

- [`ypy-prompt`](../extensions/prompt.md) and [`ypy-confirm`](../reference.md#core-attributes)
- [Delete row](delete-row.md) and [dialogs with layers](dialogs-layers.md)
