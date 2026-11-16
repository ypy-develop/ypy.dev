---
order: 10
id: "DOC-index"
type: "index"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/index", "lang/en"]
title: "Docs"
description: "One page on YPY, from the first attribute to security and configuration: install it, make requests, swap HTML, handle errors and history, serve it."
---

# YPY documentation

YPY is a hypermedia framework by HYZIS. This page is the whole manual in one reading order, from the first attribute to
configuration and the server. The names of every attribute, header, event and setting are in the
[reference](reference.md), the optional scripts are in [extensions](extensions/index.md), and the reasons behind the
design are in the [essays](essays/index.md).

## Introduction

An HTML element says in its attributes which request to make, when to make it and where to put the HTML that comes
back. The server answers with HTML fragments instead of JSON. You add one `<script>` tag and build nothing.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<button ypy-get="/hello" ypy-target="#out" ypy-swap="innerHTML">Say hello</button>
<div id="out"></div>
```

Read it as a sentence: on click, `GET /hello`, and put the answer inside `#out`. If the server answers
`<p>Hello</p>`, that paragraph appears and the rest of the page stays as it was, including the address bar.

Four questions cover most of what a page needs:

| Question | Attribute | Default |
|---|---|---|
| Which request? | `ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` | none |
| When? | `ypy-trigger` | `click`; `change` for inputs and selects; `submit` for forms |
| Where does the answer go? | `ypy-target` | the element itself |
| How is it put in? | `ypy-swap` | `innerMorph` |

Attributes are `ypy-<name>` or `data-ypy-<name>`, headers are `YPY-<Name>`, events are `ypy:<phase>:<verb>` and the
global object is `ypy`.

### Why this shape

A page that asks the server for HTML needs no client-side router, no state store and no JSON contract to keep in sync
with the markup. The server already knows how to draw the page, so it draws the piece that changed. The default swap
is a morph, so the focus, the typed text and the scroll position of what was already there survive. The core stays
under 18,432 bytes min+gzip (the budget the build checks), and each extension under 6,144 bytes. The
[essays](essays/index.md) go into when this fits and when it does not.

If you already have a project written for an older library, read [Migration](migration.md).

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

The package has the same `dist/` files as the repository build, and `https://unpkg.com/@ypy-dev/ypy@0.1.1/dist/ypy.min.js`
serves the same minified core.

### Download a copy

To build the files yourself you need Go 1.26 or later. The build writes them to `dist/`:

```sh
go run ./cmd/ypy build
```

| File | Use it when |
|---|---|
| `dist/ypy.min.js` | you want the core alone and add the extensions you need |
| `dist/ypy.max.min.js` | you want the core and the official extensions that are safe to ship, in one file |
| `dist/ypy.esm.min.js` | you load it as an ES module (`import ypy from "./ypy.esm.min.js"`); it also sets `window.ypy` |
| `dist/ext/<name>.min.js` | you want one extension, loaded after the core |
| `dist/ypy.css`, `dist/ypy.min.css` | you want the small stylesheet for the indicator, the progress bar and the opt-in transitions |
| `dist/ypy.d.ts` | you want the TypeScript types of `window.ypy` |

Each `.min.js` has a source map next to it. If you copy only the `.js` file into your own `static/` directory, copy
from `dist/embed/` instead. Those files are the same code without the `sourceMappingURL` comment, so the browser does
not request a map that is not there. Editors can autocomplete the names too: `ypy gen` writes the data for VS Code,
JetBrains and a JSON Schema of the configuration from the same table the [reference](reference.md) is built from.

## Requests

`ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` and `ypy-query` each issue the request of their name to the
URL in the value. When the method and the URL come from different places, set `ypy-method` and `ypy-action`.

```html
<form ypy-post="/items" ypy-target="#list" ypy-swap="beforeend">
  <input name="title" required>
  <button>Add</button>
</form>
<ul id="list"></ul>
```

A form sends its fields. Any other element sends its own `name` and `value` if it has them, plus the values of the
form around it. `ypy-ignore` makes YPY skip an element and everything inside it.

## Triggers

`ypy-trigger` takes a comma-separated list of events, each with optional modifiers. Without it, the trigger is the
natural event of the element: `click`, `change` for inputs and selects, `submit` for forms.

```html
<input name="q" ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms, search">
```

That input waits 300 ms after the last key and asks only when the value changed.

### Modifiers

| Modifier | Effect |
|---|---|
| `once` | only the first time |
| `changed` | only when the value of the element changed |
| `delay:<time>` | waits; a new event restarts the wait |
| `throttle:<time>` | the first event goes at once, the following ones wait for the interval |
| `from:<selector>` | listens on another element (`document`, `window`, `closest <selector>`, and so on) |
| `target:<selector>` | only when `event.target` matches |
| `prevent`, `stop`, `halt` | `preventDefault()`, `stopPropagation()`, or both |
| `capture`, `passive` | the listener options of the same name |

### Filters

An event can carry a filter in brackets, such as `click[ctrlKey]`. The filter is code, so it runs only when
`config.callbacks` allows author code; see [Security](#security). The [`ypy-expr`](extensions/expr.md) extension
evaluates filters with its own small language, with no `eval`.

### Special events

| Event | Fires |
|---|---|
| `load` | when the element is processed |
| `revealed` | the first time the element scrolls into view |
| `intersect` | each time the element meets the viewport (`root:<selector>`, `threshold:<number>`) |
| `every <time>` | on a timer, for example `every 2s` |

```html
<!-- load the next page when this row comes into view -->
<tr ypy-get="/rows?page=2" ypy-trigger="revealed" ypy-swap="afterend"></tr>
```

### Polling

`every` makes a fixed interval. A plain interval keeps asking in a background tab. The [`ypy-poll`](extensions/poll.md)
extension slows it down there, lets the server choose the next interval and lets it end the polling with a `286`. The
[polling example](examples/polling.md) shows it end to end.

### Load polling

Combine `load` with `every` to fetch once at once and then keep fetching:

```html
<div ypy-get="/stats" ypy-trigger="load, every 30s"></div>
```

## Request indicators

While a request runs, the element that sent it gets the class `ypy-request` (and `data-ypy-state="request"`).
`ypy-indicator` names other elements to mark instead. An element with the class `ypy-indicator` stays hidden until it,
or an ancestor, has `ypy-request`.

```html
<button ypy-post="/report" ypy-indicator="#spinner">Build report</button>
<span id="spinner" class="ypy-indicator">Working...</span>
```

To lock controls while the request runs, load [`ypy-disable`](extensions/disable.md). To draw a bar for page
navigation, load [`ypy-progress`](extensions/progress.md). The classes are listed in
[CSS classes](reference.md#css-classes).

## Targets

`ypy-target` is a CSS selector, `this`, or a relative form: `closest <selector>`, `find <selector>`, `next [selector]`,
`previous [selector]`. When the selector matches nothing, YPY logs it and fires `ypy:target:missing`. A response can
retarget itself with the `YPY-Retarget` header.

`ypy-main` marks the main target, the element a swap falls back to when its own target is missing. Without a marker,
`<main>` is the main target. `ypy-fallback` turns that fallback on or off for one element, and `config.targetFallback`
does it for the page.

## Swapping

`ypy-swap` says how the answer enters the target. When you write nothing, the swap is a morph: the new HTML is merged
into what is there, so elements that exist on both sides stay the same DOM nodes.

### Styles

| Style | What it does |
|---|---|
| `innerMorph` (default) | merges the response into the children of the target |
| `outerMorph` | merges the response into the target itself |
| `innerHTML`, `outerHTML` | replaces the children, or the target, with nothing kept |
| `outerSync` | copies the attributes of the target and replaces its children |
| `beforebegin`, `afterbegin`, `beforeend`, `afterend` | inserts next to or inside the target |
| `textContent` | sets the text; the response is not parsed as HTML |
| `delete`, `none` | removes the target; does nothing |

Extensions add styles: `upsert`, `download`, and `key:` and `limit:` for lists ([`ypy-list`](extensions/list.md)).

### Morph

A morph walks both trees and keeps what the visitor did. Focus, typed text, a selection, the scroll position of a
block, a playing video, an open `<details>` or `<dialog>` all survive. A checked box or a chosen option stays until
the response changes it. To let the response win on one element, mark it in the response with `ypy-morph-force`.

Give an `id` to anything that must keep its identity, such as a field or a panel. Elements without an `id` match by
position and tag. `config.morphScanLimit` bounds how far a morph looks for the matching node, so lower it for long
lists, or write `ypy-swap="innerHTML"` for a region that is replaced whole.

A default swap that meets `document.body`, like a boosted link, replaces the page instead of morphing it, so the text
typed on the old page does not leak into the new one. To refresh the same page by morphing, write the style:
`ypy-swap="innerMorph"`.

### View transitions

Add `transition:true` to a swap, or set `config.transitions`, to wrap it in the View Transitions API. The
[`ypy-transitions`](extensions/transitions.md) extension fires `ypy:before:viewTransition` and
`ypy:after:viewTransition` around it, and without it the swap runs at once with no animation.

### Options

Modifiers follow the style: `ypy-swap="innerHTML swap:200ms settle:100ms show:top"`.

| Option | Effect |
|---|---|
| `swap:<time>` | delay between the request ending and the swap; the target has `ypy-swapping` meanwhile |
| `settle:<time>` | delay before the settle phase ends; the new content has `ypy-added` meanwhile |
| `scroll:<top or bottom>`, `show:<top or bottom>` | scrolls the target, or the page, after the swap |
| `transition:true` | wraps the swap in a view transition |

`ypy-status:<code>` changes `swap:`, `target:`, `select:`, `push:`, `replace:` and `transition:` for one status or a
range such as `5xx`; see [Status codes](#status-codes-and-errors).

## Synchronization

When an element asks for a request while another one from the same place is still running, `ypy-sync` decides what
happens. The default is `queue first`: the new request waits, unless one is already waiting. The other strategies are
`drop`, `abort`, `replace`, `queue last`, `queue all`, `latest` and `parallel`.

```html
<input name="q" ypy-get="/search" ypy-trigger="input changed delay:200ms" ypy-sync="this:replace">
```

An old response never overwrites a newer one. [`ypy-dedupe`](extensions/dedupe.md) joins identical GETs into one
request.

## CSS transitions

YPY adds a class at each step of a swap, so CSS can animate it. The target gets `ypy-swapping` before the swap,
the new content gets `ypy-added` and the target gets `ypy-settling` until the settle delay ends. To let the old content
leave before the new one arrives, add a swap delay.

```html
<div id="panel" class="ypy-fade" ypy-get="/panel" ypy-trigger="click from:#refresh"
     ypy-swap="innerHTML swap:200ms">...</div>
```

`ypy.css` has two opt-in transitions. `ypy-fade` fades the target out while it swaps and `ypy-fade-in` fades the new
content in. Keep the element across the swap, with a stable `id`, if the transition should run from its old state to
its new one.

## Out-of-band swaps

A response can update more than the target. Put each extra piece in a `<ypy-partial>`, or mark an element with
`ypy-swap-oob`. YPY swaps the main content first, then puts each extra piece where its `id` points.

```html
<li>New item</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">3 items</ypy-partial>
<ypy-partial id="toast"><p>Saved</p></ypy-partial>
```

### `<ypy-partial>`

The element takes `ypy-target` or an `id` to choose the destination and `ypy-swap` to choose the style. A partial with
no `ypy-swap` merges as `innerMorph`.

### `ypy-swap-oob` and `ypy-select-oob`

`ypy-swap-oob="true"` on an element of the response swaps it, as an `outerMorph`, into the page element with the same
`id`. The value can be a style, or `<style>:<selector>` to choose the target. `ypy-select-oob` on the request picks
elements of the response by selector and swaps each one out of band.

To make the whole response all or nothing, the server sends `YPY-Group: atomic`.

## Selecting content

`ypy-select` takes one part of the response by CSS selector and drops the rest. It is the way to reuse one template
for a page and for its fragments, as in the [template fragments essay](essays/template-fragments.md).

```html
<button ypy-get="/orders" ypy-target="#orders" ypy-select="#orders">Reload</button>
```

The server can choose a part itself with the `YPY-Reselect` header.

## Preserving content

`ypy-preserve` says how much of an element a swap leaves alone. With no value it keeps the node by `id`. `all` keeps
the attributes and the children, `attrs` stops the morph from changing attributes (`ypy-preserve-attrs="class,style,aria-*"`
lists which), `children` keeps the content, and `false` turns the protection off. `ypy-morph-skip` and
`ypy-morph-skip-children` freeze an element or its children during a morph.

```html
<video id="player" ypy-preserve src="/intro.mp4" controls></video>
```

The [`ypy-preserve` extension](extensions/preserve.md) adds words for the visitor's own state (typed value, selection, focus, scroll), and
[`ypy-persist`](extensions/persist.md) keeps an element across boosted navigations. The
[preserve file inputs example](examples/preserve-file-inputs.md) shows a file field that survives a validation error.

## Parameters

A form sends its fields. Any other element sends its own `name` and `value`, plus the values of the form around it. A
`GET` puts them in the query string. Every other method sends them in the body.

```html
<button ypy-post="/votes" ypy-vals='{"item": 42}' ypy-include="#comment">Vote</button>
```

`ypy-vals` adds values from JSON, `ypy-include` adds the values of other elements by selector, and `ypy-headers` adds
request headers.

### File upload

Set `ypy-encoding="multipart/form-data"` and the request sends files.

```html
<form ypy-post="/files" ypy-encoding="multipart/form-data" ypy-target="#result">
  <input type="file" name="doc">
  <button>Send</button>
</form>
```

For a progress bar while the file goes out, load [`ypy-upload`](extensions/upload.md). See the
[file upload example](examples/file-upload.md).

### Extra values

`ypy-vals` and `ypy-include` are the two ways to send what is not in a form: a fixed value, or the value of an element
elsewhere. `ypy-config` changes the fetch options of one element (`timeout`, `credentials`, `cache`); it cannot change
`mode`.

## Confirming requests

`ypy-confirm` asks before the request. The value is the message.

```html
<button ypy-delete="/items/7" ypy-confirm="Delete this item?" ypy-target="closest li" ypy-swap="delete">Delete</button>
```

The `ypy:confirm` event lets you replace the browser dialog with your own. The [`ypy-prompt`](extensions/prompt.md)
extension asks for a text and sends it in the `YPY-Prompt` header. See [confirm and prompt](examples/confirm-and-prompt.md).

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

A bare attribute on a container stays on the container. Inheritance is always explicit, so reading one element tells
you what it does.

## Boosting

`ypy-boost` turns the links and forms inside an element into requests that swap the body and push the address. The
page does not reload, so scripts and styles stay loaded. If JavaScript fails to load, the links and forms still work
as usual, which makes boosting a progressive enhancement.

```html
<body ypy-boost="true">
  <nav><a href="/orders">Orders</a> <a href="/people">People</a></nav>
  <main id="main">...</main>
</body>
```

Set `ypy-boost="false"` on a subtree to leave it alone. The [`ypy-nav`](extensions/nav.md) extension boosts the whole
page with one setting, prefetches links on hover and marks the current link. [`ypy-head`](extensions/head.md) merges
the `<head>` of the response. After a boosted navigation, focus moves to the main heading and a live region announces
the new title. `ypy-focus` chooses another place, `ypy-announce` sets the message, and `ypy-scroll` keeps, resets or
restores the scroll position. See the [progressive enhancement essay](essays/progressive-enhancement.md).

## SSE and WebSockets

Server-Sent Events and WebSockets belong to extensions, so the core stays small.

```html
<!-- Server-Sent Events: each message is swapped into the target -->
<div ypy-sse:connect="/feed" ypy-target="#feed"></div>

<!-- WebSocket: messages in, a form that sends -->
<div ypy-ws:connect="/chat" ypy-target="#messages">
  <ul id="messages"></ul>
  <form ypy-ws:send><input name="text"><button>Send</button></form>
</div>
```

[`ypy-sse`](extensions/sse.md) and [`ypy-ws`](extensions/ws.md) are plain transports. [`ypy-stream`](extensions/stream.md)
speaks the YPY streaming protocol: HTML patches, signal patches, redirects and events over one response. For an answer
that arrives token by token, use [`ypy-model`](extensions/model.md). The
[live updates example](examples/live-updates.md) puts them together.

## History

A swap that pushes or replaces the address writes a history entry that YPY can restore. `ypy-push-url` pushes a URL
after the swap, `ypy-replace-url` replaces the current entry, and `true` uses the URL of the request. A boosted GET
pushes by default. The server can do the same with the `YPY-Push-Url` and `YPY-Replace-Url` headers.

```html
<a ypy-get="/orders?page=2" ypy-target="#orders" ypy-push-url="true">Next page</a>
```

On back and forward, YPY asks the server for the page again with `YPY-History-Restore-Request: true`, and the server
has to answer with a whole page. `config.history` is `true`, `false` or `"reload"`.

### History element

`ypy-history-elt` marks the element that is restored, instead of `body`.

### Snapshots

[`ypy-history-cache`](extensions/history-cache.md) keeps snapshots for instant back navigation. A widget that changes
the DOM of a snapshot, such as a chart, needs [`ypy-preserve`](#preserving-content) or an `ypy:` event to restore
itself.

### Turning history off

`config.history = false` turns history management off for the page. `ypy-push-url="false"` suppresses the push of one boosted request.

## Requests and responses

### Request headers

Every YPY request carries `YPY-Request: true`. The other headers say who asked and what is being replaced.

| Header | Meaning |
|---|---|
| `YPY-Request` | always `true` on a YPY request |
| `YPY-Source` | the element that triggered the request |
| `YPY-Target` | the element the response is swapped into |
| `YPY-Current-URL` | the current URL of the browser |
| `YPY-Request-Type` | `partial` for a targeted swap, `full` when the target is `body` or `ypy-select` is used |
| `YPY-Boosted` | present on a boosted request |
| `YPY-History-Restore-Request` | present when restoring a page after back or forward |

### Response headers

| Header | Effect |
|---|---|
| `YPY-Trigger` | fires events on the page after the response |
| `YPY-Push-Url`, `YPY-Replace-Url` | pushes or replaces the address |
| `YPY-Redirect` | redirects with a full page load |
| `YPY-Location` | navigates with a request, without a reload |
| `YPY-Refresh` | reloads the page when the value is `true` |
| `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect` | overrides the target, the swap style or the selected part |
| `YPY-Group` | `atomic` makes the swap all or nothing |

The complete tables, with the extension headers, are in the reference: [request headers](reference.md#request-headers)
and [response headers](reference.md#response-headers).

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

The form convention: a `422` means the form is not valid and is rendered back into the form, with the focus on the first
invalid field. A `303` after a POST means done, and YPY swaps the page it redirects to. A network failure, a timeout or
an exception fires `ypy:error`. A 404 or a 500 is an answer, so it goes through the rules above. To repeat a request
that failed, load [`ypy-retry`](extensions/retry.md).

### 4xx and 5xx

A client error (4xx) is usually something the visitor can fix, so render it next to the control: `ypy-error-target`
for the place and `ypy-error-focus` for the field. A server error (5xx) is usually not theirs to fix, so send it to one
visible region with `ypy-status:5xx`. `ypy-error-swap` sets the style used for either.

### Order of operations

For a request that reaches the server and succeeds, the events fire in this order: `ypy:confirm`,
`ypy:config:request`, `ypy:before:request`, `ypy:before:response`, `ypy:after:request`, `ypy:before:swap`,
`ypy:before:settle` and `ypy:after:settle` for each target, `ypy:after:swap` once, then the history events and
`ypy:finally:swap` and `ypy:finally:request`. The address is written after the swap, never before.

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

See the [inline validation example](examples/inline-validation.md).

## Animations

The classes of [CSS transitions](#css-transitions) cover most animations. For a movement between two states of the
same element, use a [view transition](#view-transitions). `ypy.css` carries two opt-in fades, and the
[animations example](examples/animations.md) has a complete page. Under `prefers-reduced-motion`, turn the delays
off in your own stylesheet.

## Extensions

The core covers what nearly every page needs. Anything else is an [extension](extensions/index.md), a script you load
after the core.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-poll.min.js"></script>
<script src="ext/ypy-validate.min.js"></script>
```

Or load `ypy.max.min.js` to get most of the official extensions in one request. The ones meant for development or for
narrow cases are left out; the [extensions page](extensions/index.md) lists them. An extension registers itself with
`ypy.registerExtension`, and `config.extensions` is an allow list: once it is set, only the names in it may register.

A few that most pages want: [`ypy-poll`](extensions/poll.md), [`ypy-nav`](extensions/nav.md),
[`ypy-validate`](extensions/validate.md), [`ypy-cache`](extensions/cache.md), [`ypy-layers`](extensions/layers.md) and
[`ypy-signals`](extensions/signals.md).

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

Initialize a third-party widget for new content with `ypy.onLoad`, add a header for every request in
`ypy:config:request`, and change how a swap behaves in `ypy:before:swap`. The full list, with the `detail` of each, is
in the reference: [events](reference.md#events).

To see what happens, set `ypy.config.logAll = true`. Every YPY event goes to the console.

## Debugging

Three tools, lightest first:

- `config.logAll` logs every event.
- [`ypy-dev`](extensions/dev.md), loaded only in development: errors with the element and the request, an overlay,
  a latency and offline simulator, and warnings for markup that does not do what it looks like.
- [`ypy-inspector`](extensions/inspector.md): a floating panel with requests, swaps, events and signals as they
  happen.

Keep both extensions out of production. To reproduce a problem, reduce the page to the element, the target and the
response; the [`ypy-test`](extensions/test.md) extension adds `ypy.idle()` to browser tests, a promise that resolves
when no request, swap or timer is pending.

## Scripting

`ypy-on:<event>="..."` runs JavaScript when a DOM event fires, and `ypy-on::<event>` is the shorthand for the `ypy:`
events.

```html
<button ypy-on:click="this.classList.toggle('active')">Toggle</button>
<div ypy-get="/panel" ypy-on::after:swap="console.log('swapped')"></div>
```

Author code is off by default (`config.callbacks` is `none`), so these attributes do nothing until you allow code. An
event listener in a script is the safer choice. [`ypy-signals`](extensions/signals.md) keeps state in the browser
without a build step and without `eval`:

```html
<form ypy-signals="name:'', agree:false" ypy-post="/signup">
  <input ypy-bind="name">
  <label><input type="checkbox" ypy-bind="agree"> I agree</label>
  <p ypy-show="$name.length > 0">Hello, <b ypy-text="$name"></b></p>
  <button ypy-attr:disabled="!$agree || $name == ''">Send</button>
</form>
```

## Third-party JavaScript and web components

A widget that owns its DOM, such as a chart, an editor or a web component, must survive a swap. `ypy-preserve` keeps a
node by `id`. `ypy-morph-skip` and `ypy-morph-skip-children` freeze an element or its children during a morph.
`config.morphThirdParty` lists the elements another script owns, by default `[data-reactroot],[data-v-app],[x-data]`: a
response that gives them with no content leaves their children alone.

Start a widget once for each new element with `ypy.onLoad`:

```js
ypy.onLoad("[data-chart]", (el) => new Chart(el));
```

`ypy.process(root)` makes YPY read markup that a script inserted. The full API is in the reference:
[JavaScript API](reference.md#javascript-api). For Alpine, load [`ypy-alpine-compat`](extensions/alpine-compat.md).

## Caching

Three kinds of caching, each separate.

- The HTTP cache of the browser. Send `Vary: YPY-Request` on every response that changes with that header, so the
  fragment and the page do not mix.
- Responses kept by the page, with [`ypy-cache`](extensions/cache.md): show the copy at once, ask the server whether
  it is still current, and swap again only if it is not. Nothing is cached unless the element says `ypy-cache`.
- Snapshots for the back button, with [`ypy-history-cache`](extensions/history-cache.md).

[`ypy-ptag`](extensions/ptag.md) lets the server skip a swap when a polled fragment did not change.

## Security

YPY turns attributes into behavior, so an attacker who can put markup in your page does not need a script. A `ypy-post`
on a button is enough. Three layers defend against that, and none of them replaces escaping on the server.

1. Escape everything on the server. Fragments are HTML like any other. Never let user text become part of a tag, an
   attribute name or a `ypy-*` value, and strip `ypy-*` and `data-ypy-*` from user content.
2. Keep author code off. `config.callbacks` is `none` by default, so `ypy-on`, `js:` values and bracket filters do
   nothing.
3. Keep requests on your origin. `config.mode` is `same-origin`.

### Content Security Policy and nonce

With [`ypy-csp`](extensions/csp.md), author code runs only with the nonce of the page: set `config.callbacks` to
`nonce`. Scripts in responses follow `config.scripts`: `auto` lets them run, `nonce` runs only those with the nonce of
the page, and `block` runs none. A strict policy needs no inline script or style, so keep your handlers in files.

### CSRF

The [`ypy-csrf`](extensions/csrf.md) extension reads a token from `<meta name="ypy-csrf">` and sends it with every
request that is not a `GET` or `HEAD` and goes to your own origin. It never sends the token to another origin. When a
response carries `YPY-CSRF`, the token is replaced for the next request.

Read the [security essay](essays/security.md) for the reasoning behind these defaults.

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
it. Every key, with its type and default, is in [Configuration](reference.md#configuration).

## Server side

YPY needs nothing special from the server: it answers HTTP with HTML. Any language that can write a response works.

The browser sends `YPY-Request: true` on each request, so the server can tell a fragment request from a page load and
answer with the piece instead of the whole page. When the same URL returns both, send `Vary: YPY-Request` so that
caches keep them apart.

```http
GET /orders?page=2 HTTP/1.1
YPY-Request: true
YPY-Target: div#orders
```

```http
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
YPY-Push-Url: /orders?page=2
YPY-Trigger: orders:loaded

<ul>...</ul>
```

Use `422` for an invalid form, `303` after a successful POST and `204` for "nothing to swap". The headers the server can
send are in [Response headers](#response-headers). For Go there is a package that writes these for you, streams
fragments and tests handlers without a browser: [`github.com/ypy-develop/go`](https://pkg.go.dev/github.com/ypy-develop/go).
