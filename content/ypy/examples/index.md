---
order: 40
title: Examples
description: 25 interface patterns built with YPY attributes, from click to edit to infinite scroll, each with the page HTML and the Go handler that answers it.
---

# Examples

Each page is one pattern: the HTML with its `ypy-*` attributes, the Go handler that answers it, and a note on why it
works. The Go code uses the official SDK, [`github.com/ypy-develop/go`](../index.md#server-side). If your server is not
Go, the HTML half carries over unchanged, and the [server recipes](../index.md#server-side) show how to answer YPY requests in
other stacks.

## What every example assumes

- The script. Every example needs the core:

  ```html
  <script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
  ```

  The same file is also served at `https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js`. When an example needs an extension, it
  shows the second `<script>`. The `ypy.max` build already includes most of them.
- The default swap. Without `ypy-swap`, the response is morphed into the target: its children are updated in place
  (`innerMorph`). The examples that replace the element itself say `ypy-swap="outerHTML"`.
- Attributes are not inherited. Write `ypy-target` and `ypy-swap` on the element that makes the request. See
  [Attribute inheritance](../index.md#attribute-inheritance).
- The server. Handlers answer with an HTML fragment through `ypy.Fragment`, and the router is wrapped in
  `ypy.Middleware` so every answer carries `Vary: YPY-Request`. User data goes through `html/template`, which escapes it.
- CSRF. A `POST`, `PUT` or `DELETE` needs a CSRF check in a real application. The snippets leave it out to stay short;
  [CSRF](../index.md#csrf) shows how to add it.
- No script. Where it changes the result, the example says what a visitor without JavaScript gets. The approach is to
  start from links and forms that work and let YPY make them faster.

## The patterns

| Example | What it shows | Extension needed |
|---|---|---|
| [Click to edit](click-to-edit.md) | Swap a read-only block for a form, and back | none |
| [Bulk update](bulk-update.md) | One form of checkboxes, one answer that updates the table and a message | none |
| [Click to load](click-to-load.md) | A Load more button that adds rows and moves itself | none |
| [Delete row](delete-row.md) | Confirm, delete on the server, remove the row | none |
| [Edit row](edit-row.md) | A table row that becomes inputs and back | none |
| [Lazy loading](lazy-loading.md) | Load a slow part after the page, at load or when seen | [`ypy-lazy`](../extensions/lazy.md), optional |
| [Inline validation](inline-validation.md) | Check a field on the server when the user leaves it | [`ypy-validate`](../extensions/validate.md) |
| [Infinite scroll](infinite-scroll.md) | The last row asks for the next page when it appears | none |
| [Active search](active-search.md) | Search while typing, with a delay | none |
| [Progress bar](progress-bar.md) | Follow a long job by asking for its progress | [`ypy-poll`](../extensions/poll.md), optional |
| [Value select](value-select.md) | A second select that fills from the first | none |
| [Animations](animations.md) | Fade and slide content in and out with CSS classes | none |
| [File upload](file-upload.md) | Send a file with a progress bar | [`ypy-upload`](../extensions/upload.md), optional |
| [Preserve file inputs](preserve-file-inputs.md) | Keep the chosen file when a form comes back with errors | none |
| [Reset user input](reset-user-input.md) | Clear a form after a successful submit | none |
| [Confirm and prompt](confirm-and-prompt.md) | Ask the user before a request, or for a typed answer | [`ypy-prompt`](../extensions/prompt.md), for the prompt |
| [Dialogs with layers](dialogs-layers.md) | Modal, drawer and popup content loaded from the server | [`ypy-layers`](../extensions/layers.md) |
| [Tabs](tabs.md) | Load one panel at a time and mark the current tab | none |
| [Keyboard shortcuts](keyboard-shortcuts.md) | Start a request from a key press | none |
| [Drag and drop sort](drag-and-drop-sort.md) | Reorder a list by hand and save the order | [`ypy-sort`](../extensions/sort.md) |
| [Updating other content](updating-other-content.md) | One response that changes several places | none |
| [Session expired](session-expired.md) | Send the user to sign in when a request is refused | none |
| [Live updates](live-updates.md) | The server pushes changes to the page | [`ypy-sse`](../extensions/sse.md) |
| [Optimistic UI](optimistic-ui.md) | Show the result before the server answers, and undo it on failure | [`ypy-optimistic`](../extensions/optimistic.md) |
| [Polling](polling.md) | Ask again every few seconds, and stop when the server says so | [`ypy-poll`](../extensions/poll.md), optional |

An extension marked optional improves a pattern that already works with the core alone.

## Where next

- [Docs](../index.md): the long page that goes through every part in order.
- [Reference](../reference.md): attributes, headers, events and settings in tables.
- [Extensions](../extensions/index.md): the optional scripts named in the table above.
- [Server](../index.md#server-side): the Go SDK, and recipes for Arandu, Django, Rails and Laravel.
