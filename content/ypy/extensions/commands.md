---
title: "Client commands from the server: the ypy-commands extension"
description: "The server sends small client-side effects (highlight a row, move focus, hide a banner, fire an event) in one YPY-Commands header, from a fixed set."
order: 60
---
# Client commands from the server: the `ypy-commands` extension

Some answers need small side effects that are not HTML: highlight a row, move the focus to a field, hide a banner, tell the
rest of the page that something was saved. `ypy-commands` lets the server send them in one response header,
`YPY-Commands`, as a list of typed commands. There is no `eval` and no script. Each command is one of a fixed set, and the
extension checks every field before it runs. The extension is part of `ypy.max`, and without it the header is ignored.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-commands.min.js"></script>

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

## The header

`YPY-Commands` holds a JSON list. Each command is an object with a `type` and a `target`, a CSS selector. A command acts on
every element the selector matches, in document order, and the commands run in the order of the list.

- They run after the swap, in `ypy:finally:request`, so a target can be an element the response brought. They also run for
  a response that is not swapped, such as a `204` or an error status.
- A command whose selector matches nothing is skipped without a warning, because the element may be gone.
- Write non-ASCII text as `\uXXXX` inside the JSON, because a header value is not UTF-8 text.

## Commands

| `type` | Fields | What it does |
|---|---|---|
| `addClass` | `class`: one or more names, separated by spaces | adds the classes |
| `removeClass` | `class` | removes the classes |
| `toggleClass` | `class`, optional `force` | toggles each class; with `force` it adds or removes |
| `setAttribute` | `name`, `value` (string, number or boolean) | sets the attribute |
| `removeAttribute` | `name` | removes the attribute |
| `focus` | none | calls `focus()` |
| `scrollIntoView` | optional `behavior`, `block`, `inline` | scrolls the element into view; defaults `auto`, `start`, `nearest` |
| `remove` | none | takes the element out, with its timers and listeners; `ypy:before:cleanup` fires |
| `dispatch` | `event`, optional `detail`, optional `bubbles` (default `true`) | dispatches a cancelable `CustomEvent` on the target, or on `document` without a `target` |
| `show`, `hide` | none | removes or sets the `hidden` attribute |

```json
[
  {"type": "toggleClass", "target": ".card", "class": "open", "force": true},
  {"type": "scrollIntoView", "target": "#row-4", "behavior": "smooth", "block": "center"},
  {"type": "dispatch", "event": "task:saved", "detail": {"id": 4}}
]
```

The page listens with `document.addEventListener('task:saved', evt => evt.detail.id)`.

## What is refused

A command is not run when its `type` is unknown, a field is missing or of the wrong kind, or the selector is not valid. A
console warning says why, `ypy:commands:ignored` fires, and the next command runs.

`setAttribute` cannot write code, because the header must not become a way around the page's CSP. These are refused: event
handlers (`onclick` and the like, in any case), `style` and `srcdoc`, attributes of YPY itself (`ypy-*`, `data-ypy-*` and
the custom prefix), a value that is a `javascript:` or `vbscript:` URL or a `data:text/html` URL, and a name with spaces,
quotes, `<`, `>`, `/` or `=`. An ordinary `href`, `title`, `aria-*` or `data-*` attribute is fine. A header that is not
valid JSON, is not a list, or holds more commands than `commands.max` is ignored as a whole.

## Events and configuration

| Event | On | Detail |
|---|---|---|
| `ypy:before:command` | the element that sent the request | `ctx`, `command`, `targets`. Cancelable: cancel it to skip that command |
| `ypy:commands:ignored` | the element that sent the request | `command` and `reason` |

The `reason` is `not valid JSON`, `not a list`, `more than <max>`, `unknown command`, `not allowed`, `invalid <field>` or
`invalid target`.

| Key | Default | Meaning |
|---|---|---|
| `commands.max` | `100` | most commands in one list; a longer list is ignored as a whole |

`ypy.commands.run(list, source)` runs a list from JavaScript. `list` is an array or a JSON string, and `source` is the
element the events fire on.

## Pitfalls

- Commands are not a way to run code. For page logic, `dispatch` an event and let a listener of your page do the work.
- `show` and `hide` use the `hidden` attribute, and a CSS rule that sets `display` wins over it. Add `[hidden] { display: none }`
  for such elements.
- A command that targets something the swap removed finds nothing and is skipped. One that targets something the swap
  replaced acts on the new element.
- `focus` needs a focusable element, so on a `<div>` without `tabindex` it does nothing.
- A command that matches many elements acts on all of them. Use an `id` for one.
- Servers and proxies limit header size. Keep the list short, and for anything bigger send HTML.
