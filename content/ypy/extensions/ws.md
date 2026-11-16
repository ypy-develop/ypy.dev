---
order: 490
title: "WebSocket: the ypy-ws extension"
description: "Opens a WebSocket from an element, swaps incoming messages into the page and sends form values back over the same connection."
---

# WebSocket: the `ypy-ws` extension

A WebSocket is a two-way connection: the server can send at any time, and the page can send without a new request.
`ypy-ws` connects it to the attributes you already use. An element opens the socket, incoming messages are swapped into
the page, and another element sends its values through it.

Use it when messages go both ways at a high rate, such as a chat. When the page only listens, [`ypy-sse`](sse.md) or
[`ypy-stream`](stream.md) is simpler, because they are plain HTTP.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-ws.js"></script>
```

The extension is part of `ypy.max`.

## Connect and send

```html
<div ypy-ws:connect="/chat/7/socket" ypy-target="#log" ypy-swap="beforeend">
  <ul id="log"></ul>

  <form ypy-ws:send>
    <input name="text" autocomplete="off" required>
    <button>Send</button>
  </form>
</div>
```

- `ypy-ws:connect` opens the socket on its `ypy-trigger` (default `load`). A relative URL, or one that starts with
  `http://` or `https://`, becomes `ws://` or `wss://`.
- `ypy-ws:send` sends the element's values on `submit` for a form, `change` for a field and `click` for anything else.
  Set `ypy-trigger` to change that. It uses the nearest `ypy-ws:connect` above it.
- The socket closes when the element with `ypy-ws:connect` leaves the page.

## Messages from the server

| The server sends | What happens |
|---|---|
| a text message that is not JSON | it is swapped into the target as HTML |
| JSON with a `content` field | `content` is swapped; the object may also carry `target`, `swap` and `select` for this message |
| JSON with no `content` | nothing is swapped; listen to `ypy:ws:before:message:incoming` to use it |

```json
{"content": "<li>Ana: hello</li>", "target": "#log", "swap": "beforeend"}
```

An empty message does not clear the target.

## Messages to the server

`ypy-ws:send` collects what a request would: the form values and `ypy-vals`. It sends them as one JSON object. Fields
with several values, such as checkboxes, become arrays. The same `YPY-*` headers that a request would carry go in a
`headers` member.

```json
{"text": "hello", "headers": {"YPY-Request": "true", "YPY-Current-URL": "https://example.com/chat/7"}}
```

To send something else, set `detail.message.data` in `ypy:ws:before:message:outgoing`: it is sent as it is.

Messages sent before the socket is open wait in a queue and go out as soon as it opens.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-ws:connect` | URL | open the WebSocket on the trigger |
| `ypy-ws:send` | none | send this element's values over the socket of the nearest `ypy-ws:connect` |
| `ypy-ws-connect`, `ypy-ws-send` | | old spellings; copied to the new names with a console warning |

## Events

| Event | Cancelable | `detail` | Fires |
|---|---|---|---|
| `ypy:ws:before:connection` | yes | `connection` | before a socket is made and before each reconnect; cancel it to stop (a `ypy:ws:close` with reason `cancelled` follows) |
| `ypy:ws:after:connection` | no | `connection` | the socket opened; queued messages go out right after |
| `ypy:ws:close` | no | `connection`, `reason`, `code` | the socket closed; `reason` is `closed`, `cancelled` or `removed`; `code` is `null` unless the server closed it |
| `ypy:ws:error` | no | `connection`, `error`, `url` | a socket error, a failed send, a full queue, or a send with no open socket |
| `ypy:ws:before:message:outgoing` | yes | `connection`, `message`, `cancelled`, `waitUntil` | before a send; edit `message.headers` or `message.values`, set `message.data`, or cancel |
| `ypy:ws:after:message:outgoing` | no | `connection`, `message` | after the message was written to the socket |
| `ypy:ws:before:message:incoming` | yes | `connection`, `message`, `cancelled`, `waitUntil` | before an incoming message is swapped; cancel it to ignore it |
| `ypy:ws:after:message:incoming` | no | `connection`, `message` | after it was processed |

`detail.waitUntil(promise)` delays the send or the processing until the promise settles.

## Settings

Set them in `ypy.config.ws`, or for one element with `ypy-config='{"ws":{...}}'`.

| Setting | Default | Meaning |
|---|---|---|
| `ws.reconnect` | `true` | reconnect after a close with one of `reconnectCodes` |
| `ws.reconnectCodes` | `1001, 1005, 1006, 1011, 1012, 1013, 1014` | close codes that trigger a reconnect; any other ends the connection |
| `ws.reconnectDelay` | `500` | first wait, in ms or `"2s"`; doubles on each attempt |
| `ws.reconnectMaxDelay` | `60000` | longest wait |
| `ws.reconnectMaxAttempts` | unlimited | attempts in a row before giving up |
| `ws.reconnectJitter` | `0.3` | random spread of each wait (`0` is off) |
| `ws.pauseOnBackground` | `true` | close while the tab is hidden, reconnect when it shows |
| `ws.heartbeat` | `0` (off) | close and reopen when no message arrives for this long, such as `45s` |
| `ws.maxOutgoingMessagesQueueSize` | `100` | messages that may wait for the socket; the next one fires `ypy:ws:error` |
| `ws.protocols` | none | sub-protocols for the `WebSocket` constructor |

The browser does not show protocol ping frames to the page, so with `ws.heartbeat` the server has to send a text message
now and then.

## Server

The Go SDK has no WebSocket helper, and the standard library has no WebSocket server. Use a WebSocket library, and keep
the messages in the shapes above. The rules for the HTML are the same as for any answer: escape what users wrote.

```go
// the JSON the page sends with ypy-ws:send on the form above
type chatIn struct {
	Text    string            `json:"text"`
	Headers map[string]string `json:"headers"`
}

// what you send back: HTML to swap, with an optional target for this message
type chatOut struct {
	Content string `json:"content"`
	Target  string `json:"target,omitempty"`
	Swap    string `json:"swap,omitempty"`
}
```

Check the `Origin` of the handshake on the server, and authenticate the request that opens the socket. A WebSocket
handshake is a `GET`, so [`ypy-csrf`](csrf.md) adds no token to it, and the browser's CORS rules for `fetch` do not
apply to it.

## Notes

- Swaps from a socket do not use view transitions.
- A server that closes the socket with a code outside `reconnectCodes` (for example `1000` or `1008`) ends the
  connection for good. That is the way to tell the page not to come back.

## See also

- [Server-Sent Events: `ypy-sse`](sse.md)
- [Streaming: `ypy-stream`](stream.md)
- [Security](../reference/security.md)
