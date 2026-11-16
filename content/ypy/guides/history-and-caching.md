---
title: "History and caching"
description: "Why a cache can serve a fragment on a back-button restore, how Vary: YPY-Request prevents it, how to check it with curl, and what the client does anyway."
order: 50
---
# History and caching

When someone presses the back button, they should see the page as the server renders it now, not a stale fragment. This
guide covers the server side of that.

## The problem

A YPY request from a link or form carries `YPY-Request: true` and usually gets a fragment. The restore request after a
back button does not carry that header and must get a whole page. Both go to the same URL, and a browser or proxy cache
keys on the URL, so a fragment stored for the first request can be served for the second.

## The fix: `Vary: YPY-Request`

Send `Vary: YPY-Request` on every response whose body depends on that header, fragments and pages alike. Caches then
keep one copy per value of the header.

```go
// Go, net/http
func render(w http.ResponseWriter, r *http.Request, page, fragment templ.Component) {
	w.Header().Add("Vary", "YPY-Request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Header.Get("YPY-Request") == "true" {
		_ = fragment.Render(r.Context(), w)
		return
	}
	_ = page.Render(r.Context(), w)
}
```

Use `Header().Add` so you do not drop a `Vary` that something else set. If your pages must not be cached at all, add
`Cache-Control: no-cache` (revalidate every time) or `no-store`.

## What the client does anyway

The restore request uses `cache: "no-cache"`, so the browser asks the server before it uses a copy (see
[History](../reference/history.md)). If the answer is still not a whole page, YPY reloads the page and fires
`ypy:history:restore:fragment`. Both are fallbacks and do not replace `Vary`: a reload costs a second request, and a
proxy between you and the user may not honor `no-cache` the way the browser does.

## Checking your server

```sh
curl -sI -H 'YPY-Request: true' https://example.com/items | grep -i vary
curl -sI https://example.com/items | grep -i vary
```

Both must list `YPY-Request`. The restore request is a plain `GET` with `YPY-History-Restore-Request: true` and no
`YPY-Request`, so it gets the page.

## Tailwind

None of this depends on your CSS. Here is a history element in a Tailwind layout:

```html
<main ypy-history-elt class="mx-auto max-w-3xl p-6">
  <a href="/items" ypy-boost class="text-sky-700 underline">Items</a>
</main>
```

## The `ypy-history-cache` extension

`ypy-history-cache` restores pages from `sessionStorage` without asking the server, so what it shows can be stale by
design. It does nothing unless you load it. If you do, keep `Vary` and `no-cache` for the pages it does not cover, and
expect the saved copy of a page to be older than the server's.
