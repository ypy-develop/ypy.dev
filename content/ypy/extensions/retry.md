---
title: "ypy-retry"
description: "Repeats a request that could not reach the server, and with on:error also one answered with a 5xx. Off until the extension is loaded."
order: 350
---
# `ypy-retry`

A request that fails on a bad connection often works a second later. `ypy-retry` repeats it: by default when the
request could not reach the server, and with `on:error` also when the server answered with a 5xx. It is off until you
load the official [`ypy-retry` extension](index.md#ypy-retry) (`ext/ypy-retry.js` after the core; `ypy.max` already
includes it). Without the extension the attribute does nothing and breaks nothing; the core only classifies the failure
(`ctx.errorKind`) and tracks the offline state of the page.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-retry.js"></script>
```

If you set `ypy.config.extensions`, list `ypy-retry` in it, or the extension is refused.

```html
<button ypy-get="/inbox" ypy-retry="attempts:3 delay:1s backoff:2">Refresh</button>
```

The attribute is read as HCON, like `ypy-swap` and `ypy-trigger`. `ypy-retry` and `data-ypy-retry` are the same
attribute. It is not inherited unless you write `ypy-retry:inherited` on the container:

```html
<main ypy-retry:inherited="attempts:2 delay:500ms">
  <button ypy-get="/a">A</button>   <!-- retried -->
  <button ypy-get="/b" ypy-retry="false">B</button>   <!-- not retried -->
</main>
```

## Options

| Option | Default | Meaning |
|---|---|---|
| `attempts` | `0`, or `3` for a bare `ypy-retry` | how many times a failed request is repeated |
| `on` | `network` | what counts as a failure: `network`, `error`, `always` or `never` (see [What is repeated](#what-is-repeated)) |
| `delay` | `1s` | wait before the first repeat (`1000`, `500ms`, `1s`) |
| `backoff` | `2` | each wait is this many times longer than the one before; `1` keeps it constant |
| `jitter` | `0` | each wait varies at random by up to this fraction (`0.2`) or percentage (`20%`), up and down |
| `max` | `30s` | the longest wait between two attempts |
| `methods` | `GET HEAD QUERY` | narrows the methods that are repeated; a space or comma separated string or an array |
| `unsafe` | `false` | repeat every method, including `POST`, `PUT`, `PATCH` and `DELETE` |

`ypy-retry="false"`, `ypy-retry="on:never"` and `ypy-retry="attempts:0"` turn it off for an element. The same options are the defaults for
the whole page in `ypy.config.retry`, and the attribute overrides them:

```js
ypy.config.retry = {attempts: 2, delay: '500ms'}
```

With `delay:1s backoff:2` the waits are 1 s, 2 s, 4 s, ... up to `max`.

## Jitter

Many clients that fail together and retry on the same schedule hit the server together again. `jitter` spreads them: each
wait is multiplied by a random number between `1 - jitter` and `1 + jitter`, and then capped by `max`. `jitter:0.2` (or
`jitter:20%`) makes a 1 s wait last between 0.8 s and 1.2 s; `jitter:0` keeps the waits exact. The value is a fraction
from 0 to 1 or a percentage; it is `0` unless you set it, per element or for the page:

```js
ypy.config.retry = {attempts: 3, on: 'error', jitter: 0.2}
```

`delay` in the `ypy:request:retry` event is the wait that was chosen, with the jitter applied.

## What is repeated

`on` says what counts as a failure:

| `on` | Repeated |
|---|---|
| `network` (default) | a failure to get a response: `ctx.errorKind` is `offline` or `network` (see [`ypy:offline`](../reference/events/ypy-offline.md)) |
| `error` | the same, and a response with status 500 or more |
| `always` | the same, and a response with status 400 or more |
| `never` | nothing: the same as `false` |

An unknown value is read as `network`. Never repeated, in any mode: a timeout and an abort. A response that is not an
error (status below 400) is never repeated either.

```html
<div ypy-get="/feed" ypy-retry="attempts:3 on:error">...</div>
```

When the attempts run out, the last response is not changed: `ypy:response:error` and the error flow of the core run as
without `ypy-retry`, once, for that last response. The answers that were repeated are discarded without being read, and
fire no event of their own. `ypy:request:retry` has `status`, the status of the answer that is repeated; it is absent
after a network failure.

`on:always` repeats a `404` or a `422` too, which rarely helps: use it for an endpoint that answers `4xx` while it
warms up, not as a general setting.

### The modes at a glance

| What fails | `network` (default) | `error` | `always` | `never` |
|---|---|---|---|---|
| no response (offline, server unreachable, a `fetch` that rejected) | repeated | repeated | repeated | no |
| `5xx` | no | repeated | repeated | no |
| `4xx` (`404`, `422`, `429`...) | no | no | repeated | no |
| `2xx` and `3xx` | no | no | no | no |
| a timeout or an abort | no | no | no | no |

A request is repeated only when the mode says so and `attempts` is not used up and the method is allowed (see
[Methods](#methods)). `attempts` counts repeats: `attempts:3` is at most four requests. The mode can be set for the
page (`ypy.config.retry = {attempts: 2, on: 'error'}`) and changed per element; without `ypy-retry` on the element,
nothing is repeated, whatever the page default says for `on`.

One example for each mode, and what the server does.

```html
<!-- network: the server was unreachable; a 503 is shown at once -->
<div ypy-get="/feed" ypy-retry="attempts:3">...</div>

<!-- error: also a 500, 502, 503... -->
<div ypy-get="/report" ypy-retry="attempts:4 delay:2s on:error">...</div>

<!-- always: the endpoint answers 404 until a job is ready -->
<div ypy-get="/jobs/42/result" ypy-retry="attempts:5 delay:1s backoff:1 on:always">...</div>

<!-- never: off for this element, even inside an ypy-retry:inherited container -->
<button ypy-post="/pay" ypy-retry="on:never">Pay</button>
```

The server asks for a pause with `Retry-After` (see [Retry-After](#retry-after)). A Go handler that is overloaded
answers like this:

```go
func Report(w http.ResponseWriter, r *http.Request) {
	if overloaded() {
		w.Header().Set("Retry-After", "3") // seconds, or an HTTP date
		http.Error(w, "busy", http.StatusServiceUnavailable)
		return
	}
	ypy.Fragment(w, renderReport(r))
}
```

```http
HTTP/1.1 503 Service Unavailable
Retry-After: 3
```

With `on:error` the client waits at least 3 seconds (more when the backoff is longer) and asks again; with the default
`on:network` the `503` goes to the page as it is.

### Retry-After

A `503` or a `429` that is repeated waits at least as long as its `Retry-After` header says, in seconds or as an HTTP
date. The wait is the longer of the backoff and the header, and jitter never shortens the header. When the header asks for
more than `max`, the request is not repeated and the answer goes to the page as it is, because repeating earlier than the
server asked would only be refused again. A date in the past, or a value that is not valid, is ignored. A `429` is
repeated only with `on:always`; a `503`, with `on:error` or `on:always`. For a response from another origin the browser
shows `Retry-After` only if the server lists it in `Access-Control-Expose-Headers`.

### Methods

A method that is not idempotent is not repeated, whatever `on` says, because the first request may have reached the
server and run: a repeated `POST` can create a second order, also after a `500`. Opt in with `unsafe:true` only
where the server handles duplicates (an idempotency key, for example):

```html
<form ypy-post="/orders" ypy-retry="attempts:2 on:error unsafe:true">...</form>
```

`unsafe` is also needed when `methods` lists a method that is not safe; the list can only narrow the safe ones. With
`unsafe:true` every method is repeated and `methods` is not read.

## How the wait works

The extension wraps `ctx.fetch` (from its `ypy_before_request` hook) with the loop, and the core calls it once, so the
repeat happens inside the same request, not as a new one:

- the request keeps its slot in the queue (`ypy-sync`), so a second click waits behind it;
- `ypy:before:request` and `ypy:finally:request` fire once; `ypy:request:retry` fires before each repeat with
  `{attempt, delay, ctx, status}` (`attempt` is 1 for the first repeat) and can be cancelled with `preventDefault()`;
- the wait ends early when the window fires `online`, otherwise after `delay`;
- the request timeout is the deadline for all attempts together, and an abort ends the wait at once;
- if the element is removed from the page (swapped out, cleaned up) the request is aborted quietly, with no
  timer or listener left behind;
- `ypy:error` fires once, after the last attempt, with `ctx.errorKind` set, when the failure was a network failure. After a
  repeated 5xx the last answer is handled like any response with that status. A repeat that works fires no error.

The same body is sent again; a request with a stream as its body cannot be repeated.

## Pitfalls

- A repeated `POST` can run twice. Leave `unsafe` off unless the server de-duplicates (an idempotency key).
- `on:always` repeats a `422` and a `404`, which will not change. Use it only for endpoints that answer `4xx` while they
  warm up.
- A `Retry-After` longer than `max` stops the retry and shows the answer; raise `max` when the server asks for long
  pauses.
- The request timeout covers all attempts together, so `attempts:3` with `delay:10s` needs a timeout longer than the
  waits.
- A request with a stream as its body, or one that is aborted, is not repeated.

## Example with Tailwind CSS

```html
<div ypy-get="/feed" ypy-trigger="load, every 30s" ypy-retry="attempts:3"
     class="[.ypy-offline_&]:opacity-60">
  ...
</div>
```
