---
order: 50
title: "Extensions"
description: "The 48 official YPY extensions: what each one does, which are in the ypy.max bundle, how to load them and how to write your own."
---

# Official extensions

Extensions live in `ext/` next to the core. Load the `.min.js` file with a `<script>` after the core, or load `ypy.max`,
which bundles the ones marked below. Each extension also ships as a readable `.js` file with the same name, which is
useful when you want to step through it in a debugger. An extension registers itself with `ypy.registerExtension(name, hooks)`. When
`ypy.config.extensions` is set, only the names in that list can register; any other is refused.

A feature that not every page needs is an extension, so the core only carries what every page pays for. The size
budgets are 18,432 bytes min+gzip for the core and 6,144 bytes for each extension, and `ypy build --sizes` fails the
build when one goes over.

## All extensions

There are 48 extensions in `ext/`. Each page lists the attributes, events, headers and settings of its extension.

| Extension | What it does | Loads |
|---|---|---|
| [`ypy-alpine-compat`](alpine-compat.md) | Lets Alpine.js and YPY share a page: Alpine starts on new fragments and keeps its state through a morph. | `ypy.max` or `ext/ypy-alpine-compat.min.js` |
| [`ypy-boot`](boot.md) | Start-up order: an extension or script waits for what it `requires`, and `ypy:ready` fires when everything is in. | `ext/ypy-boot.min.js` only |
| [`ypy-browser-indicator`](browser-indicator.md) | Shows the browser's own loading indicator (tab spinner, stop button) while a request runs. | `ypy.max` or `ext/ypy-browser-indicator.min.js` |
| [`ypy-cache`](cache.md) | Keeps GET answers and revalidates them, with tags and expiry from the server. | `ypy.max` or `ext/ypy-cache.min.js` |
| [`ypy-commands`](commands.md) | The `YPY-Commands` response header: a closed list of small page commands. | `ypy.max` or `ext/ypy-commands.min.js` |
| [`ypy-csp`](csp.md) | Content Security Policy: nonce gate, Trusted Types and safe evaluation. | `ext/ypy-csp.min.js` only |
| [`ypy-csrf`](csrf.md) | Puts a CSRF token on every same-origin unsafe request and takes the renewed one from the server. | `ypy.max` or `ext/ypy-csrf.min.js` |
| [`ypy-dedupe`](dedupe.md) | Joins identical requests that are already in flight. | `ext/ypy-dedupe.min.js` only |
| [`ypy-dev`](dev.md) | Development mode: errors with context, overlay, latency and offline simulation, lint warnings. | `ext/ypy-dev.min.js` only |
| [`ypy-disable`](disable.md) | Disables chosen elements while a request runs, and makes the target inert. | `ypy.max` or `ext/ypy-disable.min.js` |
| [`ypy-download`](download.md) | The `download` swap style: saves the response as a file, with progress events. | `ypy.max` or `ext/ypy-download.min.js` |
| [`ypy-expr`](expr.md) | The expression language of trigger filters, without `eval`. | `ext/ypy-expr.min.js` only |
| [`ypy-head`](head.md) | Merges the `<head>` of a full-page response into the page. | `ext/ypy-head.min.js` only |
| [`ypy-history-cache`](history-cache.md) | Keeps page snapshots for the back and forward buttons. | `ypy.max` or `ext/ypy-history-cache.min.js` |
| [`ypy-inspector`](inspector.md) | A floating panel that records every request and event. | `ext/ypy-inspector.min.js` only |
| [`ypy-layers`](layers.md) | Modals, drawers and popups on `<dialog>`, in a stack. | `ypy.max` or `ext/ypy-layers.min.js` |
| [`ypy-lazy`](lazy.md) | Loads a region when it is about to be seen, alone or in a group. | `ypy.max` or `ext/ypy-lazy.min.js` |
| [`ypy-list`](list.md) | `key:` and `limit:` for the `append` and `prepend` swap styles: lists with no duplicates and a size cap. | `ypy.max` or `ext/ypy-list.min.js` |
| [`ypy-live`](live.md) | Reactive expressions that run again when the DOM, an input or a swap changes what they read. | `ypy.max` or `ext/ypy-live.min.js` |
| [`ypy-lock`](lock.md) | Freezes the security settings and `ypy.config` after start-up. | `ypy.max` or `ext/ypy-lock.min.js` |
| [`ypy-model`](model.md) | The client side of a language-model answer: an untrusted-content wall, announcements, follow and stop. | `ypy.max` or `ext/ypy-model.min.js` |
| [`ypy-multipart`](multipart.md) | Handles long-lived multipart responses, part by part. | `ext/ypy-multipart.min.js` only |
| [`ypy-native`](native.md) | The web side of a native app shell: visits, native components, `YPY-Native`. | `ext/ypy-native.min.js` only |
| [`ypy-nav`](nav.md) | Global boost, prefetch, the current link and `ypy-once`. | `ypy.max` or `ext/ypy-nav.min.js` |
| [`ypy-optimistic`](optimistic.md) | Shows the result at once and undoes it if the request fails. | `ypy.max` or `ext/ypy-optimistic.min.js` |
| [`ypy-pending`](pending.md) | Shows placeholder content in the target while a request is pending. | `ypy.max` or `ext/ypy-pending.min.js` |
| [`ypy-persist`](persist.md) | Keeps elements alive across navigations and reloads the page when tracked assets change. | `ypy.max` or `ext/ypy-persist.min.js` |
| [`ypy-poll`](poll.md) | Smarter `every` polling: slower in a hidden tab, interval set by the server, status 286 stops it. | `ypy.max` or `ext/ypy-poll.min.js` |
| [`ypy-preload`](preload.md) | Starts a GET before the click and reuses the answer. | `ypy.max` or `ext/ypy-preload.min.js` |
| [`ypy-preserve`](preserve.md) | Keeps the value, caret, focus and scroll of fields and containers through a swap; waits for an input method. | `ypy.max` or `ext/ypy-preserve.min.js` |
| [`ypy-progress`](progress.md) | A thin progress bar at the top of the page for navigations and chosen requests. | `ypy.max` or `ext/ypy-progress.min.js` |
| [`ypy-prompt`](prompt.md) | Asks a question first and sends the answer in the `YPY-Prompt` header. | `ext/ypy-prompt.min.js` only |
| [`ypy-ptag`](ptag.md) | A per-element tag so a poll can be answered with `304` when nothing changed. | `ext/ypy-ptag.min.js` only |
| [`ypy-retry`](retry.md) | Repeats a failed request, with backoff, jitter and `Retry-After`. | `ypy.max` or `ext/ypy-retry.min.js` |
| [`ypy-scope`](scope.md) | The `scope`, `layer` and `shadow:` selectors. | `ypy.max` or `ext/ypy-scope.min.js` |
| [`ypy-signals`](signals.md) | Client state without a build step: signals, bindings, computed values. | `ypy.max` or `ext/ypy-signals.min.js` |
| [`ypy-sort`](sort.md) | Drag and keyboard reordering, sent to the server. | `ypy.max` or `ext/ypy-sort.min.js` |
| [`ypy-sse`](sse.md) | Server-Sent Events: swaps messages into the page and fires named events. | `ypy.max` or `ext/ypy-sse.min.js` |
| [`ypy-stream`](stream.md) | The client of the YPY stream protocol: HTML patches, signals, redirects and events over one response. | `ypy.max` or `ext/ypy-stream.min.js` |
| [`ypy-switch`](switch.md) | Shows or hides parts of a form by the value of a field. | `ypy.max` or `ext/ypy-switch.min.js` |
| [`ypy-targets`](targets.md) | Swaps one response into every element that matches a selector. | `ypy.max` or `ext/ypy-targets.min.js` |
| [`ypy-telemetry`](telemetry.md) | Sends every YPY failure to the page's own error handler. | `ypy.max` or `ext/ypy-telemetry.min.js` |
| [`ypy-test`](test.md) | In-page hooks for browser tests: `ypy.idle()` and a busy marker. | `ext/ypy-test.min.js` only |
| [`ypy-transitions`](transitions.md) | Runs swaps inside the View Transitions API. | `ypy.max` or `ext/ypy-transitions.min.js` |
| [`ypy-upload`](upload.md) | Progress of requests that send a file. | `ypy.max` or `ext/ypy-upload.min.js` |
| [`ypy-upsert`](upsert.md) | The `upsert` swap style: replace by id, insert the rest, optional sort. | `ypy.max` or `ext/ypy-upsert.min.js` |
| [`ypy-validate`](validate.md) | Checks fields on the server while the visitor types. | `ypy.max` or `ext/ypy-validate.min.js` |
| [`ypy-ws`](ws.md) | WebSocket: swaps incoming messages and sends form values back. | `ypy.max` or `ext/ypy-ws.min.js` |

## Files and bundles

The same extensions by file, with whether `ypy.max` carries them:

| Extension | File | In `ypy.max` | What it adds |
|---|---|---|---|
| `ypy-retry` | `ext/ypy-retry.min.js` | yes | automatic retry of failed requests: no connection, optionally 5xx; jitter and `Retry-After` |
| `ypy-lock` | `ext/ypy-lock.min.js` | yes | the configuration lock (`locked:true`, `ypy.config.lock()`) |
| `ypy-csp` | `ext/ypy-csp.min.js` | no | nonce gate, Trusted Types and safe eval |
| `ypy-stream` | `ext/ypy-stream.min.js` | yes | client of the streaming protocol (SSE): ypy-patch-elements, ypy-patch-signals, ypy-redirect, ypy-event |
| `ypy-model` | `ext/ypy-model.min.js` | yes | a language-model answer in a page: `ypy-untrusted` content wall, start and end announcements, `ypy-follow`, `ypy-stop` |
| `ypy-sse` | `ext/ypy-sse.min.js` | yes | Server-Sent Events |
| `ypy-ws` | `ext/ypy-ws.min.js` | yes | WebSocket |
| `ypy-preload` | `ext/ypy-preload.min.js` | yes | preloading of links and requests |
| `ypy-browser-indicator` | `ext/ypy-browser-indicator.min.js` | yes | the browser's own loading indicator |
| `ypy-download` | `ext/ypy-download.min.js` | yes | the `download` swap style |
| `ypy-pending` | `ext/ypy-pending.min.js` | yes | placeholder content while a request is pending |
| `ypy-targets` | `ext/ypy-targets.min.js` | yes | the `ypy-targets` attribute: one response swapped into several elements |
| `ypy-live` | `ext/ypy-live.min.js` | yes | reactive expressions |
| `ypy-history-cache` | `ext/ypy-history-cache.min.js` | yes | history snapshot cache, off in `ypy.max` until turned on ([history-cache](history-cache.md)) |
| `ypy-upsert` | `ext/ypy-upsert.min.js` | yes | the `upsert` swap style |
| `ypy-list` | `ext/ypy-list.min.js` | yes | `key:<attr>` and `limit:<n>` for the `append` and `prepend` swap styles ([list](list.md)) |
| `ypy-alpine-compat` | `ext/ypy-alpine-compat.min.js` | yes | Alpine.js interoperability |
| `ypy-head` | `ext/ypy-head.min.js` | no | merging of `<head>` content |
| `ypy-multipart` | `ext/ypy-multipart.min.js` | no | multipart responses, handled in the extension |
| `ypy-prompt` | `ext/ypy-prompt.min.js` | no | the `ypy-prompt` attribute |
| `ypy-ptag` | `ext/ypy-ptag.min.js` | no | per-element polling tags: the server can skip an unchanged swap |
| `ypy-poll` | `ext/ypy-poll.min.js` | yes | smart `every` polling: slower in a hidden tab, interval set by the server, status 286 stops it ([poll](poll.md)) |
| `ypy-signals` | `ext/ypy-signals.min.js` | yes | client state |
| `ypy-dev` | `ext/ypy-dev.min.js` | no | development mode: errors with context, error overlay, latency and offline simulator, warnings, reload of templates |
| `ypy-inspector` | `ext/ypy-inspector.min.js` | no | the floating inspector panel |
| `ypy-layers` | `ext/ypy-layers.min.js` | yes | modals, drawers and popups on `<dialog>`, in a stack ([layers](layers.md)) |
| `ypy-native` | `ext/ypy-native.min.js` | no | the web side of a native app shell: visits, native components, `YPY-Native` ([native](native.md)) |
| `ypy-nav` | `ext/ypy-nav.min.js` | yes | global boost and prefetch ([navigation](nav.md)) |
| `ypy-progress` | `ext/ypy-progress.min.js` | yes | progress bar for navigations ([navigation](nav.md)) |
| `ypy-persist` | `ext/ypy-persist.min.js` | yes | elements kept across navigations, reload on tracked asset change ([navigation](nav.md)) |
| `ypy-lazy` | `ext/ypy-lazy.min.js` | yes | regions loaded when they become visible, alone or in a group ([lazy](lazy.md)) |
| `ypy-cache` | `ext/ypy-cache.min.js` | yes | response cache with revalidation ([cache](cache.md)) |
| `ypy-optimistic` | `ext/ypy-optimistic.min.js` | yes | optimistic UI, undone when the request fails ([optimistic](optimistic.md)) |
| `ypy-validate` | `ext/ypy-validate.min.js` | yes | field validation on the server ([validate](validate.md)) |
| `ypy-upload` | `ext/ypy-upload.min.js` | yes | upload progress for requests that send a file ([upload](upload.md)) |
| `ypy-csrf` | `ext/ypy-csrf.min.js` | yes | CSRF token on every same-origin unsafe request, renewed by the server ([csrf](csrf.md)) |
| `ypy-telemetry` | `ext/ypy-telemetry.min.js` | yes | error reports for the page's own error handler ([telemetry](telemetry.md)) |
| `ypy-switch` | `ext/ypy-switch.min.js` | yes | show and hide parts of a form by the value of a field ([switch](switch.md)) |
| `ypy-commands` | `ext/ypy-commands.min.js` | yes | the `YPY-Commands` response header: a closed list of small page commands ([commands](commands.md)) |
| `ypy-sort` | `ext/ypy-sort.min.js` | yes | drag and keyboard reordering sent to the server ([sort](sort.md)) |
| `ypy-dedupe` | `ext/ypy-dedupe.min.js` | no | joins identical requests that are already in flight ([dedupe](dedupe.md)) |
| `ypy-expr` | `ext/ypy-expr.min.js` | no | the expression language for trigger filters, without `eval` ([expr](expr.md)) |
| `ypy-boot` | `ext/ypy-boot.min.js` | no | start-up: `requires`, `provides`, `ypy:ready` ([boot](boot.md)) |
| `ypy-test` | `ext/ypy-test.min.js` | no | in-page testing hooks: `ypy.idle()`, `data-ypy-busy` ([testing](test.md)) |
| `ypy-scope` | `ext/ypy-scope.min.js` | yes | the `scope`, `layer` and `shadow:` selectors ([scope](scope.md)) |
| `ypy-transitions` | `ext/ypy-transitions.min.js` | yes | View Transitions for swaps with `transition:true` ([transitions](transitions.md)) |
| `ypy-preserve` | `ext/ypy-preserve.min.js` | yes | `value`, `selection`, `focus` and `scroll` modes, and the wait for an input method ([preserve](preserve.md)) |
| `ypy-disable` | `ext/ypy-disable.min.js` | yes | `ypy-disable` and `ypy-inert`: hold elements back while a request runs ([disable](disable.md)) |

## `ypy-retry`

Repeats a failed request. Attribute `ypy-retry`, setting `ypy.config.retry`, event `ypy:request:retry`. See
[`ypy-retry`](retry.md) and [offline state](../reference/events/ypy-offline.md).

The core classifies failures (`ctx.errorKind` is `offline`, `network` or `timeout`) and tracks whether the page is
offline (the `ypy-offline` class, `data-ypy-state="offline"`, `ypy:offline` and `ypy:online`). It calls `ctx.fetch` once
per request. `ypy-retry` wraps `ctx.fetch` from its `ypy_before_request` hook, so the loop runs inside the request: the
request keeps its queue slot and its indicators, and the request events fire once. Without the extension,
`ypy-retry` and `config.retry` are ignored.

## `ypy-lock`

Freezes `mode`, `callbacks`, `extensions`, `nonce`, `inlineScriptNonce`, the approved extension list and the
`ypy.config` object. Turn it on with `locked:true` in the `ypy-config` meta tag or with `ypy.config.lock()`. See
[Security](../reference/security.md#the-ypy-config-meta-tag-and-its-lock).

The lock is its own file instead of part of `ypy-csp` because it is useful without nonces, and `ypy-csp` blocks every
element on a page that has no nonce. The lock closes when the extension starts, so load it right after the core. On a
page that sets `config.extensions`, list `ypy-lock` in it: if it is refused, nothing is locked.

## `ypy-dev` and `ypy-inspector`

These two are for development and are not in `ypy.max`, since `ypy.max` is a build you may ship. `ypy-dev` logs,
warns and can slow requests down on purpose. The inspector keeps every request and event in memory. Load each with its
own tag after the core. See [`ypy-dev`](dev.md) and [the inspector](inspector.md).

## Writing an extension

An extension is a script that runs after the core and calls `ypy.registerExtension(name, hooks)`. It may use two
things: the public API (`ypy.*`) and the object the core hands to `init`, described below. Members written with a
double underscore are private fields in the built core (`dist/`), so an extension cannot reach them there; they exist
only in the source build the tests use. An extension that reads `ypy.__x` works from `src/` and breaks in `dist/`.
`ypy test --dist` runs the extension and security suites against the built files, and a unit test of the build
rejects an extension that names a private member.

If an extension needs something from the core that the API below does not offer, add it to the API under a public
name instead of reaching for a private one.

### Extension API

`init(api)` receives the object once, when the extension registers. Keep it in a variable; the same object is
the one every extension gets.

| Member | What it does |
|---|---|
| `attr(elt, name)` | The element's own attribute `name` (written as `ypy-x`), in whichever prefix it carries (`ypy-`, `data-ypy-`, `config.prefix`). Never inherited. `null` when absent. |
| `prefixSelector(selector)` | The selector with `ypy-` replaced by every prefix, for example `[ypy-ignore]` becomes `[ypy-ignore],[data-ypy-ignore]`. |
| `attributeValue(elt, name, default, collector, allowInheritance)` | The value of an attribute, read from the element or inherited from an ancestor by the rules of [attribute inheritance](../reference/inheritance.md). |
| `HCON` | The parser of the attribute value syntax: `HCON.parse(text)`, `HCON.split(text)`, `HCON.merge(text, target)`. |
| `triggerEvent(elt, name, detail)` | Dispatches an event on `elt` (the name carries the `ypy:` prefix) and runs the extensions' `ypy_*` hooks. Returns `false` when a listener cancelled it. |
| `parseTriggerSpecs(spec)`, `determineMethodAndAction(elt, evt)` | The parsing of `ypy-trigger` and the choice of verb and URL for an element. |
| `createRequestContext(elt, evt)`, `collectFormData(elt, form, submitter, validate, isGet)`, `getAttributeObject(elt, name, callback, scope)` | The request context, the form values of a request and the evaluated `ypy-vals`-style attributes. |
| `insertContent(task)`, `morph(oldNode, fragment, innerHTML)`, `isSoftMatch(a, b)` | The swap and the morph the core uses, for extensions that add a swap style. |
| `onTrigger(elt, spec, handler)`, `ypyProp(elt)` | Listens to a trigger specification (with its modifiers) on an element; the core's state record of the element (`elt._ypy`), created when missing. |
| `executeJavaScript(elt, vars, code, expression, isAsync, compile)` | Runs author code under `config.callbacks`; the only place that compiles code. |
| `ttPolicy()` | The `ypy` Trusted Types policy, or `null` when the CSP forbids it. |
| `initSecurity(ttPolicy, syncFn, asyncFn, ignoreFn)` | For the nonce extension only: replaces the policy, the constructors that compile author code and, with `ignoreFn(elt)`, the test of whether an element sits under an `ypy-ignore` marker. Once an extension calls it, `config.nonce` lets the page process. |

