---
title: "Go SDK: streaming, signals, headers, CSRF, crawlers and tests"
description: "The parts of the SDK beyond a basic handler. The handler example and the install steps are in the repository README; the request and response helpers are in the documentation index."
order: 20
---
# Go SDK: streaming, signals, headers, CSRF, crawlers and tests

The parts of the SDK beyond a basic handler. The handler example and the install
steps are in the repository README; the [request and response helpers](README.md) are
in the documentation index. The wire format of the streaming part is
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

## Streaming

```go
func (h *Jobs) Events(w http.ResponseWriter, r *http.Request) {
	s, err := ypy.NewStream(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer s.Close() // always: it stops the keepalive

	s.PatchElements(`<p id="status">Working</p>`)           // morph by id
	s.PatchSignals(map[string]any{"progress": 50})          // JSON Merge Patch
	s.PatchElements(`<li>done</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#log"))
	s.Event("job:done", map[string]int{"id": 7})
	s.Redirect("/jobs/7")
	<-s.Done() // returns when the client leaves
}
```

To send the same events to many clients, mount a `Broadcaster` as the handler of
the stream route and publish from anywhere:

```go
room := ypy.NewBroadcaster(ypy.WithSnapshot(func(r *http.Request) ([]ypy.Event, error) {
	ev, err := ypy.PatchElementsEvent(renderMessages()) // the full current state
	return []ypy.Event{ev}, err
}))
defer room.Close()

mux.Handle("GET /room/events", room)

room.PatchElements(`<li>hello</li>`, ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"))
```

The snapshot is what makes reconnection safe: every connection starts with the
full state, and the client morphs it in. A client that cannot keep up is dropped
and reconnects. Nothing blocks and no goroutine outlives its client.

### Lists: `WithKey` and `WithLimit`

A chat, a log or a feed adds items at one end of a list. Two options on an append
or prepend patch keep that list correct, with no extension on the page:

```go
room.PatchElements(`<li data-id="42">Ana: hello</li>`,
	ypy.WithMode(ypy.ModeAppend), ypy.WithSelector("#messages"),
	ypy.WithKey("data-id"), ypy.WithLimit(100))
```

- `WithKey(attr)`: an item whose attribute has the value of an item already in the
  list replaces it where it stands; otherwise it is added. The same value twice in
  one patch keeps the last. An item without the attribute is always added.
- `WithLimit(n)`: after the patch the list has at most `n` items; the oldest go
  (the first ones for an append, the last ones for a prepend).
- The key is applied before the limit, so updating an item already shown never
  pushes another one out. Every element child of the target counts as an item, so
  keep a header or an empty-state message outside the list.

They are the rules of the `key:` and `limit:` modifiers of
[`ypy-list`](../ypy/extensions/list.md); the stream client does not need that extension.
`PatchElementsEvent` returns an error for `WithKey` or `WithLimit` in any mode but
append and prepend, for a limit below 1 and for a key that is not an attribute name
(letters, digits, `_`, `.` and `-`). With no selector, the patch is the list element
itself (matched by `id`) with the new items inside it. The snapshot of a `Broadcaster` still works: after a
reconnection the client morphs the whole list, and the next patches go on from it.

### Refresh: ask every client to fetch again

A patch carries the same HTML to everyone. When a change looks different to each
user (a cart, a balance, a list filtered by permission, a page with a CSRF token),
send `ypy-refresh` instead. It carries no HTML: each client fetches its own copy,
with its own session, and the answer is morphed in.

```html
<div id="room" ypy-stream="/room/events" ypy-history-elt>
  <section id="cart" ypy-get="/cart">...</section>
</div>
```

```go
// after an order changes: every connected client fetches its own #cart
room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))

room.Refresh() // no selector: every client fetches its page again
```

`Stream.Refresh` does the same on one stream, and `RefreshEvent` builds the value
for `Send` and `Publish`. Options:

| Option | Meaning |
|---|---|
| `WithRefreshSelector(sel)` | the elements to fetch again; each one that matches must have `ypy-get`. Without it, the page |
| `WithRefreshDebounce(d)` | events with the same selector inside `d` of the first become one request. Default 250 ms (`DefaultRefreshDebounce`), 0 to 1 minute |
| `WithRefreshJitter(d)` | each client waits a random time up to `d` after the debounce, so a thousand clients do not arrive together. 0 to 1 minute |

A value outside 0 to 1 minute makes `RefreshEvent` return an error. The SDK always
writes the `debounce` line, because an SSE frame with no data line is not delivered.

What the client requests is a normal YPY request (same headers, cookies, signals):

- **With a selector**, a `GET` to the `ypy-get` URL of each matching element, swapped
  into the children of that element. The handler answers a fragment, as for any
  `ypy-get`. An element without `ypy-get` is reported through `ypy:stream:error`.
- **Without a selector**, a `GET` to the path and query of the page, with
  `YPY-Request-Type: full`. `q.IsPartial()` is false, so the same handler that
  answers a browser answers the page, and the client morphs the element marked
  `ypy-history-elt` (or the body) out of it. Do not answer a bare fragment there.

The client fires the cancelable `ypy:stream:refresh` on the element that opened the
stream just before the request. Do not use it for a change that is the same for
everyone: a patch is cheaper. A refresh reaches the clients connected when the
change happens; a client that connects later gets the snapshot. See
[Streaming](../ypy/extensions/stream.md) for the client side.

## Signals

```go
type Search struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
}

s, err := ypy.ReadSignals[Search](r) // ErrNoSignals when the request has none
ypy.PatchSignals(w, map[string]any{"page": 1, "draft": nil}) // null deletes
```

The client sends signals in the `ypy-signals` query parameter (GET), as a JSON
body (POST, PUT, PATCH) or in the `ypy-signals` form field.

### Signed read-only signals

A signal declared with `ypy-signals.server` cannot be written by the page, but
the browser sends every signal back, so a user can send any value. To trust one,
sign it when you send it and verify it when it returns (HMAC-SHA256 over the
name and the canonical JSON of the value, constant-time comparison):

```go
signer, _ := ypy.NewSigner(newKey, oldKey) // sign with newKey; accept both while pages with old values live
signer = signer.Bind(sessionID)             // optional: a value signed for one session fails in another

// Sending: the values and their signatures, to the page or to a stream.
decl, _ := signer.Declaration(map[string]any{"plan": "free", "seats": 3}) // <div ypy-signals.server="{{decl}}">
ypy.PatchSignedSignals(w, signer, map[string]any{"plan": "pro"})          // or: patch, _ := signer.SignPatch(...); stream.PatchSignals(patch)

// Receiving: the check happens before anything is decoded.
got, err := ypy.ReadSignedSignals[Account](r, signer, "plan", "seats")
if errors.Is(err, ypy.ErrSignatureMissing) || errors.Is(err, ypy.ErrSignatureInvalid) {
	http.Error(w, "forbidden", http.StatusForbidden) // errors.As with a *ypy.SignatureError gives the signal name
	return
}
```

The signatures travel in one reserved signal, `ypySig`; the client needs no
change. Name the signals you rely on: a signal that is absent or unsigned fails.
The format is in section 3.3 of [`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md). In tests,
`ypytest.SignedSignals(signer, values)` sends a signed request.

## Extension headers

The extensions `ypy-validate`, `ypy-layers` and `ypy-cache` talk through headers:

```go
q := ypy.FromRequest(r)
if q.IsValidation() { // YPY-Validate: the name of the field being checked; change nothing
	if msg := problems(r)[q.Validate]; msg != "" {
		ypy.ValidationError(w, "<span>"+html.EscapeString(msg)+"</span>")
		return
	}
	ypy.NoContent(w)
	return
}
if q.InLayer() { /* YPY-Layer: "modal", "drawer" or "popup": answer the content only */ }

ypy.OpenLayer(w, "/items/7/confirm", ypy.LayerOptions{Mode: ypy.LayerDrawer, CloseOn: []string{"deleted"}}) // YPY-Open-Layer
ypy.ExpireCache(w, "/reports*", "/dashboard")                                                              // YPY-Cache-Expire
```

`OpenLayer` writes the bare URL when it has no options and JSON (ASCII) otherwise;
`OpenLayerHTML` shows markup you already have. `KeepCache` sends `false`. In tests:
`ypytest.Validate("email")`, `ypytest.Layer(ypy.LayerModal)`, `res.AssertOpenLayer`,
`res.AssertCacheExpire`.

## CSRF protection

`ypy.NewCSRF` protects every unsafe request (anything but GET, HEAD, OPTIONS and
TRACE) with a token bound to the session: an HMAC-SHA256 of the session id, so it
needs no storage. The Go side and the [`ypy-csrf`](../ypy/extensions/csrf.md)
extension of the page are two halves of one contract:

1. The server renders the token of the session in the page head:
   `<meta name="ypy-csrf" content="TOKEN">` (`csrf.Meta(r)`).
2. The extension (`ext/ypy-csrf.js`, listed in `ypy.config.extensions` when you set
   it) sends that token in `X-CSRF-Token` on every `POST`, `PUT`, `PATCH`, `DELETE`
   and `QUERY` that goes to the page's own origin, even when the element inherits no
   header. It never sends it to another origin. A plain form with no client sends it
   in the hidden field `csrf_token` (`csrf.Field(r)`).
3. `csrf.Middleware` reads the header, else the field, and compares it with the
   token of the request's session. It also checks `Origin`.
4. When the token must change, the server answers with `YPY-CSRF: <new token>` and
   the extension replaces the content of the meta tag. The next request sends it.

```go
csrf, err := ypy.NewCSRF(key, // 32 random bytes or more
	ypy.CSRFSession(func(r *http.Request) string { return sessions.ID(r) }), // optional, see below
	ypy.CSRFPreviousKeys(oldKey))                                              // optional, for rotation
handler := ypy.Middleware(csrf.Middleware(mux)) // 403 with a clear body when the token is missing or wrong
```

```html
<head>{{ .CSRFMeta }}</head>                <!-- csrf.Meta(r): <meta name="ypy-csrf" content="..."> -->
<form method="post">{{ .CSRFField }}</form> <!-- csrf.Field(r): a plain form, with no client -->
```

Put `csrf.Meta(r)` (and `csrf.Field(r)` for forms that must work without the
client) in the page data. A visitor with no session gets a random one in a
cookie (`HttpOnly`, `SameSite=Lax`, `__Host-` prefix over TLS or with
`ypy.CSRFSecure()`), so a login form is protected too; `CSRFSession` gives the
application's own session id, which wins when it is not empty. The middleware also
adds `Vary: Cookie`, since the page depends on the cookie.

The middleware answers `403` with a plain text body that says what to do:

| Error | Cause | Answer |
|---|---|---|
| `ErrCSRFOrigin` | `Origin` is another origin (and not in `CSRFTrustedOrigins`), or no `Origin` and `Sec-Fetch-Site: cross-site` | 403 |
| `ErrCSRFMissing` | no token in the header or in the field. With `ypy-csrf` loaded, the page has no meta tag: the client fires `ypy:csrf:missing` | 403 |
| `ErrCSRFInvalid` | the token is from another session (a sign-in in another tab), or its session cookie did not arrive | 403 with `YPY-CSRF` holding the current token, so the client retries with it |
| valid, made with a previous key | key rotation | the request goes on, and the response carries `YPY-CSRF` |

`CSRFOnFailure(fn)` replaces the 403 body. `csrf.Check(r)` is the same check as an
error, for one handler outside the middleware. Behind a proxy that rewrites `Host`,
list the public origin with `CSRFTrustedOrigins`.

When the session changes, the page still holds the old token. Call
`csrf.RenewFor(w, newSessionID)` (or `csrf.Renew(w, r)`) in the sign-in handler:

```go
mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
	id := sessions.Start(w, r.PostFormValue("email")) // the new session
	csrf.RenewFor(w, id)                               // YPY-CSRF: the client updates its meta tag
	ypy.Redirect(w, r, "/")
})
```

A stream that opens with GET must not change state (it needs no token); an action
started from a stream is a POST like any other. This is not for APIs authenticated
by a header token (no cookie, no CSRF). The format is section 3.4 of
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

Test it with `ypytest.CSRFToken`, taking the token of a session with `TokenFor`:

```go
func TestCreateNeedsTheToken(t *testing.T) {
	csrf, _ := ypy.NewCSRF(key, ypy.CSRFSession(func(r *http.Request) string { return "user-7" }))
	h := csrf.Middleware(mux)
	form := ypytest.Form(url.Values{"title": {"Buy milk"}})

	ypytest.Do(h, ypytest.NewRequest("POST", "/items", form)).AssertStatus(t, http.StatusForbidden)

	token := csrf.TokenFor("user-7")
	ypytest.Do(h, ypytest.NewRequest("POST", "/items", form, ypytest.CSRFToken(token))).
		AssertStatus(t, http.StatusOK)

	stale := ypytest.Do(h, ypytest.NewRequest("POST", "/items", form, ypytest.CSRFToken(csrf.TokenFor("user-8"))))
	stale.AssertStatus(t, http.StatusForbidden)
	stale.AssertHeader(t, ypy.HeaderCSRF, token) // the client picks the right token up
}
```

## Fragments and crawlers

A fragment is a piece of a page. A search engine or an AI crawler that reaches one
(through a cache that ignores `Vary`, or through a leaked link) must not put it in an
index. The SDK marks fragments for you:

| Where | What it adds |
|---|---|
| `ypy.Fragment`, `FragmentStatus`, `ValidationError`, `Partial` | `X-Robots-Tag: noindex` and `Vary: YPY-Request` |
| `ypy.Middleware` | `Vary: YPY-Request` on every response, and `X-Robots-Tag: noindex` on the answer to a YPY request that expects a fragment (`q.IsPartial()`) |

Neither touches a page. The header is left out when the markup given to `Fragment`
is a whole document (a doctype or an `<html>` element), so a handler that writes
its full page that way is not removed from search results; a browser request, a
history restore and a full-content request never get it from the middleware; and a
value the handler already put in `X-Robots-Tag` is kept.

A fragment that is a page of its own, one a crawler should index, opts out. The
`Vary` header stays, because it keeps caches from mixing fragment and page:

```go
ypy.Fragment(w, docsListHTML, ypy.AllowIndex())

handler := ypy.NewMiddleware(ypy.AllowIndex())(mux) // instead of ypy.Middleware
```

The header is not a YPY header and the client ignores it. In a test,
`res.AssertNoIndex(t)` checks that the answer has a `noindex` (or `none`) that is
not limited to one crawler (`googlebot: noindex` does not count) and `YPY-Request` in
`Vary`.

## Testing without a browser

```go
func TestCreate(t *testing.T) {
	res := ypytest.Do(handler, ypytest.NewRequest("POST", "/items",
		ypytest.Target("ul#items"),
		ypytest.Form(url.Values{"title": {"Buy milk"}})))

	res.AssertStatus(t, http.StatusOK)
	res.AssertText(t, "li", "Buy milk")                    // CSS selector
	res.AssertPartial(t, "#count", "1 items")              // <ypy-partial>
	res.AssertTrigger(t, ypy.HeaderTrigger, "item:saved", map[string]any{"count": 1})
	res.AssertVary(t, ypy.HeaderRequest)
}
```

`ypytest` also reads signals patches (`res.AssertSignals`), the events of a
stream (`res.AssertPatchElements`, `res.AssertRefreshEvent`, `res.Events()`) and runs
handlers that never return (`ypytest.OpenStream`). Its HTML parser is small and
made for the markup a handler writes; it uses only the standard library.

### Fragment or page: which helper

A handler that serves a URL the client pushes on the history has to answer two
shapes. A YPY request wants a fragment, with no layout. A browser request (a reload,
a shared link) and the history restore after back or forward want the full page. The
classic defect is one answer for both: the user reloads `/items` and sees a bare
list. Tests for it:

| Helper | What it checks | Use it when |
|---|---|---|
| `res.AssertFragment(t)` | no doctype and no `html`, `head` or `body` element; an empty body (a 204) passes | you test one answer to a YPY request (a form post, a partial) |
| `res.AssertFullPage(t)` | an `html` and a `body` element | you test one answer to a browser request or a `Full()` request |
| `ypytest.AssertFragmentAndPage(t, h, url, opts...)` | three `GET`s to the same URL (YPY with `opts`, plain with `NotYPY()`, `HistoryRestore()`): fragment, page, page, and `YPY-Request` in `Vary` on all three. Returns the fragment and the page | the URL is a page that also serves a fragment (the routes with `ypy-push-url`, boosted links, `ypy-history-elt`) |
| `res.AssertNoIndex(t)` | `X-Robots-Tag: noindex` and `Vary: YPY-Request` | the answer is a fragment that must stay out of search engines |

```go
func TestItemsPage(t *testing.T) {
	fragment, page := ypytest.AssertFragmentAndPage(t, handler, "/items", ypytest.Target("#items"))
	fragment.AssertSelector(t, "li", 3)
	fragment.AssertNoIndex(t)
	page.AssertText(t, "title", "Items")
}
```

`AssertFragmentAndPage` only makes `GET`s; test a `POST` that answers a fragment with
`AssertFragment`. The options go to the YPY request only. A request that refreshes
the page from a stream (`ypy-refresh` with no selector) is a YPY request that wants
the full content, which `ypytest.Full()` builds: `AssertFullPage` on its answer
checks that the handler does not send a bare fragment there.

The other assertions check the `YPY-*` response headers that steer the swap, with
the exact value: `AssertRetarget`, `AssertReswap`, `AssertReselect`,
`AssertPushURL`, `AssertReplaceURL`, `AssertRedirect`, `AssertLocation` and
`AssertRefresh`.

### Testing a refresh and a list

```go
func TestOrderRefreshesTheCart(t *testing.T) {
	room := ypy.NewBroadcaster()
	defer room.Close()
	s := ypytest.OpenStream(t, room, ypytest.NewRequest("GET", "/room/events"))
	defer s.Close()
	s.WaitStarted()

	room.Refresh(ypy.WithRefreshSelector("#cart"), ypy.WithRefreshJitter(2*time.Second))

	got, ok := s.Next().Refresh() // Selector, Debounce (250 ms by default) and Jitter
	if !ok || got.Selector != "#cart" || got.Jitter != 2*time.Second {
		t.Fatalf("refresh = %+v", got)
	}
}
```

For a handler that writes a few events and returns, `res.AssertRefreshEvent(t, "#cart")`
(`""` for the page) on the recorded response is enough. The patch of a list decodes
with `ev.Elements()`: `p.Key` and `p.Limit` hold the `WithKey` and `WithLimit` values.

## Conformance vectors

`spec/vectors/` in the repository holds the vectors that the SDK and the `ypy-stream`
extension are both tested with. Each is a text file of sections opened by `=== name`:

| Section | Content |
|---|---|
| `about` | one line, what the vector shows |
| `page` | the HTML of the container (`<div id="room" ypy-stream="/room" ypy-history-elt>`) before the stream |
| `input` | the bytes of the stream, exactly as a server writes them |
| `dom` | the HTML of the container afterwards |
| `signals`, `events` | the signal patches and `ypy-event`s the client must have applied, in order |
| `respond`, `requests` | the URLs the client may request (`/cart => HTML`; `@page` is the page itself) and the requests it must make, such as `GET /cart` |

`index.txt` lists the vectors: the plain patches (`elements-by-id`, `elements-modes`,
`elements-svg`, `signals`, `event`), the lists (`list-append-limit`,
`list-prepend-limit`, `list-key`, `list-key-limit`), the refresh (`refresh-selector`,
`refresh-page`, `refresh-debounce`, `refresh-jitter`) and `unknown-things`, which
says that a client ignores an event, a key or a field it does not know and that an
`id:` line changes nothing. Only `input` is required.

The two sides check different things:

- **The SDK** (`vectors_test.go`) builds the stream of every vector with its own
  functions and compares the bytes with `input`. After a deliberate change of the wire
  format, `go test -run TestVectors -update` rewrites the `input` sections; read the
  diff. It also fails when `index.txt`, the files and the builders disagree, and when
  an event has no `data` line, which an SSE client would not deliver.
- **The client** feeds `input` to `ypy-stream`, one byte at a time and in one piece,
  and compares the page, the signals, the events and the requests with the other
  sections. A new event or key in the protocol comes with a vector.

A server written in another language (or in Go without this SDK) uses them as a
conformance suite: for each name in `index.txt`, write the stream the vector
describes and compare it with `input`. The check is the same in any language:

```go
func TestMyServerAgreesWithTheVectors(t *testing.T) {
	names, _ := os.ReadFile("vectors/index.txt") // a copy of spec/vectors
	for _, name := range strings.Fields(string(names)) {
		want := section(filepath.Join("vectors", name+".txt"), "input") // the text after "=== input"
		got := myServerStream(name)                                    // the bytes your server writes for this case
		if got != want {
			t.Errorf("%s:\n got %q\nwant %q", name, got, want)
		}
	}
}
```

Take the `about` line as the description of what to write; the Go builders in
`vectors_test.go` show the call that produces each one. The two sides share the files
by copy (`spec/vectors` is copied to the client's fixtures), so a vector never
describes what only one of them does.
