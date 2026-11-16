---
order: 90
title: "ypytest"
description: "The API of the ypytest package: build requests like the YPY client, assert on responses, partials, signals and streams."
---

# `ypytest`

`github.com/ypy-develop/go/ypytest` tests handlers without a browser. It builds requests the way the YPY client sends
them, runs the handler in memory and gives you the answer to assert on. It uses only the standard library. For a walk
through the tests of a whole application, read [Testing your app](testing.md); this page is the API.

## Build a request

```go
req := ypytest.NewRequest("POST", "/items",
	ypytest.Target("ul#items"),
	ypytest.Source("form#new"),
	ypytest.Form(url.Values{"title": {"Buy milk"}}))
```

`NewRequest(method, target string, opts ...Option)` sets `YPY-Request: true` and `YPY-Request-Type: partial`. The
options change what the request says:

| Option | Effect |
|---|---|
| `Target(sel)`, `Source(sel)` | `YPY-Target`, `YPY-Source` |
| `Full()` | `YPY-Request-Type: full`, a request that wants the whole page content |
| `NotYPY()` | a plain browser request, with no YPY headers |
| `Boosted()`, `HistoryRestore()` | a boosted link or form; a restore after back or forward |
| `CurrentURL(u)`, `Prompt(answer)` | `YPY-Current-URL`, `YPY-Prompt` |
| `Validate(field)`, `Layer(mode)` | what `ypy-validate` and `ypy-layers` send |
| `Native(platform, version, caps...)`, `NativeUserAgent(...)` | a request from a native app shell |
| `Form(values)`, `Query(values)`, `Body(contentType, r)` | the body or the query string |
| `Signals(v)`, `SignedSignals(signer, v, names...)` | signals in the request, signed or not |
| `CSRFToken(token)` | the CSRF header |
| `Header(name, value)` | any other header |

## Run it

`ypytest.Do(handler, req)` returns a `*Response`. `ypytest.Wrap(rec)` wraps an `httptest.ResponseRecorder` you already
have. The response gives `Status()`, `Header(name)`, `Body()` and `Document()`, a small parsed tree you can query with
`Select` or `First`.

```go
res := ypytest.Do(handler, req)

res.AssertStatus(t, http.StatusOK)
res.AssertText(t, "li", "Buy milk")                 // CSS selector
res.AssertSelector(t, "li", 3)                      // how many match
res.AssertPartial(t, "#count", "1 items")           // a <ypy-partial> and its text
res.AssertTrigger(t, ypy.HeaderTrigger, "item:saved", map[string]any{"count": 1})
res.AssertVary(t, ypy.HeaderRequest)
```

## What you can assert

| Area | Methods |
|---|---|
| Status, headers, body | `AssertStatus`, `AssertHeader`, `AssertNoHeader`, `AssertVary`, `AssertContains` |
| HTML | `AssertText`, `AssertSelector`, `Select` |
| Partials | `Partials`, `PartialFor`, `AssertPartial` |
| Navigation headers | `AssertRedirect`, `AssertSeeOther`, `AssertLocation`, `AssertPushURL`, `AssertReplaceURL`, `AssertRefresh` |
| Swap headers | `AssertRetarget`, `AssertReswap`, `AssertReselect`, `Triggers`, `AssertTrigger` |
| Page or fragment | `AssertFragment`, `AssertFullPage`, `ypytest.AssertFragmentAndPage` |
| Crawlers and CSP | `AssertNoIndex`, `AssertNoInline` |
| Extensions | `AssertOpenLayer`, `AssertCacheExpire`, `AssertCacheKept`, `AssertVaryNative`, `AssertNativeNavigate` |
| Signals | `SignalsPatches`, `Signals`, `AssertSignals` |
| Streams | `Events`, `AssertStream`, `ElementsPatches`, `AssertPatchElements`, `AssertRedirectEvent`, `AssertDispatchedEvent`, `AssertRefreshEvent` |

## One URL, two answers

A route the client pushes on the history must answer a fragment to the YPY client and the full page to a browser and to a
history restore. `AssertFragmentAndPage` makes three `GET`s to the URL and checks all three, plus `Vary: YPY-Request`:

```go
fragment, page := ypytest.AssertFragmentAndPage(t, handler, "/items", ypytest.Target("#items"))
fragment.AssertSelector(t, "li", 3)
page.AssertText(t, "title", "Items")
```

## Signals

```go
handler := ypy.StreamFunc(func(s *ypy.Stream, r *http.Request) error {
	return s.PatchSignals(map[string]any{"count": 2, "draft": nil})
})
res := ypytest.Do(handler, ypytest.NewRequest("GET", "/"))
res.AssertSignals(t, json.RawMessage(`{"count":1,"draft":"x"}`), map[string]any{"count": 2})
```

`AssertSignals(t, before, want)` applies the patches of the response to `before` as a JSON Merge Patch and compares the
result.

## Streams that never end

A `Broadcaster` or a long `StreamFunc` does not return until the client leaves. `OpenStream` runs it in the background and
lets the test read events as they arrive:

```go
room := ypy.NewBroadcaster()
defer room.Close()

s := ypytest.OpenStream(t, room, ypytest.NewRequest("GET", "/room/events"))
defer s.Close()
s.WaitStarted() // status and headers of the open stream

room.Refresh(ypy.WithRefreshSelector("#cart"))
got, ok := s.Next().Refresh()
// ok is true, got.Selector is "#cart"
```

`Next` waits for the next event, `NextWithin(d)` gives up after `d`, `Comments` returns the keepalive comments, and
`WaitDone` waits for the handler to return. An `Event` decodes itself with `Elements`, `Signals`, `Redirect`, `Refresh`
and `Dispatched`. `ParseEvents(body)` does the same for a recorded body.

The package also has the helpers `AssertRedirect`, `AssertFragment`, `AssertNoIndex` and the like as functions that take
a `*Response`, for code that prefers that form. See the examples in the package: `go doc github.com/ypy-develop/go/ypytest`.
