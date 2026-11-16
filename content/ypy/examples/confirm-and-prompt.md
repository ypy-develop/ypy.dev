---
order: 160
title: Confirm and prompt
description: Ask yes or no with ypy-confirm, or collect a reason with ypy-prompt that reaches the server in a header. A cancel sends no request.
---

# Confirm and prompt

Two small questions before a request goes out. `ypy-confirm` asks yes or no. `ypy-prompt` asks for a line of text and
sends the answer to the server in the `YPY-Prompt` header. If the visitor says no, or cancels, no request is made.

## The page

`ypy-prompt` is an extension, so it needs its own script. `ypy-confirm` is part of the core.

```html
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

## What the server answers

| Request | Status | Body |
|---|---|---|
| `DELETE /items/5` | `200` | Empty. A `204` would not be swapped, and the row would stay |
| `POST /orders/12/cancel` with `YPY-Prompt: Customer%20changed%20their%20mind` | `200` | `<p>Order cancelled: Customer changed their mind</p>` |
| `POST /orders/12/cancel` with an empty `YPY-Prompt` | `422` | `<p role="alert">A reason is required.</p>` |

The header value is URI-encoded by the extension, so the server decodes it before use. The reason is user input: check
it, and escape it before it goes back into HTML. Every answer is a fragment with `Vary: YPY-Request`.

## How it works

- `ypy-confirm="…"` shows the browser's `confirm()` dialog with that text before the request. Cancel drops the
  request. `ypy-swap="delete"` removes the target whatever the answer holds, so the reply can be empty.
- `ypy-prompt="…"` shows `prompt()`. A cancel aborts the request. Any other answer, even an empty one, is sent as the
  `YPY-Prompt` header.
- Never trust that the header came from the dialog: a request can carry it without one.
- `ypy:prompt` fires on the element after the visitor answers. Cancel the event to abort the request.
- `ypy-error-target` sends the `422` to the message element, because the button is not inside a form.

## Variations

- Your own dialog for `confirm`: listen to `ypy:confirm`, call `event.preventDefault()`, open a `<dialog>` and call
  `issueRequest()` or `dropRequest()` from its buttons. The request waits until one of them runs. Put this in a script
  file your page loads:

  ```js
  document.addEventListener('ypy:confirm', (event) => {
    event.preventDefault();
    const { issueRequest, dropRequest } = event.detail;
    const dialog = document.querySelector('#sure');
    dialog.addEventListener('close', () => (dialog.returnValue === 'yes' ? issueRequest() : dropRequest()), { once: true });
    dialog.showModal();
  });
  ```

- A dialog for `prompt`: define `window.ypyPrompt` as a function that returns a string, or `null` to cancel. It must be
  synchronous.
- A confirmation written as code: `ypy-confirm="js:…"` runs an expression, which may return a promise. It needs
  `config.callbacks` turned on; see [security](../index.md#security).
- A bigger decision deserves a page of its own: open it in a [layer](dialogs-layers.md) instead.

## See also

- [`ypy-prompt`](../extensions/prompt.md) and [confirming requests](../index.md#confirming-requests)
- [Delete row](delete-row.md) and [dialogs with layers](dialogs-layers.md)
