---
order: 190
title: Keyboard shortcuts
description: Start a request from Alt+N or the slash key with a ypy-trigger event filter, or use the HTML accesskey attribute when you want no script at all.
---

# Keyboard shortcuts

A button can start its request on a key press as well as on a click. `ypy-trigger` listens on any element with `from:`,
and a filter in square brackets decides which key counts. Filters are author code, so they need `config.callbacks`
turned on. When you do not want that, an HTML `accesskey` gives a shortcut with no script at all.

## The page

```html
<meta name="ypy-config" content='{"callbacks":"all"}'>

<button ypy-get="/notes/new" ypy-target="#editor" ypy-swap="innerHTML"
        ypy-trigger="click, keydown[altKey&&key=='n'] from:body prevent">
  New note <kbd>Alt N</kbd>
</button>

<button ypy-get="/search/panel" ypy-target="#editor" ypy-swap="innerHTML"
        ypy-trigger="click, keydown[key=='/'] from:body target:body prevent">
  Search <kbd>/</kbd>
</button>

<div id="editor"></div>
```

The first button starts on a click or on Alt+N from anywhere on the page. The second starts on `/` only when no field has
the focus: `target:body` lets the event through only when it comes from `<body>` itself, so typing a slash into an input
does nothing.

## What the server answers

The server does not know the request came from a key. It answers the address like any other, `200` with a fragment and
`Vary: YPY-Request`:

```html
<form action="/notes" method="post">
  <label for="note">Note</label>
  <textarea id="note" name="body" autofocus></textarea>
  <button>Save</button>
</form>
```

## How it works

- `keydown[altKey&&key=='n']` is an event name with a filter. The text in the brackets is an expression that sees the
  properties of the event it names (`altKey`, `key`). When it is false the trigger does nothing.
- `from:body` moves the listener to `<body>`. Other values are `document`, `window`, `closest`, `find`, `next`,
  `previous` and `outside`.
- `prevent` calls `preventDefault()` only when the filter passed, so other keys keep working.
- `config.callbacks` is `"none"` by default. The values are `"all"`, or `"nonce"` to run author code only through the
  [`ypy-csp`](../extensions/csp.md) extension on a page with a strict CSP.
- On a page whose CSP forbids `'unsafe-eval'`, add the [`ypy-expr`](../extensions/expr.md) extension: it runs filters
  without `eval` and without `new Function`.
- A shortcut shown on a button is also a hint for people who use a keyboard. Keep the button reachable with Tab.

## Variations

- No script: `<button ypy-get="/notes/new" ypy-target="#editor" accesskey="n">New note</button>`. The browser decides
  the modifier keys (Alt, Alt+Shift or Ctrl+Alt, by browser and system), and the click it produces starts the request.
- Debounce a search while typing: `ypy-trigger="keyup changed delay:300ms"`. See [Active search](active-search.md).

## See also

- [Triggers](../index.md#triggers) and [`ypy-trigger`](../reference.md#core-attributes)
- [Events](../reference.md#events)
- [Security](../index.md#security), for what `callbacks` allows
