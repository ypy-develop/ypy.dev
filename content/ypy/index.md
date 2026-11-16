---
order: 10
id: "DOC-index"
type: "index"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/index", "lang/en"]
title: "Docs"
description: "One long page on YPY: install it, make requests, swap HTML, handle errors and history, then security, testing and configuration, with links to the details."
---

# YPY documentation

This page covers YPY in the order you are likely to need it, from the first attribute to security and testing. Each
section is short and points to the page with the details: the [reference](reference/attributes/index.md), the
[examples](examples/index.md), the [extensions](extensions/index.md) and the [server pages](server/index.md). There is
also a Portuguese version, [Documentação do YPY](pt-BR/index.md).

## YPY in a nutshell

With YPY, an HTML element says in its attributes which request to make, when to make it, and where to put the HTML
that comes back. The server answers with HTML fragments instead of JSON. You add one `<script>` tag; there is nothing
to build and nothing to install.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<button ypy-get="/hello" ypy-target="#out" ypy-swap="innerHTML">Say hello</button>
<div id="out"></div>
```

Read it as a sentence: on click, `GET /hello`, and put the answer inside `#out`. If the server answers
`<p>Hello</p>`, that paragraph appears in the page and the rest stays as it was, including the address bar.

Four questions cover most of what a page needs:

| Question | Attribute | Default |
|---|---|---|
| Which request? | `ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` | none |
| When? | `ypy-trigger` | `click`; `change` for inputs and selects; `submit` for forms |
| Where does the answer go? | `ypy-target` | the element itself |
| How is it put in? | `ypy-swap` | `innerMorph` |

Attributes are `ypy-<name>` or `data-ypy-<name>`, headers are `YPY-<Name>`, events are `ypy:<phase>:<verb>`, and the
global object is `ypy`.

If you want to build a first page and a first server, start with [Getting started](guides/getting-started.md). If you
want the model in one page, read [Core concepts](guides/core-concepts.md).

## Installing

### From a CDN

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<!-- or -->
<script src="https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js"></script>
```

In production, serve the file from your own origin. A strict Content Security Policy is easier to write that way, and
your page no longer depends on a third party being up.

### From npm

```sh
npm install @ypy-dev/ypy@0.1.1
```

The package has the same `dist/` files as the repository build, and `https://unpkg.com/@ypy-dev/ypy@0.1.1/dist/ypy.min.js` serves the same minified core.

### Download a copy

To build the files yourself, you need Go 1.26 or later. The build writes them to `dist/`:

```sh
go run ./cmd/ypy build
```

| File | Use it when |
|---|---|
| `dist/ypy.min.js` | the core alone; you add the extensions you need |
| `dist/ypy.max.min.js` | the core and the official extensions that are safe to ship, in one file |
| `dist/ypy.esm.min.js` | you load it as an ES module (`import ypy from "./ypy.esm.min.js"`); it also sets `window.ypy` |
| `dist/ext/<name>.min.js` | one extension, loaded after the core |
| `dist/ypy.css`, `dist/ypy.min.css` | the small stylesheet for the indicator, the progress bar and the opt-in transitions |
| `dist/ypy.d.ts` | the TypeScript types of `window.ypy` |

Each `.min.js` has a source map next to it. If you copy only the `.js` file into your own `static/` directory, copy
from `dist/embed/` instead. Those files are the same code without the `sourceMappingURL` comment, so the browser does
not request a map that is not there.

### Extensions

The core covers what nearly every page needs. Anything else is an [extension](extensions/index.md), a script you load
after the core.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-poll.min.js"></script>
<script src="ext/ypy-validate.min.js"></script>
```

Or load `ypy.max.min.js` to get most of the official extensions in one request. The ones meant for development or for
narrow cases are left out (`ypy-dev`, `ypy-inspector`, `ypy-test`, `ypy-csp` and a few more); the
[table of extensions](extensions/index.md) lists them. The build enforces a size budget: 18,432 bytes min+gzip for the
core and 6,144 bytes for each extension (`go run ./cmd/ypy build --sizes`).

## Requests

### Verbs and URLs

`ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` and `ypy-query` each issue the request of their name to the
URL in the value. When the method and the URL come from different places, use `ypy-method` with `ypy-action`.

```html
<form ypy-post="/items" ypy-target="#list" ypy-swap="beforeend">
  <input name="title" required>
  <button>Add</button>
</form>
<ul id="list"></ul>
```

A form sends its fields. Any other element sends its own `name` and `value` if it has them, plus the values of the form
around it. [`ypy-vals`](reference/attributes/index.md) adds values, [`ypy-include`](reference/attributes/index.md) adds the values of other
elements, `ypy-headers` adds headers and `ypy-encoding="multipart/form-data"` sends files.

```html
<button ypy-post="/votes" ypy-vals='{"item": 42}' ypy-include="#comment">Vote</button>
```

### Triggers

`ypy-trigger` takes a comma-separated list of events, each with optional modifiers.

```html
<!-- on load, then every 30 seconds -->
<div ypy-get="/stats" ypy-trigger="load, every 30s"></div>

<!-- active search: wait 300 ms after the last key, only if the value changed -->
<input name="q" ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms, search">

<!-- load the next page when this row comes into view -->
<tr ypy-get="/rows?page=2" ypy-trigger="revealed" ypy-swap="afterend"></tr>
```

| Special trigger | Fires |
|---|---|
| `load` | when the element is processed |
| `revealed` | the first time the element scrolls into view |
| `intersect` | each time the element meets the viewport (`root:<selector>`, `threshold:<number>`) |
| `every <time>` | on a timer, for example `every 2s` |

| Modifier | Effect |
|---|---|
| `once` | only the first time |
| `changed` | only when the value of the element changed |
| `delay:<time>` | waits; a new event restarts the wait |
| `throttle:<time>` | the first event goes at once, the following ones wait for the interval |
| `from:<selector>` | listens on another element (`document`, `window`, `closest <selector>`, ...) |
| `target:<selector>` | only when `event.target` matches |
| `prevent`, `stop`, `halt` | `preventDefault()`, `stopPropagation()`, or both |
| `capture`, `passive` | the listener options of the same name |

An event with a filter in brackets (`click[ctrlKey]`) runs code, so it needs `config.callbacks` turned on; see
[Security](#security). A plain `every 30s` keeps polling in a background tab. The [`ypy-poll`](extensions/poll.md)
extension slows it down there, and the [polling example](examples/polling.md) shows how.

### Indicators

While a request runs, the element that sent it gets the class `ypy-request` (and `data-ypy-state="request"`).
`ypy-indicator` names other elements to mark instead. An element with the class `ypy-indicator` is hidden until it, or
an ancestor, has `ypy-request`.

```html
<button ypy-post="/report" ypy-indicator="#spinner">Build report</button>
<span id="spinner" class="ypy-indicator">Working...</span>
```

To lock controls while the request runs, load [`ypy-disable`](extensions/disable.md) and name them with
`ypy-disable`. The classes are listed in [CSS classes](reference/css-classes.md).

### Targets

`ypy-target` is a CSS selector, `this`, or a relative form: `closest <selector>`, `find <selector>`,
`next [selector]`, `previous [selector]`. When the selector matches nothing, YPY logs it and fires
[`ypy:target:missing`](reference/events/target.md). A response can retarget itself with the `YPY-Retarget` header.

`ypy-select` takes one part of the response by selector and drops the rest. `ypy-main` marks the main target, the
element a swap falls back to when its own target is missing ([`ypy-main`](reference/attributes/ypy-main.md)).

### Swapping

`ypy-swap` says how the answer enters the target. The default is a morph: the new HTML is merged into what is there, so
elements that exist on both sides stay the same DOM nodes. Focus, typed text, scroll position and open `<details>`
survive.

| Style | What it does |
|---|---|
| `innerMorph` (default) | merges the response into the children of the target |
| `outerMorph` | merges the response into the target itself |
| `innerHTML`, `outerHTML` | replaces the children, or the target, with nothing kept |
| `outerSync` | morphs the attributes of the target and replaces its children |
| `beforebegin`, `afterbegin`, `beforeend`, `afterend` | inserts next to or inside the target |
| `textContent` | sets the text; the response is not parsed as HTML |
| `delete`, `none` | removes the target; does nothing |

Modifiers follow the style: `ypy-swap="innerHTML swap:200ms settle:100ms show:top"`. They set delays (`swap:`,
`settle:`), scrolling (`scroll:`, `show:`), view transitions (`transition:true`) and a few more. Extensions add styles:
`upsert`, `download`, and `key:` and `limit:` for lists ([`ypy-list`](extensions/list.md)). The full rules are in
[Swap styles and morph](reference/swap-and-morph.md).

The [click to edit](examples/click-to-edit.md) example swaps a whole row for a form and back.

### Several places from one response

A response can update more than the target. Put each extra piece in a `<ypy-partial>`, or mark an element with
`ypy-swap-oob`. YPY swaps the main content first, then puts each extra piece where its `id` points.

```html
<li>New item</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">3 items</ypy-partial>
<ypy-partial id="toast"><p>Saved</p></ypy-partial>
```

`ypy-select-oob` on the request picks elements of the response by selector and swaps each one out of band. To make
the whole response all or nothing, the server sends `YPY-Group: atomic` ([Atomic swaps](reference/atomic.md)). See
[`<ypy-partial>`](reference/tags/ypy-partial.md), [Multiple targets](guides/multiple-targets.md) and the
[updating other content](examples/updating-other-content.md) example.

### Requests that overlap

When an element asks for a request while another one from the same place is still running, `ypy-sync` decides what
happens. The default is `queue first`: the new request waits, unless one is already waiting. The other strategies are
`drop`, `abort`, `replace`, `queue last`, `queue all`, `latest` and `parallel`.

```html
<input name="q" ypy-get="/search" ypy-trigger="input changed delay:200ms" ypy-sync="this:replace">
```

An old response never overwrites a newer one. See [`ypy-sync`](reference/attributes/ypy-sync.md) and the
[request queue events](reference/events/request-queue.md). [`ypy-dedupe`](extensions/dedupe.md) joins identical GETs
into one request.

### Confirming

`ypy-confirm` asks before the request. The value is the message.

```html
<button ypy-delete="/items/7" ypy-confirm="Delete this item?" ypy-target="closest li" ypy-swap="delete">Delete</button>
```

The `ypy:confirm` event lets you replace the browser dialog with your own. The [`ypy-prompt`](extensions/prompt.md)
extension asks for a text and sends it in the `YPY-Prompt` header. See [confirm and prompt](examples/confirm-and-prompt.md).

### Handlers in markup

`ypy-on:<event>="..."` runs JavaScript when a DOM event fires, and `ypy-on::<event>` is the shorthand for the
`ypy:` events. Author code is off by default (`callbacks` is `none`). An event listener in a script is the safer
choice, or the [signals](#scripting) extension, which never compiles text into code.

## Attribute inheritance

An attribute reaches the elements below it only when you say so, with the `:inherited` modifier. `:append` adds the
value of a descendant to the inherited ones.

```html
<section ypy-target:inherited="#results" ypy-swap:inherited="innerHTML">
  <button ypy-get="/people">People</button>
  <button ypy-get="/places">Places</button>
</section>
<div id="results"></div>
```

A bare attribute on a container stays on the container. See [Attribute inheritance](reference/inheritance.md).

## Navigation and boost

`ypy-boost` turns the links and forms inside an element into requests that swap the body and push the address. The
page does not reload, so scripts and styles stay loaded. If JavaScript fails to load, the links and forms still work as
usual.

```html
<body ypy-boost="true">
  <nav><a href="/orders">Orders</a> <a href="/people">People</a></nav>
  <main id="main">...</main>
</body>
```

Set `ypy-boost="false"` on a subtree to leave it alone. The [`ypy-nav`](extensions/nav.md) extension boosts the whole
page with one setting (`config.nav.boost`), prefetches links on hover, marks the current link, and the
[`ypy-progress`](extensions/progress.md) extension draws a progress bar for the navigation. Keep an element across
navigations with [`ypy-persist`](extensions/persist.md), and merge the `<head>` of the response with
[`ypy-head`](extensions/head.md).

After a boosted navigation, the focus moves to the main heading and a live region announces the new title; see
[Accessibility](#accessibility).

## Real time

Three extensions let the server update the page without a click.

```html
<!-- Server-Sent Events: each message is swapped into the target -->
<div ypy-sse:connect="/feed" ypy-target="#feed"></div>

<!-- WebSocket: messages in, a form that sends -->
<div ypy-ws:connect="/chat" ypy-target="#messages">
  <ul id="messages"></ul>
  <form ypy-ws:send><input name="text"><button>Send</button></form>
</div>

<!-- The YPY stream: HTML patches, signal patches, redirects and events over one response -->
<div id="room" ypy-stream="/room/events"></div>
```

[`ypy-sse`](extensions/sse.md) and [`ypy-ws`](extensions/ws.md) are plain transports. [`ypy-stream`](extensions/stream.md)
speaks the YPY streaming protocol: a Go server writes it with `ypy.NewStream`, or fans it out to many pages with
`ypy.Broadcaster` ([server streaming](server/streaming.md)). For a language model answer that arrives token by token,
use [`ypy-model`](extensions/model.md) on the page and [model streaming](server/model-streaming.md) on the server. The
[live updates](examples/live-updates.md) example puts them together.

## History

A swap that pushes or replaces the address writes a history entry that YPY can restore. `ypy-push-url` pushes a URL
after the swap, `ypy-replace-url` replaces the current entry, and `true` uses the URL of the request. A boosted
GET pushes by default. The server can do the same with the `YPY-Push-Url` and `YPY-Replace-Url` headers.

```html
<a ypy-get="/orders?page=2" ypy-target="#orders" ypy-push-url="true">Next page</a>
```

On back and forward, YPY asks the server for the page again with `YPY-History-Restore-Request: true`, and the server
has to answer with a whole page. Mark the part to restore with `ypy-history-elt`, and keep snapshots for
instant back navigation with [`ypy-history-cache`](extensions/history-cache.md). `config.history` is `true`, `false`
or `"reload"`. See [History](reference/history.md) and [History and caching](guides/history-and-caching.md).

## Requests and responses

### What the server receives

Every YPY request carries `YPY-Request: true`. The other request headers say who asked and what is being replaced.

| Header | Meaning |
|---|---|
| `YPY-Request` | always `true` on a YPY request |
| `YPY-Source` | the element that triggered the request |
| `YPY-Target` | the element the response is swapped into |
| `YPY-Current-URL` | the current URL of the browser |
| `YPY-Request-Type` | `partial` for a targeted swap, `full` when the target is `body` or `ypy-select` is used |
| `YPY-Boosted` | present on a boosted request |
| `YPY-History-Restore-Request` | present when restoring a page after back or forward |

A server that answers a fragment to a YPY request and a page to a normal one, at the same URL, sends
`Vary: YPY-Request`. Without it a cache can hand the fragment to a page load.

### What the server can say

| Header | Effect |
|---|---|
| `YPY-Trigger` | fires events on the page after the response |
| `YPY-Push-Url`, `YPY-Replace-Url` | pushes or replaces the address |
| `YPY-Redirect` | redirects with a full page load |
| `YPY-Location` | navigates with a request, without a reload |
| `YPY-Refresh` | reloads the page when the value is `true` |
| `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect` | overrides the target, the swap style or the selected part |
| `YPY-Group` | `atomic` makes the swap all or nothing |

The complete tables, with the extension headers, are in [Headers](reference/headers.md). The Go helpers for each are
in [Requests and responses](server/requests-and-responses.md).

### Status codes and errors

A response with status 204 or 304 is not swapped (`config.noSwap`). A status of 400 or more fires
`ypy:response:error`. Where the response lands depends on these rules, weakest first: the form convention,
`ypy-error-target`, `ypy-error-swap` and `ypy-error-focus`, the response headers, and `ypy-status:<code>`.

```html
<form ypy-post="/signup" ypy-target="#page"
      ypy-error-target="#errors" ypy-error-focus="first-invalid"
      ypy-status:5xx="target:#fatal swap:innerHTML">
  ...
</form>
<div id="errors"></div>
<div id="fatal"></div>
```

The form convention: a `422` means the form is not valid and is rendered back into the form, with the focus on the
first invalid field; a `303` after a POST means done, and YPY swaps the page it redirects to. A network failure, a
timeout or an exception fires `ypy:error`. A 404 or a 500 is an answer, so it goes through the rules above. See
[`ypy-error-*`](reference/attributes/ypy-error.md), [Forms](guides/forms.md) and, for repeating requests that failed,
[`ypy-retry`](extensions/retry.md).

### Order of operations

For a request that reaches the server and succeeds, the events fire in this order: `ypy:confirm`,
`ypy:config:request`, `ypy:before:request`, `ypy:before:response`, `ypy:after:request`, `ypy:before:swap`,
`ypy:before:settle` and `ypy:after:settle` for each target, `ypy:after:swap` once, then the history events and
`ypy:finally:swap` and `ypy:finally:request`. The address is written after the swap, never before. The full table,
with what each event is for, is in [Events](reference/events/index.md); the numbers behind each phase are in
[Timing](reference/timing.md).

## Validation

Validation runs in three layers that you can combine.

1. The browser's own constraint validation. A form without `novalidate` is checked before the request, and the
   browser shows its messages. `ypy-validate="true"` turns it on for other elements, and `ypy-validate="false"` turns
   it off.
2. The server's answer to the whole form: a `422` with the form and its errors.
3. The [`ypy-validate`](extensions/validate.md) extension, which sends one field (or a batch) to the server when the
   user leaves it and puts the answer next to the field.

```html
<form ypy-post="/signup" ypy-validate-fields="/signup/validate">
  <input id="email" name="email" type="email">
  <small id="email-error"></small>
  <button>Create account</button>
</form>
```

See the [inline validation](examples/inline-validation.md) example and the [Go recipe](server/recipes/go.md).

## Animations

YPY adds a class at each step of a swap, so CSS can animate it: `ypy-request` while the request runs, `ypy-swapping`
on the target before the swap, `ypy-settling` after it, and `ypy-added` on the new content until it settles. Add a swap delay
and the old content has time to leave before the new one arrives.

```html
<div id="panel" class="ypy-fade" ypy-get="/panel" ypy-trigger="click from:#refresh"
     ypy-swap="innerHTML swap:200ms">...</div>
```

`ypy.css` has two opt-in transitions: `ypy-fade` fades the target out while it swaps, `ypy-fade-in` fades the new
content in. For a view transition, use `transition:true` in the swap or `config.transitions`. See
[Styling](reference/styling.md) and the [animations example](examples/animations.md).

## Extensions

An extension registers itself with `ypy.registerExtension`. `config.extensions` is an allow list: once it is set,
only the names in it may register. The [official extensions](extensions/index.md) include polling, caching, lazy
loading, validation, uploads, layers (modal, drawer, popup), sorting, signals and streaming. To choose, load, restrict or
write one, see [Using and writing extensions](extensions/using-and-writing.md).

## Events and logging

Everything YPY does is announced with a DOM event named `ypy:<phase>:<verb>`. Events that carry a request put its
context in `event.detail.ctx`. Some can be cancelled with `event.preventDefault()`.

```js
document.addEventListener("ypy:config:request", (e) => {
  e.detail.ctx.request.headers["X-App-Version"] = "42";
});

document.addEventListener("ypy:after:swap", (e) => {
  // the new content is in the DOM and processed
});
```

Initialize a third-party widget for new content with `ypy.onLoad`, and add a header for every request in
`ypy:config:request`. The [Events guide](reference/events/index.md) has the life of a request and recipes; the
[event list](reference/events/list.md) has every event with its `detail`.

To see what happens, set `ypy.config.logAll = true`: every YPY event goes to the console.

## Debugging

Three tools, lightest first:

- `config.logAll` logs every event.
- [`ypy-dev`](extensions/dev.md), loaded only in development: errors with the element and the request, an overlay,
  a latency and offline simulator, and warnings for markup that does not do what it looks like.
- [`ypy-inspector`](extensions/inspector.md): a floating panel with requests, swaps, events and signals as they
  happen (Alt+Shift+Y).

Keep both extensions out of production.

## Scripting

### Signals

[`ypy-signals`](extensions/signals.md) keeps state in the browser without a build step and without `eval`.

```html
<form ypy-signals="name:'', agree:false" ypy-post="/signup">
  <input ypy-bind="name">
  <label><input type="checkbox" ypy-bind="agree"> I agree</label>
  <p ypy-show="$name.length > 0">Hello, <b ypy-text="$name"></b></p>
  <button ypy-attr:disabled="!$agree || $name == ''">Send</button>
</form>
```

The server reads the signals from each request (`ypy.ReadSignals`) and changes them with a patch
(`ypy.PatchSignals`); see [server signals](server/signals.md). [`ypy-live`](extensions/live.md) adds reactive
expressions for markup that is not a form, and [`ypy-expr`](extensions/expr.md) is the expression language of trigger
filters, with no `eval`.

### The JavaScript API

`ypy.ajax(verb, path, options)` issues a request from a script, `ypy.trigger(element, event, detail)` fires an event,
`ypy.on` adds a listener, `ypy.find` and `ypy.findAll` query the document, and `ypy.process(root)` makes YPY read new
markup that a script inserted. The full list, with `ypy.config`, is in [JavaScript API](reference/javascript-api.md).

### Third-party components

A widget that owns its DOM (a chart, an editor, a Web Component) must survive a swap. `ypy-preserve` keeps a node by
`id`; `ypy-morph-skip` and `ypy-morph-skip-children` freeze an element or its children during a morph;
`config.morphThirdParty` lists the elements another script owns. See
[Third-party components](guides/third-party-components.md), [`ypy-preserve`](reference/attributes/ypy-preserve.md)
and the [`ypy-preserve` extension](extensions/preserve.md).

## Caching

Three kinds of caching, each separate.

- The HTTP cache of the browser. Send `Vary: YPY-Request` on every response that changes with that header, so the
  fragment and the page do not mix. See [History and caching](guides/history-and-caching.md).
- Responses kept by the page, with [`ypy-cache`](extensions/cache.md): show the copy at once, ask the server whether
  it is still current, and swap again only if it is not. Nothing is cached unless the element says `ypy-cache`.
- Snapshots for the back button, with [`ypy-history-cache`](extensions/history-cache.md).

[`ypy-ptag`](extensions/ptag.md) lets the server skip a swap when a polled fragment did not change.

## Security

YPY turns attributes into behavior, so an attacker who can put markup in your page does not need a script. A
`ypy-post` on a button is enough. Three layers defend against that, and none of them replaces escaping on the server.

1. Escape everything on the server. Fragments are HTML like any other. Never let user text become part of a tag,
   an attribute name or a `ypy-*` value, and strip `ypy-*` and `data-ypy-*` from user content.
2. Keep author code off. `config.callbacks` is `none` by default, so `ypy-on`, `js:` values and bracket filters do
   nothing. With [`ypy-csp`](extensions/csp.md) code runs only with the nonce of the page.
3. Keep requests on your origin. `config.mode` is `same-origin`, and the [`ypy-csrf`](extensions/csrf.md)
   extension sends a token with every unsafe request ([server CSRF](server/csrf.md)).

Scripts in responses follow `config.scripts` (`auto`, `nonce` or `block`). A strict CSP needs no inline script or
style: the Go SDK can move the CSS and scripts of your templates into hashed files ([assets](server/assets.md)). Read
the [Security checklist](guides/security-checklist.md) before you ship, the [Security reference](reference/security.md)
for the settings, and the [essay on security](essays/security.md) for the reasoning.

## Accessibility

YPY handles the repetitive part of accessibility for a swap. The part that depends on your content is yours.

- `aria-busy="true"` is set on the target while a request runs (`config.ariaBusy`, `ypy-busy="false"` to opt out).
- A managed live region announces what changed. The core ships no text, so you set the messages
  (`config.announce.messages`) or write them per element with `ypy-announce`.
- After a boosted navigation, the focus goes to the main heading, `ypy-focus` chooses it elsewhere, and
  `ypy-error-focus` places it after an error.
- `ypy-scroll` keeps, resets or restores the scroll position, and `ypy-inert` makes the target `inert` while a request
  runs (with [`ypy-disable`](extensions/disable.md)).

See [Accessibility](guides/accessibility.md), [`ypy-announce`](reference/attributes/ypy-announce.md),
[`ypy-busy`](reference/attributes/ypy-busy.md), [`ypy-focus`](reference/attributes/ypy-focus.md) and
[`ypy-scroll`](reference/attributes/ypy-scroll.md).

## Configuring

Settings live in `ypy.config`. Set them from a script, or before the core loads, with a meta tag:

```html
<meta name="ypy-config" content='{"history": false, "defaultTimeout": 10000, "callbacks": "none"}'>
```

```js
ypy.config.defaultSwap = "innerHTML";
ypy.config.classes.request = "is-loading";
```

One request can change its fetch options with `ypy-config` on the element (`timeout`, `credentials`, `cache`). The
`mode` cannot be changed per element. [`ypy-lock`](extensions/lock.md) freezes the configuration after the page sets
it. Every key, with its type and default, is in [Configuration](reference/config/index.md).

## Testing

The server side of a YPY page is plain HTTP, so you can test it without a browser. In Go, `ypytest` builds the request
a browser would send, runs your handler, and lets you assert on the fragment, the status and the headers. It can also
read a stream.

```go
req := ypytest.NewRequest(http.MethodPost, "/items",
	ypytest.Target("#items"), ypytest.Form(url.Values{"title": {"Milk"}}))
res := ypytest.Do(handler, req)
res.AssertStatus(t, http.StatusOK)
res.AssertText(t, "li.item", "Milk")
```

For the page itself, the [`ypy-test`](extensions/test.md) extension adds `ypy.idle()` to browser tests: a promise that
resolves when no request, swap or timer is pending. See [Testing your app](server/testing.md).

## The `ypy` command

One Go binary, no Node. In the repository, run `go run ./cmd/ypy <command>`.

| Command | What it does | Read |
|---|---|---|
| `ypy build` | writes `dist/`: bundles, minified files, extensions; `--sizes` checks the size budgets | [Installing](#installing) |
| `ypy test` | runs the browser suites of the repository | [Testing your app](server/testing.md) |
| `ypy ci` | runs every gate | [Contributing: testing](contributing/testing.md) |
| `ypy bench` | measures performance in a real browser | [Performance](essays/performance.md) |
| `ypy check` | finds old uses that break or change | [`ypy check`](migration/upgrade-check.md) |
| `ypy migrate` | converts a project written for the old library | [`ypy migrate`](migration/migrate.md) |
| `ypy gen` | generates the editor data from the table of names | [Editor support](reference/editors.md) |

## Where to go next

| I want to... | Read |
|---|---|
| try YPY on a small page | [Getting started](guides/getting-started.md) (plain HTML, or Go and Arandu) |
| understand the model | [Core concepts](guides/core-concepts.md), [Events](reference/events/index.md) |
| look up an attribute, header, event or setting | [Attributes](reference/attributes/index.md), [Headers](reference/headers.md), [Event list](reference/events/list.md), [JavaScript API](reference/javascript-api.md), [Configuration](reference/config/index.md) |
| copy a pattern | [Examples](examples/index.md): [click to edit](examples/click-to-edit.md), [active search](examples/active-search.md), [infinite scroll](examples/infinite-scroll.md), [inline validation](examples/inline-validation.md), [file upload](examples/file-upload.md), [dialogs](examples/dialogs-layers.md), [tabs](examples/tabs.md), [sorting](examples/drag-and-drop-sort.md) |
| build a form that validates | [Forms](guides/forms.md), [`ypy-error-*`](reference/attributes/ypy-error.md), [`ypy-validate`](extensions/validate.md), the [recipe](server/recipes/go.md) of your stack |
| update several places from one response | [Multiple targets](guides/multiple-targets.md), [`<ypy-partial>`](reference/tags/ypy-partial.md) |
| make the back button and caching work | [History and caching](guides/history-and-caching.md), [History](reference/history.md) |
| serve YPY from my backend | [Server](server/index.md): [Go](server/recipes/go.md), [Arandu](server/recipes/arandu.md), [Django](server/recipes/django.md), [Rails](server/recipes/rails.md), [Laravel](server/recipes/laravel.md) |
| move an existing project to YPY | [Migration guide](migration/index.md), [what is different](migration/differences.md), then [`ypy migrate`](migration/migrate.md) and [`ypy check`](migration/upgrade-check.md) |
| ship it safely | [Security checklist](guides/security-checklist.md), [Security](reference/security.md), [Accessibility](guides/accessibility.md) |
| make it fast | [Performance](essays/performance.md), [the performance plan](performance/plan.md) |
| receive live updates from the server | [Streaming](extensions/stream.md), [server streaming](server/streaming.md), [Signals](extensions/signals.md) |
| get autocomplete and types in my editor | [Editor support](reference/editors.md) |
| update the page without leaving it half changed | [Atomic swaps and navigation](reference/atomic.md), [Timing](reference/timing.md) |
| open a menu, a dialog or a modal | [Portals guide](guides/portals.md), [Layers](extensions/layers.md) |
| use or write an extension | [Using and writing extensions](extensions/using-and-writing.md), [Official extensions](extensions/index.md) |
| use a widget library (Bootstrap, React, Alpine, a Web Component) | [Third-party components](guides/third-party-components.md), [`ypy-preserve`](reference/attributes/ypy-preserve.md) |
| test my server without a browser | [Testing your app](server/testing.md) |
| know why hypermedia, and when not | [Why hypermedia](essays/why-hypermedia.md), [When to use YPY](essays/when-to-use-ypy.md), [No build step](essays/no-build-step.md), [Progressive enhancement](essays/progressive-enhancement.md), [Template fragments](essays/template-fragments.md) |
| work on YPY itself | [Core map](contributing/core-map.md), [Coding standards](contributing/coding-standards.md) |

### Conventions of these pages

- Names are `ypy-*` and `data-ypy-*` (attributes), `YPY-*` (headers), `ypy:*` (events) and `ypy` (the global object).
- Examples show the HTML of the page and, where the server matters, the Go code with the SDK
  `github.com/ypy-develop/go`. The [Go SDK pages](../go/README.md) and its [guide](../go/guide.md) cover the SDK in
  detail.
- The documentation is in English and Brazilian Portuguese. The Portuguese pages mirror these paths under
  [`pt-BR/`](pt-BR/index.md).
