---
order: 20
title: HTTP headers
description: "The request headers YPY sends and the response headers it reads, each with its effect and, for the ones a server sets, a Go example."
---

# HTTP headers

YPY talks to the server with ordinary HTTP headers. The browser side adds request headers, so the server can
tell a YPY request from a page load. The server answers with response headers when it wants to change what the
client does with the answer. All names start with `YPY-`.

## Request headers

The client sends these. A server reads them to decide what to answer, usually a fragment instead of a whole page.

| Header | Value | What it tells the server |
|---|---|---|
| `YPY-Request` | `true` | The request comes from YPY. It is always `true`. A server that varies a page by it should send `Vary: YPY-Request`. |
| `YPY-Source` | `tag#id`, for example `button#submit` | The element that triggered the request. |
| `YPY-Target` | `tag#id`, for example `div#results` | The element the response goes into. |
| `YPY-Current-URL` | an absolute URL | The URL the browser shows now. |
| `YPY-Request-Type` | `partial` or `full` | `partial` for a targeted swap, `full` when the target is `body` or `ypy-select` is used. |
| `YPY-Boosted` | `true` | Present when the request comes from [`ypy-boost`](attributes/index.md). |
| `YPY-History-Restore-Request` | `true` | Present when the request restores a page after back or forward. The server must answer with a whole page. |

A history restore request leaves `YPY-Request` out on purpose, so a server that answers a YPY request with a
fragment does not send one to a page that needs the whole document.

With the Go SDK, `ypy.FromRequest(r)` reads all of them into one struct:

```go
func (h *Items) List(w http.ResponseWriter, r *http.Request) {
	req := ypy.FromRequest(r)
	if req.IsYPY && !req.Boosted && req.RequestType == ypy.RequestPartial {
		ypy.Fragment(w, renderRows(h.store.All())) // only the rows
		return
	}
	renderPage(w, h.store.All()) // the whole page
}
```

`ypy.Middleware` adds `Vary: YPY-Request` to every response. See the [server section](../server/index.md).

## Response headers

The server sets these. The client reads them after the response arrives and before it swaps.

| Header | Value | Effect |
|---|---|---|
| `YPY-Trigger` | an event name, a comma separated list, or a JSON object of names and details | Fires client events after the response. |
| `YPY-Push-Url` | a URL, or `false` | Pushes a URL into the history. It overrides the history attributes of the element. |
| `YPY-Replace-Url` | a URL, or `false` | Replaces the current history entry. It overrides the history attributes of the element. |
| `YPY-Redirect` | a URL | Redirects with a full page load. |
| `YPY-Location` | a URL, or a JSON object with `path` and request options | Navigates with an AJAX request, without a full reload. |
| `YPY-Refresh` | `true` | Reloads the page. |
| `YPY-Retarget` | a CSS selector | Replaces the target of the swap. |
| `YPY-Reswap` | a swap spec, as in `ypy-swap` | Replaces the swap style. |
| `YPY-Reselect` | a CSS selector | Replaces `ypy-select`: the part of the response that is swapped. |
| `YPY-Group` | `atomic` | Makes the swap all or nothing. If one task of the response (main swap, out-of-band element or partial) has no target, none is applied and the history is not updated. See [Atomic swaps](atomic.md). |

### Examples in Go

Each helper of the SDK sets one header. They go before the body is written.

```go
// Fire an event the page listens for, with a detail.
ypy.Trigger(w, "item:saved", map[string]any{"count": 12})

// The form answered with a list, but this request should fill #notice instead.
ypy.Retarget(w, "#notice")
ypy.Reswap(w, "outerHTML")
ypy.Reselect(w, "#notice")

// After a successful change, show the new address.
ypy.PushURL(w, "/items/12")
ypy.ReplaceURL(w, "/items?page=2") // or replace the current entry
```

Navigation helpers:

```go
// Full page load to another URL. A browser without YPY gets a 303.
ypy.Redirect(w, r, "/login")

// Navigate without a reload, into a target, with the swap style of your choice.
ypy.Location(w, ypy.LocationOptions{Path: "/items/12", Target: "#main", Swap: "innerHTML"})

// Reload the page, for example after a deploy marker changed.
ypy.Refresh(w)

// All or nothing for a response with a main swap and out-of-band parts.
ypy.AtomicGroup(w)
```

`ypy.Trigger` can be called more than once; the events are merged into one `YPY-Trigger` header. The client
sees the event after the response arrives, so a listener can be on any element or on `document`:

```html
<script>
  document.body.addEventListener("item:saved", (e) => {
    console.log("items:", e.detail.count)
  })
</script>
```

The SDK also has `ypy.TriggerAfterSwap` and `ypy.TriggerAfterSettle`. The 0.1.1 client does not read a header
for them, so use `ypy.Trigger` and listen for `ypy:after:swap` or `ypy:after:settle` when the code must wait for
the swap.

## Added by extensions

The headers below exist only when the extension is loaded. The page of each extension has the details.

| Header | Direction | Extension | Value and effect |
|---|---|---|---|
| `YPY-Cache-Expire` | response | [`ypy-cache`](../extensions/cache.md) | Marks stored copies old, by path pattern or `tag:<name>`, or `false` to keep everything. An old copy is still shown at once and is always revalidated. |
| `YPY-Cache-Evict` | response | [`ypy-cache`](../extensions/cache.md) | Deletes stored copies by path pattern or `tag:<name>`. |
| `YPY-Commands` | response | [`ypy-commands`](../extensions/commands.md) | A JSON list of typed client commands that run after the swap, with no `eval`. |
| `YPY-CSRF` | response | [`ypy-csrf`](../extensions/csrf.md) | A new CSRF token for the next request. |
| `YPY-Download` | response | [`ypy-download`](../extensions/download.md) | A URL the extension fetches and saves as a download. |
| `YPY-Layer` | request | [`ypy-layers`](../extensions/layers.md) | The kind of layer (`modal`, `drawer` or `popup`) the request fills or starts in. Send `Vary: YPY-Layer` when the answer is cacheable. |
| `YPY-Open-Layer` | response | [`ypy-layers`](../extensions/layers.md) | Opens a layer instead of swapping. A JSON object with `url` or `html` and options, or just a URL. |
| `YPY-Lazy-Group` | request | [`ypy-lazy`](../extensions/lazy.md) | The group name of a request that loads several lazy regions at once. |
| `YPY-Last-Part-ID` | request | [`ypy-multipart`](../extensions/multipart.md) | Sent on a reconnect: the id of the last part fully handled. |
| `YPY-Part-ID` | response | [`ypy-multipart`](../extensions/multipart.md) | Part header that identifies the part. |
| `YPY-Swap` | response | [`ypy-multipart`](../extensions/multipart.md) | Swap style of a part, or of the whole stream. `YPY-Reswap` wins over it. |
| `YPY-Select` | response | [`ypy-multipart`](../extensions/multipart.md) | Selector of the content to swap for a part. `YPY-Reselect` wins over it. |
| `YPY-Push` | response | [`ypy-multipart`](../extensions/multipart.md) | Alias of `YPY-Push-Url` on a part. |
| `YPY-Native` | request | [`ypy-native`](../extensions/native.md) | Platform, bridge version and native components of an app shell: `<platform>; v=<n>; caps=<a,b>`. Decide on `caps`, not on the version. |
| `YPY-Purpose` | request | [`ypy-nav`](../extensions/nav.md) | `prefetch` on every prefetch request, so the server can tell it from a visit. |
| `YPY-Poll-Interval` | response | [`ypy-poll`](../extensions/poll.md) | New polling interval for the element, in milliseconds or as a time such as `5s`. |
| `YPY-Prompt` | request | [`ypy-prompt`](../extensions/prompt.md) | The answer to `ypy-prompt`, URI-encoded. |
| `YPY-PTag` | request and response | [`ypy-ptag`](../extensions/ptag.md) | The polling tag of the element. The server sets a new one in the response and the client sends it back. |
| `YPY-Validate` | request | [`ypy-validate`](../extensions/validate.md) | The name of the field being checked. The server answers 422 with the message markup, or 204 when the field is valid, and saves nothing. |

Go helpers for the extension headers exist where it makes sense:

```go
// Layers: open a form in a modal instead of swapping.
ypy.OpenLayer(w, "/items/new")

// Cache: mark copies old after a change, or delete them.
ypy.ExpireCache(w, "/reports*", "tag:products")
ypy.EvictCache(w, "/cart")

// Field validation: a check must not save anything.
if ypy.FromRequest(r).IsValidation() {
	// answer 422 with the message, or 204
}
```

Everything the SDK sets or reads is in the [server section](../server/index.md).
