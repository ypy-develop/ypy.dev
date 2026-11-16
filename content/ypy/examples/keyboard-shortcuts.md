---
order: 190
title: Keyboard shortcuts
description: Start a request from Alt+N or the slash key with a ypy-trigger event filter, or use the HTML accesskey attribute when you want no script.
---

# Keyboard shortcuts

A button can start its request on a key press as well as on a click. `ypy-trigger` listens on any element with
`from:`, and a filter in square brackets decides which key counts. Filters run author code, so they need
`callbacks` turned on. When you do not want that, an HTML `accesskey` gives a shortcut with no script at all.

## The page

```html
<meta name="ypy-config" content='{"callbacks":"all"}'>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

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

The first button starts on a click or on Alt+N from anywhere on the page. The second starts on `/` only when no field
has the focus: `target:body` fires only when the key event comes from `<body>` itself, so typing a slash into an input does
nothing.

## The server (Go)

The server does not know the request came from a key. It answers the URL like any other:

```go
func (h *Handler) NewNote(w http.ResponseWriter, r *http.Request) {
	ypy.Fragment(w, `<form action="/notes" method="post">
<label>Note <textarea name="body" autofocus></textarea></label>
<button>Save</button>
</form>`)
}
```

## How it works

- `keydown[altKey&&key=='n']` is an event name with a filter. The text in the brackets is an expression that sees the
  properties of the event it names (`altKey`, `key`). When it is false the trigger does nothing.
- `from:body` moves the listener to `<body>`. Other values are `document`, `window`, `closest`, `find`, `next`,
  `previous` and `outside`.
- `prevent` calls `preventDefault()` only when the filter passed, so other keys keep working.
- `config.callbacks` is `"none"` by default. The values are `"all"`, or `"nonce"` to run author code only through the
  [`ypy-csp`](../extensions/csp.md) extension on a page with a strict CSP.
- A shortcut shown on a button is also a hint for people who use a keyboard. Keep the button reachable with Tab.

## Variations

- No script: `<button ypy-get="/notes/new" ypy-target="#editor" accesskey="n">New note</button>`. The browser
  decides the modifier keys (Alt, Alt+Shift or Ctrl+Alt, by browser and system), and the click it produces starts the request.
- Run a bit of script on a key with `ypy-on`: `<input id="q" ypy-on:keydown="if (event.key === 'Escape') this.value = ''">`.
  It needs `callbacks` as well.
- Debounce a search while typing: `ypy-trigger="keyup changed delay:300ms"`. See [Active search](active-search.md).

## See also

- [`ypy-trigger` in the attribute reference](../reference/attributes/index.md)
- [Events](../reference/events/index.md)
- [Security checklist](../guides/security-checklist.md) for what `callbacks` allows
- [Accessibility](../guides/accessibility.md)
