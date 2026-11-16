---
title: "Development mode: the ypy-dev extension"
description: "For development only: errors that name the element and request, a latency and offline simulator, warnings for markup that does nothing and an error overlay."
order: 100
---
# Development mode: the `ypy-dev` extension

`ypy-dev` gives you feedback while you build a page. It reports errors with the element and request that caused them,
simulates a slow or missing connection, warns about markup that does nothing and about old markup YPY no longer reads,
shows failed requests in an overlay, and reloads templates when they change.

> Use it in development only. It logs a lot, it can slow requests down on purpose, and it scans the markup of every
> swap. It is not part of `ypy.max`, so you load it with its own tag in the development layout. On a host that does not
> look like a development host (anything but `localhost`, `127.*`, `[::1]`, `*.local`, `*.test` and `*.localhost`) it
> warns once in the console that it is loaded.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-dev.min.js"></script>
```

When the page sets `config.extensions`, list `ypy-dev` in it, or it is refused. Only the public API and the extension
API are used, so it works with the built core. It is 5,953 bytes min+gzip, close to the 6,144-byte budget every extension has.

Contents: [Messages](#messages) · [Errors with context](#errors-with-context) · [Error overlay](#error-overlay) ·
[Requests cut by `ypy-sync`](#requests-cut-by-ypy-sync) · [Simulator](#latency-and-offline-simulator) ·
[Warnings for incorrect use](#warnings-for-incorrect-use) · [Legacy use](#warnings-for-legacy-use) ·
[Reload of templates](#reload-of-templates) · [Options](#options) · [JavaScript API](#javascript-api)

## Turning it on

Load the script only in the development layout, so a production build never carries it. A Go template, for example:

```html
{{ if .Dev }}<script src="/static/ypy-dev.min.js"></script>{{ end }}
```

Loaded, it prints one line at start, `ypy-dev: development helpers are on. Do not ship ypy-dev to production.`, and
starts what `ypy.config.dev` asks for (latency, offline, reload). `dev:false` in the configuration turns every
check off while the file stays loaded.

## Messages

Every message goes to three places:

- the console, as `ypy-dev: <text>`, with an object that has the element and the data of the case (warning
  level for a hint, error level for a failure, info level for the reload);
- `ypy.dev.messages`, which keeps the last 200;
- the `ypy:dev:message` event, fired on the element (on `document` when the element is gone), so a test or the
  [inspector](inspector.md) can read it. The detail is `{level, code, text, elt, ...}`: `code` is a stable name (the tables
  below), `text` is the sentence, and the other keys depend on the code (`ctx`, `selector`, `status`, `attribute`,
  `suggestion`, `guide`).

```js
document.addEventListener('ypy:dev:message', e => {
  if (e.detail.code === 'unknown-attribute') fail(e.detail.text)   // in an end-to-end test of a template
})
```

A warning about the markup is given once for each element and attribute; a swap that brings the same markup again
reports it again, because it is a new element.

## Errors with context

The core already logs the failure. `ypy-dev` adds what you need to find the cause: the request, the element that
sent it, the target and the start of the response.

| Code | When | What the message says |
|---|---|---|
| `target-missing` | an out-of-band element, a `<ypy-partial>` or a `ypy-target` has no element to go into | the selector, who asked for it (`an out-of-band element of the response`, `a <ypy-partial> of the response` or `ypy-target of <button ...>`), the request and its status, the first 120 characters of the response |
| `unknown-swap` | a swap style that does not exist | the style, the closest real style (`did you mean "beforeend"?`) and the list of styles |
| `empty-response` | a successful response with an empty body (`204` and `304` are fine) | the request, the element, the target that gets nothing, and "answer 204 to skip the swap" |
| `status-404` ... `status-599` | a response with status 400 or more (`warn` for 4xx, `error` for 5xx) | the request, the element, the target and swap the error goes to, a hint for 404, 422 and 5xx, the start of the response |
| `status-401`, `status-403`, `status-419` | the server refused the request for who is asking | the likely cause: `401` an expired session or no sign-in, `403` no permission or a failed CSRF check, `419` an expired session or a missing or stale CSRF token, with what to check. The message has `auth: true` |
| `sync-aborted`, `sync-dropped` | a `POST`, `PUT`, `PATCH` or `DELETE` that `ypy-sync` cut (see below) | the request, the element and the strategy |
| `timeout` | the request ran past its limit | the request and the limit in milliseconds, and where to change it (`config.defaultTimeout`, `ypy-config="timeout:..."`) |
| `network` | `fetch` failed and the browser is online | the request and the browser's reason; the usual causes (server down, wrong URL, cross-origin while `config.mode` is `same-origin`) |
| `offline` | `fetch` failed and the browser is offline (or [simulated](#latency-and-offline-simulator)) | the request |
| `error` | any other error | the message, the source element and the target |
| `blocked` | author code (`ypy-on`, `js:` values, `[filter]`) did not run because of `config.callbacks` | the element, the reason (`callbacks-none`) and the start of the code |

The unknown swap style and the other errors of a swap are reported once, not once for the swap and again for the
request.

Example, a `404` from a button:

```
ypy-dev: GET /gone (from <button ypy-get="/gone" ypy-target="#out">) answered 404; target <div#out>, swap innerHTML.
  Hint: check the URL and the route. Response starts: "<h1>Not here</h1>"
```

## Error overlay

A request that fails is easy to miss in the console. With the overlay on, a panel appears over the page with what you need
to start debugging:

- the status and its text (`500 Internal Server Error`), or `Request failed (network)` when no response came;
- the method and the URL;
- the element that sent the request, like `<button#save ypy-post="/save">`;
- the cause and what to check (the same sentence as the console message, so a `401` or `419` says "session expired or CSRF");
- the first 600 characters of the response body, as plain text.

It shows the last failure, one panel at a time. Press `Esc` or the Close button to close it; the next failure opens it
again. It appears for a response with status 400 or more, a network failure, an offline failure and a timeout. It does not
change the page and does not take the focus.

By default it is on only on a development host (the same hosts that `ypy-dev` accepts without a warning: `localhost`,
`127.*`, `[::1]`, `*.local`, `*.test`, `*.localhost`). Set `dev.overlay:true` to force it on, or `dev.overlay:false` to
keep it off; the console message and `ypy:dev:message` do not depend on it.

```html
<meta name="ypy-config" content="dev.overlay:false">
```

Like the [inspector](inspector.md), it is a Shadow DOM on `<html>`, outside `<body>`, so a swap never removes it. The
style is an adopted style sheet, not a `style` attribute or a `<style>` tag, and the texts are set with `textContent`,
never as markup, so a response that contains HTML or script is shown as text. A strict CSP and Trusted Types are fine.
The element has the id `ypy-dev-overlay` and `ypy-ignore`. `ypy.dev.overlay.close()` closes it from a script.

## Requests cut by `ypy-sync`

[`ypy-sync`](../reference/attributes/ypy-sync.md) decides what happens when a second request starts while one is running. For a read
(`GET`) that is harmless. For a `POST`, `PUT`, `PATCH` or `DELETE` it is not: an aborted request may already have been
processed by the server, and a dropped one is an action of the user that was never sent. `ypy-dev` listens to
`ypy:request:aborted` and `ypy:request:dropped` and warns for these four methods:

| Code | When | Strategy in the message |
|---|---|---|
| `sync-aborted` | the request was in flight and a newer request replaced it (`replace`) or displaced an `abort` request | `replace` or `abort` |
| `sync-dropped` | the request was never sent: the line was busy with `drop` or `abort`, or the waiting line was full, or a `replace` cleared it | `drop`, `abort`, `replace` or `queue-overflow` |

```
ypy-dev: POST /save (from <button#save ypy-post="/save">) was aborted in flight by ypy-sync replace: the server may
  have run it. Use "queue all" for a request that must complete
```

A timeout, an abort that the page asked for (`ypy:abort`), a cleanup of the element and a request that waits in the queue
do not warn. The message has `{ctx, strategy, reason, method}`.

## Latency and offline simulator

Test a slow or a missing connection without leaving the page.

```js
ypy.dev.latency(300)        // every request waits 300 ms before it is sent
ypy.dev.latency(100, 800)   // a random wait between 100 and 800 ms for each request
ypy.dev.latency(0)          // off

ypy.dev.offline(true)       // the browser is offline: requests fail, the page enters the offline state
ypy.dev.offline(false)      // back online
```

Both can also be set at start with the configuration tag: `<meta name="ypy-config" content="dev.latency:300,dev.offline:true">`.
A range needs the JSON form of the tag: `<meta name="ypy-config" content='{"dev":{"latency":[100,800]}}'>`.

- Latency waits inside the request, after `ypy:before:request` and before `fetch`. The wait counts against the
  request timeout (a `timeout:` shorter than the latency ends in a `timeout` error) and an aborted request stops
  waiting at once. Indicators, `ypy-sync`, `ypy-pending` and `ypy-busy` behave as with a slow server.
- Offline makes `fetch` fail with a `TypeError` as a dead network does, without sending anything. It also makes
  `navigator.onLine` report `false` and fires the window `offline` event, so `<html>` gets the
  `ypy-offline` class and `data-ypy-state="offline"` and the `ypy:offline` event fires (see
  [offline state](../reference/events/ypy-offline.md)); the requests that fail carry `ctx.errorKind === "offline"`.
  `offline(false)` removes the override and fires `online`.
- With [`ypy-retry`](retry.md) every attempt goes through the simulator, whichever extension was
  loaded first.
- The simulator wraps requests made by YPY (`ctx.fetch`). It does not touch other `fetch` calls, `EventSource`,
  `WebSocket`, or the connections of `ypy-sse` and `ypy-ws`. For those, use the browser's own network throttling.

## Warnings for incorrect use

Run on every element the core processes (the page at load and each swapped fragment). They look at the markup; they
change nothing. Elements under `ypy-ignore` are skipped.

| Code | What it catches | Example of the message |
|---|---|---|
| `unknown-attribute` | a `ypy-` or `data-ypy-` attribute that nothing reads, close to one that exists | `unknown attribute "ypy-gett" on <button>; did you mean "ypy-get"?` |
| `implicit-inheritance` | an attribute on a container that its descendants will not inherit | `"ypy-target" on <div#box ...> does not reach <button ypy-get="/x"> (2 element(s) below): an attribute is inherited only with "ypy-target:inherited"` |
| `target-no-match` | a `ypy-target` that matches nothing right now | `ypy-target "#nope" on <button ...> matches nothing right now (fine if the element exists before the request is sent)` |
| `target-invalid` | a `ypy-target` that is not a selector | |
| `callbacks-none` | `ypy-on:...` while `config.callbacks` is `"none"` (the default) and `ypy-signals` is not loaded | `"ypy-on:click" on <button ...> will not run: config.callbacks is "none" ...` |
| `prefix-conflict` | an element with `ypy-x` and `data-ypy-x` and different values | `<button ...> has "ypy-get" and "data-ypy-get" with different values; "ypy-get" wins` |

How the unknown attribute is decided: the name without its prefix and its `:modifier` or `.modifier`
(`ypy-status:404` is `status`, `ypy-signals.local` is `signals`) is compared with the list of attributes YPY and the
official extensions read. It is reported only when it is one or two edits away from a name on that list (one edit
for a name of up to four letters), so an attribute of your own, or one of an extension this version does not know,
stays quiet. Register such a name with `ypy.dev.known('layers')` to tell it from a typo.

`ypy-target` is checked at processing time, so a target that a later swap creates is reported as "right now"; the
`target-missing` error above is the one that matters when a request really fails.

## Warnings for legacy use

Old markup can carry attributes and event names that YPY no longer reads. `ypy-dev` points at each
one it sees, says what to write instead, and links the page of the [upgrade guide](../migration/upgrade-check.md) that
explains the change. `ypy migrate` converts them; see the [migration guide](../migration/index.md).

| Code | What it catches |
|---|---|
| `removed-attribute` | an attribute that YPY does not read at all: `ypy-vars`, `ypy-params`, `ypy-request`, `ypy-disabled-elt`, `ypy-disinherit`, `ypy-inherit`, `ypy-ext`, `ypy-sse`, `ypy-ws`, `ypy-sse-connect`, `ypy-sse-close`, `ypy-sse-swap`, `ypy-ws-connect`, `ypy-ws-send` |
| `legacy-event` | an event written in kebab-case, `ypy-on::after-swap` or `ypy-on:ypy:after-swap`; the current name is `after:swap` |

```
ypy-dev: removed attribute "ypy-vars" on <button ypy-get="/x">: use ypy-vals with the js: prefix, for example
  ypy-vals="js:{a: 1}". Guide: https://ypy.dev/docs/guides/upgrade-check#attributes
```

The link base is `ypy.config.dev.docs` (default `https://ypy.dev/docs/guides/upgrade-check`); the anchor is the section of
the guide (`#attributes`, `#events`, `#extension-attributes-swap-styles-and-event-sources`, `#behavior-changes`).
`ypy check` finds the same constructs in your files; `ypy-dev` finds the ones that are in the page you are running.

## Reload of templates

When a template changes on the server, the page refreshes itself and keeps its state (typed text, focus, scroll,
open dialogs) because the new page is morphed into the old one. Turn it on with the endpoint of your development
server:

```html
<meta name="ypy-config" content='dev.reload:"/__ypy/reload"'>
```

or `dev.reload:true` for the default path `/__ypy/reload`, or `ypy.dev.reload('/__ypy/reload')` from a script;
`ypy.dev.reload(false)` closes it. It is off unless you ask, so a page without the endpoint does not retry a `404`
forever.

### The contract of the endpoint

The server side belongs to your development server; any server can do it, and the handler below is enough. It is a plain
[Server-Sent Events](https://html.spec.whatwg.org/multipage/server-sent-events.html) stream.

| | |
|---|---|
| Request | `GET /__ypy/reload` (any same-origin path; the page is told which). It has no body and no authentication, so serve it only in development. |
| Response | status `200`, `Content-Type: text/event-stream`, `Cache-Control: no-cache`, kept open |
| `retry: 1000` | recommended in the first message, so the browser reconnects quickly after a restart |
| `: ping` | a comment line every 15 seconds, to keep proxies from closing the stream |
| `event: templates` | a template changed. `data:` is optional (a JSON object with `paths`, for logs). The page fetches itself again and morphs the body. Events that arrive within 50 ms count once |
| `event: reload` | a change that needs a full load (a script, a stylesheet, the layout). The page calls `location.reload()` |

```
retry: 1000

event: templates
data: {"paths":["views/orders.html"]}

event: reload
data: {"paths":["static/app.js"]}
```

A minimal handler with the Go standard library only (the real one adds the file watcher):

```go
func reload(changes <-chan string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        f := w.(http.Flusher)
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        fmt.Fprint(w, "retry: 1000\n\n")
        f.Flush()
        tick := time.NewTicker(15 * time.Second)
        defer tick.Stop()
        for {
            select {
            case <-r.Context().Done():
                return
            case <-tick.C:
                fmt.Fprint(w, ": ping\n\n")
            case path := <-changes:
                fmt.Fprintf(w, "event: templates\ndata: {\"paths\":[%q]}\n\n", path)
            }
            f.Flush()
        }
    }
}
```

### What the page does

- On `templates` it sends `GET <current path and query>` with `ypy.ajax`, `target:'body'` and `swap:'innerMorph'`:
  a normal YPY request (`YPY-Request`, `YPY-Request-Type: full`), so the server answers the full page as it does for a
  boosted link. The title is updated like any swap.
- If the page could not be swapped (the server is rebuilding, or it answered with an error) it tries again every
  400 ms, five times, then loads the page fully with `location.reload()`, so a template error page shows up.
- If the stream drops and comes back, the server restarted: the page refreshes once. The first connection never does.
- Each action is logged at info level (`reload`, `reload-on`).

## Options

`ypy.config.dev`, from the configuration tag (`<meta name="ypy-config" content="dev.latency:300,dev.reload:true">`) or
from a script before the extension loads.

| Key | Value | Default |
|---|---|---|
| `latency` | milliseconds, or `[min, max]` | off |
| `offline` | `true` starts offline | off |
| `reload` | `true` (the path `/__ypy/reload`) or a URL | off |
| `lint` | `false` turns the markup and legacy warnings off | on |
| `log` | `false` keeps the console quiet; messages are still recorded and fired | on |
| `overlay` | `true` or `false`: the [error overlay](#error-overlay) | on for a development host, off elsewhere |
| `docs` | base URL of the guide pages in legacy warnings | `https://ypy.dev/docs/guides/upgrade-check` |

`dev:false` (the whole value) turns the extension off.

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.dev.latency(min, max = min)` | Sets the simulated latency; `0` turns it off |
| `ypy.dev.offline(on = true)` | Starts or ends the simulated offline state |
| `ypy.dev.reload(urlOrBoolean)` | Starts the reload listener (`true` for the default path), or closes it with `false` |
| `ypy.dev.known(...names)` | Adds attribute names (without `ypy-`) that are not typos |
| `ypy.dev.messages` | The messages so far (the last 200) |
| `ypy.dev.clear()` | Empties `ypy.dev.messages` |
| `ypy.dev.simulation` | `{min, max, offline}`, the state of the simulator |
| `ypy.dev.overlay.close()` | Closes the error overlay |

Event: `ypy:dev:message`, with `{level, code, text, elt, ...}`.
