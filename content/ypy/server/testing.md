---
order: 80
id: "GUIDE-testing"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en", "topic/testing"]
title: "Testing your app"
description: "Test handlers, forms, signals and streams without a browser, with the ypytest package and the ypy-test extension."
---

# Testing your app

The server side of a YPY page is plain HTTP: a request with a few `YPY-*` headers in, a fragment and a few headers
out. You can test it **without a browser**, in milliseconds, with `ypytest`, a package of the Go SDK
(`github.com/ypy-develop/go/ypytest`). For the rest (the page itself, your scripts) you use a browser test, and the
`ypy test` command runs those for YPY's own suites.

| What | Tool | Speed |
|---|---|---|
| a handler answers the right fragment, status, headers, partials, signals | `ypytest` in `go test` | milliseconds |
| a stream sends the right events | `ypytest.OpenStream` in `go test` | milliseconds |
| the browser swaps it, the focus lands, the CSS applies | a browser test (Playwright, chromedp) in your own suite | seconds |
| a custom extension in the framework | `ypy test` (the repository's own suites) | seconds |

## The first test

The handler under test (from the [Go recipe](recipes/go.md)): it answers a fragment and two extras to YPY, a
`422` for an empty form, and a `303` to a browser without the script.

```go
package items

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

// Store is what the handler needs; a test passes a fake.
type Store interface {
	Add(title string) (count int)
}

type Items struct{ Store Store }

func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		ypy.ValidationError(w, `<form id="new"><p role="alert">Title is required</p></form>`)
		return
	}
	n := h.Store.Add(title)

	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/items")
		return
	}
	ypy.Fragment(w, `<li class="item">`+html.EscapeString(title)+`</li>`)
	ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, fmt.Sprintf("%d items", n))
	ypy.Trigger(w, "item:saved", map[string]any{"count": n})
}
```

The test builds a request **as the YPY client would send it**, runs the handler, and asserts on the response:

```go
package items_test

import (
	"net/http"
	"net/url"
	"testing"

	"example.com/app/items"
	ypy "github.com/ypy-develop/go"
	"github.com/ypy-develop/go/ypytest"
)

type fakeStore struct{ titles []string }

func (f *fakeStore) Add(title string) int {
	f.titles = append(f.titles, title)
	return len(f.titles)
}

// handler wires the route the way the app does, middleware included.
func handler(store items.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", (&items.Items{Store: store}).Create)
	return ypy.Middleware(mux)
}

func TestCreateAnswersAFragmentAndPartials(t *testing.T) {
	store := &fakeStore{}
	req := ypytest.NewRequest(http.MethodPost, "/items",
		ypytest.Target("ul#items"),
		ypytest.Form(url.Values{"title": {"Buy <milk>"}}))

	res := ypytest.Do(handler(store), req)

	res.AssertStatus(t, http.StatusOK)
	res.AssertVary(t, "YPY-Request")
	res.AssertText(t, "li.item", "Buy <milk>")
	res.AssertContains(t, "Buy &lt;milk&gt;") // the title was escaped
	res.AssertPartial(t, "#count", "1 items")
	res.AssertTrigger(t, "YPY-Trigger", "item:saved", map[string]any{"count": 1})
}

func TestCreateWithoutTitleIs422InsideTheForm(t *testing.T) {
	req := ypytest.NewRequest(http.MethodPost, "/items", ypytest.Form(url.Values{"title": {"  "}}))

	res := ypytest.Do(handler(&fakeStore{}), req)

	res.AssertStatus(t, http.StatusUnprocessableEntity)
	res.AssertText(t, "form#new p[role=alert]", "Title is required")
}

func TestCreateWithoutScriptRedirects(t *testing.T) {
	req := ypytest.NewRequest(http.MethodPost, "/items", ypytest.NotYPY(),
		ypytest.Form(url.Values{"title": {"Buy milk"}}))

	res := ypytest.Do(handler(&fakeStore{}), req)

	res.AssertSeeOther(t, "/items")
}
```

Run it with `go test ./...`. The middleware is part of what you test: `AssertVary` fails if a route forgot it.

## Building the request

`ypytest.NewRequest(method, target, options...)` returns an `*http.Request` that carries `YPY-Request: true` and
`YPY-Request-Type: partial`, like a real one. Options change what the client says:

| Option | Sets |
|---|---|
| `Target("ul#items")` | `YPY-Target` |
| `Source("form#new")` | `YPY-Source` |
| `CurrentURL(u)` | `YPY-Current-URL` |
| `Full()` | `YPY-Request-Type: full` (the response replaces the page content) |
| `Boosted()` | `YPY-Boosted: true`, and a full request, as a boosted link or form sends |
| `HistoryRestore()` | `YPY-History-Restore-Request: true` and **no** `YPY-Request`: the server must answer a full page |
| `NotYPY()` | no YPY headers: an ordinary browser, a crawler, a user with no script |
| `Prompt(answer)` | `YPY-Prompt`, URI-encoded as the client sends it |
| `Form(values)`, `Query(values)`, `Body(type, r)` | the body, or the query |
| `Signals(v)` | the client's [signals](../extensions/signals.md), in the JSON body (`POST`, `PUT`, `PATCH`) or the query (`GET`, `DELETE`) |
| `Header(name, value)` | any other header, for example a CSRF token or a cookie |

## Asserting on the response

`ypytest.Do(handler, req)` runs the handler and returns a `*Response`. Methods that fail the test take the `t`:

| Method | Checks |
|---|---|
| `AssertStatus(t, code)` | the status |
| `AssertHeader(t, name, value)`, `AssertNoHeader(t, name)` | a header, or its absence |
| `AssertVary(t, "YPY-Request")` | `Vary` lists the name |
| `AssertSeeOther(t, url)` | a `303` with that `Location` |
| `AssertRedirect(t, url)` | `YPY-Redirect` is `url`, with no `Location` |
| `AssertTrigger(t, "YPY-Trigger", name, detail)` | an event the server asks the client to fire; also `YPY-Trigger-After-Swap` and `-After-Settle` |
| `AssertSelector(t, css, n)` | how many elements match |
| `AssertText(t, css, text)` | the text of the first match, white space collapsed |
| `AssertContains(t, s)` | the body contains a string (for example the escaped form of user input) |
| `AssertPartial(t, "#count", text)`, `PartialFor`, `Partials` | a `<ypy-partial>` goes to that place |
| `Document()`, `Select(t, css)` | the parsed body, to walk it yourself |

Selectors are the subset tests need: type, `#id`, `.class`, attribute selectors, `>`, `+`, `~`, descendant,
lists, `:first-child`, `:last-child`, `:nth-child(N)`, `:empty`, `:not()`. Anything else is an error, not a silent
miss. The HTML parser is deliberately small; it reads the fragments your handlers write, not arbitrary pages.

## Table-driven: one handler, three audiences

A route that serves a fragment, a page and a restore is three behaviors. Test each:

```go
func TestCreateByAudience(t *testing.T) {
	form := ypytest.Form(url.Values{"title": {"x"}})
	tests := []struct {
		name   string
		opts   []ypytest.Option
		status int
	}{
		{"ypy", []ypytest.Option{form}, http.StatusOK},
		{"plain browser", []ypytest.Option{ypytest.NotYPY(), form}, http.StatusSeeOther},
		{"empty title", []ypytest.Option{ypytest.Form(url.Values{"title": {""}})}, http.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ypytest.Do(handler(&fakeStore{}), ypytest.NewRequest(http.MethodPost, "/items", tt.opts...))
			res.AssertStatus(t, tt.status)
		})
	}
}
```

## A CSRF and a security test

Add the header your framework reads and check a refusal looks like one:

```go
func TestRejectsWithoutToken(t *testing.T) {
	res := ypytest.Do(app.Handler(), ypytest.NewRequest(http.MethodPost, "/items"))
	res.AssertStatus(t, http.StatusForbidden) // or 419, as your framework answers
}

func TestAcceptsWithToken(t *testing.T) {
	req := ypytest.NewRequest(http.MethodPost, "/items", ypytest.Header("X-CSRF-Token", token(t)),
		ypytest.Form(url.Values{"title": {"x"}}))
	ypytest.Do(app.Handler(), req).AssertStatus(t, http.StatusOK)
}
```

And one that no user text becomes markup: post a title like `<button ypy-post="/delete">` and `AssertContains` its
escaped form. The [security checklist](../guides/security-checklist.md) lists what else to cover.

## Signals

A handler that reads and patches [signals](../extensions/signals.md):

```go
func add(w http.ResponseWriter, r *http.Request) {
	cart, err := ypy.ReadSignals[struct {
		Qty int `json:"qty"`
	}](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ypy.PatchSignals(w, map[string]any{"qty": cart.Qty + 1})
}
```

is tested with `Signals(v)` on the way in and `res.AssertSignals(t, before, want)` on the way out, where `before` is
the signals document the client had; the helper applies the patch (JSON Merge Patch) and compares the result:

```go
func TestAddRaisesTheQuantity(t *testing.T) {
	req := ypytest.NewRequest(http.MethodPost, "/cart/add", ypytest.Signals(map[string]any{"qty": 1}))
	res := ypytest.Do(http.HandlerFunc(add), req)
	res.AssertSignals(t, json.RawMessage(`{"qty":1}`), map[string]any{"qty": 2})
}
```

## Streams

A handler that does not return until the client leaves (`ypy.StreamFunc`, `ypy.NewStream`, a `Broadcaster`) is run
against an open stream. The test reads events as they arrive:

```go
func jobEvents() http.Handler {
	return ypy.StreamFunc(func(s *ypy.Stream, r *http.Request) error {
		s.PatchElements(`<p id="status">Working</p>`)
		s.PatchSignals(map[string]any{"progress": 50})
		<-s.Done() // until the client leaves
		return nil
	})
}

func TestJobStream(t *testing.T) {
	s := ypytest.OpenStream(t, jobEvents(), ypytest.NewRequest(http.MethodGet, "/jobs/7/events"))
	s.WaitStarted()

	first := s.Next() // waits up to ypytest.DefaultWait, fails the test otherwise
	patch, ok := first.Elements()
	if !ok || !strings.Contains(patch.Elements, `id="status"`) {
		t.Fatalf("first event = %s", first)
	}
	second := s.Next()
	if _, ok := second.Signals(); !ok {
		t.Fatalf("second event = %s", second)
	}
}
```

For a handler that ends by itself, `Do` plus `res.AssertStream(t)`, `res.ElementsPatches()`, `res.SignalsPatches()` and
`res.AssertPatchElements(t, selector, contains)` read the whole body at once. `OpenStream` closes the stream when the
test ends.

## What `ypytest` cannot see

It does not run JavaScript, apply CSS or swap anything: it tests **what the server says**. Whether the browser puts
the answer where you want is a browser test: keep a few (a form with a `422`, a list with a `303`, the back button),
written with the tool you already use, and run them against the real build in `dist/`. For YPY's own code, `ypy test`
runs the suites in Chrome, Firefox and Safari, and `ypy test --dist` runs them against the minified files:

```sh
go run ./cmd/ypy test                       # the main suite, in Chrome
go run ./cmd/ypy test --browsers chrome,firefox
go run ./cmd/ypy test --dist                # the built files, as a user loads them
go run ./cmd/ypy ci                         # every gate, as in CI
```

## See also

[Go recipe](recipes/go.md) · [Security checklist](../guides/security-checklist.md) · [Forms](../guides/forms.md) ·
[Events](../reference/events/index.md) · the package documentation: `go doc github.com/ypy-develop/go/ypytest`
