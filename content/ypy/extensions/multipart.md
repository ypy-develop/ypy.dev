---
order: 230
title: "Multipart responses: the ypy-multipart extension"
description: "Reads a multipart/mixed response part by part and swaps each part as it arrives, so one HTTP response can carry a series of page updates."
---

# Multipart responses: the `ypy-multipart` extension

One HTTP response can hold several documents, one after another, separated by a boundary: `multipart/mixed`. A server
that sends one part now and another when it is ready gives the page updates over a single long response, with ordinary
HTTP and no WebSocket. `ypy-multipart` reads each part as it arrives and swaps it into the page, like a response.

If you are choosing a real-time channel, look at [`ypy-sse`](sse.md) and [`ypy-stream`](stream.md) first. Pick
multipart when the server already produces `multipart/mixed` or when each update needs its own headers.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-multipart.min.js"></script>
```

The extension is not in `ypy.max`. It adds
`multipart/mixed` and `multipart/parallel` to the `Accept` header of requests.

## Two ways to use it

- A long connection: an element with `ypy-multipart:connect="/updates"` opens a `GET` on its `ypy-trigger` (default
  `load`) and keeps the response open. It reconnects when it ends or fails.
- An answer to a request: any request whose response has `Content-Type: multipart/mixed` is read part by part.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-multipart:connect` | a URL | Opens the connection. The parts are swapped like responses. |
| `ypy-multipart:close` | an event name | Closes the connection when that event fires on the element. |

## What a part can say

Each part has its own headers and body. It is swapped like a response, and these headers apply to it:

| Header | Meaning |
|---|---|
| `YPY-Part-ID` | Identifies the part. When the part and the ones before it were handled, this becomes the next `YPY-Last-Part-ID`. |
| `YPY-Swap` | The swap style for this part. Set on the whole response it applies to every part. `YPY-Reswap` wins. |
| `YPY-Target`, `YPY-Retarget` | The target of this part. `YPY-Retarget` wins. |
| `YPY-Select` | A selector of the content to swap. `YPY-Reselect` wins. |
| `YPY-Push` | Same as `YPY-Push-Url`. |
| `YPY-Trigger`, `YPY-Redirect`, `YPY-Location`, `YPY-Refresh` | Run for the part, as for any response. See the [response headers](../reference.md#request-headers). |

On a reconnect the request carries `YPY-Last-Part-ID`, so the server can continue after the last part that arrived.

## Events

| Event | Cancelable | Detail | When |
|---|---|---|---|
| `ypy:multipart:before:connection` | yes | `connection` | Before a connection is used, and before each reconnect. |
| `ypy:multipart:after:connection` | no | `connection` | A response was accepted. |
| `ypy:multipart:before:part` | yes | `connection`, `ctx`, `part`, `cancelled`, `waitUntil(promise)` | Before a part is handled. Cancel it to skip the part. |
| `ypy:multipart:after:part` | no | `connection`, `ctx`, `part` | After a part was swapped. |
| `ypy:multipart:error` | no | `connection`, `error`, `status` | A connection attempt, a reconnect or the parsing failed. |
| `ypy:multipart:close` | no | `connection`, `reason` | The connection was cleaned up. `reason` is `cancelled`, `ended`, `removed`, `part` or `cleanup`. |

## Settings

Set them under `multipart` in `ypy.config`, or for one request with `ypy-config='{"multipart":{...}}'`.

| Key | Default | Meaning |
|---|---|---|
| `multipart.reconnect` | `true` for `ypy-multipart:connect`, `false` for a request | Reconnect after the stream ends or fails. |
| `multipart.reconnectDelay` | `500` | Wait before the first reconnect. It doubles each time. |
| `multipart.reconnectMaxDelay` | `60000` | Longest wait between attempts. |
| `multipart.reconnectMaxAttempts` | unlimited | Attempts in a row before giving up. |
| `multipart.reconnectJitter` | `0.3` | Random spread of each wait, as a fraction. `0` turns it off. |
| `multipart.connectTimeout` | `ypy.config.defaultTimeout` | Timeout of one connection attempt. `0` waits without limit. |
| `multipart.pauseOnBackground` | `true` for `ypy-multipart:connect` | Close while the tab is hidden and reconnect when it shows. |

## Example

```html
<section id="jobs" ypy-multipart:connect="/jobs/updates">
  <p>Waiting for updates...</p>
</section>
```

## Notes

- Flush after each part. A part that sits in a server buffer does not reach the page.
- The connection stays open for a long time. Count it against the connection limit of your server and proxy.
- A `YPY-Redirect` in a part must stay on the origin of the page, unless `config.mode` is `cors`.

## See also

[`ypy-sse`](sse.md), [`ypy-stream`](stream.md), [`ypy-ws`](ws.md), [Headers](../reference.md#request-headers),
[Streaming on the server](../index.md#server-side).
