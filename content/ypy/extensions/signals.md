---
title: "Signals: the ypy-signals extension"
description: "Client state without a build step or eval: named signals bound to inputs and read by attributes, which the server reads and patches as JSON."
order: 370
---
# Signals: the `ypy-signals` extension

Some page state has no business on the server: whether a box is ticked, what is typed in a field, whether a panel is
open. `ypy-signals` keeps that state in the page as named values called signals. Inputs bind to them and attributes
read them, and the server can read them from every request and change them with a JSON patch. It needs no build step
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

The extension never compiles text into code. Its expressions are parsed by a small parser into closures, so it
works with `callbacks:"none"` (the default) and with a strict CSP (no `unsafe-eval`, no inline
script, no inline style) and Trusted Types.

Contents: [Scopes](#scopes-and-names) · [Attributes](#attributes) · [Expressions](#expressions) ·
[Binding and cadence](#ypy-bind-and-its-cadence) · [Handlers](#ypy-on) · [Requests](#signals-in-requests-the-wire-format) ·
[Server patches](#updating-from-the-server) · [Read-only signals](#server-owned-read-only-signals) ·
[Unsaved state](#unsaved-state-ypy-dirty) · [Persistence](#persistence) · [Swaps and morphs](#swaps-and-morphs) ·
[Security](#security) · [JavaScript API](#javascript-api) · [Events](#events) · [Limits](#limits)

## Scopes and names

An element with `ypy-signals` is a scope: its signals are visible to the element and its descendants. Scopes nest.
A name is looked up in the nearest scope that has it. Put signals shared by the whole page on `<body>`.

- `$name` reads a signal; `$user.address.city` and `$list[0]` read inside it. A path that does not exist is
  `undefined`, never an error. Writes replace the value along the path (copy on write), so a change is always a new
  object.
- A name that starts with `_` is local: it is never sent to the server, at any depth (`_open`, `{k:1,_tmp:2}`).
- A signal nobody declared is created, when something writes it, in the outermost scope that reaches the
  writer. The same rule places a new key that a server patch brings.
- An element outside every scope speaks for the top-level scopes of the page (for example a button in the page
  header that posts the global signals).
- The names `constructor`, `prototype` and `__proto__` are refused everywhere.
- Every scope has a read-only local signal `$_dirty` (see [unsaved state](#unsaved-state-ypy-dirty)).
- Attribute names are lower case in HTML. When a signal name sits in an attribute *name*
  (`ypy-bind:user-name`, `ypy-computed:total-price`), a hyphen stands for a capital: `userName`, `totalPrice`. In an
  attribute *value* write the name as it is.

## Attributes

All of them also exist with the `data-` prefix (`data-ypy-text`). They are read when the element is processed, after
a swap included.

| Attribute | Value | What it does |
|---|---|---|
| `ypy-signals` | `count:0, name:'x'` (the braces are optional; JSON works too) | Declares signals in a scope. Values are literals. An existing signal keeps its value when the markup is processed again (a morph does not reset the state). |
| `ypy-signals.server` | same | Server-owned signals: read-only for the client ([below](#server-owned-read-only-signals)). |
| `ypy-signals.local`, `ypy-signals.session` | same | Persisted in `localStorage` or `sessionStorage`; the element needs an `id` ([below](#persistence)). |
| `ypy-computed:name` | expression | A derived, read-only signal. It follows its dependencies and is not sent to the server. |
| `ypy-bind` or `ypy-bind:name` | path (`user.name`) | Two-way binding with a form control ([below](#ypy-bind-and-its-cadence)). |
| `ypy-text` | expression | Sets `textContent` (text, never markup). Objects show as JSON. |
| `ypy-show` | expression | Hides the element (`display:none`) while the value is false. |
| `ypy-class:name` or `ypy-class` | expression | `ypy-class:active="$on"` toggles one class; `ypy-class="{red: $bad, wide: $w > 3}"` toggles several. A class that has a dot, such as `w-1.5`, goes in the object form. |
| `ypy-attr:name` or `ypy-attr` | expression | Sets an attribute; `false`, `null` and `undefined` remove it, `true` sets it empty. `ypy-attr="{title: $t}"` sets several. Refused: `on*` attributes, `srcdoc`, `ypy-*`/`data-ypy-*`, and `javascript:`/`vbscript:` URLs in `href`, `src`, `action`, `formaction`, `xlink:href`, `poster`. |
| `ypy-on:event` | statements | Event handler ([below](#ypy-on)). |
| `ypy-signals-send="false"` | | On the element or an ancestor: its requests do not carry the signals. |

Effects (`ypy-text`, `ypy-show`, `ypy-class`, `ypy-attr`, the field side of `ypy-bind`) run once when wired and again,
in a batch, when a signal they read changes. They track exactly what they read: an effect that read `$a` does not
run when `$b` changes. Writes made in one handler, patch or request run each effect once, in a microtask. Effects
that keep waking each other stop after 100 rounds with an error in the console.

## Expressions

One language for every attribute. Handlers (`ypy-on`) also take statements.

```
statement  = assignment | expression                          (handlers only: ";" between statements)
assignment = signal ( "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "??=" ) expression
           | ( "++" | "--" ) signal | signal ( "++" | "--" )
expression = ternary
ternary    = binary [ "?" expression ":" expression ]
binary     = unary { op unary }      lowest to highest:
             "||" "??"  /  "&&"  /  "==" "!=" "===" "!=="  /  "<" ">" "<=" ">="  /  "+" "-"  /  "*" "/" "%"
unary      = ( "!" | "-" | "+" ) unary | primary
primary    = number | string | "true" | "false" | "null" | "undefined"
           | "$" name { "." key | "?." key | "[" expression "]" }       a signal, read by path
           | name "(" [ expression { "," expression } ] ")"            a registered function
           | "(" expression ")" | "[" list "]" | "{" key ":" expression { "," ... } "}"
```

- Numbers: `12`, `1.5`, `2e3`. Strings: `"..."` or `'...'` with JSON escapes (`\n`, `\"`, `\'`, `é`).
- `&&`, `||` and `??` short-circuit. `==` and `!=` are the loose comparisons, as in JavaScript.
- A bare name is an error. The only names are `true`, `false`, `null`, `undefined`, signals (`$name`) and functions given
  to `ypy.signals.register`. There is no `window`, `document`, `this`, `eval`, no method call on a value, no
  arrow function, template literal, regular expression, `new`, `typeof`, `in`, comma operator or comment.
- Reading is lenient: only own properties are read, so `$text.length` and `$list[1]` work, and `$o.toString`,
  `$list.map` and `$o.constructor` are `undefined`.
- Assignment and `++`/`--` are statements of handlers. In `ypy-text` and the other read-only places they are an error.
- Registered functions are the only way out of the language. `ypy.signals.register('upper', s => String(s).toUpperCase())`
  makes `upper($name)` available. The function runs as normal page code; it receives plain values.
- In handlers `$event` is a plain object: `type`, `key`, `code`, `detail`, `value` and `checked` of the target, and
  `shiftKey`, `ctrlKey`, `altKey`, `metaKey`. It never holds the element or the window. The name `event` is
  reserved inside handlers.

An expression that fails to parse is reported (`ypy:signals:error`, and the console) and its attribute does nothing;
the rest of the page keeps working. A runtime error (an unknown function, a refused write) is reported the same way.

## `ypy-bind` and its cadence

`ypy-bind="name"` (or `ypy-bind:name`) keeps a form control and a signal equal. The signal fills the field;
the field writes the signal.

| Control | Signal |
|---|---|
| text, search, email, textarea, any other `input` | text; a number if the signal was a number |
| `type="checkbox"` | boolean |
| `type="radio"` | the `value` of the checked radio (give each radio of the group the same `ypy-bind`) |
| `<select>` | the selected `value` |
| `<select multiple>` | an array of the selected values |
| `contenteditable` (not a form control) | the text content |

The cadence depends on the modifier. Without one, the field writes the signal on `change` (when the user commits: leaving a text
field, picking in a select or checkbox), so typing does not run effects on every key.

| Modifier | The field writes the signal |
|---|---|
| (none) | on `change` |
| `.live` | on every `input` |
| `.blur` | on `blur` only |
| `.debounce.300ms` | after the user pauses for 300 ms (a `change` writes at once). `.debounce` alone means 300 ms |
| `.throttle.300ms` | at most once per 300 ms, with the latest value (trailing). `.throttle` alone means 300 ms |

Times are `Nms` or `Ns`. Whatever the cadence, a value typed and not yet written is written before a request
starts, so the server never receives a stale field. A write to a read-only signal is refused and the field shows the
signal again.

A number signal stays a number, and clearing a number field does not write `0` back into it.

## `ypy-on`

`ypy-on:click="$count++; $open = !$open"` runs statements when the event fires. The event name is the part after
the colon (`ypy-on:keydown`, `ypy-on:input`). `ypy-on::before:request` is the form of a `ypy:` event, as in the core.

The core also has a `ypy-on`, and the two share a rule. The core's `ypy-on` runs JavaScript (and only when `config.callbacks`
allows it). Inside a scope, `ypy-on` belongs to this extension and takes statements of this language: the
extension tells the core to skip those elements, so no JavaScript is compiled for them. Outside a scope the core's
`ypy-on` is left alone. The rule is about the markup: an element is "inside a scope" when it or an ancestor has
`ypy-signals`, `ypy-signals.server`, `ypy-signals.local` or `ypy-signals.session`.

| Modifier | |
|---|---|
| `.prevent`, `.stop` | `preventDefault()`, `stopPropagation()` |
| `.once` | run one time |
| `.window` | listen on `window` (`ypy-on:resize.window`) |
| `.debounce.300ms` | run after a pause in the events |
| `.throttle.300ms` | run at most once per 300 ms (leading edge) |

The listener is registered in the capture phase on the element itself, so on an element that also starts a request
(`ypy-post`, `ypy-get`) the handler runs first and the request sees the new values.

## Signals in requests: the wire format

This section is the contract for servers and SDKs; the Go SDK follows it (`ypy.SignalsParam`, protocol §3.2).

A request whose element has signals in reach carries them, as one JSON object, in the form field or parameter
named `ypy-signals`:

| Request | Where |
|---|---|
| `GET`, `DELETE` | query parameter `ypy-signals=<url-encoded JSON>` (the core puts the fields of these verbs in the URL) |
| other verbs (`POST`, `PUT`, `PATCH`, `QUERY`) | form field `ypy-signals` in the body, `application/x-www-form-urlencoded` or `multipart/form-data` (the core never sends a JSON body) |

A server that reads signals tries the query, a JSON body and then the form field, and uses the first present.

What is in the object: the signals of every scope around the element (the inner scope wins on a name clash),
including the server-owned ones. The computed ones and any key starting with `_` are left out, at any depth. If there
is nothing to send, there is no parameter. Server-owned values come back so the server can verify its signature
([below](#server-owned-read-only-signals)).

Rules:

- Signals are sent only to the same origin as the page. A request to another origin carries none.
- `ypy-signals-send="false"` (on the element or an ancestor) turns it off for those requests.
- A GET puts the signals in the URL: they reach server logs, proxies and, with `ypy-push-url`, the history entry.
  Keep secrets out of signals, and use a POST or `ypy-signals-send="false"` for requests that must not expose them.
  A server should also set `Vary` accordingly when it caches such a GET, and limit what it reads (the Go SDK reads at most 1 MiB).
- The name `ypy-signals` is a single constant in the extension (`FIELD` at the top of `ext/ypy-signals.js`) and in the
  SDK.

## Updating from the server

A response whose `Content-Type` is `application/json` is a JSON Merge Patch ([RFC 7396](https://www.rfc-editor.org/rfc/rfc7396))
applied to the signals around the request's element. It is not swapped into the page, whatever the status code
(a `422` with `{"emailError":"taken"}` works). A response of any other type is handled by the core as usual.

```json
{"count": 5, "user": {"name": "Bo", "age": null}, "tags": ["a"], "_local": 1}
```

- An object member is merged into the signal's object, recursively. `null` removes the member. An array or a scalar
  replaces the value. At the top level `null` removes the signal.
- A key goes to the nearest scope that already has it; a new key to the outermost scope that reaches the element.
- Computed signals ignore a patch. Server-owned signals accept it (the server is their authority). The names
  `__proto__`, `constructor` and `prototype` are skipped at every depth.
- The patch fires `ypy:signals:patch` on the element first; a listener can cancel it. JSON that is not an object, or
  not JSON at all, fires `ypy:signals:error` and changes nothing.
- Values that arrive in a patch count as saved for [unsaved state](#unsaved-state-ypy-dirty).
- `ypy.signals.patch(object, element, onlyIfMissing)` applies a patch from code (for example the `ypy-patch-signals`
  event of the streaming protocol). With `onlyIfMissing` true, signals that already exist are left as they are. A JSON
  *response* has no carrier for that option; it belongs to the streaming event.

## Server-owned (read-only) signals

`ypy-signals.server="plan:'free', planSig:'3fa9c1'"` declares signals that the client cannot write: an
expression (`$plan = 'pro'`), a binding or `ypy-on` is refused, a `ypy:signals:readonly` event fires with a warning,
and a bound field shows the signal again. They change only when the server patches them or renders the scope's markup
again (the markup of a `.server` declaration always overwrites, the markup of an ordinary one never does).

Read-only is a convenience of the UI and gives no security, because whoever controls the browser can send any value.
The real protection is on the server: it signs the values it owns and verifies them when they come back. The client
needs no extra feature for that. Put the signature in a server-owned signal beside the values
(`planSig` above, for example an HMAC of the canonical JSON of `plan` with a server secret, ideally bound to the
session). Server-owned signals are sent with every request like the others, so the server recomputes the HMAC from
`plan` and compares it with `planSig`, and rejects the request when they differ. A patch updates value and
signature together.

## Unsaved state (`ypy-dirty`)

A scope is dirty while a client-owned signal (not computed, not server-owned, not `_local`) differs from what was
last saved. "Saved" is the declared value; it moves when

- a response with a 2xx status to anything but a GET arrives for a request from inside the scope (the server
  stored the form); an error status or a GET (a search) does not count;
- the server patches a key (that key becomes the saved value).

While dirty, the scope element gets the class `ypy-dirty` (rename it with `ypy.config.classes.dirty`), the token
`dirty` in `data-ypy-state`, and its `$_dirty` signal is `true`, so `ypy-show="$_dirty"` and
`ypy-class:warn="$_dirty"` work anywhere inside. A value restored from storage that differs from the declared one is
unsaved.

## Persistence

`ypy-signals.local="draft:''"` stores the signals of that declaration in `localStorage` under the key
`ypy-signals:<id>` (`ypy-signals.session` uses `sessionStorage`). The element must have an `id`, otherwise a warning is
logged and the values stay in memory. The stored value replaces the declared default when the scope is created, but only for
keys the markup declares, so stored data cannot add signals. Broken or unavailable storage is ignored. Server-owned
signals are never persisted. Local signals may be persisted: `ypy-signals.local="_theme:'dark'"`.

## Swaps and morphs

The state of a scope lives with its element. A morph that keeps the element (the default swap)
keeps the signals; a swap into a descendant keeps the scope and wires the new content. Because a morph
copies the server's text, classes and styles onto bound elements, the extension runs the effects of the swapped tree
again once it is processed, so the signals win over what the markup said. A swap that replaces the scope element
(`outerHTML`, or a morph that cannot match it) creates a new scope from the declared values: give it an `id` and
persist it, or keep the scope above the swap target.

When an element leaves the page its effects, listeners, pending debounced writes and computed signals are dropped
(on the core's cleanup event). An element removed by other code without a cleanup is dropped the next time one
of its effects runs or its `window` handler fires. `ypy.signals.stats()` reports what is alive; the tests swap
content 500 times and expect the same number afterwards.

`ypy-ignore` keeps the signal attributes of a subtree inert.

## Security

| Concern | What the extension does |
|---|---|
| Code from attributes | None is compiled. No `eval`, no `new Function`, no `setTimeout(string)`. Works with `callbacks:"none"`, with a strict CSP and with Trusted Types. |
| Reaching globals | The language has no identifiers other than signals, literals and registered functions. `window`, `globalThis`, `document`, `this`, `constructor(...)` are errors. |
| Prototype access | Reads use own properties only; `constructor`, `prototype` and `__proto__` are refused as names, path keys, object-literal keys, patch members and stored keys. A write copies along the path and never mutates a foreign object. Arrays grow only one item at a time and `length` cannot be written. |
| Markup into the page | `ypy-text` writes text. There is no HTML-writing attribute. `ypy-attr` refuses event-handler attributes, `srcdoc`, `ypy-*` attributes and script URLs. |
| Cost | An expression is at most 1000 characters (a declaration 20000), nested at most 32 levels. Compiled expressions are cached (at most 256). Effects that loop stop after 100 rounds. |
| Leaking signals | Sent only to the same origin; `_` signals and computed ones are never sent; `ypy-signals-send="false"`. |
| Injected markup | An attacker who can write markup can bind and display signals and call registered functions, but cannot run code or reach the network from an expression. A nonce policy of the page (`ypy-csp` with `callbacks:"nonce"`) is not needed by this extension and does not change it; register only functions that are safe to call from any element. |

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.signals.register(name, fn)` | Makes `fn` callable from expressions as `name(...)`. The name is a plain identifier. A function registered later is found when the expression runs. |
| `ypy.signals.get(path, element = document.body)` | The value at `"user.name"` (a leading `$` is allowed), seen from `element`. |
| `ypy.signals.patch(object, element = document.body, onlyIfMissing)` | Applies a merge patch like a server response does. |
| `ypy.signals.stats()` | `{nodes, pending}`: live computed and effect nodes, and bindings with an uncommitted value. |

## Events

All bubble from the element and run the extension hooks of the core, like every `ypy:` event.

| Event | Detail | When |
|---|---|---|
| `ypy:signals:patch` | `{patch}` | before a patch is applied; cancelable |
| `ypy:signals:readonly` | `{name, warn}` | a write to a read-only signal was refused |
| `ypy:signals:error` | `{error}` | an expression did not parse, a handler or effect threw, a patch was not usable |

## Limits

- Size: the extension is 6,009 bytes min+gzip, which leaves 135 bytes of the 6,144-byte budget.
- Dependencies are tracked per top-level signal: an effect that reads `$user.name` runs again when anything in
  `user` changes.
- A signal added to a nearer scope later does not take over names that effects already resolved in an outer scope.
- `ypy-on` inside a scope cannot run JavaScript; put such an element outside the scope or register a function.
- `$event.target` is not available by design; read `$event.value` or `$event.checked`.
- Checkbox groups into one array are not provided; use one boolean per checkbox, or `select multiple`.
