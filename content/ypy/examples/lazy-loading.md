---
order: 60
title: Lazy loading
description: Load a slow chart or report after the page is shown: on load, when it scrolls into view, or with the ypy-lazy extension and a placeholder.
---

# Lazy loading

Some parts of a page are slow or rarely seen: a chart, a report, a list of recent activity. The page is sent without
them, with a placeholder in their place, and the browser asks for the content afterwards. This page shows three ways:
at load, when the part scrolls into view, and with the `ypy-lazy` extension.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-lazy.min.js"></script>

<!-- 1. as soon as the page is processed -->
<section ypy-get="/reports/summary" ypy-trigger="load" ypy-swap="outerHTML">
  <p>Loading the summary...</p>
</section>

<!-- 2. when it scrolls into view -->
<section ypy-get="/reports/activity" ypy-trigger="revealed" ypy-swap="outerHTML">
  <p>Recent activity appears here.</p>
</section>

<!-- 3. the extension: one attribute, the request starts a little before the element is visible -->
<section ypy-lazy="/reports/map" class="min-h-64">
  <a href="/reports/map">Open the map</a>
</section>
```

The `ypy-lazy` script is only needed for the third block. Its URL is `dist/ext/ypy-lazy.min.js` next to the core.

## The server (Go)

```go
package main

import (
	"net/http"
	"time"

	ypy "github.com/ypy-develop/go"
)

func reports(mux *http.ServeMux) {
	mux.Handle("GET /reports/summary", ypy.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(800 * time.Millisecond) // stands for a slow query
		ypy.Fragment(w, `<section><h2>Summary</h2><p>1,204 sales this week.</p></section>`)
	})))
	mux.Handle("GET /reports/activity", ypy.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ypy.Fragment(w, `<section><h2>Activity</h2><ul><li>Ada paid an invoice</li></ul></section>`)
	})))
	mux.Handle("GET /reports/map", ypy.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ypy.Fragment(w, `<section><h2>Map</h2><p>The map goes here.</p></section>`)
	})))
}
```

The same URLs work as full pages for a browser without the script: the placeholder in block 3 holds a real link, and
you can answer a request that is not a YPY request (`ypy.FromRequest(r).IsYPY` is false) with a page.

## How it works

- `ypy-trigger="load"` fires when YPY processes the element, which is when the page is ready. `ypy-swap="outerHTML"`
  replaces the placeholder with the answer. Without a target the element itself is the target.
- `revealed` fires once, when the element scrolls into the viewport. The observer is disconnected afterwards. The
  element is replaced by the answer, so the request cannot repeat.
- While a request is in flight YPY sets the `ypy-request` class on the element, so you can style the placeholder, and
  `aria-busy="true"` so assistive technology knows the region is loading. See [`ypy-busy`](../reference.md#additional-attributes).
- `ypy-lazy` is a short form of "get this URL when the element is about to be seen". The request is a `GET` and the
  answer replaces the element. It also supports a `<template>` skeleton, an error template with a retry button, and
  one request for several regions. See [`ypy-lazy`](../extensions/lazy.md).
- Give the placeholder a height (`min-h-64` here), or many placeholders collapse and come into view together.

## Variations

- `ypy-trigger="intersect once threshold:0.5"` fires when half of the element is visible. `revealed` is the same
  without the options.
- A lazy request is a normal `GET`. Do not use it for an action that changes state.

## See also

- [Click to load](click-to-load.md) and [Infinite scroll](infinite-scroll.md)
- [`ypy-lazy`](../extensions/lazy.md)
- [`ypy-trigger`](../reference.md#core-attributes)
