---
title: "Development mode: the ypy-dev extension"
description: "For development only: errors that name the element and request, a latency and offline simulator, warnings for markup that does nothing, an error overlay."
order: 100
---
# Development mode: the `ypy-dev` extension

`ypy-dev` gives you feedback while you build a page. It reports errors with the element and request that caused them,
simulates a slow or missing connection, warns about markup that does nothing and about old markup YPY no longer reads,
shows failed requests in an overlay, and reloads templates when they change.

> Use it in development only. It logs a lot, it can slow requests down on purpose and it scans the markup of every swap.
> It is not part of `ypy.max`. On a host that does not look like a development host (anything but `localhost`, `127.*`,
> `[::1]`, `*.local`, `*.test` and `*.localhost`) it warns once in the console that it is loaded.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-dev.min.js"></script>
```

Load it only in the development layout, so a production build never carries it. When the page sets `config.extensions`,
list `ypy-dev` in it. `dev:false` in the configuration turns every check off while the file stays loaded.

## Messages

Every message goes to the console as `ypy-dev: <text>`, to `ypy.dev.messages` (the last 200) and to the `ypy:dev:message`
event, fired on the element (on `document` when the element is gone). The detail is `{level, code, text, elt, ...}`, where
`code` is a stable name, so a test or the [inspector](inspector.md) can read it.

```js
document.addEventListener('ypy:dev:message', e => {
  if (e.detail.code === 'unknown-attribute') fail(e.detail.text)
})
```

A warning about markup is given once for each element and attribute.

## Errors with context

The core already logs the failure. `ypy-dev` adds the request, the element that sent it, the target and the start of the
response.

| Code | When |
|---|---|
| `target-missing` | an out-of-band element, a `<ypy-partial>` or a `ypy-target` has no element to go into |
| `unknown-swap` | a swap style that does not exist; the message suggests the closest real one |
| `empty-response` | a successful response with an empty body (`204` and `304` are fine) |
| `status-404` to `status-599` | a response with status 400 or more, with a hint for 404, 422 and 5xx |
| `status-401`, `status-403`, `status-419` | the server refused who is asking: an expired session, no permission, or a missing or stale CSRF token |
| `sync-aborted`, `sync-dropped` | a `POST`, `PUT`, `PATCH` or `DELETE` that `ypy-sync` cut |
| `timeout`, `network`, `offline` | the request ran past its limit, failed online, or failed offline |
| `error` | any other error |
| `blocked` | author code (`ypy-on`, `js:` values, `[filter]`) did not run because of `config.callbacks` |

```
ypy-dev: GET /gone (from <button ypy-get="/gone" ypy-target="#out">) answered 404; target <div#out>, swap innerHTML.
  Hint: check the URL and the route. Response starts: "<h1>Not here</h1>"
```

[`ypy-sync`](../index.md#synchronization) is harmless for a GET. For a write, an aborted request may already have been
processed by the server and a dropped one is an action that never left, so these two codes warn for the four write
methods. A timeout, a `ypy:abort`, a cleanup and a request waiting in the queue do not warn.

## Error overlay

A request that fails is easy to miss in the console. With the overlay on, a panel appears over the page with the status
and its text (or `Request failed (network)`), the method and URL, the element that sent the request, the cause and what to
check, and the first 600 characters of the body as plain text. It shows the last failure. `Esc` or the Close button
closes it. It is on by default only on a development host. `dev.overlay:true` forces it on, `dev.overlay:false` keeps it
off, and `ypy.dev.overlay.close()` closes it from a script.

It is a Shadow DOM on `<html>`, so a swap never removes it. The style is an adopted style sheet and the texts are set
with `textContent`, so a strict CSP and Trusted Types are fine and a response with HTML is shown as text.

## Latency and offline simulator

```js
ypy.dev.latency(300)        // every request waits 300 ms before it is sent
ypy.dev.latency(100, 800)   // a random wait between 100 and 800 ms
ypy.dev.latency(0)          // off

ypy.dev.offline(true)       // requests fail and the page enters the offline state
ypy.dev.offline(false)
```

Set them at start with `<meta name="ypy-config" content="dev.latency:300,dev.offline:true">`. A range needs the JSON form
of the tag: `content='{"dev":{"latency":[100,800]}}'`.

Latency waits inside the request and counts against its timeout. Offline makes `fetch` fail as a dead network does,
makes `navigator.onLine` report `false`, and gives `<html>` the `ypy-offline` class and `data-ypy-state="offline"`. With
[`ypy-retry`](retry.md) every attempt goes through the simulator. Only requests made by YPY are simulated, so use the
browser's network throttling for `EventSource`, `WebSocket` and other `fetch` calls.

## Warnings for incorrect use

These run on every element the core processes, change nothing, and skip anything under `ypy-ignore`.

| Code | What it catches |
|---|---|
| `unknown-attribute` | a `ypy-` or `data-ypy-` attribute that nothing reads, close to one that exists (`ypy-gett`) |
| `implicit-inheritance` | an attribute on a container that its descendants will not inherit without `:inherited` |
| `target-no-match`, `target-invalid` | a `ypy-target` that matches nothing right now, or is not a selector |
| `callbacks-none` | `ypy-on:...` while `config.callbacks` is `"none"` and `ypy-signals` is not loaded |
| `prefix-conflict` | `ypy-x` and `data-ypy-x` with different values on one element |

An attribute is reported as unknown only when it is one or two edits away from a name YPY or an official extension reads
(one edit for a name of up to four letters), so an attribute of your own stays quiet. Register a name that is not a typo
with `ypy.dev.known('layers')`.

## Warnings for legacy use

Old markup can carry attributes and event names YPY no longer reads. `ypy-dev` names the replacement and links the
[migration page](../migration.md). `ypy migrate` converts them in your files; `ypy-dev` finds the ones in the page you are
running.

| Code | What it catches |
|---|---|
| `removed-attribute` | attributes YPY does not read at all, from the older library (the [migration page](../migration.md) lists them) |
| `legacy-event` | an event written in kebab-case, such as `ypy-on::after-swap`; the current name is `after:swap` |

## Reload of templates

When a template changes on the server, the page refreshes itself and keeps its state (typed text, focus, scroll, open
dialogs), because the new page is morphed into the old one. Turn it on with the endpoint of your development server:

```html
<meta name="ypy-config" content='dev.reload:"/__ypy/reload"'>
```

`dev.reload:true` uses the path `/__ypy/reload`, and `ypy.dev.reload(false)` closes the listener. It is off unless you ask.

The endpoint belongs to your development server, in any language. It is a plain server-sent events stream and has no
authentication, so serve it only in development.

| | |
|---|---|
| Response | status `200`, `Content-Type: text/event-stream`, `Cache-Control: no-cache`, kept open |
| `retry: 1000` | recommended first, so the browser reconnects quickly after a restart |
| `: ping` | a comment line every 15 seconds, so proxies keep the stream |
| `event: templates` | a template changed. `data:` is optional (a JSON object with `paths`). The page fetches itself again and morphs the body. Events within 50 ms count once |
| `event: reload` | a script, stylesheet or layout changed. The page calls `location.reload()` |

```
retry: 1000

event: templates
data: {"paths":["views/orders.html"]}
```

On `templates` the page sends a normal YPY request for its own path with `target:'body'` and `swap:'innerMorph'`, so the
server answers the full page as for a boosted link. If it cannot be swapped, the page tries again every 400 ms, five
times, then reloads fully, so a template error page shows up. A stream that drops and comes back means the server
restarted, so the page refreshes once.

## Options

`ypy.config.dev`, from the configuration tag or from a script before the extension loads.

| Key | Value | Default |
|---|---|---|
| `latency` | milliseconds, or `[min, max]` | off |
| `offline` | `true` starts offline | off |
| `reload` | `true` or a URL | off |
| `lint` | `false` turns the markup and legacy warnings off | on |
| `log` | `false` keeps the console quiet; messages are still recorded and fired | on |
| `overlay` | `true` or `false` | on for a development host |
| `docs` | base URL of the migration page used in legacy warnings | the YPY migration page |

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.dev.latency(min, max = min)` | Sets the simulated latency; `0` turns it off |
| `ypy.dev.offline(on = true)` | Starts or ends the simulated offline state |
| `ypy.dev.reload(urlOrBoolean)` | Starts the reload listener, or closes it with `false` |
| `ypy.dev.known(...names)` | Adds attribute names (without `ypy-`) that are not typos |
| `ypy.dev.messages`, `ypy.dev.clear()` | The messages so far, and emptying them |
| `ypy.dev.simulation` | `{min, max, offline}` |
| `ypy.dev.overlay.close()` | Closes the error overlay |

The extension is 5,953 bytes min+gzip, close to its 6,144-byte budget.
