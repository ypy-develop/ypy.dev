---
order: 40
title: Examples
description: Interface patterns built with YPY attributes, from click to edit to infinite scroll, each with the HTML, what the server answers and a demo you can use.
---

# Examples

Each page is one pattern. It says what the pattern does, shows the HTML with its `ypy-*` attributes, shows what the
server answers (an HTML fragment, a status, the headers) and, for the twelve first ones, runs the pattern on the page
itself.

## Read the demos

- The demos are real. Each one sends requests to this site, which answers with HTML fragments, and YPY swaps them into
  the page. Nothing is simulated in the browser.
- What you change is kept for you alone, in a cookie that lasts an hour. No other visitor sees it, and there is no
  database.
- The server side is described, not coded. Any server that answers an HTML fragment works, in any language. The
  [Docs](../index.md#server-side) page lists what a server has to know.

## What every example assumes

- The script. Every example needs the core:

  ```html
  <script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
  ```

  The same file is also served at `https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js`. When an example needs an extension, it
  names it. The `ypy.max` build already includes most of them.
- The default swap. Without `ypy-swap`, the response is morphed into the target: its children are updated in place
  (`innerMorph`). The examples that replace the element itself say `ypy-swap="outerHTML"`.
- Attributes are not inherited, unless they end in `:inherited`. Write `ypy-target` and `ypy-swap` on the element that
  makes the request. See [Attribute inheritance](../index.md#attribute-inheritance).
- Answers are fragments with `Vary: YPY-Request`, so a cache never hands one to a visitor who asked for a page.
- CSRF. A `POST`, `PUT` or `DELETE` needs a CSRF token. A form carries it in a hidden `_token` field; a button that is
  not in a form sends it as the `X-CSRF-Token` header with `ypy-headers`. See [CSRF](../index.md#csrf).
- No script of yours. The approach is to start from links and forms that work and let YPY make them faster.

## The patterns

| Example | What it shows | Demo | Extension needed |
|---|---|---|---|
| [Click to edit](click-to-edit.md) | Swap a read-only block for a form, and back | yes | none |
| [Bulk update](bulk-update.md) | One form of checkboxes, one answer that updates the table and a message | yes | none |
| [Click to load](click-to-load.md) | A Load more button that adds rows and moves itself | yes | none |
| [Delete row](delete-row.md) | Confirm, delete on the server, remove the row | yes | none |
| [Edit row](edit-row.md) | A table row that becomes inputs and back | yes | none |
| [Lazy loading](lazy-loading.md) | Load a slow part after the page, at load or when seen | yes | [`ypy-lazy`](../extensions/lazy.md), optional |
| [Inline validation](inline-validation.md) | Check a field on the server when the user leaves it | yes | [`ypy-validate`](../extensions/validate.md), optional |
| [Infinite scroll](infinite-scroll.md) | The last row asks for the next page when it appears | yes | none |
| [Active search](active-search.md) | Search while typing, with a delay | yes | none |
| [Progress bar](progress-bar.md) | Follow a long job by asking for its progress | yes | [`ypy-poll`](../extensions/poll.md), optional |
| [Value select](value-select.md) | A second select that fills from the first | yes | none |
| [Tabs](tabs.md) | Load one panel at a time and mark the current tab | yes | none |
| [Animations](animations.md) | Fade and slide content in and out with CSS classes | no | none |
| [File upload](file-upload.md) | Send a file with a progress bar | no | [`ypy-upload`](../extensions/upload.md), optional |
| [Preserve file inputs](preserve-file-inputs.md) | Keep the chosen file when a form comes back with errors | no | none |
| [Reset user input](reset-user-input.md) | Clear a form after a successful submit | no | none |
| [Confirm and prompt](confirm-and-prompt.md) | Ask the user before a request, or for a typed answer | no | [`ypy-prompt`](../extensions/prompt.md), for the prompt |
| [Dialogs with layers](dialogs-layers.md) | Modal, drawer and popup content loaded from the server | no | [`ypy-layers`](../extensions/layers.md) |
| [Keyboard shortcuts](keyboard-shortcuts.md) | Start a request from a key press | no | none |
| [Drag and drop sort](drag-and-drop-sort.md) | Reorder a list by hand and save the order | no | [`ypy-sort`](../extensions/sort.md) |
| [Updating other content](updating-other-content.md) | One response that changes several places | no | none |
| [Session expired](session-expired.md) | Send the user to sign in when a request is refused | no | none |
| [Live updates](live-updates.md) | The server pushes changes to the page | no | [`ypy-sse`](../extensions/sse.md) |
| [Optimistic UI](optimistic-ui.md) | Show the result before the server answers, and undo it on failure | no | [`ypy-optimistic`](../extensions/optimistic.md) |
| [Polling](polling.md) | Ask again every few seconds, and stop when the server says so | no | [`ypy-poll`](../extensions/poll.md), optional |

An extension marked optional improves a pattern that already works with the core alone.

## Where next

- [Docs](../index.md): the long page that goes through every part in order.
- [Reference](../reference.md): attributes, headers, events and settings in tables.
- [Extensions](../extensions/index.md): the optional scripts named in the table above.
