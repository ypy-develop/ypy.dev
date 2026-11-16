---
title: "Documentation"
description: "Short guides for the YPY Go SDK. The API reference is the package documentation (go doc github.com/ypy-develop/go); the wire format of the streaming part is spec/PROTOCOL.md."
order: 10
---
# Documentation

Short guides for the YPY Go SDK. The API reference is the package documentation
(`go doc github.com/ypy-develop/go`); the wire format of the streaming part is
[`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md). Versão em português:
[`pt-BR/README.md`](pt-BR/README.md).
Streaming (lists and `ypy-refresh` included), signals, the extension headers, CSRF
protection, the `X-Robots-Tag` header of fragments, testing and the conformance
vectors are in the [SDK guide](guide.md).

## Requests and responses

- `ypy.FromRequest(r)` returns a `ypy.Request`: `IsYPY`, `Boosted`,
  `HistoryRestore`, `Target`, `Source`, `RequestType` (`full` or `partial`),
  `CurrentURL`, `Prompt` (already decoded), `Validate` and `Layer`.
  `IsPartial()` says whether to answer a fragment, `IsValidation()` whether the
  request is a field check of `ypy-validate`, `InLayer()` whether it fills a layer.
- `ypy.Middleware` stores it in the context (`ypy.FromContext`) and adds
  `Vary: YPY-Request`. Use it on every route that can answer both a fragment and
  a page.
- A history restore arrives **without** `YPY-Request` and with
  `YPY-History-Restore-Request: true`: answer a full page.

| Helper | Sets |
|---|---|
| `Redirect(w, r, url)` | `YPY-Redirect` and 204 for a YPY request, 303 otherwise |
| `Location(w, LocationOptions{...})` | `YPY-Location` (a path, or JSON when options are set) |
| `PushURL`, `ReplaceURL` | `YPY-Push-Url`, `YPY-Replace-Url` (`"false"` prevents) |
| `Refresh` | `YPY-Refresh: true` |
| `Reswap`, `Retarget`, `Reselect` | `YPY-Reswap`, `YPY-Retarget`, `YPY-Reselect` |
| `Trigger`, `TriggerAfterSwap`, `TriggerAfterSettle` | `YPY-Trigger*`; a JSON object name to detail, ASCII-safe; calls add up |
| `OpenLayer(w, url, LayerOptions{...})`, `OpenLayerHTML` | `YPY-Open-Layer`: the bare URL, or ASCII JSON when options are set; unknown modes and sides are refused |
| `ExpireCache(w, patterns...)`, `KeepCache` | `YPY-Cache-Expire`: patterns joined with commas (calls add up), or `false` |
| `Fragment`, `FragmentStatus` | an HTML fragment, with `X-Robots-Tag: noindex` and `Vary: YPY-Request` ([why](#fragments-and-search-engines)) |
| `Partial(w, PartialSpec{Target, Swap, ID}, markup)` | one `<ypy-partial>`, attributes escaped; same headers as `Fragment` |
| `ValidationError` | 422 with the form and its errors; same headers as `Fragment` |
| `SeeOther` | 303 with `Location` |
| `NoContent` | 204 |

Markup is written as it is: escape user data when you build it.

### Validation and layers

A field check of `ypy-validate` is a `POST` with the values of the whole form and
`YPY-Validate: <field>`, usually to the handler of the form itself. Test
`q.IsValidation()` first, answer with `ValidationError` (422) or `NoContent`, and
return before the action runs. A request that fills a layer carries `YPY-Layer`;
add `Vary: YPY-Layer` if the answer depends on it. `OpenLayer` closes the loop from
the server side, and `ExpireCache` goes on the response the page receives (after a
303 the client only sees the final response).

### Fragments and search engines

A fragment is a piece of a page. If a crawler reaches one (through a cache that
ignores `Vary`, or through a leaked link), it must not index it. So `Fragment`,
`FragmentStatus`, `ValidationError` and `Partial` add `X-Robots-Tag: noindex` and
`Vary: YPY-Request`, and `ypy.Middleware` adds the same `noindex` to the answer of a
YPY request that expects a fragment (`IsPartial()`). The header is left out when:

- the markup is a whole document (it starts with a doctype or an `<html>` element);
- the handler already set `X-Robots-Tag` (its value is kept);
- you pass `ypy.AllowIndex()` to the helper, or `ypy.NewMiddleware(ypy.AllowIndex())`
  instead of `ypy.Middleware`. `Vary: YPY-Request` is added anyway.

`ypy.HeaderRobotsTag` and `ypy.RobotsNoIndex` are the constants. The header is not
read by the client. Test it with `res.AssertNoIndex(t)`.

## Streams

`NewStream(w, r, opts...)` starts the stream and returns an error when the writer
cannot flush. Options: `WithKeepAlive` (15 s), `WithRetry`, `WithWriteTimeout`
(10 s, renewed on each write, so a server-wide `WriteTimeout` does not cut the
stream). Always `defer s.Close()`. `StreamFunc` wraps a function as a handler and
does that for you.

Events: `PatchElements(markup, opts...)` with `WithMode`, `WithSelector`,
`WithNamespace`, `WithViewTransition`; `PatchSignals(patch, WithOnlyIfMissing())`;
`Redirect(url, WithRedirectMode(...))`; `Event(name, detail, WithEventTarget(sel))`;
`Refresh(opts...)` with `WithRefreshSelector`, `WithRefreshDebounce` (250 ms) and
`WithRefreshJitter`. `WithKey(attr)` and `WithLimit(n)` turn an append or prepend
patch into a keyed, bounded list. The same builders exist as values
(`PatchElementsEvent`, `RefreshEvent`, ...) for `Send` and `Broadcaster.Publish`. Single-line values with a line break are refused, which
stops frame injection.

`Broadcaster` options: `WithSnapshot` (the full state for each new connection),
`WithBuffer` (queue per client, 32), `WithStreamOptions`. Publish after updating
your state. `Close` disconnects everyone and waits for the handlers. `Broadcaster.Refresh`,
`PatchElements`, `PatchSignals`, `Redirect` and `Event` publish to every client.

## Signals

`ReadSignals[T]` looks in the `ypy-signals` query parameter, then in a JSON body,
then in the `ypy-signals` form field (one constant, `ypy.SignalsParam`). A JSON
body is put back so you can read it again. `PatchSignals(w, patch)` answers
`application/json`.

### Signed signals

`ypy.NewSigner(key, previousKeys...)` makes a `Signer` (keys of 32 bytes or more,
copied; it signs with the first and verifies with all). `Bind(id)` ties signatures
to a session or user. A signature is `v1.` plus the base64url HMAC-SHA256 of the
binding, the signal name and the RFC 8785 canonical JSON of its value, so key order
and number formatting do not matter (numbers go through a float64: carry values
above 2^53 as strings). Signatures live in the reserved signal `ypySig` (an object
from signal name to signature), which merge patches update with the value.

| | |
|---|---|
| `Sign(name, value)`, `Verify(name, value, sig)` | one signal |
| `SignPatch(patch, names...)` | the patch plus `ypySig`; send it with `PatchSignals`, `Stream.PatchSignals` or `Broadcaster.PatchSignals`. `null` deletes the signature too; `_local` names are not signed |
| `PatchSignedSignals(w, signer, patch, names...)` | `SignPatch` and `PatchSignals` |
| `Declaration(values, names...)` | the value of a `ypy-signals.server` attribute (escape it as an attribute) |
| `VerifySignals(doc, names...)` | checks the document `RawSignals` returns; no names is an error, an absent signal or signature is a failure |
| `ReadSignedSignals[T](r, signer, names...)` | `ReadSignals` that verifies first |

Failures are `*SignatureError` (field `Signal`) wrapping `ErrSignatureMissing` or
`ErrSignatureInvalid`; answer 403. Rotation: deploy with the new key first and the
old after it, drop the old key when no page holds its values. A signature does not
expire: sign a version or a time with the value if replay matters. Object values
are merged by the client, so prefer strings, numbers, booleans and arrays. The
specification is section 3.3 of `spec/PROTOCOL.md`.

## CSRF

`ypy.NewCSRF(key, opts...)` returns a `*CSRF` (immutable, safe for concurrent use).
Its token is `v1.` and the base64url HMAC-SHA256, with the server key, of the
session id; a token of one session fails in another and nothing is stored.

| Call | What it does |
|---|---|
| `Middleware(next)` | refuses unsafe requests with a 403 text body (`ErrCSRFMissing`, `ErrCSRFInvalid`, `ErrCSRFOrigin`), gives every request a session, adds `Vary: Cookie` |
| `Check(r)` | the same check as an error, to protect one handler |
| `Token(r)`, `TokenFor(session)` | the token of the request's session, or of an application session id |
| `Meta(r)`, `Field(r)` | `<meta name="ypy-csrf" ...>` (with `data-header` when the header is not `X-CSRF-Token`) and a hidden `csrf_token` input, as `template.HTML` |
| `Renew(w, r)`, `RenewFor(w, session)` | write `YPY-CSRF`, so the `ypy-csrf` extension updates its meta tag |

Options: `CSRFSession(fn)` (the application's session id; an empty result falls
back to the cookie of the middleware), `CSRFPreviousKeys(keys...)`,
`CSRFHeaderName(name)`, `CSRFSecure()`, `CSRFTrustedOrigins(origins...)`,
`CSRFOnFailure(fn)`.

The token is read from the header (`X-CSRF-Token`), else from the form field
`csrf_token`. The `Origin` header must be the request's own host or a trusted
origin; without `Origin`, `Sec-Fetch-Site: cross-site` is refused. Behind a proxy
that rewrites `Host`, list the public origin with `CSRFTrustedOrigins`.

A 403 for an invalid token carries `YPY-CSRF` with the token of the current
session (when the request came from an accepted origin), and a token made with a
previous key passes and is answered with the current one: the client updates
itself in both cases. After a sign-in call `RenewFor(w, newSessionID)`.

Tests: `ypytest.CSRFToken(token)` sends the header. The wire format is section 3.4
of [`spec/PROTOCOL.md`](https://github.com/ypy-develop/go/blob/main/spec/PROTOCOL.md).

## Tests

`ypytest.NewRequest(method, url, opts...)` options: `Target`, `Source`,
`CurrentURL`, `Prompt`, `Boosted`, `Full`, `HistoryRestore`, `NotYPY`, `Header`,
`Form`, `Query`, `Body`, `Signals`, `SignedSignals`, `Validate`, `Layer`.

`ypytest.Do(handler, req)` returns a `Response`: `Status`, `Header`, `Body`,
`Document()`, and assertions `AssertStatus`, `AssertHeader`, `AssertNoHeader`,
`AssertVary`, `AssertRedirect`, `AssertSeeOther`, `AssertTrigger`,
`AssertSelector`, `AssertText`, `AssertContains`, `AssertPartial`,
`AssertSignals`, `AssertStream`, `AssertPatchElements`, `AssertRedirectEvent`,
`AssertDispatchedEvent`, `AssertRefreshEvent`, `AssertNoIndex`, `AssertOpenLayer` (with `OpenLayer(t)` for the options),
`AssertCacheExpire`, `AssertCacheKept`, and the ones below (`AssertFragment`,
`AssertFullPage`, `AssertRetarget`, `AssertReswap`, `AssertReselect`,
`AssertPushURL`, `AssertReplaceURL`, `AssertLocation`, `AssertRefresh`). `ParseHTML` and `Node.Select` take a CSS selector:
type, `#id`, `.class`, attributes, the four combinators, lists and the pseudo
classes `:first-child`, `:last-child`, `:nth-child(N)`, `:empty`, `:not()`.
Anything else is an error, not a silent miss.

Each stream event decodes with `Elements()` (mode, selector, `Key`, `Limit`,
elements), `Signals()`, `Redirect()`, `Dispatched()` and `Refresh()`
(selector, debounce, jitter).

For a handler that does not return, `ypytest.OpenStream(t, handler, req)` runs it
and gives `Next()`, `NextWithin(d)`, `Comments()` and `Close()`; it fails the
test when the handler does not return after `Close` (a leaked goroutine).

The parser is small and forgiving, made for the markup a handler writes. It uses
only the standard library, which keeps the module free of dependencies.

### Fragment or full page

A YPY request is answered with a fragment, with no layout. A browser request,
and the request that restores a page after back or forward, must be answered with
the full page, and both answers must send `Vary: YPY-Request`. The classic defect
is a handler that answers the fragment to a refresh of a URL the client pushed on
the history, so the user sees a bare list with no layout.

```go
func TestItemsPage(t *testing.T) {
	fragment, page := ypytest.AssertFragmentAndPage(t, handler, "/items",
		ypytest.Target("#items"))
	fragment.AssertSelector(t, "li", 3)
	page.AssertText(t, "title", "Items")
}
```

`AssertFragmentAndPage(t, h, url, opts...)` makes three GET requests: a YPY one
(with `opts`), an ordinary one and a history restore one. It fails unless the
first answer is a fragment, the other two are full pages and all three list
`YPY-Request` in `Vary`. It returns the fragment and the page. To check a single
answer, use `res.AssertFragment(t)` (no doctype, no `html`, `head` or `body`
element; an empty body passes) and `res.AssertFullPage(t)` (an `html` and a
`body` element). Each also exists as a function: `ypytest.AssertFragment(t, res)`.

### Swap, target and navigation headers

These check the `YPY-*` response headers against the exact value; an absent
header fails with a message that says so. Each is a method of `Response` and a
function (`ypytest.AssertRetarget(t, res, "#errors")`):

| Assertion | Header |
|---|---|
| `AssertRetarget(t, selector)` | `YPY-Retarget` |
| `AssertReswap(t, style)` | `YPY-Reswap` |
| `AssertReselect(t, selector)` | `YPY-Reselect` |
| `AssertPushURL(t, url)` | `YPY-Push-Url` (`"false"` checks the push is prevented) |
| `AssertReplaceURL(t, url)` | `YPY-Replace-Url` (same) |
| `AssertRedirect(t, url)` | `YPY-Redirect`, and no `Location` header |
| `AssertLocation(t, path)` | `YPY-Location`, as a path or as JSON |
| `AssertRefresh(t)` | `YPY-Refresh: true` |

`res.Location(t)` returns the whole `ypy.LocationOptions` (target, swap, select)
of `YPY-Location`.

## Conformance vectors

`spec/vectors/*.txt` in the repository are the bytes of a stream and the page they
must produce, shared by the SDK and the `ypy-stream` extension. `index.txt` lists
them; section 11 of `spec/PROTOCOL.md` defines the format, and the
[guide](guide.md#conformance-vectors) shows how a server written in another
language checks itself against them.
