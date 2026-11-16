---
title: "Navigation"
description: "Navigation without a router: boosted links and forms, prefetch before the click, current-link marking, a progress bar, persistent elements and deploy detection."
order: 250
---
# Navigation

A site that navigates with requests and swaps keeps its scripts, styles and open widgets from one page to the next,
and it needs no router to do it. Four official [extensions](index.md) handle the rest. They boost every link and form,
fetch the next page before the click, mark the current link and draw a progress bar. They also keep chosen elements
alive across pages, notice a new deploy and load regions when they scroll into view.

Each extension is optional. None does anything until you load it and, where the text says so, turn it on.

| Extension | File | What it adds |
|---|---|---|
| `ypy-nav` | `ext/ypy-nav.js` | global boost (`config.nav.boost`), prefetch (`config.nav.prefetch`, `ypy-prefetch`), the current link (`aria-current`, `ypy-current`) and `ypy-once` |
| `ypy-progress` | `ext/ypy-progress.js` | the progress bar (`ypy-progress`, `--ypy-progress-*`) |
| `ypy-persist` | `ext/ypy-persist.js` | `ypy-persist` and `ypy-track="reload"` |
| `ypy-lazy` | `ext/ypy-lazy.js` | `ypy-lazy` regions, alone or in groups ([lazy](lazy.md)) |

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-nav.js"></script>
<script src="ext/ypy-progress.js"></script>
<script src="ext/ypy-persist.js"></script>
<script src="ext/ypy-lazy.js"></script>
```

If you set `ypy.config.extensions`, list the names you load (`ypy-nav`, `ypy-progress`, `ypy-persist`,
`ypy-lazy`) or they are refused. The files measure 2,923 bytes (`ypy-nav`), 932 (`ypy-progress`), 731 (`ypy-persist`) and 1,497 (`ypy-lazy`)
min+gzip, and each uses only the public API and the [extension API](index.md#writing-an-extension), so each also works from
`dist/`.

Everything is an attribute or a configuration key. None of it needs a Custom Element, so it works inside
`<table>`, `<tbody>`, `<ul>` and the like.

## Global boost

```html
<meta name="ypy-config" content='{"nav": {"boost": true}}'>
```

or `ypy.config.nav = {boost: true}` from a script. Now every link and form of the page is navigated with a
request and a swap, with no attribute on any of them. The request is the one an element with `ypy-boost`
sends: the same headers (`YPY-Request`, `YPY-Boosted`, `YPY-Target`...), the same default target (the body),
the same history entry, the same focus, scroll and announcement rules, the same error handling. Set
`ypy-target`, `ypy-swap`, `ypy-select` or `ypy-push-url` on a link to change them, as on any element.

The extension listens once on the document, so links and forms that arrive in a response or from your own
script are covered without any processing.

### What the browser keeps

Boost leaves to the browser everything the browser does better, and the click or submit goes through as if
YPY were not there:

| Case | Why |
|---|---|
| `target` other than `_self` (and `formtarget` on a submitter) | another window or frame |
| `download` | a file, not a page |
| a link or action on another origin, `mailto:`, `tel:`, `javascript:` | not a page of this site |
| a link to a fragment of the same page (`#section`, or the same URL with a hash) | scrolling is native |
| a click with Ctrl, Meta, Shift or Alt, or with a button other than the main one | "open in a new tab" and friends |
| `method="dialog"` or `formmethod="dialog"` | closes a `<dialog>` |
| an unknown `method` value | HTML treats it as `get` |
| a click or submit that another listener already cancelled | someone else handled it |
| an element YPY already runs (`ypy-get`, `ypy-post`, `ypy-boost`...) | it keeps its own request |
| an invalid form | the browser shows its messages and does not submit |

### Forms follow HTML

The method and the action of a form are the HTML ones. The submitter wins over the form:

```html
<form action="/items/7" method="post">
  <button>Save</button>
  <button formmethod="delete" formaction="/items/7/remove">Remove</button>
</form>
```

- `formmethod` and `formaction` of the clicked button override `method` and `action`.
- `put`, `patch` and `delete` are accepted as `method` and `formmethod` values. A browser without JavaScript
  treats them as `get`, so the page still works as a plain form.
- No `action` means the current URL. A `get` form puts its fields in the query string.
- `target`/`formtarget`, `method="dialog"` and an action on another origin stay with the browser.

### Turning it off

`ypy-boost="false"` on a link, a form or any ancestor turns boost off for everything inside it (also
`ypy-boost:inherited="false"` and `data-ypy-boost="false"`). So does `ypy-ignore`.

```html
<aside ypy-boost="false">
  <a href="/export.csv">Export</a>    <!-- a normal link -->
</aside>
```

The core `ypy-boost` attribute is unchanged: it still boosts what it is written on, with or without
`config.nav`.

### Security

- The request goes through the normal pipeline, so `selfRequestsOnly` / `config.mode`, the `javascript:`
  guard and every extension hook apply. Only same-origin links and forms are taken.
- With `config.nonce` set, global boost is off: it needs no attributes, and the nonce gate is about
  attributes. Use `ypy-boost` with `ypy-nonce` where you need a boosted element under the gate. See
  [Security](../reference/security.md).

## Prefetch

```js
ypy.config.nav = {boost: true, prefetch: 'hover'}
```

The page a link opens is requested a moment before the click, and the click uses that response. Values of
`nav.prefetch`:

| Value | When |
|---|---|
| `"hover"` (or `true`) | the pointer rests on the link for `prefetchDelay` ms (65 by default), or the link gets keyboard focus (Tab) and keeps it for the same time; `touchstart` prefetches at once |
| `"visible"` | the link comes into view |
| `"hover visible"` | both |
| `false` (default) | never |

An attribute does the same for one link or a region (the nearest `ypy-prefetch` above the link, or on it,
decides):

```html
<nav ypy-prefetch="hover visible">...</nav>
<a href="/reports" ypy-prefetch>Reports</a>          <!-- bare means hover -->
<a href="/logout" ypy-prefetch="false">Sign out</a>
<section ypy-prefetch="false">...</section>          <!-- nothing in here -->
```

Only prefetch what is safe to ask for twice. A prefetch is a plain `GET`, issued before the user has asked for
anything. A `GET` that changes state, such as a logout link or an "unsubscribe" link, needs
`ypy-prefetch="false"`.

### What is prefetched

Only a link YPY would navigate: the same cases as the table above (own window, same origin, not a fragment,
not a download), a link inside a region that did not turn boost off, and either global boost or a core
`ypy-boost`. Never a form, and never form data:

- `GET` only, and only on the origin of the page;
- the fetch `mode` of the page (`config.mode`) and `credentials: "same-origin"`. An `ypy-boost` value or
  `ypy-config` on the link cannot change them, and an action rewritten to another origin is not requested;
- no body, no query from the form the link sits in, and no `ypy-vals`.

The request carries the headers a boosted click carries, plus `Sec-Purpose: prefetch`, `Purpose: prefetch`
and `YPY-Purpose: prefetch`. A script cannot set `Sec-Purpose` (the browser drops every header that begins
with `Sec-`), so a server that wants to tell a prefetch from a visit should read `Purpose` or `YPY-Purpose`.

### When the response is used

The click uses the prefetched response only when its own request is the same: the same URL and the same
headers (the source element, the target and the current URL are part of them). When anything differs, for
example an `ypy-headers` on the link or a header an extension adds, the click asks the server normally. A
response is used once. A response that is not a success is never kept.

### Budget and cancellation

| Setting | Default | Meaning |
|---|---|---|
| `prefetchDelay` | `65` | ms of hover before the request |
| `prefetchTtl` | `10000` | ms an unused response is kept; a request still on its way at that time is aborted |
| `prefetchMax` | `10` | responses alive at once (in flight or kept); the oldest is dropped, and aborted if it is still on its way |
| `prefetchParallel` | `2` | requests in flight at once for `"visible"`; the others wait their turn |

- Leaving the link before the delay ends cancels the request. Leaving it while the request is still on its
  way aborts it. The same holds for the keyboard: moving the focus away cancels a prefetch that focus started.
  A finished response stays for `prefetchTtl`.
- A focus that comes from the pointer (a click, a tap) is left to the hover and the touch, so one visit does not
  request twice.
- Nothing is prefetched when the browser asks to save data (`navigator.connection.saveData`).
- After a boosted navigation everything that was prefetched is discarded: it was built for the old page.

`ypy-preload` (the older extension) prefetches `ypy-get` elements. For navigation use `ypy-prefetch`; the two
do not need each other.

## The current link

With `ypy-nav` loaded, a link that points to the URL of the page (same origin, same path and query, no fragment)
gets `aria-current="page"` and the class `ypy-current`. Nothing is written in the markup, and the mark is renewed
on its own:

- when the page is processed, and for every link in content that arrives by a swap;
- after each history push or replace, so the links outside the swapped region move too;
- on back and forward.

```html
<nav class="flex gap-4">
  <a href="/orders" class="text-slate-600 aria-[current=page]:font-semibold aria-[current=page]:text-slate-900">Orders</a>
  <a href="/customers" class="text-slate-600 aria-[current=page]:font-semibold aria-[current=page]:text-slate-900">Customers</a>
</nav>
```

On `/orders` the first link has `aria-current="page"`. Style it with the attribute selector (above, with a Tailwind
variant) or with `.ypy-current`.

- A link with a fragment (`/orders#top`), a `download`, a `target` other than `_self`, or another origin is never
  marked.
- The extension owns `aria-current="page"` on links: a server-rendered one is moved when the page changes. Other
  values (`step`, `location`) are left alone.
- `nav.current: false` turns the feature off.

## `ypy-once`

A script that comes in a swapped response runs every time the response is swapped in. Add `ypy-once` to run it a
single time per page:

```html
<script ypy-once src="/assets/chart.js"></script>
<script ypy-once="map-setup">window.initMap()</script>
```

When a later response brings the same script again, it is taken out before the swap, so it does not run twice. The
script is told apart by the value of `ypy-once` when it has one, else by its `src`, else by its text. A script that
came with the page counts as already run. A script without the attribute is not touched.

Use a name (`ypy-once="map-setup"`) for inline code whose text changes between responses: the name makes it the
same script.

## Progress bar

```html
<script src="ext/ypy-progress.js"></script>
```

A thin bar runs along the top of the page while a boosted navigation (or a history restore) waits for the
server. It shows only after a delay (200 ms), so a fast answer shows nothing, and it fills to the end when
the last request finishes. No CSS is required.

| To | Do |
|---|---|
| show it for the requests of one element or region | `ypy-progress` |
| hide it for a navigation | `ypy-progress="false"` on the element or a region |
| show it for every request | `ypy.config.progress = {all: true}` |
| change the delay or the name | `ypy.config.progress = {delay: 300, label: 'Chargement'}` |

Theme it with CSS variables on `:root`:

```css
:root {
  --ypy-progress-color: #0ea5e9;      /* default #2563eb */
  --ypy-progress-height: 4px;         /* 3px */
  --ypy-progress-shadow: none;        /* a glow in the same color */
  --ypy-progress-z: 50;               /* 2147483000 */
  --ypy-progress-max: 80%;            /* how far it goes before the response, 90% */
  --ypy-progress-duration: 6s;        /* time to get there, 10s */
}
```

The bar is one `<ypy-progress-bar>` element appended to `<html>`, so a swap of the body never removes it. It is
a shadow root with its own stylesheet, adopted with the Constructable Stylesheet API: there is no `<style>`
element and no `style` attribute, so a strict CSP needs nothing extra, and the CSS of the page, the reset of
Tailwind included, cannot reach it. Because the host is a child of `<html>`, set the variables on `:root`.
It has `role="progressbar"` and an `aria-label` (`"Loading"`, or `config.progress.label`); it is in the page
only while it shows. With `prefers-reduced-motion` it does not animate. Pair it with
[`ypy-announce`](../reference/attributes/ypy-announce.md), which tells screen reader users what changed.

## `ypy-persist`

An element with an `id` and `ypy-persist` stays in the page across a navigation. A video keeps playing, an
iframe keeps its state, a field keeps what was typed, a third-party widget keeps its instance.

```html
<audio id="player" ypy-persist src="/radio.mp3" controls></audio>
```

When a navigation (a boosted link or form, back or forward) brings a response with an element of the same
`id`, the live element takes the place of the new one. The attribute goes on the live element only; the
server does not have to mark the next page. If the response has no such `id`, the element is removed like
any other. A swap that is not a navigation (a button with `ypy-get`, an out-of-band update) never keeps
anything. `ypy-persist="false"` turns it off for an element.

It is the same mechanism as [`ypy-preserve`](../reference/attributes/ypy-preserve.md), which you can also use: the node is
moved, not recreated. Use `ypy-preserve` on the copy in the response when the server decides; use
`ypy-persist` on the live element when the page decides. An element the navigation never swaps (a header
outside the target) needs neither.

## `ypy-track="reload"`

A tracked element names a value that changes when the site is deployed: a version meta tag, the `href` of
the stylesheet, the `src` of the script bundle.

```html
<head>
  <meta name="app-version" content="2026.11.15-a1b2c3" ypy-track="reload">
  <link rel="stylesheet" href="/assets/app-4f9d2c.css" ypy-track="reload">
  <script src="/assets/app-8c1a77.js" ypy-track="reload"></script>
</head>
```

When a response that is a whole page does not carry the same tracked elements with the same values (the
`content`, `href`, `src`, `value` or text; an element added or gone counts), the page was loaded before a
deploy. YPY fires the cancelable `ypy:track:reload` event on the element that sent the request, with
`detail: {ctx, url, from, to}`, and unless a listener cancels it, loads `url` with a normal page load
instead of swapping the response. The user lands on the page they asked for, with the new code.

```js
document.addEventListener('ypy:track:reload', evt => {
  // evt.detail.url: where the page loads; evt.preventDefault() swaps the response instead
})
```

It applies only to a `GET` whose response is a page (it begins with a doctype, `<html>`, `<head>` or `<body>`)
and succeeded. A fragment, an error and the answer to a `POST` never trigger it, so a deploy is noticed on the
next navigation. A page that tracks nothing and a response that tracks nothing cost one attribute lookup.

## `ypy-lazy`

```html
<section ypy-lazy="/reports/summary" class="min-h-32 animate-pulse">
  <a href="/reports/summary">Open the summary</a>
</section>
```

When the element is about to be seen (`config.lazy.margin`, `"200px"` by default) YPY asks for the URL with a
`GET` and the response replaces the element (`outerHTML`). `ypy-swap`, `ypy-target`, `ypy-select` and
the other request attributes of the element work as on `ypy-get`: `ypy-swap="innerHTML"` keeps the element
and fills it. It is requested once.

It is an attribute, so it works where a Custom Element cannot:

```html
<table>
  <tbody>
    <tr ypy-lazy="/orders?page=2"><td colspan="3"><a href="/orders?page=2">Show more orders</a></td></tr>
  </tbody>
</table>
```

The response is `<tr>` rows; they replace the placeholder row. The element needs a box to be seen: an
`<option>`, or anything with `display: none` or `display: contents`, never loads.

### Placeholder

The content of the element is the placeholder, shown until the response arrives. It is also what a visitor
without JavaScript sees, so put a link there. While the request runs the core marks the element
(`ypy-request` class, `data-ypy-state="request"`, `aria-busy="true"`), which is enough to style a skeleton.

For a skeleton that only JavaScript visitors get, `ypy-lazy-placeholder` names a `<template>` (or an element)
whose content replaces the children as soon as the element is processed:

```html
<template id="skeleton"><div class="h-24 animate-pulse rounded bg-slate-200"></div></template>
<div ypy-lazy="/panel" ypy-lazy-placeholder="#skeleton"><a href="/panel">Open the panel</a></div>
```

### Errors

When the request fails (no connection, a status of 400 or more, a swap error), the element gets
`data-ypy-state="error"` and the event `ypy:lazy:error` fires on it, with `{ctx, error}`. The
body of an error response is not swapped in (it would replace the placeholder with an error page);
`ypy-error-swap="outerHTML"` asks for it. `ypy-lazy-error` names a `<template>` whose content replaces the
children, and a `ypy-lazy-retry` element in it loads the region again when clicked:

```html
<template id="failed">
  <p class="text-red-700">The summary did not load.</p>
  <button ypy-lazy-retry class="underline">Try again</button>
</template>
<div ypy-lazy="/reports/summary" ypy-lazy-error="#failed">Loading...</div>
```

A region inside `ypy-ignore` is not started. The `ypy:before:init` event fires on a lazy region before it
starts, as on any element YPY initializes, so a gate such as the nonce extension can stop it.

### Groups

Regions with the same `ypy-lazy-group` that come into view within `lazy.window` (50 ms) share one `GET`, which
carries `?group=<name>&id=<id>&id=<id>` and the header `YPY-Lazy-Group`, and the server answers with one
`<ypy-partial id="...">` per region. `ypy-lazy-group-url` names the URL of that request, and `ypy:lazy:group` fires
before it is sent. A region without an `id` loads alone. The whole feature, with the server side, is in
[Lazy regions](lazy.md).

## Progressive enhancement

Without JavaScript, or without the extensions, nothing here changes how the page works:

- links and forms are links and forms: they navigate and submit as HTML says, including `target`,
  `download`, `formaction` and `formmethod`;
- `ypy-lazy` regions show their content, which should hold a link to the same URL;
- `ypy-persist` and `ypy-track` are plain attributes; a full page load replaces everything, so nothing is
  stale;
- no bar, no prefetch. The `put`, `patch` and `delete` values of `method` act as `get`.

The tests run these cases with the extensions absent (`test/tests/ext/ypy-nav-enhancement.js`).

## Server notes

- A server that answers `YPY-Request: true` with a fragment and a plain visit with a page must say so with
  `Vary: YPY-Request` ([History and caching](../guides/history-and-caching.md)). Boost and prefetch send
  `YPY-Request: true`, so a prefetched response and a navigation see the same variant.
- Read `Purpose: prefetch` or `YPY-Purpose: prefetch` to skip analytics or per-user side effects for a
  prefetch. Do not rely on `Sec-Purpose`.
- Send the deploy value (`ypy-track`) in every full page, including the error pages that users can reach by a
  navigation.

## Configuration summary

| Key | Extension | Default |
|---|---|---|
| `nav.boost` | `ypy-nav` | `false` |
| `nav.prefetch` | `ypy-nav` | `false` |
| `nav.prefetchDelay`, `nav.prefetchTtl`, `nav.prefetchMax`, `nav.prefetchParallel` | `ypy-nav` | `65`, `10000`, `10`, `2` |
| `progress.delay`, `progress.all`, `progress.label` | `ypy-progress` | `200`, `false`, `"Loading"` |
| `lazy.margin`, `lazy.window` | `ypy-lazy` | `"200px"`, `50` |
| `nav.current` | `ypy-nav` | `true` |

| Attribute | Extension | Where |
|---|---|---|
| `ypy-boost="false"` | `ypy-nav` | a link, a form or a region |
| `ypy-prefetch` (`hover`, `visible`, `false`) | `ypy-nav` | a link or a region |
| `ypy-progress` (`false`) | `ypy-progress` | an element or a region |
| `ypy-persist` (`false`) | `ypy-persist` | an element with an `id` |
| `ypy-track="reload"` | `ypy-persist` | an element of the page, usually in `<head>` |
| `ypy-lazy`, `ypy-lazy-placeholder`, `ypy-lazy-error`, `ypy-lazy-retry`, `ypy-lazy-group`, `ypy-lazy-group-url` | `ypy-lazy` | a region |
| `ypy-once` (a name) | `ypy-nav` | a `<script>` in swapped content |

| Event | Extension | When |
|---|---|---|
| `ypy:track:reload` | `ypy-persist` | a page response carries other tracked values; cancel it to swap instead |
| `ypy:lazy:error` | `ypy-lazy` | a lazy region failed to load |
| `ypy:lazy:group` | `ypy-lazy` | the request of a group of lazy regions is about to be sent |
