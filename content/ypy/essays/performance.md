---
title: "Performance"
description: "Where a YPY page spends its time: bytes shipped, requests made and DOM swapped. The real size budgets, the knobs to turn, and how to measure with ypy bench."
order: 80
id: "GUIDE-performance"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en", "topic/performance"]
---

# Performance

On a YPY page, time goes to three places: the bytes you ship, the requests you make and the DOM you swap. Most of the
cost is usually in the network and the server, not in YPY itself. This page goes through the three in that order and
ends with how to measure, so you tune from numbers.

## Ship few bytes

| Do | Why |
|---|---|
| Load `ypy.min.js` and only the extensions you use, not `ypy.max.min.js` | the core is 18,316 B with `gzip -9`, and every extension has a 6,144 B budget; `ypy.max.min.js` is 65,716 B because it carries the official extensions that are safe to ship |
| Serve the `.br` file next to each `.min.js` (`ypy.min.js.br`) with `Content-Encoding: br` | the build writes a Brotli copy of every file, and the core is 16,512 B in that form |
| Give the script a long cache life with a hashed or versioned URL | the file only changes when you rebuild |
| Load the script in `<head>` without `async` | YPY processes the page when it runs, so a late script means a late first request |

`go run ./cmd/ypy build --sizes` prints the table and fails when the core passes 18,432 B or an extension passes
6,144 B (measured min+gzip with Go's compressor, where the core is at 18,410 B). The budget holds as the code changes.

## Ask for little

- Return a fragment, not a page. An answer to a `YPY-Request` should be the piece that changes. When a whole page
  must be narrowed down, `ypy-select` does it on the client, but rendering just the piece on the server is cheaper.
- Push instead of polling when you can. `ypy-trigger="every 5s"` costs a request each time, while `ypy-sse:connect`
  keeps one connection open and the server sends changes ([Extensions](../extensions/using-and-writing.md)). If you
  must poll, add `ypy-ptag` so the server can answer "unchanged" and the swap is skipped.
- Debounce typing. `ypy-trigger="input changed delay:300ms"` waits for the last keystroke and skips repeated values.
  Use `throttle:1s` for scroll and resize.
- Let YPY order concurrent requests. With `ypy-sync="this:replace"` (or `drop`, `abort`), someone who clicks fast
  sends one useful request and not a queue ([`ypy-sync`](../reference/attributes/ypy-sync.md)).
- Prefetch what the user is about to open. The `ypy-preload` extension fetches a link on `mousedown` or `touchstart`,
  so the click lands on an answer that is already there.
- Use the HTTP cache for what is the same for everyone. A `GET` fragment that does not depend on the user can send
  `Cache-Control: public, max-age=60` with `Vary: YPY-Request`. With an `ETag`, the browser revalidates and the server
  can answer `304` with no body.
- Keep the history cache small, or off. By default the back button asks the server for the page again
  ([History](../reference/history.md)). The `ypy-history-cache` extension shows a snapshot first, which pays off only
  where the page is expensive to render ([History and caching](../guides/history-and-caching.md)).

## Swap little

- Aim the target at the smallest element that changes: `ypy-target="#count"` for a counter, not the whole `<main>`.
  When one response must update two or three small places, use `<ypy-partial>`
  ([Multiple targets](../guides/multiple-targets.md)).
- Keep the default morph for lists and forms. It keeps the DOM nodes that did not change, so the browser does not
  recreate them and the user keeps focus and scroll. When a big region is replaced entirely, as with a new page,
  `innerHTML` is cheaper than a merge that finds nothing to keep.
- Give repeated elements a stable `id` (`id="row-42"`). The morph matches by `id`, so a reordered list moves nodes
  instead of rewriting them.
- Protect heavy widgets such as a map, a chart or an editor with `ypy-preserve`, so a swap does not rebuild them
  ([Third-party components](../guides/third-party-components.md)).
- Mark static regions with `ypy-ignore`. YPY then does not process the subtree, which adds up in a large table of
  plain cells.
- Keep your own listeners on `ypy:*` events cheap. A listener on `document` runs for every event of that name, so
  filter early and return.

## Fewer requests with the server

- Return `204` when nothing changes: no body and no swap.
- Send `YPY-Trigger` instead of making a second request when the only reason for it is to tell another part of the
  page.
- Answer the redirect after a `POST` with `303`. The browser's fetch follows it, and YPY swaps the final page in one
  round trip as the user sees it.
- Compress responses as you would any HTML, with `gzip` or `br` in your server or proxy.

## Measure

`ypy bench` runs scenarios in a real Chrome and stores JSON you can compare:

```sh
go run ./cmd/ypy bench                                   # measures dist/ypy.min.js
go run ./cmd/ypy bench --ref <commit>                    # a reference commit, built in a temporary directory
go run ./cmd/ypy bench compare bench/results/A.json bench/results/B.json
go run ./cmd/ypy bench profile swap.innerHTML.1000 --build dist/ypy.js   # CPU profile with function names
```

The scenarios cover swaps of 10, 100 and 1000 rows, morph, processing, attribute reads, the cost of one request and a
session of 500 requests with a heap check. The numbers measure script time, not layout or paint, so they show what
YPY itself costs. [The performance plan](../performance/plan.md) explains how to read a result (compare back to back;
`REGRESSION` is flagged above 5 %), and [the baseline](../performance/baseline.md) has the starting point.

For your own pages, use the browser's Performance and Network panels. Look for requests that return more than the
target needs, requests that repeat, and a `YPY-Request` answer that is slower than the page it replaces.

`ypy ci bench-smoke` is an opt-in gate. It compares deterministic metrics (counts and sizes) with the last stored
result. Time itself is not a gate, because shared machines are too noisy for it.

## Checklist

- [ ] only the extensions the page uses, with `.br` served
- [ ] polling replaced by SSE, or guarded with `ypy-ptag`
- [ ] `delay`, `throttle` and `changed` on search and scroll triggers
- [ ] `ypy-sync` on buttons that can be clicked twice
- [ ] the smallest target, and stable `id`s on list rows
- [ ] `ypy-preserve` on heavy widgets and `ypy-ignore` on static regions
- [ ] `Vary: YPY-Request` and cache headers on the fragments that can be cached
- [ ] measured with `ypy bench` or the browser

## See also

[Core concepts](../guides/core-concepts.md) · [Extensions](../extensions/using-and-writing.md) · [History and caching](../guides/history-and-caching.md) ·
[Swap styles and morph](../reference/swap-and-morph.md) · [Security checklist](../guides/security-checklist.md)
