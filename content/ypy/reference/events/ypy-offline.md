---
title: "Offline state: ypy:offline, ypy:online, ypy:request:retry, ctx.errorKind"
description: "How YPY tells a browser with no connection from a server that answered with an error: page state, events, the retry event and ctx.errorKind."
order: 40
---
# Offline state: `ypy:offline`, `ypy:online`, `ypy:request:retry`, `ctx.errorKind`

YPY tells "the browser has no connection" apart from "the server answered with an error".

## The page state

The core listens to the window `online` and `offline` events and reads the state once at load.
While the browser is offline, `<html>` carries both of these:

```html
<html class="ypy-offline" data-ypy-state="offline">
```

When the connection comes back both are removed. `data-ypy-state` is a space-separated list of states, so
select it with `~=`: `html[data-ypy-state~="offline"]`.

The class name is `config.classes.offline` (default `ypy-offline`). Set it empty to get the attribute only:

```js
ypy.config.classes.offline = 'is-offline'
```

### Events

| Event | Fired on | When | Detail |
|---|---|---|---|
| `ypy:offline` | `document` (bubbles to `window`) | the page went offline | `{}` |
| `ypy:online` | `document` | the connection came back, or a response proved it | `{}` |
| `ypy:request:retry` | the source element | a failed request is about to be repeated | `{attempt, delay, ctx}` |

Each of the first two fires once per change, never twice in a row. `ypy:request:retry` is fired by the
[`ypy-retry` extension](../../extensions/index.md#ypy-retry) and is cancelable: call
`event.preventDefault()` and the request gives up, reporting the failure as usual (see
[`ypy-retry`](../../extensions/retry.md)).

### Example with Tailwind CSS

```html
<p class="hidden [.ypy-offline_&]:block rounded bg-amber-100 px-3 py-2 text-amber-900">
  You are offline. Changes will be sent when the connection is back.
</p>
```

The `ypy-offline:hidden` variants of the Tailwind plugin come with the styling plugin; the arbitrary variant
above works with Tailwind 3.1 and later without it.

```js
document.addEventListener('ypy:offline', () => console.log('offline'))
document.addEventListener('ypy:online', () => console.log('back online'))
```

### The limit of `navigator.onLine`

`navigator.onLine === false` means the device has no network interface. `true` does not mean the server can
be reached: a VPN, a captive portal or a broken route all report `true`, and Chromium and WebKit disagree in
those cases. YPY therefore never trusts a `true`. Two safeguards follow from that:

- a request that gets any response while the page is marked offline clears the mark and fires `ypy:online`;
- a failure with `navigator.onLine === true` is not `offline` but `network`.

## `ctx.errorKind`

A request that fails before it gets a response sets `ctx.errorKind`:

| Value | Cause |
|---|---|
| `offline` | `fetch` rejected and `navigator.onLine` was `false` |
| `network` | `fetch` rejected and `navigator.onLine` was `true` (server down, DNS, refused, blocked) |
| `timeout` | the request timeout (`config.defaultTimeout` or `ypy-config="timeout:..."`) aborted it |

A user abort leaves it unset and stays silent. An HTTP error (4xx, 5xx) is a response, not a network failure, and
never gets one of these three values. `ypy:error` keeps firing with `{ctx, error}` for all of them, so listeners
written before this feature work unchanged; read `ctx.errorKind` to tell the cases apart:

```js
document.addEventListener('ypy:error', ({detail: {ctx}}) => {
  if (ctx.errorKind === 'offline') showToast('No connection')
  else if (ctx.errorKind === 'timeout') showToast('The server took too long')
  else if (ctx.errorKind === 'network') showToast('The server cannot be reached')
})
```

`ctx.errorKind` is data on the request context, not a visible change.
