---
order: 390
title: "Server-Sent Events: the ypy-sse extension"
description: "Opens an event stream from an element, swaps each unnamed message into the page, turns named messages into DOM events and reconnects when it drops."
---

# Server-Sent Events: the `ypy-sse` extension

A Server-Sent Events (SSE) stream is one long `GET` response in which the server sends messages when it has them.
With `ypy-sse`, an element opens such a stream. Each unnamed message is swapped into the page like a response, and each
named message becomes a DOM event.

This extension reads the plain SSE format. For the YPY stream protocol, with `ypy-patch-elements`, signals and
redirects, use [`ypy-stream`](stream.md).

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-sse.js"></script>
```

The extension is part of `ypy.max`.

## Open a stream

```html
<div ypy-sse:connect="/room/7/messages" ypy-target="#messages" ypy-swap="beforeend"></div>
<ul id="messages"></ul>
```

The element connects when its `ypy-trigger` fires (`load` by default) with a `GET` to the URL. The `ypy-target`,
`ypy-swap` and `ypy-select` of the element say where each message goes.

A request can also turn into a stream: when the answer to any `ypy-get` has the content type `text/event-stream`, the
extension takes over the body.

```html
<button ypy-get="/jobs/42/log" ypy-target="#log" ypy-swap="beforeend">Follow the job</button>
```

## Messages

| What the server sends | What happens |
|---|---|
| a message with only `data:` | the data is swapped into the target as HTML |
| a message with `event: <name>` | a DOM event with that name fires on the element, with `detail.data` and `detail.id` |
| `id:` | remembered, and sent back as `Last-Event-ID` when the stream reconnects |
| `retry:` | replaces `sse.reconnectDelay` |
| a line that starts with `:` | a comment; it counts as activity for the heartbeat |

An empty message never clears the target.

```html
<div ypy-sse:connect="/jobs/42/events" ypy-sse:close="done" ypy-target="#status"></div>
<script>
  document.addEventListener("progress", e => console.log(e.detail.data))
</script>
```

When a message with the event name given in `ypy-sse:close` arrives, the stream closes. Without it the stream ends when
the element leaves the page or the server closes the connection.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-sse:connect` | URL | open the stream on the trigger (default `load`) |
| `ypy-sse:close` | event name | close the stream when this named event arrives |
| `ypy-sse-connect`, `ypy-sse-close` | | old spellings; the extension copies them to the new names and warns in the console |
| `ypy-sse-swap` | | no longer used: unnamed messages are swapped, named ones are events; it only warns |

## Events

| Event | Cancelable | `detail` | Fires |
|---|---|---|---|
| `ypy:sse:before:connection` | yes | `connection` | before each connection and each reconnect; cancel it, or set `detail.connection.cancelled`, to stop |
| `ypy:sse:after:connection` | no | `connection` (`status`) | when a `text/event-stream` answer was accepted |
| `ypy:sse:before:message` | yes | `connection`, `message`, `cancelled`, `waitUntil` | for each message, before it is used; cancel it to ignore the message |
| `ypy:sse:after:message` | no | `connection`, `message` | after a message was swapped or its event dispatched |
| `ypy:sse:error` | no | `connection`, `error`, `status`, `url` | a connection or the stream failed; a stream you aborted does not fire it |
| `ypy:sse:close` | no | `connection`, `reason` | once, when the connection is cleaned up; `reason` is `cancelled`, `message`, `ended`, `removed` or `cleanup` |
| `ypy:release` | no | `data`, `id` | a named event the server can send to release the request slot while the stream stays open |

## Settings

Set them in `ypy.config.sse`, or for one element with `ypy-config='{"sse":{...}}'`.

| Setting | Default | Meaning |
|---|---|---|
| `sse.reconnect` | `true` for `ypy-sse:connect`, `false` for a request turned into a stream | reconnect after the stream ends or fails |
| `sse.reconnectDelay` | `500` | first wait before a reconnect, in ms or `"2s"`; doubles on each attempt |
| `sse.reconnectMaxDelay` | `60000` | longest wait |
| `sse.reconnectMaxAttempts` | unlimited | attempts in a row before giving up |
| `sse.reconnectJitter` | `0.3` | random spread of each wait, as a fraction (`0` is off) |
| `sse.connectTimeout` | `defaultTimeout` | timeout of one connection attempt (`0` is none) |
| `sse.pauseOnBackground` | `true` for `ypy-sse:connect` | close while the tab is hidden, reconnect when it shows again |
| `sse.heartbeat` | `0` (off) | close and reopen when no byte arrives for this long, such as `45s` |
| `sse.releaseOn` | `immediate` for `ypy-sse:connect`, `end` otherwise | when the request gives back its queue slot and indicators: `immediate`, `first` or `end` |

## Server

The format is plain `text/event-stream`. Write each line of HTML on its own `data:` line, end the message with a blank
line and flush.

Escape user text when you render the HTML. When the client reconnects it sends `Last-Event-ID`:
read it and replay what the visitor missed.

The YPY stream protocol uses named events, so an element with `ypy-sse:connect` would receive those as DOM events, not
as swaps. Use [`ypy-stream`](stream.md) for a server that speaks it, and this extension for servers that already speak
plain SSE.

## Notes

- Every `ypy-sse:connect` holds one connection. Browsers limit the connections per host over HTTP/1.1; HTTP/2 raises
  the limit a lot.
- Messages from the network are HTML. The same rules apply as for any response: the server escapes what users wrote.
- A proxy that buffers responses delays the stream. Turn buffering off for the route.
- A stream that pauses in a background tab frees one server connection per hidden tab.

## See also

- [Streaming: `ypy-stream`](stream.md)
- [WebSocket: `ypy-ws`](ws.md)
- [Live updates (example)](../examples/live-updates.md)
