---
title: "Streaming: the ypy-stream extension"
description: "The client side of the YPY streaming protocol: one long response that patches HTML and signals, redirects and fires events, with reconnect."
order: 400
---
# Streaming: the `ypy-stream` extension

The server pushes changes to a page over one long-lived HTTP response. The page opens the stream, reads it as it
arrives and applies each event: patch some HTML, patch the signals, send the user somewhere, fire an event. This is
the client side of the YPY protocol over server-sent events; the wire format is
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/develop/spec/PROTOCOL.md) in the Go SDK repository, which
any language can implement with plain text and flushes.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-signals.min.js"></script>   <!-- optional: ypy-patch-signals needs it -->
<script src="ext/ypy-stream.min.js"></script>

<div id="room" ypy-stream="/room/events">
  <ul id="messages"></ul>
  <b ypy-text="$online"></b> online
</div>
```

```go
// The server, with the Go SDK (github.com/ypy-develop/go).
room := ypy.NewBroadcaster(ypy.WithSnapshot(func(r *http.Request) ([]ypy.Event, error) {
	ev, err := ypy.PatchElementsEvent(renderMessages()) // the full current state: a fat morph
	return []ypy.Event{ev}, err
}))
defer room.Close()
mux.Handle("GET /room/events", room)

// from any handler: every connected page changes
room.PatchElements(`<li>hello</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"))
room.PatchSignals(map[string]any{"online": room.Len()})
```

For one client, `ypy.NewStream(w, r)` gives the same methods (`PatchElements`, `PatchSignals`, `Redirect`, `Event`)
on a single connection. See the [Go SDK README](https://github.com/ypy-develop/go#streaming).

The extension is `ext/ypy-stream.js`: 4,331 bytes min+gzip, inside the 6,144-byte budget every extension has. It uses
only the public API and the extension API.

Contents: [Opening a stream](#opening-a-stream) · [Events](#events) · [`ypy-patch-elements`](#ypy-patch-elements) ·
[Lists](#lists-key-and-limit) · [`ypy-patch-signals`](#ypy-patch-signals) · [`ypy-redirect`](#ypy-redirect) ·
[`ypy-event`](#ypy-event) · [`ypy-refresh`](#ypy-refresh) · [Reconnection](#reconnection) ·
[Pausing in a background tab](#pausing-in-a-background-tab) · [Silent connections](#silent-connections-and-keepalive) ·
[Closing](#closing) · [Routing by Content-Type](#routing-by-content-type) · [JavaScript API](#javascript-api) ·
[Client events](#client-events) · [Configuration](#configuration) · [Security](#security) · [Limits](#limits) ·
[Pitfalls](#pitfalls)

## Opening a stream

There are three ways, and they all open the same stream.

The first is the `ypy-stream` attribute. The value is the URL and the verb is GET. Other verbs are modifiers:
`ypy-stream.post`, `.put`, `.patch` and `.delete`. The stream opens on `load` (the default) or on the trigger the
element names with `ypy-trigger`. All spellings also exist with the `data-` prefix.

```html
<div ypy-stream="/feed"></div>                                   <!-- on load -->
<button ypy-stream="/feed" ypy-trigger="click">Follow</button>   <!-- on click -->
<form ypy-stream.post="/jobs" ypy-trigger="submit">...</form>    <!-- POST with the form's values -->
```

The request is an ordinary YPY request, built by the same code as every other: the `YPY-*` headers, `ypy-headers`,
`ypy-vals`, `ypy-include`, the form values and, with the [signals extension](signals.md), the current signals
(protocol section 3.2). Each connection, the first and every reconnection, builds it again, so the signals are the
current ones. The `Accept` header is `text/event-stream`.

An element opens one stream at a time: the trigger of an element whose stream is live does nothing. Content that a
swap brings in is scanned like any other, so a `ypy-stream` element in a response opens when it arrives.

The second is any request whose answer is an event stream. A `<button ypy-get="/feed">` that gets `Content-Type: text/event-stream`
with status `200` turns into a stream; the answer is not swapped into the button. See
[routing](#routing-by-content-type).

The third is code: `ypy.ajax('GET', '/feed', {source: element})` makes the same request as the second way.

The stream is not a request in the sense of `ypy-sync`: it lets go of the request queue of its element as soon as the
headers arrive, so the element can send other requests while the stream is open. Opening again from an element that
has a live stream (a second click on a `ypy-get` button, say) closes the old stream and opens the new one.

## Events

The body is UTF-8 text in the SSE framing: lines end in LF, CRLF or CR, a blank line ends a frame, `:` starts a
comment (the keepalive), the fields are `event`, `data`, `id` and `retry`. The decoder follows the WHATWG rules,
including a missing space after the colon, a BOM at the start, and a frame cut at any byte, even in the middle of
a multi-byte character or between the CR and the LF of a line end. A last frame without its blank line is dropped.

A YPY event is a frame with an `event` name and `data` lines of the form `data: <key> <value>`; the key has no spaces
and the value is everything after the first space. A key repeated on several lines has its values joined with a line
feed, which is how multi-line HTML travels (one `elements` line per line of HTML). Unknown event names and unknown keys
are ignored, so the protocol can grow without breaking old pages. Events are applied one at a time, in order; a
patch that waits for a view transition holds the ones behind it.

An event that fails (bad JSON, an unknown mode, a selector that does not parse, a redirect that is refused) fires
`ypy:stream:error` on the streaming element and the stream goes on with the next event.

```
event: ypy-patch-elements
data: mode append
data: selector #messages
data: elements <li>how are you?</li>

```

## `ypy-patch-elements`

| Key | Value | Default |
|---|---|---|
| `mode` | `outer`, `inner`, `replace`, `append`, `prepend`, `before`, `after`, `remove` | `outer` |
| `selector` | a selector; the extended selectors of YPY work (`this`, `closest .x`, `find .y`, ...) | none |
| `namespace` | `html`, `svg`, `mathml` | `html` |
| `useViewTransition` | `true` | off |
| `key` | an attribute name, for `append` and `prepend` only: see [Lists](#lists-key-and-limit) | none |
| `limit` | a whole number of 1 or more, for `append` and `prepend` only | none |
| `elements` | the HTML | empty |

| Mode | Effect on each matched element | Swap style |
|---|---|---|
| `outer` | morph it into the one element of `elements` | `outerMorph` |
| `inner` | morph its children into `elements` | `innerMorph` |
| `replace` | replace it with the nodes of `elements`, no morph | `outerHTML` |
| `append` | add the nodes as its last children | `beforeend` |
| `prepend` | add them as its first children | `afterbegin` |
| `before` | add them before it | `beforebegin` |
| `after` | add them after it | `afterend` |
| `remove` | remove it (`elements` is not used) | `delete` |

A patch is a swap. It goes through `ypy.swap`, the same code as the answer to a request, so everything the swap does
applies: the morph keeps focus, selection, scroll and what the user typed (`ypy-preserve` and `ypy-morph-skip` are
honored), scripts are re-created and so follow the nonce gate and Trusted Types, `ypy-swap-oob` elements and
`<ypy-partial>` in the HTML go to their own targets, and the `ypy:before:swap` / `ypy:after:swap` events fire on the
streaming element with `ctx.target` set. The `replace` and insertion modes do not keep that state.

Which elements a patch reaches depends on the `selector`. With one, every element that matches (relative to the streaming element for the extended
selectors, to the document for plain ones). Without one, each top-level element of `elements` goes to the page element
with the same `id`:

- `outer` and `replace` take the top-level element itself; `remove` removes the element with that id;
- `inner`, `append`, `prepend`, `before` and `after` take its children, so the top-level element is a named
  container: `elements <ul id="messages"><li>new</li></ul>` with `mode append` adds `<li>new</li>` at the end of
  `#messages`.

A top-level element with no `id`, an `id` that is not on the page, or a selector that matches nothing cannot be
placed. The extension fires `ypy:target:missing` on the streaming element, with `detail.type` `"stream"` and
`detail.selector` (empty for an element with no id), and goes on with the next event. The event is cancelable; unless a
listener cancels it, a warning goes to the console.

```js
document.addEventListener('ypy:target:missing', e => {
  if (e.detail.type === 'stream') { e.preventDefault(); resync() }   // ask for the full state again
});
```

With the namespace `svg` or `mathml` the fragment is parsed inside that namespace, so `<circle r="4"/>` becomes an
SVG element and `<mi>x</mi>` a MathML one.

With `useViewTransition true` the patch runs inside `document.startViewTransition` when the
browser has it, and plainly when it does not.

### Lists: `key` and `limit`

An `append` or `prepend` patch that sends one item at a time makes a list that grows without end and shows the same
item twice when the server repeats it. `key` and `limit` turn the matched element into a bounded, keyed list. They
follow the rules of the `key` and `limit` modifiers of [`ypy-list`](list.md), and this extension does not need
`ypy-list` for them.

- `key` names an attribute (`id`, `data-id`; letters, digits, `_`, `.` and `-`; a bare `key` with no value means `id`).
  A new item whose attribute has the value of an item already in the list replaces that item where it stands. The same
  value twice in one patch keeps the last one. An item without the attribute is always added.
- `limit` is the most children the list keeps after the patch. The oldest go: the first ones for `append`, the last
  ones for `prepend`. What leaves is removed through the swap, so it is cleaned up like anything a swap removes.
- The key is applied before the limit, so a replaced item pushes nothing out.
- In any other mode they are ignored. A `key` that is not an attribute name, or a `limit` that is not a whole number
  of 1 or more, fires `ypy:stream:error` and the patch is not applied.

```html
<ul id="messages">
  <li data-id="7">hi</li>
</ul>
```

```
event: ypy-patch-elements
data: mode append
data: selector #messages
data: key data-id
data: limit 50
data: elements <li data-id="7">hi (edited)</li>
data: elements <li data-id="8">how are you?</li>

```

The first `<li>` is replaced where it stands, the second is added at the end, and when the list now has more than 50
items the first ones are removed. The Go side:

```go
ev, err := ypy.PatchElementsEvent(`<li data-id="8">how are you?</li>`,
	ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"),
	ypy.WithKey("data-id"), ypy.WithLimit(50))
```

`WithKey` and `WithLimit` are refused by the SDK with any mode but `append` and `prepend`. A client that reconnects
gets the snapshot (an `outer` morph of the whole list) and is back in line with the server.

## `ypy-patch-signals`

| Key | Value | Default |
|---|---|---|
| `signals` | a JSON object on one line: a JSON Merge Patch (RFC 7396) | required |
| `onlyIfMissing` | `true` | off |

```
event: ypy-patch-signals
data: onlyIfMissing true
data: signals {"theme":"light","page":1}
```

With the [`ypy-signals`](signals.md) extension loaded, the patch is applied with
`ypy.signals.patch(object, element, onlyIfMissing)` in the scopes around the streaming element, exactly like the body
of a JSON response: a key goes to the nearest scope that has it, `null` deletes, objects merge, local signals (`_`
prefix) may be named. With `onlyIfMissing true` the signals that exist keep their value at every level, and only the
absent ones are set; it is how a server sends defaults.

Without the signals extension the patch is not lost: the extension fires `ypy:stream:signals` on the streaming element
with `{patch, onlyIfMissing}`, for the page's own state code.

## `ypy-redirect`

| Key | Value | Default |
|---|---|---|
| `url` | where to go | required |
| `mode` | `navigate`: through YPY, pushing the URL; `replace`: the same, replacing the history entry; `hard`: a full page load | `navigate` |

`navigate` and `replace` issue a `GET` through the client (`ypy.ajax`, which swaps the answer into the page as
a boosted navigation does and updates the history). `hard` sets `location.href`. The stream is closed first and does
not reconnect.

The same origin rule as for any request applies: with `selfRequestsOnly` on (the default), a URL on another origin is
refused with a `ypy:stream:error`. A URL that is not `http:` or `https:` (`javascript:`, `data:`) is always refused.

## `ypy-event`

| Key | Value | Default |
|---|---|---|
| `name` | the event name | required |
| `selector` | the element to fire it on (the first match) | the document |
| `detail` | a JSON value on one line | none |

A DOM `CustomEvent` with that exact name (no `ypy:` prefix, like the names of `YPY-Trigger`), `detail` as its detail,
that bubbles and can be cancelled. A `selector` that matches nothing is reported like a patch's.

```
event: ypy-event
data: name message:new
data: detail {"id":42}
```

```js
document.addEventListener('message:new', e => notify(e.detail.id));
```

## `ypy-refresh`

| Key | Value | Default |
|---|---|---|
| `selector` | the elements to fetch again | none: the page |
| `debounce` | milliseconds, 0 to 60000 | 250 |
| `jitter` | milliseconds, 0 to 60000 | 0 |

The event carries no HTML. It says "something changed" and each client fetches its own copy, with its own session,
cookies and signals. Use it when a patch cannot be the same for everyone: a balance, a list filtered by permission, a
page with a CSRF token. For a change that looks the same to everyone, a patch is lighter.

```
event: ypy-refresh
data: selector #cart
data: debounce 250
data: jitter 1000

```

- Without `selector`, the client sends a `GET` to its own page (path and query) and morphs the answer into the element
  marked `ypy-history-elt`, or into the body when there is none: the element a history restore uses. Focus, selection,
  scroll and what the user typed survive.
- With `selector`, for every element that matches and has a `ypy-get`, the client makes that request, with that
  element as the source, and morphs the answer into the element's children (`innerMorph`). An element with no `ypy-get`
  fires `ypy:stream:error`; a selector that matches nothing fires `ypy:target:missing`.
- The request is an ordinary YPY request: `YPY-*` headers, current signals, CSRF token, `selfRequestsOnly`.
- Debounce: events with the same selector (the page counts as one) that come within `debounce` ms of the first make
  one request, when that window ends. The next event after the request starts a new window, so a change that arrives
  while the answer is on its way is not lost. Each selector has its own window.
- Jitter: a random wait between 0 and `jitter` ms is added after the debounce, so the many clients that got the same
  event do not ask at the same instant. A value above 60000 is 60000; one that is not a number is the default.
- Just before the request the client fires the cancelable `ypy:stream:refresh` on the streaming element, with
  `detail.selector`. `preventDefault()` skips that refresh. A refresh still waiting when the stream closes is dropped.

```html
<div id="room" ypy-stream="/room/events" ypy-history-elt>
  <div id="cart" ypy-get="/cart">...</div>
</div>
```

```go
// when an order changes, every connected page fetches its own cart within 2 seconds
room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))
```

`Refresh` is also a method of `ypy.Stream` for one client. The SDK always writes the `debounce` line, because a frame
with no `data` line is not delivered by an event stream reader. `WithRefreshDebounce` sets it (default 250 ms). A
server with many clients should use `jitter`.

## Reconnection

A stream ends when the network fails, a proxy cuts it, the server restarts or the server closes it. Unless a stop
condition holds, the client reconnects.

- The request is built again by the same code as the first (see [opening](#opening-a-stream)), with the current
  signals and, when the server sent an `id`, a `Last-Event-ID` header with the last one. An `id` line without data is
  remembered too; an empty `id` clears it.
- The delay starts at the last `retry` value the server sent (1 second without one) and doubles on every failed
  attempt, up to 30 seconds. Each wait is a random 75 to 100 percent of that, so a crowd of clients does not
  come back together. It starts over once a connection has delivered an event. A `retry` smaller than 10 ms is read as
  10 ms.
- Offline: while `navigator.onLine` is false the client waits for the `online` event and does not count that wait
  as failed attempts. The core fires `ypy:offline` and `ypy:online` as for any request.
- What the server sends on every connection, first or not, is the full state: a `ypy-patch-elements` in mode `outer`
  (or `inner` with a selector) for each region, with the complete HTML, and a `ypy-patch-signals` with the complete
  signals. The client applies them as morphs, which are idempotent and keep focus and typed text, so repeating the
  state is invisible when nothing changed and puts the page right when something was missed. The client keeps no
  record of what it missed. The Go `Broadcaster` does this with `WithSnapshot`.

The client does not reconnect when:

- the answer to a reconnection is `204` (the server's way to say "no more");
- the answer has any other status than `200`, or is not an event stream: it is an answer like any other, handled by
  the core (an error answer goes to the error handling of `ypy-error-*`), and not retried by this rule;
- a `ypy-redirect` was applied;
- the element that opened the stream left the page, or the page code closed the stream.

A request that gets no answer at all (the network is down) is retried with the delay above.

## Pausing in a background tab

A page in a hidden tab holds a connection and gets events nobody sees. With `pauseOnBackground` the extension closes the
connection while `document.hidden` is true and opens it again when the tab is shown. It is off by default.

```html
<div ypy-stream="/room/events" ypy-config='{"stream":{"pauseOnBackground":true}}'>...</div>
```

```js
ypy.config.stream = { pauseOnBackground: true };   // every stream of the page; an element can set it to false
```

- Hiding closes the connection and does not reconnect. It is not a failure: no attempt is counted and there is no
  wait. `ypy:stream:pause` fires on the element.
- Showing reconnects at once, with no delay and no backoff (the attempts start over). The request is built again,
  with the current signals and `Last-Event-ID` when the server sent an `id`. `ypy:stream:resume` fires, then the new
  connection fires `ypy:stream:open`.
- The server answers with the full state, as for any reconnection, so what happened meanwhile is applied. A server that
  numbers its events with `id` may send only what is new, if it can prove the result is the same.
- The stop conditions still hold while paused: an element removed, `ypy.stream.close()` and a `ypy-redirect` end the
  stream for good.

```
GET /room/events          (after the tab is shown again)
Accept: text/event-stream
Last-Event-ID: 12
```

## Silent connections and keepalive

A client cannot tell a quiet stream from a dead one, and a proxy or a sleeping phone can leave a connection that neither
fails nor ends. `ypy-stream` does not time the silence out itself: the server keeps the connection alive. The Go SDK
writes a comment line (`: keepalive`) when a stream has been silent for 15 seconds, so proxies and the browser keep it
open, and the decoder ignores comments.

```go
mux.Handle("GET /room/events", ypy.NewBroadcaster(
	ypy.WithStreamOptions(
		ypy.WithKeepAlive(20*time.Second),  // 0 or less turns the comment off; the default is 15s
		ypy.WithRetry(2*time.Second),       // the reconnection delay the client starts from
		ypy.WithWriteTimeout(10*time.Second),
	),
))
```

For one client, pass the same options to `ypy.NewStream(w, r, ...)`. A server that does not send comments should send
an event often enough for its proxies. The other event-stream and socket extensions have a client-side heartbeat:
`sse.heartbeat` of [`ypy-sse`](index.md) and `ws.heartbeat` of `ypy-ws` (a time, `0` by default) close and reopen a
connection that got no data for that long, and reopen it at once when the page comes back (`pageshow`, `online`, a
visible tab). Use them for elements that open a stream with `ypy-sse:connect` or a socket with `ypy-ws:connect`:

```html
<div ypy-sse:connect="/ticker" ypy-config='{"sse":{"heartbeat":"45s"}}'>...</div>
```

## Closing

A stream closes, with no reconnection, when its element leaves the page: removed from the DOM by anyone (the extension
watches for it), swapped out (the core cleanup calls it), or closed with `ypy.stream.close(element)`. The connection
is aborted, the timers stop and nothing stays registered: 200 streams opened and closed leave no connection, timer or
listener behind (`test/tests/ext/ypy-stream.js`). The browser closes the streams of a page that unloads.

## Routing by Content-Type

| Content-Type of the answer | What happens |
|---|---|
| `text/html` | swap or morph into the target (core) |
| `application/json` | a signals patch ([`ypy-signals`](signals.md)) |
| `text/event-stream`, status `200` | read as a stream by this extension |

The extension takes over in the core's `ypy:before:response` hook, before the body is consumed. It cancels the normal
processing of that answer, so nothing is swapped into the requesting element.

When `ypy-sse` is loaded too, note that both extensions answer to `text/event-stream`, and the first one that
registered takes the answer. Load `ypy-stream` before `ypy-sse`. An element with `ypy-sse:connect` (or the old `ypy-sse-connect`) is left
to `ypy-sse`; every other element's event streams are read as YPY streams. If you relied on a plain `ypy-get` whose
answer is an event stream being read by `ypy-sse` (unnamed messages swapped into the element), add `ypy-sse:connect`
to it, or do not load `ypy-stream`.

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.stream.close(element)` | Closes the stream of an element (reason `api`). It does not reconnect. |
| `ypy.stream.stats()` | `{open}`: the number of streams that are open, for leak tests and tooling. |

## Client events

All bubble from the streaming element and run the extension hooks of the core, like every `ypy:` event. `stream` is
the connection record (`{elt, m, url, id, retry, att}`; treat it as read-only).

| Event | Detail | When |
|---|---|---|
| `ypy:stream:open` | `{stream}` | a connection delivered its headers and is being read (first or after a reconnection) |
| `ypy:stream:reconnect` | `{stream, delay}` | the client is about to wait `delay` ms before the next attempt |
| `ypy:stream:pause` | `{stream}` | the tab went to the background and the stream, which has `pauseOnBackground`, closed its connection |
| `ypy:stream:resume` | `{stream}` | the tab is visible again and the stream opens its connection now, with `Last-Event-ID` |
| `ypy:stream:refresh` | `{stream, selector}` | a `ypy-refresh` is about to make its request; cancelable, `preventDefault()` skips it |
| `ypy:stream:close` | `{stream, reason}` | the stream ended for good: `api`, `removed`, `redirect`, `replaced`, `status <code>` or `error` |
| `ypy:stream:error` | `{stream, error}` | an event failed, or the read broke (the client reconnects after a read error) |
| `ypy:stream:signals` | `{patch, onlyIfMissing}` | a signals patch arrived and the signals extension is not loaded |
| `ypy:target:missing` | `{selector, type: "stream", stream}` | a patch or event had no element to go to; cancelable |

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `stream.pauseOnBackground` | `false` | close the connection while the tab is hidden and open it again when it is shown ([details](#pausing-in-a-background-tab)) |

`ypy.config.stream` sets it for the page, and `ypy-config='{"stream":{...}}'` on the streaming element overrides it for
that element.

## Security

- A patch passes through the same pipeline as a swapped response: scripts follow the nonce gate and Trusted Types
  (`ypy-csp`), and the `ypy` Trusted Types policy parses the HTML. If the CSP blocks the policy, patches without a
  selector and namespace patches fail with an error, and so does every swap.
- The request is a normal YPY request, so `selfRequestsOnly` (the fetch mode of the core) applies to the URL that
  opens the stream and the extension cannot lift it.
- A redirect to another origin or to a URL that is not `http(s)` is refused.
- A stream opened with GET must not change state; use `ypy-stream.post` with the application's CSRF protection for
  that. The signals travel in the query on GET: keep secrets out of them, or use `ypy-signals-send="false"` on the
  element (protocol section 7).
- The server must escape user data in `elements`, as for any HTML it sends.
- Keys that look like prototype members (`__proto__`, `constructor`) are plain data in the event decoder.

## Limits

- HTTP/1.1 allows six connections per host, so a page with a stream per widget runs out. Use HTTP/2 and one stream per
  page.
- The extension is 4,331 bytes min+gzip, against a budget of 6,144 bytes.
- Events are applied in order, one at a time. A server that sends faster than the page can morph is slowed by the
  browser's own flow control.
- A page with an open stream is generally not eligible for the browser's back-forward cache.

## Pitfalls

- A patch with no `selector` needs its top-level elements to have an `id` that exists on the page. Without it, or with a
  typo, nothing changes and only `ypy:target:missing` tells you.
- `append` without `key` adds the same item again every time the server repeats it. For a chat, a log or a feed, send
  `key` and `limit`.
- `ypy-refresh` with a `selector` needs a `ypy-get` on the matched element; the stream element itself does not need one.
- A refresh of the page needs an answer that contains the element marked `ypy-history-elt` (or a body).
- Keep `ypy-stream` before `ypy-sse` in the page, or an event stream is read by the wrong extension.
- A server that closes the stream with status `200` and no events makes the client reconnect forever. To stop it, answer
  `204` to the next request.
- With `pauseOnBackground`, a page in a hidden tab gets nothing, so do not use it for a stream whose events must be
  counted (a notification sound, a title badge).
