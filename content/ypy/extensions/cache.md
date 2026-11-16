---
title: "Cache with revalidation: the ypy-cache extension"
description: "Shows the saved copy of a response at once, asks the server whether it is still current and swaps again only if it is not (stale-while-revalidate)."
order: 50
---
# Cache with revalidation: the `ypy-cache` extension

`ypy-cache` shows the copy of a response the page already has, at once, then asks the server whether the copy is still
current and swaps again only if it is not (stale-while-revalidate). A click on a tab you visited a minute ago paints
immediately, and if the server says nothing changed (`304`), nothing happens on screen.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-cache.min.js"></script>

<nav>
  <a ypy-get="/reports" ypy-target="#main" ypy-cache>Reports</a>
  <a ypy-get="/people"  ypy-target="#main" ypy-cache>People</a>
</nav>
<main id="main"></main>
```

Without the extension `ypy-cache` does nothing and breaks nothing. See [Caching](../index.md#caching) for how it fits with
the HTTP cache and history snapshots.

## Opting in

Nothing is cached unless you ask, because a cache that is on by default would show stale polling and search results.

| How | Effect |
|---|---|
| `ypy-cache` on the element that sends the request | cache this element's `GET` |
| `ypy-cache:inherited` on a container | every request inside it |
| `ypy-cache="false"` | off for this element, even against `config.cache.all` |
| `ypy.config.cache = {all: true}` | every `GET` of the page |
| `ypy-cache-tags="a, b"` | labels the copy ([tags](#tags)) |

The attribute value takes options that override `ypy.config.cache`:

| Option | Default | Meaning |
|---|---|---|
| `ttl` | `10m` | a copy older than this is not shown; it is only used to ask the server with its validators |
| `fresh` | `0` | a copy younger than this is shown without asking the server (the response's `Cache-Control: max-age` wins) |
| `private` | `false` | also keep responses marked `Cache-Control: private` |
| `offline` | `false` | keep the copy in `sessionStorage` and show it when the network fails |
| `size` | `50` | most copies held; the least recently used goes first |
| `bytes` | `2097152` | most characters of response text held (2 MiB) |

```html
<a ypy-get="/reports" ypy-target="#main" ypy-cache="ttl:1h fresh:30s">Reports</a>
```

## What is kept

Only a `200` response to a `GET` to the page's own origin is kept, under the full URL (fragment excluded) plus the
`YPY-Request-Type` header. Never kept, whatever the attributes say: a response with `Cache-Control: no-store`, one with
`Vary: *`, one larger than half of `config.cache.bytes`, and anything to another origin or with another method. A response
marked `Cache-Control: private` is kept only with the `private` option. Pages made for one user usually carry `private`, so
keeping them is your decision, and calling `ypy.cache.clear()` on sign-out is your job.

## Revalidation

When a copy is shown and is not fresh, the extension asks the server after the swap, with `If-None-Match` (from `ETag`),
`If-Modified-Since` (from `Last-Modified`) and the same `YPY-*` headers as the original request. While it runs the target
has the class `ypy-revalidating`.

| The server answers | What the page does |
|---|---|
| `304 Not Modified`, or `200` with the same body | nothing; the copy is renewed |
| `200`, new body | the new content is swapped into the same target and the copy is replaced |
| another status, or blocked, or timed out | nothing on screen; on a status other than 200 or 304 the copy is dropped |

The second swap is quiet: no history entry, no scroll, no focus move, and the URL is untouched. It uses the context of the
original request, so `ypy-target`, `ypy-select`, the swap style and out-of-band elements work as before. It is dropped, and
only the copy is renewed, when the page moved on meanwhile, so a slow revalidation of tab A never paints over tab B.

## Expiry the server commands

Two response headers say that copies stopped being true, on any response. Both take a comma-separated list of URL patterns,
or `false`.

| Header | The copy | The next request |
|---|---|---|
| `YPY-Cache-Expire` | marked old, kept | shown at once, and always revalidated |
| `YPY-Cache-Evict` | deleted | goes to the server and waits, nothing is shown meanwhile |

Use `Expire` when an old screen is better than a blank one, such as a list that gained a row. Use `Evict` when the old copy
would be wrong, such as a deleted row that flashes back, or the page of another user.

```
YPY-Cache-Expire: /reports*, /dashboard
YPY-Cache-Evict: tag:private
```

| Pattern | Names |
|---|---|
| `*` | every copy |
| `/reports` | `/reports` and `/reports?anything` |
| `/reports*` | everything that starts with `/reports`, query strings included |
| `/reports/*/edit` | `*` is any run of characters, anywhere |
| `tag:products` | every copy labelled with that tag |
| `false` | nothing, and it turns off the default below |

When the response has neither header, a successful request that is not a `GET` (status below 400) evicts every copy, because
the state of the server may have changed in any way. A request with the `YPY-Validate` header (see
[`ypy-validate`](validate.md)) changes nothing. If a handler answers `303`, put the header on the final response, because
the script only sees the headers of that one.

### Tags

`ypy-cache-tags` labels the copy of the element's request so the server can name copies without knowing their URLs.

```html
<a ypy-get="/products" ypy-target="#main" ypy-cache ypy-cache-tags="products">Products</a>
<a ypy-get="/featured" ypy-target="#main" ypy-cache ypy-cache-tags="products, home">Featured</a>
```

`YPY-Cache-Expire: tag:home` marks only `/featured` old, and `tag:products` names both.

## Freshness, memory and Vary

A copy is fresh when it is younger than `fresh`, or than the `max-age` of its response. A fresh copy is shown and nothing
else happens. A copy that is not fresh, or that `YPY-Cache-Expire` marked old, is shown and revalidated. A copy older than
`ttl` is only a set of validators: on `304` it is shown, on a new `200` the new response is.

The store lives in memory and holds at most `size` copies and `bytes` characters, dropping the least recently used first. A
reload starts empty, except for offline screens.

The extension honors `Vary`. It records the value of each named request header and shows the copy only when the new request
has the same values. A server that answers differently by a header (a language, a tenant) must name it in `Vary`, as for any
HTTP cache, and a server that answers a fragment and a page at one URL sends `Vary: YPY-Request`.

## Offline screens

`ypy-cache="offline"` also writes the copy to `sessionStorage`. A screen kept this way is shown when the request cannot reach
the server, whatever its age, and `ypy:cache:hit` fires with `offline: true`.

```html
<a ypy-get="/boarding-pass" ypy-target="#main" ypy-cache="offline">Boarding pass</a>
```

This is for a few critical screens. Whatever you keep in `sessionStorage` stays readable by any script of the page, so do not
keep what a signed-out user must not see, and call `ypy.cache.clear()` on sign-out. Entries are loaded back when the
extension starts.

## Events and API

| Event | Fired on | Detail |
|---|---|---|
| `ypy:cache:hit` | the source element | `{ctx, url, age, offline}`: a copy answers the request |
| `ypy:cache:revalidated` | the source element | `{ctx, url, changed, status}`: the check ended |
| `ypy:cache:expire` | `document` | `{patterns, expired}`: copies were marked old |
| `ypy:cache:evict` | `document` | `{patterns, removed}`: copies were deleted |

A copy goes through every phase of a request as a response, including the status code rules, `YPY-Retarget`, `YPY-Reswap`,
`YPY-Reselect`, `YPY-Push-Url` and `YPY-Replace-Url`, which are stored with the copy. `YPY-Trigger`, `YPY-Redirect`,
`YPY-Refresh` and `YPY-Location` are not replayed, because they belong to one response.

```js
ypy.cache.expire('/reports*, /dashboard') // marks old; returns how many
ypy.cache.evict('tag:products')           // deletes; returns how many
ypy.cache.clear()                         // deletes every copy; call it on sign-out
ypy.cache.stats()                         // {entries, chars}
```

## Limits

Only `GET`, only `200`, only the same origin, and text only. The extension wraps `ctx.fetch`, so it works with `ypy-retry`
and `ypy-preload`, but a request another extension already answered, such as a preloaded one, is not cached. A revalidation
does not run for a request whose swap was cancelled, and its second swap needs the target to still be in the page.
