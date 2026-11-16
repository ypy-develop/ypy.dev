---
title: "Navigation"
description: "Boost every link and form, prefetch before the click, mark the current link, draw a progress bar, keep elements across pages and detect a new deploy."
order: 250
---
# Navigation

A site that navigates with requests and swaps keeps its scripts, styles and open widgets from one page to the next,
with no router. Three official [extensions](index.md) cover it: `ypy-nav` (global boost, prefetch, the current link,
`ypy-once`), `ypy-progress` (a progress bar) and `ypy-persist` (elements kept across pages, deploy detection). Load the
regions that appear on scroll with [`ypy-lazy`](lazy.md).

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-nav.min.js"></script>
<script src="ext/ypy-progress.min.js"></script>
<script src="ext/ypy-persist.min.js"></script>
```

Nothing here is active until you load the extension and, where the text says so, turn it on. Everything is an attribute
or a setting, so it works inside `<table>`, `<tbody>` and `<ul>`, where a custom element cannot.

## Global boost

```html
<meta name="ypy-config" content='{"nav": {"boost": true}}'>
```

Now every link and form of the page navigates with a request and a swap, with no attribute on any of them. The request
is the one `ypy-boost` sends: the same headers, the body as default target, the same history entry, focus, scroll and
announcement rules and error handling. Set `ypy-target`, `ypy-swap`, `ypy-select` or `ypy-push-url` on a link to change
them. The extension listens once on the document, so links in responses and in your own scripts are covered.

The browser keeps what it does better:

| Case | Why |
|---|---|
| `target` other than `_self`, `download` | another window, or a file |
| another origin, `mailto:`, `tel:`, `javascript:` | not a page of this site |
| a link to a fragment of the same page | scrolling is native |
| a click with Ctrl, Meta, Shift, Alt or a non-main button | "open in a new tab" |
| `method="dialog"`, an unknown `method` | closes a `<dialog>`; HTML treats it as `get` |
| a click another listener already cancelled | someone else handled it |
| an element that already has `ypy-get`, `ypy-post` or `ypy-boost` | it keeps its own request |
| an invalid form | the browser shows its messages |

Forms follow HTML. `formmethod` and `formaction` of the clicked button override `method` and `action`, and `put`, `patch`
and `delete` are accepted as method values. A browser without JavaScript treats them as `get`, so the page still works.

`ypy-boost="false"` on a link, a form or an ancestor turns boost off inside it, and so does `ypy-ignore`. The core
`ypy-boost` attribute is unchanged: it boosts what it is written on, with or without `config.nav`.

Boost goes through the normal pipeline, so `config.mode` and every hook apply, and only same-origin links and forms are
taken. With `config.nonce` set, global boost is off, because it needs no attributes and the nonce gate is about
attributes. See [Security](../index.md#security).

## Prefetch

```js
ypy.config.nav = {boost: true, prefetch: 'hover'}
```

The page a link opens is requested a moment before the click, and the click uses that response.

| `nav.prefetch` | When |
|---|---|
| `"hover"` (or `true`) | the pointer rests on the link for `prefetchDelay` ms (65 by default) or the link keeps keyboard focus for as long; `touchstart` prefetches at once |
| `"visible"` | the link comes into view |
| `"hover visible"` | both |
| `false` (default) | never |

`ypy-prefetch` does the same for one link or a region, and the nearest one above the link decides:

```html
<nav ypy-prefetch="hover visible">...</nav>
<a href="/logout" ypy-prefetch="false">Sign out</a>
```

A prefetch is a plain GET issued before the user asked for anything, so a GET that changes state, such as a logout link,
needs `ypy-prefetch="false"`. Only links YPY would navigate are prefetched, never a form or form data. The request is
same-origin, uses `config.mode` and carries the headers of a boosted click plus `Purpose: prefetch` and
`YPY-Purpose: prefetch`. A script cannot set `Sec-Purpose`, so a server that wants to tell a prefetch from a visit reads
`Purpose` or `YPY-Purpose`.

The click uses the prefetched response only when its own request is the same, with the same URL and headers. A response
is used once, and one that is not a success is never kept. Nothing is prefetched when the browser asks to save data, and
everything prefetched is discarded after a boosted navigation.

| Setting | Default | Meaning |
|---|---|---|
| `prefetchDelay` | `65` | ms of hover before the request |
| `prefetchTtl` | `10000` | ms an unused response is kept |
| `prefetchMax` | `10` | responses alive at once; the oldest is dropped |
| `prefetchParallel` | `2` | requests in flight at once for `"visible"` |

Leaving the link before the delay cancels the request, and leaving it while the request is on its way aborts it.

## The current link

With `ypy-nav` loaded, a link that points to the URL of the page (same origin, path and query, no fragment) gets
`aria-current="page"` and the class `ypy-current`. The mark is renewed when the page is processed, for links that arrive
in a swap, after each history push and on back and forward.

```html
<nav>
  <a href="/orders" class="aria-[current=page]:font-semibold">Orders</a>
  <a href="/customers" class="aria-[current=page]:font-semibold">Customers</a>
</nav>
```

A link with a fragment, a `download`, another `target` or another origin is never marked. The extension owns
`aria-current="page"` on links, and leaves other values (`step`, `location`) alone. `nav.current: false` turns it off.

## `ypy-once`

A script in a swapped response runs every time the response is swapped in. `ypy-once` runs it once per page:

```html
<script ypy-once src="/assets/chart.js"></script>
<script ypy-once="map-setup">window.initMap()</script>
```

When a later response brings the same script, it is taken out before the swap. Scripts are told apart by the value of
`ypy-once`, else by `src`, else by their text, and a script that came with the page counts as run. Use a name for inline
code whose text changes between responses.

## Progress bar

A thin bar runs along the top of the page while a boosted navigation or a history restore waits for the server. It shows
after a delay of 200 ms, so a fast answer shows nothing, and it fills to the end when the last request finishes.

| To | Do |
|---|---|
| show it for the requests of one element or region | `ypy-progress` |
| hide it for a navigation | `ypy-progress="false"` |
| show it for every request | `ypy.config.progress = {all: true}` |
| change the delay or the label | `ypy.config.progress = {delay: 300, label: 'Loading'}` |

Theme it with variables on `:root`: `--ypy-progress-color` (default `#2563eb`), `--ypy-progress-height` (`3px`),
`--ypy-progress-shadow`, `--ypy-progress-z`, `--ypy-progress-max` (`90%`, how far it goes before the response) and
`--ypy-progress-duration` (`10s`).

The bar is one `<ypy-progress-bar>` element on `<html>`, so swapping the body never removes it. It lives in a shadow root
with an adopted stylesheet, so a strict CSP needs nothing extra and the page's CSS cannot reach it. It has
`role="progressbar"` and an `aria-label`, and with `prefers-reduced-motion` it does not animate.

## `ypy-persist`

An element with an `id` and `ypy-persist` stays in the page across a navigation: a video keeps playing, an iframe keeps its
state, a field keeps what was typed, a widget keeps its instance.

```html
<audio id="player" ypy-persist src="/radio.mp3" controls></audio>
```

When a navigation (a boosted link or form, back or forward) brings a response with an element of the same `id`, the live
element takes the place of the new one. The attribute goes on the live element only. Without a match in the response the
element is removed like any other, and a swap that is not a navigation never keeps anything. `ypy-persist="false"` turns
it off for one element.

It is the mechanism of [`ypy-preserve`](../index.md#preserving-content). Use `ypy-preserve` on the copy in the response
when the server decides, and `ypy-persist` on the live element when the page decides.

## `ypy-track="reload"`

A tracked element holds a value that changes when the site is deployed: a version meta tag, the `href` of the stylesheet,
the `src` of the script bundle.

```html
<meta name="app-version" content="2026.11.15-a1b2c3" ypy-track="reload">
<link rel="stylesheet" href="/assets/app-4f9d2c.css" ypy-track="reload">
```

When a full-page GET response does not carry the same tracked elements with the same values, the page was loaded before a
deploy. YPY fires the cancelable `ypy:track:reload` on the element that sent the request, with
`detail: {ctx, url, from, to}`, and unless a listener cancels it, loads `url` with a normal page load. The user lands on
the page they asked for, with the new code. A fragment, an error or the answer to a POST never triggers it.

## Without JavaScript

Links and forms keep working as HTML says, including `target`, `download`, `formaction` and `formmethod`. `ypy-persist`
and `ypy-track` are plain attributes, and a full page load replaces everything, so nothing is stale. The only loss is the
bar, the prefetch and the `put`, `patch` and `delete` methods, which act as `get`.

On the server, send `Vary: YPY-Request` when one URL answers a fragment and a page ([History](../index.md#history)), so
a prefetched response and a navigation see the same variant, and send the deploy value in every full page, including error
pages.

## Summary

| Key or attribute | Extension | Default |
|---|---|---|
| `nav.boost`, `nav.prefetch`, `nav.current` | `ypy-nav` | `false`, `false`, `true` |
| `nav.prefetchDelay`, `prefetchTtl`, `prefetchMax`, `prefetchParallel` | `ypy-nav` | `65`, `10000`, `10`, `2` |
| `progress.delay`, `progress.all`, `progress.label` | `ypy-progress` | `200`, `false`, `"Loading"` |
| `ypy-boost="false"`, `ypy-prefetch`, `ypy-once` | `ypy-nav` | |
| `ypy-progress` | `ypy-progress` | |
| `ypy-persist`, `ypy-track="reload"` | `ypy-persist` | |

| Event | When |
|---|---|
| `ypy:track:reload` | a page response carries other tracked values; cancel it to swap instead |
