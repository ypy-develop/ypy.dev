---
order: 230
title: Live updates
description: A chat room where every open page receives new messages over one ypy-stream response, and a reconnecting page gets the current state first.
---

# Live updates

A chat room where a message sent by one person shows up on every open page. The page opens one long-lived response
with `ypy-stream`. The server keeps the connections in a `Broadcaster` and publishes HTML patches to all of them. A page
that reconnects gets the full current state first, so it never needs to remember what it missed.

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

The `<section>` opens `GET /room/events` when it loads and keeps it open. The form posts a message; the answer to
that request only resets the form.

## The server (Go)

```go
package room

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

type Room struct {
	hub      *ypy.Broadcaster
	messages *Log // your storage
}

func New(messages *Log) *Room {
	r := &Room{messages: messages}
	// Every new connection first receives the whole list, then live patches.
	r.hub = ypy.NewBroadcaster(ypy.WithSnapshot(func(*http.Request) ([]ypy.Event, error) {
		ev, err := ypy.PatchElementsEvent(r.list())
		return []ypy.Event{ev}, err
	}))
	return r
}

func (r *Room) list() string {
	var b strings.Builder
	b.WriteString(`<ul id="messages">`)
	for _, m := range r.messages.All() {
		b.WriteString(item(m))
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func item(m Message) string { return `<li>` + html.EscapeString(m.Body) + `</li>` }

func (r *Room) Events(w http.ResponseWriter, req *http.Request) { r.hub.ServeHTTP(w, req) }

func (r *Room) Send(w http.ResponseWriter, req *http.Request) {
	body := strings.TrimSpace(req.PostFormValue("body"))
	if body == "" {
		ypy.NoContent(w)
		return
	}
	m := r.messages.Add(body)
	r.hub.PatchElements(item(m), ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"))
	ypy.Fragment(w, `<input name="body" autocomplete="off" required><button>Send</button>`) // a cleared form
}
```

Wire it with `mux.HandleFunc("GET /room/events", room.Events)` and `mux.HandleFunc("POST /room/messages", room.Send)`,
inside `ypy.Middleware`.

## How it works

- The stream is plain server-sent events. Each event has a name (`ypy-patch-elements`) and `data:` lines. `PatchElements`
  with `ModeAppend` and a selector adds the new `<li>` at the end of `#messages`. With no mode, the patch morphs the element with the same `id`,
  which is what the snapshot does to bring a page up to date.
- `Broadcaster.Publish` never blocks. A client that cannot keep up is disconnected and comes back to a fresh snapshot.
- The connection reconnects by itself with a growing delay. It stays open while the tab is in the background, unless you
  set `stream.pauseOnBackground`, which closes it until the tab is visible again.
- Escape message text. The patch is HTML, and it goes through the same checks as a swapped response.

## Variations

- For one client only, `ypy.NewStream(w, r)` gives the same methods on one connection. Remember `defer s.Close()`.
- To change signals as well as HTML, call `PatchSignals`. To send the page somewhere, `Redirect`. To fire a named event
  on the page, `Event`.
- A plain server-sent events handler works with the `ypy-sse` extension, without the SDK. Each unnamed message is HTML
  swapped into the element: `<ul ypy-sse:connect="/ticker" ypy-swap="beforeend"></ul>`. The handler writes
  `Content-Type: text/event-stream` and frames like `data: <li>Tick</li>` followed by a blank line. Load `ypy-stream`
  before `ypy-sse` when both are on the page.
- For a state that changes, not just grows, [polling](polling.md) is simpler and needs no open connection.

## See also

- [`ypy-stream`](../extensions/stream.md) and the wire format in `spec/PROTOCOL.md` of the Go SDK
- [Streaming with the Go SDK](../server/streaming.md)
- [Polling](polling.md)
- [Streaming a model answer](../server/model-streaming.md)
