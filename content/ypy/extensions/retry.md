---
title: "ypy-retry"
description: "Repeats a request that could not reach the server, and with on:error also one answered with a 5xx. Off until the extension is loaded."
order: 350
---
# `ypy-retry`

A request that fails on a bad connection often works a second later. `ypy-retry` repeats it: by default when the request
could not reach the server, and with `on:error` also when the server answered with a 5xx. Without the extension the
attribute does nothing and breaks nothing. The core only classifies the failure (`ctx.errorKind`) and tracks the offline
state of the page.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-retry.min.js"></script>

<button ypy-get="/inbox" ypy-retry="attempts:3 delay:1s backoff:2">Refresh</button>
```

The attribute is not inherited unless you write `ypy-retry:inherited` on a container:

```html
<main ypy-retry:inherited="attempts:2 delay:500ms">
  <button ypy-get="/a">A</button>
  <button ypy-get="/b" ypy-retry="false">B</button>
</main>
```

## Options

| Option | Default | Meaning |
|---|---|---|
| `attempts` | `0`, or `3` for a bare `ypy-retry` | how many times a failed request is repeated |
| `on` | `network` | what counts as a failure: `network`, `error`, `always` or `never` |
| `delay` | `1s` | wait before the first repeat (`1000`, `500ms`, `1s`) |
| `backoff` | `2` | each wait is this many times longer than the one before; `1` keeps it constant |
| `jitter` | `0` | each wait varies at random by up to this fraction (`0.2`) or percentage (`20%`) |
| `max` | `30s` | the longest wait between two attempts |
| `methods` | `GET HEAD QUERY` | narrows the methods that are repeated |
| `unsafe` | `false` | repeat every method, including `POST`, `PUT`, `PATCH` and `DELETE` |

`ypy-retry="false"`, `on:never` and `attempts:0` turn it off for an element. `ypy.config.retry` sets the same options for the
whole page, and the attribute overrides it:

```js
ypy.config.retry = {attempts: 2, delay: '500ms', jitter: 0.2}
```

With `delay:1s backoff:2` the waits are 1 s, 2 s, 4 s and so on, up to `max`. Clients that fail together would otherwise
retry together, so `jitter` multiplies each wait by a random number between `1 - jitter` and `1 + jitter`, then caps it at
`max`.

## What is repeated

| What fails | `network` (default) | `error` | `always` | `never` |
|---|---|---|---|---|
| no response (offline, unreachable, a `fetch` that rejected) | repeated | repeated | repeated | no |
| `5xx` | no | repeated | repeated | no |
| `4xx` | no | no | repeated | no |
| `2xx` and `3xx` | no | no | no | no |
| a timeout or an abort | no | no | no | no |

An unknown `on` value is read as `network`. A request is repeated only when the mode says so, `attempts` is not used up and
the method is allowed. `attempts:3` is at most four requests. Without `ypy-retry` on the element nothing is repeated,
whatever the page default says for `on`.

`on:always` repeats a `404` or a `422` too, which rarely helps. Use it for an endpoint that answers `4xx` while it warms up,
such as a job whose result is not ready.

When the attempts run out, the last response is handled as without `ypy-retry`: `ypy:response:error` and the error flow
run once, for that response. The repeated answers are discarded and fire no event of their own.

### Retry-After

A `503` or `429` that is repeated waits at least as long as its `Retry-After` header says, in seconds or as an HTTP date.
Jitter never shortens it. When the header asks for more than `max`, the request is not repeated and the answer goes to the
page, because asking sooner would only be refused again. A `429` is repeated only with `on:always`, and a `503` with
`on:error` or `on:always`. For another origin the browser shows `Retry-After` only if the server lists it in
`Access-Control-Expose-Headers`.

```http
HTTP/1.1 503 Service Unavailable
Retry-After: 3
```

### Methods

A method that is not idempotent is not repeated whatever `on` says, because the first request may have reached the server
and run. A repeated `POST` can create a second order. Opt in with `unsafe:true` only where the server handles duplicates, for
example with an idempotency key.

```html
<form ypy-post="/orders" ypy-retry="attempts:2 on:error unsafe:true">...</form>
```

With `unsafe:true` every method is repeated and `methods` is not read.

## How the wait works

The extension wraps `ctx.fetch` from its `ypy_before_request` hook, so the repeat happens inside the same request.

- The request keeps its queue slot (`ypy-sync`), so a second click waits behind it.
- `ypy:before:request` and `ypy:finally:request` fire once. `ypy:request:retry` fires before each repeat with
  `{attempt, delay, ctx, status}` and can be cancelled with `preventDefault()`. `delay` is the wait that was chosen, jitter
  included, and `status` is absent after a network failure.
- The wait ends early when the window fires `online`. The request timeout is the deadline for all attempts together, and an
  abort ends the wait at once.
- If the element is removed from the page, the request is aborted quietly.
- `ypy:error` fires once, after the last attempt, with `ctx.errorKind` set when the failure was a network failure.

The same body is sent again, so a request with a stream as its body cannot be repeated.

## Pitfalls

- A repeated `POST` can run twice. Leave `unsafe` off unless the server de-duplicates.
- The request timeout covers all attempts, so `attempts:3` with `delay:10s` needs a timeout longer than the waits.
- A `Retry-After` longer than `max` stops the retry. Raise `max` when the server asks for long pauses.

A page can show that it is offline while retries run, for example with Tailwind:

```html
<div ypy-get="/feed" ypy-trigger="load, every 30s" ypy-retry="attempts:3" class="[.ypy-offline_&]:opacity-60">...</div>
```
