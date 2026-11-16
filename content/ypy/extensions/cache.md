---
title: "Cache with revalidation: the ypy-cache extension"
description: "Shows the saved copy of a response at once, asks the server whether it is still current and swaps again only if it is not (stale-while-revalidate)."
order: 50
---
# Cache with revalidation: the `ypy-cache` extension

`ypy-cache` shows the copy of a response the page already has, at once, then asks the server whether the copy is
still current and swaps again only if it is not (stale-while-revalidate). A click on a tab you visited a minute ago
paints immediately; if the server says nothing changed (`304`), nothing happens on screen.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-cache.js"></script>   <!-- provided by the official extension -->

<nav>
  <a ypy-get="/reports" ypy-target="#main" ypy-cache>Reports</a>
  <a ypy-get="/people"  ypy-target="#main" ypy-cache>People</a>
</nav>
<main id="main"></main>
```

If you set `ypy.config.extensions`, list `ypy-cache` in it, or the extension is refused. It is listed with the other
official extensions in [Extensions](index.md). Without the extension `ypy-cache` does nothing and breaks nothing.

Contents: [Opting in](#opting-in) · [What is kept](#what-is-kept-and-what-never-is) ·
[Revalidation](#revalidation) · [Expiry from the server](#expiry-the-server-commands) · [Tags](#tags) ·
[Freshness and memory](#freshness-and-memory) · [Vary](#vary) · [Offline screens](#offline-screens) ·
[A Go handler](#a-go-handler) · [Events](#events) · [JavaScript API](#javascript-api) · [Limits](#limits)

## Opting in

Nothing is cached unless you ask. A cache that is on by default would show stale polling results and stale search
results; an author who wants a screen cached says so.

| How | Effect |
|---|---|
| `ypy-cache` on the element that sends the request | cache this element's `GET` |
| `ypy-cache:inherited` on a container | every request inside it |
| `ypy-cache="false"` | off for this element (also against `config.cache.all`) |
| `ypy.config.cache = {all: true}` | every `GET` of the page |
| `ypy-cache-tags="a, b"` | labels the copy ([tags](#tags)); `ypy-cache-tags:inherited` on a container |

The attribute value is read as HCON, like `ypy-retry`, and its options override `ypy.config.cache`:

| Option | Default | Meaning |
|---|---|---|
| `ttl` | `10m` | a copy older than this is not shown; it is only used to ask the server with its validators |
| `fresh` | `0` | a copy younger than this is shown without asking the server (the response's `Cache-Control: max-age` wins) |
| `private` | `false` | also keep responses marked `Cache-Control: private` |
| `offline` | `false` | keep the copy in `sessionStorage` and show it when the network fails ([below](#offline-screens)) |
| `size` | `50` | most copies held; the least recently used goes first |
| `bytes` | `2097152` | most characters of response text held (2 MiB) |

```html
<a ypy-get="/reports" ypy-target="#main" ypy-cache="ttl:1h fresh:30s">Reports</a>
```

## What is kept, and what never is

Only a `200` response to a `GET` to the page's own origin is kept. Never kept, whatever the attributes say:

- a response with `Cache-Control: no-store`;
- a response with `Cache-Control: private`, unless the element has the `private` option (pages made for one user
  usually carry `private`; keeping them is your decision, and `ypy.cache.clear()` on sign-out is your job);
- a response with `Vary: *`;
- a response larger than half of `config.cache.bytes`;
- a request to another origin;
- anything but `GET`.

The key is the full URL (the fragment excluded) and the `YPY-Request-Type` header, so a partial and a full answer of
the same URL are two copies.

## Revalidation

When a copy is shown and it is not fresh, the extension asks the server after the swap, with `If-None-Match` (from
`ETag`) and `If-Modified-Since` (from `Last-Modified`), and the same `YPY-*` request headers as the original request.
While it runs the target of the request has the class `ypy-revalidating`; style it for a thin progress line or a
dimmed look.

| The server answers | What the page does |
|---|---|
| `304 Not Modified` | nothing; the copy is renewed |
| `200`, same body as the copy | nothing; the copy is renewed |
| `200`, new body | the new content is swapped into the same target, and the copy is replaced |
| any other status, or the request is blocked or times out | nothing on screen; on a status other than 200 or 304 the copy is dropped so the next visit asks first |

The second swap is a quiet one: it makes no history entry, no scroll (the position is kept), no focus
move, and the URL is not touched. It uses the context of the original request, so `ypy-target`, `ypy-select`, the
swap style (a morph keeps what the user typed and the open `<details>`) and out-of-band elements work as they did.

It is dropped, and only the copy is renewed, when the page moved on in the meantime: another request went to the
same target element, or the target left the page. So a slow revalidation of tab A never paints over tab B that the
user opened after it.

A copy older than `ttl` is not shown. The extension sends the request with the copy's validators; on `304` it shows
the copy (the server just confirmed it), on a new `200` it shows the new response. Either way no second
revalidation runs.

## Expiry the server commands

Two response headers say that copies stopped being true, on any response, a `GET` or a `POST`. Both take a comma
separated list of URL patterns, or `false`. They differ in what happens to the copy:

| Header | The copy | The next request |
|---|---|---|
| `YPY-Cache-Expire` | marked old, kept | shown at once, and always revalidated, whatever its age or `max-age` |
| `YPY-Cache-Evict` | deleted | goes to the server and waits for it, nothing is shown meanwhile |

Expiring is not removing. Use `Expire` when an old screen is better than a blank one (a list that gained a row): the
page paints at once and corrects itself. Use `Evict` when showing the old copy would be wrong, as a deleted row that
flashes back, or a page of another user.

```
YPY-Cache-Expire: /reports*, /dashboard
YPY-Cache-Evict: tag:private
```

A pattern is matched against the path and query of the copy:

| Pattern | Names |
|---|---|
| `*` | every copy |
| `/reports` | `/reports` and `/reports?anything` (a pattern without `*` is a path) |
| `/reports*` | everything that starts with `/reports`, query strings included |
| `/reports/*/edit` | `*` is any run of characters, anywhere in the pattern |
| `tag:products` | every copy labelled with that tag ([below](#tags)) |
| `false` | nothing, and it turns off the default below |

The default, when the response has neither header: a successful request that is not a `GET` (status below 400)
evicts every copy, as the state of the server may have changed in any way. `false` in either header keeps everything.
A request that carries the `YPY-Validate` header (see [`ypy-validate`](validate.md)) changes nothing and is exempt.

A copy marked old that the server then confirms with `304` becomes fresh again.

`ypy:cache:expire` fires on the document with `{patterns, expired}` and `ypy:cache:evict` with `{patterns, removed}`,
only when at least one copy was named.

### Tags

`ypy-cache-tags` labels the copy of the element's request, so the server can name copies without knowing their URLs.
The value is a comma or space separated list; `ypy-cache-tags:inherited` on a container labels every request inside it.
The tags stay on the copy after a revalidation and in a copy kept for [offline use](#offline-screens).

```html
<a ypy-get="/products" ypy-target="#main" ypy-cache ypy-cache-tags="products">Products</a>
<a ypy-get="/featured" ypy-target="#main" ypy-cache ypy-cache-tags="products, home">Featured</a>
```

```
YPY-Cache-Expire: tag:home
```

marks only `/featured` old; `tag:products` names both.

## Freshness and memory

A copy is fresh when it is younger than `fresh`, or than the `max-age` of its response when it has one.
A fresh copy is shown and nothing else happens. A copy that is not fresh, or that `YPY-Cache-Expire` marked old, is shown and revalidated. A copy older
than `ttl` is only a set of validators.

The store holds at most `size` copies and `bytes` characters of text. Showing a copy moves it to the end of the
line; when a limit is crossed the copies at the front go first (least recently used). The store lives in memory:
a reload starts empty, except for [offline screens](#offline-screens).

## Vary

The extension honors `Vary`. It records the value of every named request header at the time of the response and
shows the copy only when the new request has the same values. A server that answers differently by a header (a
language, a tenant) must name it in `Vary`, as it must for any HTTP cache. The Go SDK's `ypy.Middleware` adds
`Vary: YPY-Request`.

## Offline screens

`ypy-cache="offline"` (or `offline` in `config.cache`) also writes the copy to `sessionStorage`, key prefix
`ypy-cache:`. A screen kept this way is shown when the request cannot reach the server, whatever its age, and
`ypy:cache:hit` fires with `offline: true`. The page's [offline state](../reference/events/ypy-offline.md) is unchanged: the
`ypy-offline` class stays on `<html>` while the device is offline, and the copy is not revalidated after a network
failure.

```html
<a ypy-get="/boarding-pass" ypy-target="#main" ypy-cache="offline">Boarding pass</a>
```

This is for a few critical screens, not for the site. `sessionStorage` belongs to the tab and ends with it, but
whatever you keep there stays readable by any script of the page until then: do not keep what a signed-out user
must not see, and call `ypy.cache.clear()` when the user signs out (a successful `POST` already clears it by
default). Entries are loaded back when the extension starts, so a reload of the tab still has them.

## A Go handler

The handler sets `ETag`, answers `304` to a matching `If-None-Match`, and tells the cache when a write makes
`/reports` old. `ypy.Middleware` adds `Vary: YPY-Request`.

```go
package reports

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	ypy "github.com/ypy-develop/go"
)

type Handler struct {
	Render func(r *http.Request) (string, error) // the fragment of the page
	Save   func(r *http.Request) error
}

// Show answers GET /reports. A copy with the same tag is not sent again.
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	body, err := h.Render(r)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	sum := sha256.Sum256([]byte(body))
	etag := `"` + hex.EncodeToString(sum[:8]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache") // keep it, but always ask
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	ypy.Fragment(w, body)
}

// Create answers POST /reports with the new list, and marks every copy of the reports old.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := h.Save(r); err != nil {
		ypy.ValidationError(w, "<p>Could not save</p>")
		return
	}
	body, err := h.Render(r)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("YPY-Cache-Expire", "/reports*")
	ypy.Fragment(w, body)
}

func Wire(mux *http.ServeMux, h *Handler) http.Handler {
	mux.HandleFunc("GET /reports", h.Show)
	mux.HandleFunc("POST /reports", h.Create)
	return ypy.Middleware(mux)
}
```

The Go SDK has `ypy.ExpireCache(w, "/reports*")` for `YPY-Cache-Expire` and `ypy.KeepCache(w)` for `false`. To
delete instead of marking old, set `YPY-Cache-Evict` yourself: `w.Header().Set("YPY-Cache-Evict", "/reports*")`.

Put the header on the response the page receives: when a handler answers `303 See Other`, the browser follows
the redirect and the script only sees the headers of the final response (without the header, the default of a
successful `POST` applies and every copy is deleted).

For a page made for one user, add `Cache-Control: private` (it is then never kept unless the element has the
`private` option), or send `YPY-Cache-Evict: *` on sign-in and sign-out.

## Events

| Event | Fired on | Detail |
|---|---|---|
| `ypy:cache:hit` | the source element | `{ctx, url, age, offline}`: a copy answers the request (`age` in ms) |
| `ypy:cache:revalidated` | the source element | `{ctx, url, changed, status}`: the check ended; `changed` is true when it swapped new content |
| `ypy:cache:expire` | `document` | `{patterns, expired}`: copies were marked old |
| `ypy:cache:evict` | `document` | `{patterns, removed}`: copies were deleted (a header, the default after a non-`GET`, `ypy.cache.evict()` or `clear()`) |

A copy goes through every phase of a request as a response: `ypy:before:request`, `ypy:after:request`,
`ypy:before:swap`, `ypy:after:swap`, the status code rules, `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect`,
`YPY-Push-Url` and `YPY-Replace-Url` (stored with the copy). `YPY-Trigger`, `YPY-Redirect`, `YPY-Refresh` and
`YPY-Location` are not replayed: they are events of one response, not part of its content.

## JavaScript API

```js
ypy.cache.expire('/reports*, /dashboard') // marks old (still shown, then revalidated); returns how many
ypy.cache.evict('tag:products')           // deletes; returns how many
ypy.cache.clear()                         // deletes every copy; call it when the user signs out
ypy.cache.stats()                         // {entries, chars}
```

## Limits

- Only `GET`, only `200`, only the same origin.
- The extension wraps `ctx.fetch` from `ypy_before_request`. It works together with `ypy-retry` and
  `ypy-preload`, but a request that another extension has already answered (a preloaded one) is not cached.
- A revalidation does not run for a request whose swap was cancelled, and its second swap needs the target
  element to still be in the page. An `outerHTML` swap of the target itself (the target is replaced) renews the copy
  and leaves the page alone.
- The first display of a copy is a normal swap; a swap with `swap:` or `settle:` delays waits as usual.
- Text only: the copy is the response text, kept as a string; binary bodies are not supported.
- Size: under 2.5 kB min+gzip (`ypy build --sizes`).
