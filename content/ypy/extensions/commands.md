---
title: "Client commands from the server: the ypy-commands extension"
description: "The server sends small client-side effects (highlight a row, move focus, hide a banner, fire an event) in one YPY-Commands header, from a fixed set."
order: 60
---
# Client commands from the server: the `ypy-commands` extension

Some answers need small side effects that are not HTML: highlight a row, move the focus to a field, hide a banner, tell
the rest of the page that something was saved. `ypy-commands` lets the server send them in one response header,
`YPY-Commands`, as a list of typed commands. There is no `eval` and no script: each command is one of a fixed set, and
the extension checks every field before it runs.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-commands.js"></script>

<form ypy-post="/tasks/4" ypy-target="#row-4">
  <input name="title" value="Write the docs">
  <button>Save</button>
</form>
<input id="name">
```

The server answers with the new row and one header:

```http
HTTP/1.1 200 OK
Content-Type: text/html
YPY-Commands: [{"type":"addClass","target":"#row-4","class":"saved"},{"type":"focus","target":"#name"}]

<tr id="row-4">...</tr>
```

The row is swapped, then `#row-4` gets the class `saved` and `#name` gets the focus.

If you set `ypy.config.extensions`, list `ypy-commands` in it, or the extension is refused. It is listed with the other
official extensions in [Extensions](index.md) and it is part of `ypy.max`. Without it the header is ignored.

## The header

`YPY-Commands` holds a JSON list. Each command is an object with a `type` and a `target`, a CSS selector. The command
acts on every element the selector matches, in document order. The commands run in the order of the list.

- They run after the response was swapped (in `ypy:finally:request`), so a target can be an element the response
  brought.
- They also run for a response that is not swapped: a `204`, or an error status.
- A command whose selector matches nothing is skipped without a warning, because the element may be gone.
- Non-ASCII text must be written as `\uXXXX` inside the JSON, because a header value is not UTF-8 text.

## Commands

| `type` | Fields | What it does |
|---|---|---|
| `addClass` | `class`: one or more names, separated by spaces | adds the classes |
| `removeClass` | `class` | removes the classes |
| `toggleClass` | `class`, optional `force` (`true` or `false`) | toggles each class; with `force` it adds or removes |
| `setAttribute` | `name`, `value` (a string, number or boolean; written as a string) | sets the attribute |
| `removeAttribute` | `name` | removes the attribute |
| `focus` | none | calls `focus()` |
| `scrollIntoView` | optional `behavior`, `block`, `inline` (strings) | scrolls the element into view; the defaults are `auto`, `start` and `nearest` |
| `remove` | none | takes the element out, like the core does for removed content: its timers and listeners stop, and `ypy:before:cleanup` fires |
| `dispatch` | `event`, optional `detail`, optional `bubbles` (default `true`) | dispatches a `CustomEvent` (cancelable) on the target. The `target` is optional: without it the event goes to `document` |
| `show` | none | removes the `hidden` attribute |
| `hide` | none | sets the `hidden` attribute |

```json
[
  {"type": "toggleClass", "target": ".card", "class": "open", "force": true},
  {"type": "setAttribute", "target": "#total", "name": "aria-live", "value": "polite"},
  {"type": "scrollIntoView", "target": "#row-4", "behavior": "smooth", "block": "center"},
  {"type": "dispatch", "event": "task:saved", "detail": {"id": 4}}
]
```

The event of the last command goes to `document`, so the page listens with
`document.addEventListener('task:saved', evt => evt.detail.id)`.

## What is refused

A command is not run when its `type` is not in the table, a field is missing or of the wrong kind, or the selector
is not valid. A console warning says why, `ypy:commands:ignored` fires, and the next command runs.

`setAttribute` cannot write code, because the header must not become a way around the page's CSP. These names and
values are refused:

- an event handler (`onclick`, `onerror`, in any case),
- `style` and `srcdoc`,
- an attribute of YPY itself (`ypy-*`, `data-ypy-*`, and the custom prefix of `ypy.config.prefix` when you set one),
- a value that is a `javascript:` or `vbscript:` URL, or a `data:text/html` URL (leading spaces and line breaks do not
  hide it),
- a name with spaces, quotes, `<`, `>`, `/` or `=`.

An ordinary `href`, `title`, `aria-*` or `data-*` attribute is fine.

A header that is not valid JSON, is not a list, or holds more commands than `commands.max` is ignored as a whole.

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:before:command` | the element that sent the request | `ctx`, `command`, `targets` (the elements the selector matched). Cancelable: cancel it to skip that one command |
| `ypy:commands:ignored` | the element that sent the request | `command` (what was refused) and `reason` |

The `reason` is one of: `not valid JSON`, `not a list`, `more than <max>` (for a long list, `command` is the text
`<n> commands`), `unknown command`, `not allowed` (a blocked attribute), `invalid <field>` (for example `invalid class`
or `invalid event`) and `invalid target`.

```js
// a page that does not allow a command that hides things
document.addEventListener('ypy:before:command', evt => {
  if (evt.detail.command.type === 'hide') evt.preventDefault();
});
```

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `commands.max` | `100` | most commands in one list; a longer list is ignored as a whole |

```html
<meta name="ypy-config" content='{"commands":{"max":20}}'>
```

## From JavaScript

`ypy.commands.run(list, source)` runs a list the same way, for code that gets its commands somewhere else. `list` is an
array or a JSON string; `source` is the element the events fire on (it falls back to `document.body` when it is not in
the page).

```js
ypy.commands.run([{type: 'addClass', target: '#a', class: 'one'}]);
```

## The Go handler

The Go SDK has no helper for this header. Build the list, marshal it and set the header with `net/http`:

```go
package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type command map[string]any

// ascii writes every non-ASCII rune as \uXXXX, as a header value needs.
func ascii(b []byte) string {
	var sb strings.Builder
	for _, r := range string(b) {
		if r < 0x80 {
			sb.WriteRune(r)
		} else if r > 0xFFFF {
			r -= 0x10000
			sb.WriteString("\\u" + hex(0xD800+(r>>10)) + "\\u" + hex(0xDC00+(r&0x3FF)))
		} else {
			sb.WriteString("\\u" + hex(r))
		}
	}
	return sb.String()
}

func hex(r rune) string {
	const d = "0123456789abcdef"
	return string([]byte{d[r>>12&15], d[r>>8&15], d[r>>4&15], d[r&15]})
}

func saveTask(w http.ResponseWriter, r *http.Request) {
	// ... save the task ...
	list, _ := json.Marshal([]command{
		{"type": "addClass", "target": "#row-4", "class": "saved"},
		{"type": "focus", "target": "#name"},
	})
	w.Header().Set("YPY-Commands", ascii(list))
	w.Write([]byte(`<tr id="row-4">...</tr>`))
}
```

Set the header before the first write to the body. If the text of your commands is plain ASCII, `string(list)` is
enough. Many servers and proxies limit the size of the headers; keep the list short and, for anything bigger, send
HTML instead.

## Pitfalls

- Commands are not a way to run code. There is no command that evaluates text. If you need page logic, `dispatch`
  an event and let a listener of your page do the work.
- `show` and `hide` use the `hidden` attribute. A CSS rule that sets `display` on the element wins over it. Add
  `[hidden] { display: none }` for such elements.
- The commands run after the swap. A command that targets something the swap replaced acts on the new element.
  A command that targets something the swap removed finds nothing and is skipped.
- `focus` needs a focusable element. On a `<div>` without `tabindex` it does nothing, as in the browser.
- One list per response. The header is read once per request.
- A command that matches many elements acts on all of them. Use an `id` for one.
- An unknown command does not stop the list. Check the console during development: the warning names the command
  and the reason.
