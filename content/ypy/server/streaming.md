---
order: 20
title: "Streaming"
description: "Push changes to one page or to many with server-sent events, using Stream and Broadcaster."
---

# Streaming

A normal answer ends when the response ends. A stream stays open, and the server writes an event whenever something
changes: HTML to patch in, a signals patch, a redirect, a named event or a request to fetch again. The page side is the
[`ypy-stream`](../extensions/stream.md) extension (and [`ypy-sse`](../extensions/sse.md) for plain server-sent
messages). The wire format is in `spec/PROTOCOL.md` of the SDK repository, and any language can write it with plain
text and flushes.

```html
<div id="room" ypy-stream="/room/events">
  <ul id="messages"></ul>
</div>
```

## One client: `Stream`

```go
func (h *Jobs) Events(w http.ResponseWriter, r *http.Request) {
	s, err := ypy.NewStream(w, r)
	if err != nil { // the writer cannot flush
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer s.Close() // always: it stops the keepalive

	s.PatchElements(`<p id="status">Working</p>`)    // morph by id
	s.PatchSignals(map[string]any{"progress": 50})   // JSON Merge Patch, null deletes
	s.PatchElements(`<li>done</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#log"))
	s.Event("job:done", map[string]int{"id": 7})     // a DOM event on the page
	s.Redirect("/jobs/7")                            // navigate; ypy.WithRedirectMode(ypy.RedirectReplace) or RedirectHard
	<-s.Done()                                       // returns when the client leaves
}
```

`ypy.StreamFunc(fn)` turns a function of `(*Stream, *http.Request) error` into a handler. `NewStream` sends a keepalive
comment every 15 seconds by default; `WithKeepAlive`, `WithRetry` and `WithWriteTimeout` change it.

The patch modes are `ModeOuter` (the default, a morph), `ModeInner`, `ModeReplace`, `ModeAppend`, `ModePrepend`,
`ModeBefore`, `ModeAfter` and `ModeRemove`. Other `ElementsOption` values: `WithSelector`, `WithNamespace`,
`WithViewTransition`, and `WithKey` and `WithLimit` for lists.

## Many clients: `Broadcaster`

```go
room := ypy.NewBroadcaster(ypy.WithSnapshot(func(r *http.Request) ([]ypy.Event, error) {
	ev, err := ypy.PatchElementsEvent(renderMessages()) // the full current state
	return []ypy.Event{ev}, err
}))
defer room.Close()

mux.Handle("GET /room/events", room) // a Broadcaster is an http.Handler

// from any handler:
room.PatchElements(`<li>hello</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"))
room.PatchSignals(map[string]any{"online": room.Len()})
```

It has the same methods as `Stream` (`PatchElements`, `PatchSignals`, `Redirect`, `Event`, `Refresh`) and
`Publish(ev)` for an event you built. The snapshot is what makes a reconnection safe: every connection starts with the
full state and the page morphs it in. A client that cannot keep up is dropped and reconnects, `Publish` never blocks and
no goroutine outlives its client. `WithBuffer(n)` sets how many events wait per client.

## Lists

A chat, a log or a feed adds items at one end. Two options on an append or prepend patch keep the list correct:

```go
room.PatchElements(`<li data-id="42">Ana: hello</li>`,
	ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"),
	ypy.WithKey("data-id"), ypy.WithLimit(100))
```

`WithKey` replaces an item that has the same attribute value instead of adding it again. `WithLimit` drops the oldest
items beyond `n`. The key is applied before the limit.

## Refresh instead of a patch

A patch carries the same HTML to everyone. When each user sees something different (a cart, a balance, a list filtered
by permission), send a refresh. It has no HTML: each client fetches its own copy with its own session.

```go
room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))
room.Refresh() // no selector: every client fetches its page again
```

`WithRefreshDebounce` merges events that arrive close together (250 ms by default), and `WithRefreshJitter` spreads the
requests of a large crowd. Without a selector the client asks for the page with `YPY-Request-Type: full`, so answer the
whole page there, not a bare fragment.

## Test a stream

```go
s := ypytest.OpenStream(t, room, ypytest.NewRequest("GET", "/room/events"))
defer s.Close()
s.WaitStarted()
room.PatchElements(`<li>hi</li>`)
patch, ok := s.Next().Elements()
```

See [`ypytest`](ypytest.md). The conformance vectors in `spec/vectors/` of the SDK repository are the suite for a server
written without this SDK.
