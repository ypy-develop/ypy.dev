---
order: 230
title: Live updates
description: A chat room where every open page receives new messages over one ypy-stream response, and a reconnecting page gets the current state first.
---

# Live updates

A message sent by one person shows up on every open page: the page opens one long-lived response with `ypy-stream`, and the server writes HTML patches into it.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-stream.min.js"></script>

<section ypy-stream="/room/events">
  <ul id="messages"></ul>
</section>

<form ypy-post="/room/messages" ypy-target="this" ypy-swap="innerHTML">
  <input name="body" autocomplete="off" required>
  <button>Send</button>
</form>
```

The `<section>` opens `GET /room/events` when it loads and keeps the connection open. The form posts a message; the
answer to that request only clears the form, because the message itself arrives through the stream.

## What the server answers

`GET /room/events` answers with a stream, not a fragment:

```text
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
```

The first event is the whole current state, so a page that has just connected or reconnected is up to date. Without a
`selector`, a patch goes to the page element that has the same `id` as its top-level element:

```text
event: ypy-patch-elements
id: 41
data: elements <ul id="messages">
data: elements   <li>Helena: the deploy is out</li>
data: elements   <li>Rui: confirmed, checking the logs</li>
data: elements </ul>

```

When someone sends a message, every open connection receives one more event. `selector` and `mode` say where the new
`<li>` goes:

```text
event: ypy-patch-elements
id: 42
data: selector #messages
data: mode append
data: elements <li>Marta: logs look clean</li>

```

`POST /room/messages` answers with the form again, empty:

```html
<input name="body" autocomplete="off" required>
<button>Send</button>
```

Headers of the POST answer: `Content-Type: text/html; charset=utf-8` and `Vary: YPY-Request`. An empty message gets
`204` and nothing changes.

## How it works

- The stream is plain server-sent events. Each frame has an `event:` name and `data:` lines written as `key value`; a
  repeated key joins its values with a line feed. The names `ypy-patch-elements`, `ypy-patch-signals`, `ypy-redirect`,
  `ypy-event` and `ypy-refresh` are the ones the extension reads. An unknown name is ignored.
- `mode append` adds the new `<li>` at the end of `#messages`. With no mode, the patch morphs the element with the same
  `id`, which is what the first event does to bring the page up to date.
- The client sends `Last-Event-ID` with the last `id:` it saw when it reconnects. The server answers with the full
  state again, so the page never has to remember what it missed.
- The connection reconnects by itself with a growing delay. It stays open while the tab is in the background, unless
  you set `stream.pauseOnBackground`, which closes it until the tab is visible again.
- Escape message text on the server. A patch is HTML, and it goes through the same checks as a swapped response.

## Variations

- For one client only, any handler that writes `text/event-stream` works: the extension reads the response of the
  request whatever started it. `ypy-stream.post="/jobs/7/run"` starts a stream with a `POST`.
- To change signals as well as HTML, send `ypy-patch-signals`. To send the page somewhere, `ypy-redirect`. To fire a
  named event on the page, `ypy-event`.
- A plain server-sent events endpoint works with the `ypy-sse` extension, without the patch events. Each unnamed message
  is HTML swapped into the element: `<ul ypy-sse:connect="/ticker" ypy-swap="beforeend"></ul>`. Load `ypy-stream`
  before `ypy-sse` when both are on the page.
- For a state that changes, not just grows, [polling](polling.md) is simpler and needs no open connection.

## See also

- [`ypy-stream`](../extensions/stream.md) and [`ypy-sse`](../extensions/sse.md)
- [Streaming on the server](../index.md#sse-and-websockets)
- [Polling](polling.md)
- [Streaming a model answer](../extensions/model.md)
