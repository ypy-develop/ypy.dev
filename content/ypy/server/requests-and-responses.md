---
order: 10
title: "Requests and responses"
description: "Read what the YPY client sent and answer with fragments, partials and the response headers."
---

# Requests and responses

The client sends `YPY-*` headers with every request. The SDK reads them into a typed value and has one helper for
each way of answering. The names of all headers are in [Headers](../reference/headers.md).

## Read the request

```go
func (h *Items) Index(w http.ResponseWriter, r *http.Request) {
	q := ypy.FromRequest(r) // or ypy.FromContext(r.Context()) after ypy.Middleware
	if q.IsPartial() {
		ypy.Fragment(w, renderList()) // a YPY request that wants a piece of the page
		return
	}
	ypy.Fragment(w, renderPage()) // a browser request, or a history restore: the full page
}
```

| Field or method | Header | Meaning |
|---|---|---|
| `IsYPY` | `YPY-Request` | the YPY client made the request |
| `Boosted` | `YPY-Boosted` | it comes from a boosted link or form |
| `HistoryRestore` | `YPY-History-Restore-Request` | the page is restored after back or forward; answer a full page |
| `Target`, `Source` | `YPY-Target`, `YPY-Source` | the target element and the element that asked, as `tag#id` |
| `RequestType` | `YPY-Request-Type` | `ypy.RequestFull` or `ypy.RequestPartial` |
| `CurrentURL` | `YPY-Current-URL` | the page that made the request |
| `Prompt` | `YPY-Prompt` | the answer to a prompt, already decoded |
| `Validate`, `IsValidation()` | `YPY-Validate` | a field check from `ypy-validate`: change nothing |
| `Layer`, `InLayer()` | `YPY-Layer` | the request fills or starts inside a layer |
| `IsPartial()` | | a YPY request that is not a history restore and did not ask for the full content |

A history restore arrives without `YPY-Request`. Answer it with the whole page, or the user sees a bare fragment after
pressing back. `ypy.Middleware` adds `Vary: YPY-Request` to every response, so a cache never serves a fragment to a
page load.

## Answer with HTML

| Helper | What it writes |
|---|---|
| `Fragment(w, markup)` | 200 with an HTML fragment |
| `FragmentStatus(w, status, markup)` | the same, with the status you choose |
| `ValidationError(w, markup)` | 422 with the form and its errors; the client swaps it in |
| `Partial(w, PartialSpec{Target, Swap, ID}, markup)` | one `<ypy-partial>`, with attributes escaped |
| `SeeOther(w, url)` | 303 with `Location`, for a post with no YPY client |
| `NoContent(w)` | 204 |

Markup is written as it is. Escape user data when you build it (`html.EscapeString`, or a `html/template`).

```go
func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		ypy.ValidationError(w, `<form id="new"><p class="error">Title is required</p></form>`)
		return
	}
	n := h.store.Add(title)

	ypy.Fragment(w, `<li>`+html.EscapeString(title)+`</li>`) // the request target
	ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, fmt.Sprintf("%d items", n))
	ypy.Partial(w, ypy.PartialSpec{ID: "toast"}, "<p>Saved</p>") // with only an ID, the target is #toast
}
```

Call `Partial` after `Fragment`: the client swaps the main content first and then each partial. See
[`<ypy-partial>`](../reference/tags/ypy-partial.md) and [Multiple targets](../guides/multiple-targets.md).

## Answer with headers

| Helper | Header | Use |
|---|---|---|
| `Redirect(w, r, url)` | `YPY-Redirect` and 204, or 303 for other requests | send the user elsewhere with a full load |
| `Location(w, LocationOptions{...})` | `YPY-Location` | navigate with the client, no full load, with target, swap and select |
| `PushURL(w, url)`, `ReplaceURL(w, url)` | `YPY-Push-Url`, `YPY-Replace-Url` | change the address bar; `"false"` prevents it |
| `Refresh(w)` | `YPY-Refresh` | reload the page |
| `Retarget(w, sel)`, `Reswap(w, style)`, `Reselect(w, sel)` | `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect` | change where, how and what is swapped |
| `Trigger(w, name, payload)` | `YPY-Trigger` | fire a client event; calls add up in one header |
| `AtomicGroup(w)` | `YPY-Group: atomic` | swap everything or nothing; see [Atomic swaps](../reference/atomic.md) |
| `ExpireCache`, `EvictCache`, `KeepCache` | `YPY-Cache-Expire`, `YPY-Cache-Evict` | tell the [`ypy-cache`](../extensions/cache.md) extension what is stale |
| `OpenLayer`, `OpenLayerHTML` | `YPY-Open-Layer` | open a [layer](layers.md) |

```go
ypy.Retarget(w, "#errors")
ypy.Reswap(w, "innerHTML")
ypy.Trigger(w, "item:saved", map[string]any{"count": n}) // the event detail
ypy.Redirect(w, r, "/items")                              // 204 + YPY-Redirect for YPY, 303 otherwise
```

`TriggerAfterSwap` and `TriggerAfterSettle` exist in the SDK, but the 0.1.1 client does not read a header for them.
Use `Trigger` and listen for `ypy:after:swap` or `ypy:after:settle` when the code must wait. The details are in
[Headers](../reference/headers.md).

## Fragments and crawlers

`Fragment`, `FragmentStatus`, `ValidationError` and `Partial` add `X-Robots-Tag: noindex` and `Vary: YPY-Request`, so a
search engine that reaches a piece of a page does not index it. A markup that is a whole document (a doctype or an
`<html>` element) gets no `noindex`. A fragment that is a page of its own opts out with `ypy.AllowIndex()`, or for a
whole handler tree with `ypy.NewMiddleware(ypy.AllowIndex())`. `res.AssertNoIndex(t)` checks it in a test.

## One handler, two answers

A route the client pushes on the history answers a fragment to the YPY client and the full page to everyone else. The
usual defect is one answer for both: the user reloads `/items` and sees a bare list.
`ypytest.AssertFragmentAndPage` tests it; see [`ypytest`](ypytest.md).
