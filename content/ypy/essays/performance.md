---
order: 80
title: Performance
description: Where a YPY page spends its time: the bytes you ship, the requests you make and the DOM you swap, and the few attributes that reduce each one.
---

# Performance

On a YPY page, time goes to three places: the bytes you ship, the requests you make and the DOM you swap. Most of the
cost sits in the network and the server, and YPY itself is the smallest part. The levers below are ordered by how
often they matter.

## Ship few bytes

The core has a size budget of 18,432 B min+gzip and each extension has 6,144 B; `ypy build --sizes` fails when either
is exceeded. Load `ypy.min.js` and the extensions the page uses, serve the Brotli copy the build writes next to each
file, and give the script a versioned address so it can be cached for a long time. Put the tag in `<head>` without
`async`: YPY processes the page when it runs, so a late script means a late first request.

## Ask for little

An answer to a `YPY-Request` should be the piece that changes, not a page. Debounce typing and skip repeated values:

```html
<input name="q" ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms" ypy-sync="this:replace">
```

`delay` waits for the last keystroke, `changed` skips an unchanged value, and `ypy-sync="this:replace"` makes sure a
fast typist sends one useful request instead of a queue. Where the server can push, [`ypy-sse`](../extensions/sse.md)
keeps one connection open instead of a request every few seconds; where it must be polled, [`ypy-ptag`](../extensions/ptag.md)
lets the server answer "unchanged" and the swap is skipped. A `204` skips the swap altogether.

## Swap little

Aim the target at the smallest element that changes: `ypy-target="#count"` for a counter, not the whole `<main>`. When
one response must update two or three small places, use `<ypy-partial>`. Keep the default morph for lists and forms,
and give repeated elements a stable `id`, because the morph matches by `id` and a reordered list moves nodes instead of
rewriting them. Protect heavy widgets such as a map or an editor with `ypy-preserve`, and mark static regions with
`ypy-ignore` so YPY does not process them.

## Measure

In the YPY repository, `go run ./cmd/ypy bench` runs swap, morph, processing and request scenarios in a real Chrome and
stores JSON you can compare between commits. For your own pages, use the Performance and Network panels of the browser
and look for requests that return more than the target needs, requests that repeat, and a `YPY-Request` answer that is
slower than the page it replaces.

See [Swapping](../index.md#swapping), [Synchronization](../index.md#synchronization) and
[Caching](../index.md#caching) for the attributes above.
