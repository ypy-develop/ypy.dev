---
order: 50
title: "Extensions"
description: "The 48 official YPY extensions: what each one adds to the core, how to load them, how to restrict which ones may run and how to write your own."
---

# Extensions

The core handles requests, swaps, morph, history, errors and accessibility. Everything else is an extension: a small
script you load after the core. The core has a budget of 18,432 bytes min+gzip and each extension 6,144 bytes;
`go run ./cmd/ypy build --sizes` fails the build when one goes over.

A page that does not load an extension pays nothing for it, and its attributes do nothing. Each page below lists the
attributes, events, headers and settings of its extension. The same names, for all extensions at once, are in the
[reference](../reference.md#extensions).

## Loading extensions

Load the extension file with a `<script>` tag after the core, one tag per extension:

```html
<script src="/ypy.min.js"></script>
<script src="/ext/ypy-sse.min.js"></script>
<script src="/ext/ypy-retry.min.js"></script>
```

`ypy.max.min.js` is one file with the core and 36 of the 48 extensions, in a single request. Twelve are not in it, because
they are for development or for narrow cases: `ypy-boot`, `ypy-csp`, `ypy-dedupe`, `ypy-dev`, `ypy-expr`,
`ypy-head`, `ypy-inspector`, `ypy-multipart`, `ypy-native`, `ypy-prompt`, `ypy-ptag` and `ypy-test`. Load those with
their own tag. Each extension also ships as a readable `.js` file, which is easier to step through in a debugger.

### Allowing only some extensions

By default any script that calls `ypy.registerExtension` is accepted. On a page where that is too open (user content
near the `<head>`, third-party scripts), list the names you ship in `config.extensions`:

```html
<meta name="ypy-config" content='extensions:"ypy-lock,ypy-csp,ypy-sse,ypy-retry",locked:true'>
```

The list matches the full name an extension registers with, `ypy-` prefix included: `ypy-sse`, not `sse`. An extension
that is not on the list logs a warning and its hooks never run. When you set `locked:true`, list `ypy-lock` too, or the
lock is refused and the configuration stays writable. See [Security](../index.md#security).

## All extensions

| Extension | What it does |
|---|---|
| [`ypy-alpine-compat`](alpine-compat.md) | Lets Alpine.js and YPY share a page: Alpine starts on new fragments and keeps its state through a morph. |
| [`ypy-boot`](boot.md) | Start-up order: an extension or script waits for what it `requires`, and `ypy:ready` fires when everything is in. |
| [`ypy-browser-indicator`](browser-indicator.md) | Shows the browser's own loading indicator (tab spinner, stop button) while a request runs. |
| [`ypy-cache`](cache.md) | Keeps GET answers and revalidates them, with tags and expiry from the server. |
| [`ypy-commands`](commands.md) | The `YPY-Commands` response header: a closed list of small page commands. |
| [`ypy-csp`](csp.md) | Content Security Policy: nonce gate, Trusted Types and safe evaluation. |
| [`ypy-csrf`](csrf.md) | Puts a CSRF token on every same-origin unsafe request and takes the renewed one from the server. |
| [`ypy-dedupe`](dedupe.md) | Joins identical requests that are already in flight. |
| [`ypy-dev`](dev.md) | Development mode: errors with context, overlay, latency and offline simulation, lint warnings. |
| [`ypy-disable`](disable.md) | Disables chosen elements while a request runs, and makes the target inert. |
| [`ypy-download`](download.md) | The `download` swap style: saves the response as a file, with progress events. |
| [`ypy-expr`](expr.md) | The expression language of trigger filters, without `eval`. |
| [`ypy-head`](head.md) | Merges the `<head>` of a full-page response into the page. |
| [`ypy-history-cache`](history-cache.md) | Keeps page snapshots for the back and forward buttons. |
| [`ypy-inspector`](inspector.md) | A floating panel that records every request and event. |
| [`ypy-layers`](layers.md) | Modals, drawers and popups on `<dialog>`, in a stack. |
| [`ypy-lazy`](lazy.md) | Loads a region when it is about to be seen, alone or in a group. |
| [`ypy-list`](list.md) | `key:` and `limit:` for the `append` and `prepend` swap styles: lists with no duplicates and a size cap. |
| [`ypy-live`](live.md) | Reactive expressions that run again when the DOM, an input or a swap changes what they read. |
| [`ypy-lock`](lock.md) | Freezes the security settings and `ypy.config` after start-up. |
| [`ypy-model`](model.md) | The client side of a language-model answer: an untrusted-content wall, announcements, follow and stop. |
| [`ypy-multipart`](multipart.md) | Handles long-lived multipart responses, part by part. |
| [`ypy-native`](native.md) | The web side of a native app shell: visits, native components, `YPY-Native`. |
| [`ypy-nav`](nav.md) | Global boost, prefetch, the current link and `ypy-once`. |
| [`ypy-optimistic`](optimistic.md) | Shows the result at once and undoes it if the request fails. |
| [`ypy-pending`](pending.md) | Shows placeholder content in the target while a request is pending. |
| [`ypy-persist`](persist.md) | Keeps elements alive across navigations and reloads the page when tracked assets change. |
| [`ypy-poll`](poll.md) | Smarter `every` polling: slower in a hidden tab, interval set by the server, status 286 stops it. |
| [`ypy-preload`](preload.md) | Starts a GET before the click and reuses the answer. |
| [`ypy-preserve`](preserve.md) | Keeps the value, caret, focus and scroll of fields and containers through a swap; waits for an input method. |
| [`ypy-progress`](progress.md) | A thin progress bar at the top of the page for navigations and chosen requests. |
| [`ypy-prompt`](prompt.md) | Asks a question first and sends the answer in the `YPY-Prompt` header. |
| [`ypy-ptag`](ptag.md) | A per-element tag so a poll can be answered with `304` when nothing changed. |
| [`ypy-retry`](retry.md) | Repeats a failed request, with backoff, jitter and `Retry-After`. |
| [`ypy-scope`](scope.md) | The `scope`, `layer` and `shadow:` selectors. |
| [`ypy-signals`](signals.md) | Client state without a build step: signals, bindings, computed values. |
| [`ypy-sort`](sort.md) | Drag and keyboard reordering, sent to the server. |
| [`ypy-sse`](sse.md) | Server-Sent Events: swaps messages into the page and fires named events. |
| [`ypy-stream`](stream.md) | The client of the YPY stream protocol: HTML patches, signals, redirects and events over one response. |
| [`ypy-switch`](switch.md) | Shows or hides parts of a form by the value of a field. |
| [`ypy-targets`](targets.md) | Swaps one response into every element that matches a selector. |
| [`ypy-telemetry`](telemetry.md) | Sends every YPY failure to the page's own error handler. |
| [`ypy-test`](test.md) | In-page hooks for browser tests: `ypy.idle()` and a busy marker. |
| [`ypy-transitions`](transitions.md) | Runs swaps inside the View Transitions API. |
| [`ypy-upload`](upload.md) | Progress of requests that send a file. |
| [`ypy-upsert`](upsert.md) | The `upsert` swap style: replace by id, insert the rest, optional sort. |
| [`ypy-validate`](validate.md) | Checks fields on the server while the visitor types. |
| [`ypy-ws`](ws.md) | WebSocket: swaps incoming messages and sends form values back. |

## Writing an extension

An extension is a script that calls `ypy.registerExtension(name, hooks)`. It uses the public `ypy.*` API and the object
that `init` receives. Members written with a double underscore are private in the built files (`dist/`), so an
extension that reads `ypy.__x` works from the source and breaks in the build. If you need something the API below does
not offer, ask for it under a public name instead of reaching for a private one.

A hook is named after an event: the event name with every colon replaced by an underscore, so `ypy:before:request` is
`ypy_before_request` and `ypy:after:swap` is `ypy_after_swap`. A hook receives `(elt, detail)`, and returning `false`
cancels a cancelable event. This extension adds an attribute, `ypy-flash`, that makes the target of a swap flash:

```js
// ypy-flash.js
(() => {
  let api

  ypy.registerExtension('ypy-flash', {
    init(internalApi) {
      api = internalApi
    },

    // after every swap, look at the element that sent the request
    ypy_after_swap(elt, detail) {
      const color = api.attributeValue(detail.ctx.sourceElement, 'ypy-flash')
      if (!color) return
      const target = detail.ctx.target
      if (!(target instanceof Element)) return

      target.setAttribute('data-flash', color)
      setTimeout(() => target.removeAttribute('data-flash'), 600)
    },

    // forget what you set up when an element leaves
    ypy_before_cleanup(elt) {
      elt.removeAttribute('data-flash')
    },
  })
})()
```

```html
<script src="/ypy.min.js"></script>
<script src="/ypy-flash.js"></script>

<button ypy-post="/save" ypy-target="#panel" ypy-flash="amber">Save</button>
<div id="panel" class="panel"></div>
```

Rules for a good extension:

1. Read attributes with `api.attributeValue(elt, 'ypy-name')`, never `getAttribute('ypy-...')`. The resolver knows the
   `data-ypy-` spelling and `:inherited`.
2. Set up in `ypy_after_init`, work in the request or swap hooks, tear down in `ypy_before_cleanup`. Nothing you start
   should outlive the element.
3. Never compile text into code. Author code goes through `api.executeJavaScript`, which obeys `config.callbacks`.
4. Keep it under 6 KB min+gzip, the budget of an official extension.
5. Fire events named `ypy:<noun>:<verb>` with `api.triggerEvent`, so pages and other extensions can hook them.

### The extension API

`init(api)` receives the object once, when the extension registers. Keep it in a variable.

| Member | What it does |
|---|---|
| `attr(elt, name)` | The element's own attribute `name` (written as `ypy-x`), in whichever prefix it carries. Never inherited. `null` when absent. |
| `prefixSelector(selector)` | The selector with `ypy-` replaced by every prefix, so `[ypy-ignore]` becomes `[ypy-ignore],[data-ypy-ignore]`. |
| `attributeValue(elt, name, default, collector, allowInheritance)` | The value of an attribute, from the element or inherited from an ancestor by the rules of [attribute inheritance](../index.md#attribute-inheritance). |
| `HCON` | The parser of the attribute value syntax: `HCON.parse(text)`, `HCON.split(text)`, `HCON.merge(text, target)`. |
| `triggerEvent(elt, name, detail)` | Dispatches an event on `elt` (the name carries the `ypy:` prefix) and runs the `ypy_*` hooks. Returns `false` when a listener cancelled it. |
| `parseTriggerSpecs(spec)`, `determineMethodAndAction(elt, evt)` | The parsing of `ypy-trigger` and the choice of verb and URL for an element. |
| `createRequestContext(elt, evt)`, `collectFormData(elt, form, submitter, validate, isGet)`, `getAttributeObject(elt, name, callback, scope)` | The request context, the form values of a request and the evaluated `ypy-vals`-style attributes. |
| `insertContent(task)`, `morph(oldNode, fragment, innerHTML)`, `isSoftMatch(a, b)` | The swap and the morph the core uses, for extensions that add a swap style. |
| `onTrigger(elt, spec, handler)`, `ypyProp(elt)` | Listens to a trigger specification, modifiers included; the core's state record of the element (`elt._ypy`), created when missing. |
| `executeJavaScript(elt, vars, code, expression, isAsync, compile)` | Runs author code under `config.callbacks`. The only place that compiles code. |
| `ttPolicy()` | The `ypy` Trusted Types policy, or `null` when the CSP forbids it. |
| `initSecurity(ttPolicy, syncFn, asyncFn, ignoreFn)` | For the nonce extension only: replaces the policy and the constructors that compile author code, and tests whether an element sits under an `ypy-ignore` marker. Once an extension calls it, `config.nonce` lets the page process. |

### Testing an extension

The repository's browser suite loads an extension from `src/ext/` in a fresh page. Model yours on
`test/tests/ext/ypy-download.js`, put it in `test/tests/ext/` and run only it, then run the whole suite against the
built files:

```sh
go run ./cmd/ypy test --files 'test/tests/ext/ypy-flash.js'
go run ./cmd/ypy test --dist
```

`--dist` fails an extension that uses a private member.
