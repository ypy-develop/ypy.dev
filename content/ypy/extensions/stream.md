---
title: "Streaming: the ypy-stream extension"
description: "The client of the YPY streaming protocol: one long response that patches HTML and signals, redirects, fires events and reconnects by itself."
order: 400
---
# Streaming: the `ypy-stream` extension

The server pushes changes to a page over one long HTTP response. The page opens the stream, reads it as it arrives and
applies each event: patch some HTML, patch the signals, send the user somewhere, fire an event. The wire format is
server-sent events, and any language can write it with plain text and flushes. The Go package has helpers for it
([server side](../index.md#server-side)).

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-signals.min.js"></script>   <!-- optional: ypy-patch-signals needs it -->
<script src="ext/ypy-stream.min.js"></script>

<div id="room" ypy-stream="/room/events">
  <ul id="messages"></ul>
  <b ypy-text="$online"></b> online
</div>
```

## Opening a stream

The `ypy-stream` attribute holds the URL and the verb is GET. Other verbs are modifiers: `ypy-stream.post`, `.put`,
`.patch` and `.delete`. The stream opens on `load`, or on the trigger named by `ypy-trigger`.

```html
<div ypy-stream="/feed"></div>
<button ypy-stream="/feed" ypy-trigger="click">Follow</button>
<form ypy-stream.post="/jobs" ypy-trigger="submit">...</form>
```

Any request whose answer has `Content-Type: text/event-stream` and status `200` also turns into a stream, and so does
`ypy.ajax('GET', '/feed', {source: element})`. The request is a normal YPY request: `YPY-*` headers, `ypy-headers`,
`ypy-vals`, `ypy-include`, the form values and, with [`ypy-signals`](signals.md), the current signals. Every
reconnection builds it again. An element opens one stream at a time, and the stream lets go of the request queue as
soon as the headers arrive.

## Events

The body is UTF-8 text with the usual server-sent events framing. A YPY event has an `event` name and `data` lines of
the form `data: <key> <value>`. A key repeated on several lines has its values joined with a line feed, which is how
multi-line HTML travels. Unknown events and keys are ignored, so old pages survive new servers. Events are applied
one at a time, in order. A failed event fires `ypy:stream:error` on the element and the stream goes on.

```
event: ypy-patch-elements
data: mode append
data: selector #messages
data: elements <li>how are you?</li>

```

### `ypy-patch-elements`

| Key | Value | Default |
|---|---|---|
| `mode` | `outer`, `inner`, `replace`, `append`, `prepend`, `before`, `after`, `remove` | `outer` |
| `selector` | a selector; the extended selectors work (`closest .x`, `find .y`) | none |
| `namespace` | `html`, `svg`, `mathml` | `html` |
| `useViewTransition` | `true` | off |
| `key`, `limit` | for `append` and `prepend` only, see below | none |
| `elements` | the HTML | empty |

| Mode | Effect on each matched element | Swap style |
|---|---|---|
| `outer` | morph it into the one element of `elements` | `outerMorph` |
| `inner` | morph its children into `elements` | `innerMorph` |
| `replace` | replace it with the nodes, no morph | `outerHTML` |
| `append`, `prepend` | add the nodes as its last or first children | `beforeend`, `afterbegin` |
| `before`, `after` | add them before or after it | `beforebegin`, `afterend` |
| `remove` | remove it | `delete` |

A patch is a swap, so focus, selection, scroll and typed text survive a morph, scripts follow the nonce gate,
`<ypy-partial>` and `ypy-swap-oob` in the HTML go to their own targets, and `ypy:before:swap` and `ypy:after:swap`
fire on the streaming element.

With a `selector`, every matching element is patched. Without one, each top-level element of `elements` goes to the page
element with the same `id`. `outer`, `replace` and `remove` take that element itself; the other modes take its children,
so `elements <ul id="messages"><li>new</li></ul>` in mode `append` adds the `<li>` at the end of `#messages`. A patch
that finds no place fires the cancelable `ypy:target:missing` with `detail.type` set to `"stream"`.

`key` and `limit` make `append` and `prepend` a bounded, keyed list, with the rules of [`ypy-list`](list.md): a new
item whose `key` attribute matches an item in the list replaces it, and `limit` drops the oldest items. A chat, a log or
a feed needs both, or the same item shows twice when the server repeats it.

### `ypy-patch-signals`

`signals` is a JSON object on one line, applied as a JSON Merge Patch (RFC 7396). `onlyIfMissing true` sets only the
signals that do not exist yet, which is how a server sends defaults.

```
event: ypy-patch-signals
data: signals {"theme":"light","page":1}
```

With [`ypy-signals`](signals.md) loaded, the patch goes to the nearest scope that has each key. Without it the extension
fires `ypy:stream:signals` with `{patch, onlyIfMissing}` for your own code.

### `ypy-redirect`

`url` is required. `mode` is `navigate` (a boosted visit that pushes the URL, the default), `replace` (the same, replacing the
history entry) or `hard` (a full page load). The stream closes first and does not reconnect. A URL on another origin is
refused while `selfRequestsOnly` is on, and `javascript:` and `data:` URLs are always refused.

### `ypy-event`

Fires a `CustomEvent` with the exact `name` (no `ypy:` prefix), the JSON `detail`, on the first element that matches
`selector` or on the document. It bubbles and can be cancelled.

```
event: ypy-event
data: name message:new
data: detail {"id":42}
```

### `ypy-refresh`

The event carries no HTML. It says "something changed" and each client fetches its own copy, with its own session,
cookies and signals. Use it when a patch cannot be the same for everyone, such as a balance or a list filtered by
permission.

| Key | Value | Default |
|---|---|---|
| `selector` | the elements to fetch again | the page |
| `debounce` | milliseconds, 0 to 60000 | 250 |
| `jitter` | milliseconds, 0 to 60000 | 0 |

Without `selector`, the client sends a GET to its own address and morphs the answer into the element marked
`ypy-history-elt`, or into the body. With `selector`, each matching element that has a `ypy-get` makes that request and
morphs the answer into its children. Events with the same selector inside the `debounce` window make one request, and
`jitter` adds a random wait so a crowd of clients does not ask at the same second. The cancelable
`ypy:stream:refresh` fires just before the request.

## Reconnection

A stream that ends reconnects, unless a stop condition holds.

- The delay starts at the last `retry` value the server sent (1 second without one), doubles on every failed attempt up
  to 30 seconds, and each wait is a random 75 to 100 percent of that. It starts over after a connection delivers an
  event.
- The request carries `Last-Event-ID` when the server sent an `id`.
- While the browser is offline, the client waits for the `online` event and does not count it as a failed attempt.
- Every connection should start with the full state: a patch for each region and a signals patch. Morphs are idempotent,
  so repeating the state is invisible when nothing changed and puts the page right when something was missed.

It does not reconnect when the answer is `204` (the server's "no more"), when the status is not `200` or the answer is not
an event stream (the core handles it like any other answer), after a `ypy-redirect`, or when the element left the page
or the code closed the stream.

### Pausing in a background tab

With `pauseOnBackground` the extension closes the connection while the tab is hidden and opens it again, at once and
with the current signals, when the tab is shown. It is off by default.

```html
<div ypy-stream="/room/events" ypy-config='{"stream":{"pauseOnBackground":true}}'>...</div>
```

`ypy:stream:pause` and `ypy:stream:resume` fire around it. A page in a hidden tab gets nothing, so do not use it for a
stream whose events must be counted, such as a notification sound.

### Silent connections

A client cannot tell a quiet stream from a dead one, so the server keeps the connection alive with a comment line
(`: keepalive`), which the decoder ignores. The Go package writes one after 15 seconds of silence. For elements that use
`ypy-sse:connect` or `ypy-ws:connect`, `sse.heartbeat` and `ws.heartbeat` close and reopen a connection that got no data
for the time you set.

## Closing

A stream closes for good when its element is removed, swapped out or closed with `ypy.stream.close(element)`. The
connection is aborted and the timers stop. The browser closes the streams of a page that unloads.

## Routing by Content-Type

| Content-Type of the answer | What happens |
|---|---|
| `text/html` | swapped or morphed into the target |
| `application/json` | a signals patch ([`ypy-signals`](signals.md)) |
| `text/event-stream`, status `200` | read as a stream by this extension |

Both `ypy-sse` and `ypy-stream` answer to `text/event-stream`, and the first one registered wins. Load `ypy-stream`
before `ypy-sse`. An element with `ypy-sse:connect` is left to `ypy-sse`.

## JavaScript API and events

| Member | What it does |
|---|---|
| `ypy.stream.close(element)` | Closes the stream of an element. It does not reconnect. |
| `ypy.stream.stats()` | `{open}`: the number of open streams, for leak tests. |

| Event | When |
|---|---|
| `ypy:stream:open` | a connection delivered its headers (first or after a reconnection) |
| `ypy:stream:reconnect` | the client is about to wait `detail.delay` ms |
| `ypy:stream:pause`, `ypy:stream:resume` | the tab hid or showed, with `pauseOnBackground` |
| `ypy:stream:refresh` | a `ypy-refresh` is about to request; cancelable |
| `ypy:stream:close` | the stream ended for good, with `detail.reason` |
| `ypy:stream:error` | an event failed or the read broke |
| `ypy:stream:signals` | a signals patch arrived and `ypy-signals` is not loaded |
| `ypy:target:missing` | a patch had no element to go to; cancelable |

The one setting is `stream.pauseOnBackground` (default `false`). `ypy.config.stream` sets it for the page and `ypy-config`
on the element overrides it.

## Security and limits

- A patch goes through the same pipeline as a swapped response, so the nonce gate and Trusted Types apply, and
  `selfRequestsOnly` applies to the URL that opens the stream. The server must escape user data in `elements`.
- A stream opened with GET must not change state. Signals travel in the query on GET, so keep secrets out of them or set
  `ypy-signals-send="false"` on the element.
- HTTP/1.1 allows six connections per host. Use HTTP/2 and one stream per page.
- A page with an open stream is generally not eligible for the browser's back-forward cache.
- A server that closes the stream with status `200` and no events makes the client reconnect forever. Answer `204` to the
  next request to stop it.
- A patch without `selector` needs top-level elements with an `id` that exists on the page. A typo changes nothing and
  only `ypy:target:missing` tells you.
