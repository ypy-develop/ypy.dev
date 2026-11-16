---
title: "Signals: the ypy-signals extension"
description: "Client state without a build step or eval: named signals bound to inputs and read by attributes, which the server reads and patches as JSON."
order: 370
---
# Signals: the `ypy-signals` extension

Some page state has no business on the server: whether a box is ticked, what is typed in a field, whether a panel is
open. `ypy-signals` keeps that state in the page as named values called signals. Inputs bind to them and attributes read
them. The server can read them from every request and change them with a JSON patch. The extension needs no build step
and no `eval`.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-signals.min.js"></script>   <!-- ypy.max already includes it -->

<form ypy-signals="name:'', agree:false" ypy-post="/signup" ypy-target="#result">
  <input ypy-bind="name" placeholder="Name">
  <label><input type="checkbox" ypy-bind="agree"> I agree</label>
  <p ypy-show="$name.length > 0">Hello, <b ypy-text="$name"></b></p>
  <button ypy-attr:disabled="!$agree || $name == ''">Send</button>
  <span ypy-show="$_dirty">Unsaved</span>
</form>
```

Expressions are parsed by a small parser into closures and never compiled into code. They work with `callbacks:"none"`
(the default), a strict CSP and Trusted Types.

## Scopes and names

An element with `ypy-signals` is a scope. Its signals are visible to the element and its descendants, scopes nest, and a
name is looked up in the nearest scope that has it. Put signals shared by the whole page on `<body>`.

- `$name` reads a signal, `$user.address.city` and `$list[0]` read inside it. A path that does not exist is `undefined`,
  never an error. Writes copy along the path, so a change is always a new object.
- A name that starts with `_` is local: it is never sent to the server, at any depth.
- A signal nobody declared is created, when something writes it, in the outermost scope that reaches the writer.
- `constructor`, `prototype` and `__proto__` are refused everywhere.
- Every scope has a read-only local signal `$_dirty`.
- In an attribute name a hyphen stands for a capital: `ypy-computed:total-price` is `totalPrice`. In a value, write the
  name as it is.

## Attributes

All of them also exist with the `data-` prefix. They are read when the element is processed, after a swap included.

| Attribute | Value | What it does |
|---|---|---|
| `ypy-signals` | `count:0, name:'x'` (braces optional, JSON works) | Declares signals in a scope. An existing signal keeps its value when the markup is processed again. |
| `ypy-signals.server` | same | Server-owned signals, read-only for the client. |
| `ypy-signals.local`, `ypy-signals.session` | same | Persisted in `localStorage` or `sessionStorage`. The element needs an `id`. |
| `ypy-computed:name` | expression | A derived, read-only signal. Not sent to the server. |
| `ypy-bind` or `ypy-bind:name` | path (`user.name`) | Two-way binding with a form control. |
| `ypy-text` | expression | Sets `textContent`, never markup. |
| `ypy-show` | expression | Hides the element while the value is false. |
| `ypy-class:name` or `ypy-class` | expression | One class (`ypy-class:active="$on"`) or several (`ypy-class="{red: $bad}"`). |
| `ypy-attr:name` or `ypy-attr` | expression | Sets an attribute. `false`, `null` and `undefined` remove it. Refused: `on*`, `srcdoc`, `ypy-*` and script URLs in `href`, `src`, `action` and similar. |
| `ypy-on:event` | statements | An event handler. |
| `ypy-signals-send="false"` | | On the element or an ancestor: its requests do not carry the signals. |

Effects run once when wired and again, in a batch, when a signal they read changes. They track exactly what they read,
and effects that keep waking each other stop after 100 rounds with an error in the console.

## Expressions

One small language for every attribute. Handlers also take statements separated by `;`.

- Literals: numbers, strings (`"..."` or `'...'`), `true`, `false`, `null`, `undefined`, lists and objects.
- Operators: `! - +`, `* / %`, `+ -`, `< > <= >=`, `== != === !==`, `&&`, `||`, `??`, the ternary `a ? b : c`. `&&`, `||` and
  `??` short-circuit.
- Assignment (`=`, `+=`, `-=`, `*=`, `/=`, `%=`, `??=`) and `++`/`--` work in handlers only.
- Reading is lenient. Only own properties are read, so `$text.length` and `$list[1]` work, and `$o.toString` and
  `$list.map` are `undefined`.
- The only names are signals (`$name`) and functions given to `ypy.signals.register`, for example
  `ypy.signals.register('upper', s => String(s).toUpperCase())` makes `upper($name)` available. There is no `window`,
  `document`, `this`, method call, arrow function, template literal or `new`.
- In handlers `$event` is a plain object with `type`, `key`, `code`, `detail`, `value`, `checked` and the modifier keys. It
  never holds an element or the window.

An expression that fails to parse fires `ypy:signals:error` and its attribute does nothing. The rest of the page keeps
working.

## `ypy-bind` and its cadence

`ypy-bind="name"` keeps a form control and a signal equal.

| Control | Signal |
|---|---|
| text, search, email, textarea, other `input` | text, or a number if the signal was a number |
| `checkbox` | boolean |
| `radio` | the `value` of the checked radio (give the group the same `ypy-bind`) |
| `<select>`, `<select multiple>` | the selected value, or an array of values |
| `contenteditable` | the text content |

Without a modifier the field writes the signal on `change`, so typing does not run effects on every key.

| Modifier | The field writes the signal |
|---|---|
| `.live` | on every `input` |
| `.blur` | on `blur` only |
| `.debounce.300ms` | after a pause of 300 ms (`.debounce` alone means 300 ms) |
| `.throttle.300ms` | at most once per 300 ms, with the latest value |

A value typed and not yet written is written before a request starts, so the server never gets a stale field.

## `ypy-on`

`ypy-on:click="$count++; $open = !$open"` runs statements when the event fires. `ypy-on::before:request` is the form for
`ypy:` events. Inside a scope (an element or an ancestor with one of the `ypy-signals` attributes) `ypy-on` belongs to
this extension and takes this language, so no JavaScript is compiled. Outside a scope the core's `ypy-on` is untouched.

Modifiers: `.prevent`, `.stop`, `.once`, `.window` (listen on `window`), `.debounce.300ms` and `.throttle.300ms`. The
listener runs in the capture phase on the element itself, so on an element that also starts a request the handler runs
first and the request sees the new values.

## Signals in requests

A request whose element has signals in reach carries them as one JSON object in the parameter `ypy-signals`:

| Request | Where |
|---|---|
| `GET`, `DELETE` | query parameter `ypy-signals=<url-encoded JSON>` |
| other verbs | form field `ypy-signals` in the body, urlencoded or multipart |

The object holds the signals of every scope around the element (the inner scope wins on a clash), server-owned ones
included. Computed signals and any key starting with `_` are left out. Signals go only to the same origin.

A GET puts signals in the URL, so they reach logs, proxies and, with `ypy-push-url`, the history entry. Keep secrets out
of signals, and use a POST or `ypy-signals-send="false"` for requests that must not expose them.

## Updating from the server

A response with `Content-Type: application/json` is a JSON Merge Patch ([RFC 7396](https://www.rfc-editor.org/rfc/rfc7396))
applied to the signals around the request's element. It is not swapped into the page, whatever the status, so a `422` with
`{"emailError":"taken"}` works. Any other type is handled by the core.

```json
{"count": 5, "user": {"name": "Bo", "age": null}, "tags": ["a"]}
```

An object member merges recursively, `null` removes it, and an array or scalar replaces the value. A key goes to the
nearest scope that has it, and a new key to the outermost scope that reaches the element. Computed signals ignore a
patch. `ypy:signals:patch` fires first and can be cancelled. The [stream](stream.md) event `ypy-patch-signals` uses the
same patch, and `ypy.signals.patch(object, element, onlyIfMissing)` applies one from code.

### Server-owned signals

`ypy-signals.server="plan:'free', planSig:'3fa9c1'"` declares signals the client cannot write. An expression, binding or
handler that tries fires `ypy:signals:readonly` and the field shows the signal again. They change only when the server
patches them or renders the scope's markup again.

Read-only is a convenience of the UI and gives no security, because whoever controls the browser can send any value. The
protection is on the server: sign the values it owns, for example an HMAC of `plan` bound to the session, send the
signature in a server-owned signal, and verify it when it comes back.

## Unsaved state

A scope is dirty while a client-owned signal differs from what was last saved. The saved value is the declared one, and it
moves when a 2xx response to a non-GET request from inside the scope arrives, or when the server patches the key. While
dirty, the scope gets the class `ypy-dirty` (rename it with `ypy.config.classes.dirty`), `dirty` in `data-ypy-state`, and
`$_dirty` is `true`.

## Persistence

`ypy-signals.local="draft:''"` stores the signals in `localStorage` under `ypy-signals:<id>`, and `ypy-signals.session`
uses `sessionStorage`. The element needs an `id`. The stored value replaces the declared default only for keys the
markup declares. Server-owned signals are never persisted.

## Swaps and morphs

The state of a scope lives with its element. A morph that keeps the element keeps the signals, and the extension runs the
effects of the swapped tree again, so signals win over what the markup said. A swap that replaces the scope element
creates a new scope from the declared values, so give it an `id` and persist it, or keep the scope above the swap target.
When an element leaves the page its effects, listeners and pending writes are dropped. `ypy-ignore` keeps the signal
attributes of a subtree inert.

## Security

- No text becomes code: no `eval`, no `new Function`. The language has no identifiers other than signals, literals and
  registered functions.
- Reads use own properties only, and a write never mutates a foreign object.
- `ypy-text` writes text, and `ypy-attr` refuses event handlers, `srcdoc`, `ypy-*` and script URLs.
- An expression is at most 1000 characters and 32 levels deep.
- An attacker who can write markup can bind and display signals and call registered functions, but cannot run code or reach
  the network from an expression. Register only functions that are safe to call from any element.

## JavaScript API and events

| Member | What it does |
|---|---|
| `ypy.signals.register(name, fn)` | Makes `fn` callable from expressions as `name(...)`. |
| `ypy.signals.get(path, element)` | The value at `"user.name"`, seen from `element`. |
| `ypy.signals.patch(object, element, onlyIfMissing)` | Applies a merge patch. |
| `ypy.signals.stats()` | `{nodes, pending}`: live nodes and uncommitted bindings. |

| Event | When |
|---|---|
| `ypy:signals:patch` | before a patch is applied; cancelable |
| `ypy:signals:readonly` | a write to a read-only signal was refused |
| `ypy:signals:error` | an expression did not parse, or a handler, effect or patch failed |

## Limits

The extension is 6,009 bytes min+gzip, which leaves 135 of its 6,144-byte budget. Dependencies are tracked per top-level
signal, so an effect that reads `$user.name` runs when anything in `user` changes. `ypy-on` inside a scope cannot run
JavaScript, so put such an element outside the scope or register a function. `$event.target` is not available by design.
Checkbox groups into one array are not provided, so use one boolean per checkbox or `select multiple`.
