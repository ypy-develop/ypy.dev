---
title: "Regions that load when seen: the ypy-lazy extension"
description: "Leaves a placeholder in the page and loads the real content only when it nears the viewport, so visitors who never scroll there never request it."
order: 180
---
# Regions that load when seen: the `ypy-lazy` extension

A dashboard, a long page or a table with "show more" rows holds content that most visitors never scroll to.
`ypy-lazy` leaves a placeholder in the page and asks the server for the real content only when the placeholder is
about to come into view.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-lazy.min.js"></script>

<section ypy-lazy="/reports/summary" class="min-h-32 animate-pulse rounded bg-slate-100 p-4">
  <a href="/reports/summary" class="underline">Open the summary</a>
</section>
```

When the section is within 200 px of the window, YPY sends `GET /reports/summary` and the response replaces the
element. It is requested once.

The extension is part of `ypy.max`.

It is an attribute, never a Custom Element, so it works on `<tr>`, `<tbody>`, `<li>` and `<dd>`, where a Custom
Element is not allowed:

```html
<table>
  <tbody>
    <tr ypy-lazy="/orders?page=2"><td colspan="3"><a href="/orders?page=2">Show more orders</a></td></tr>
  </tbody>
</table>
```

The server answers with `<tr>` rows and they replace the placeholder row.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-lazy` | a URL | the `GET` made when the element is about to be seen (`data-ypy-lazy` also works) |
| `ypy-lazy-placeholder` | a selector | a `<template>` (or element) whose content replaces the children as soon as the element is processed |
| `ypy-lazy-error` | a selector | a `<template>` (or element) whose content replaces the children when the load fails |
| `ypy-lazy-retry` | none | on an element inside a failed region: loads the region again when clicked |
| `ypy-lazy-group` | a name | regions of the same group that come into view together share one request (see below) |
| `ypy-lazy-group-url` | a URL | the URL of the group request, on any region of the group |

The URL must be on the page's origin, like every YPY request. The other request attributes of the element work as on
`ypy-get`: `ypy-swap`, `ypy-target`, `ypy-select`. With `ypy-swap="innerHTML"` the element stays and is filled.

The element needs a box to be seen. An `<option>`, or anything with `display: none` or `display: contents`, never loads.
A region inside `ypy-ignore` is not started, and `ypy:before:init` fires on it before it starts, so a gate such as the
nonce extension can stop it.

## Placeholder

The content of the element is the placeholder. It is also what a visitor without JavaScript sees, so put a link there.
While the request runs the core sets the `ypy-request` class, `data-ypy-state="request"` and `aria-busy="true"`.

For a skeleton that only visitors with JavaScript get, point `ypy-lazy-placeholder` at a `<template>`:

```html
<template id="skeleton"><div class="h-24 animate-pulse rounded bg-slate-200"></div></template>
<div ypy-lazy="/panel" ypy-lazy-placeholder="#skeleton"><a href="/panel">Open the panel</a></div>
```

## Errors

If the request fails (no connection, a status of 400 or more, a swap error), the element gets
`data-ypy-state="error"` and `ypy:lazy:error` fires on it. The body of an error response is not swapped in; it
would replace the placeholder with an error page. `ypy-error-swap="outerHTML"` on the element asks for it.

```html
<template id="failed">
  <p class="text-red-700">The summary did not load.</p>
  <button ypy-lazy-retry class="underline">Try again</button>
</template>
<div ypy-lazy="/reports/summary" ypy-lazy-error="#failed">Loading...</div>
```

A click on an `ypy-lazy-retry` element loads the region again, and only while the region is in the error state; the
state is cleared when the new request starts.

## Groups: one request for several regions

A page with six report cards would send six requests as the visitor scrolls. Give the regions one `ypy-lazy-group` and
an `id` each:

```html
<div id="sales"  ypy-lazy="/reports" ypy-lazy-group="reports" class="h-40">Sales...</div>
<div id="visits" ypy-lazy="/reports" ypy-lazy-group="reports" class="h-40">Visits...</div>
```

Regions of one group that enter the view within `lazy.window` (50 ms by default) are loaded with one `GET`:

```http
GET /reports?group=reports&id=sales&id=visits HTTP/1.1
YPY-Request: true
YPY-Lazy-Group: reports
```

- The URL is the `ypy-lazy-group-url` of the first region that has one, or else the `ypy-lazy` of the first region of
  the batch. `group` and one `id` per region (in the order they came into view) are added to its query string.
- The server answers with one `<ypy-partial id="...">` per `id`. A partial with no `ypy-swap` replaces its region, the
  way an individual region would be replaced (with the region's own `ypy-swap`, `outerHTML` by default). A partial can
  name its own `ypy-swap`.
- A region the answer leaves out is an error, as if it had failed alone: `data-ypy-state="error"` and
  `ypy:lazy:error`. If the whole request fails, every region of the group is marked.
- A region without an `id`, and a region alone in its window, loads on its own with its own `ypy-lazy` URL. The
  individual mode needs nothing from the server, so the group is an optimization you can add later.
- Groups are kept apart by name: two groups in view together make two requests.

```http
HTTP/1.1 200 OK
Content-Type: text/html

<ypy-partial id="sales"><section id="sales" class="p-4">Sales: 1,204</section></ypy-partial>
<ypy-partial id="visits"><section id="visits" class="p-4">Visits: 8,310</section></ypy-partial>
```

The server answers a group with one `<ypy-partial>` per `id` it received. A partial with only an `id` targets `#id`, which is the region. Write the `id` on the card you send back too, so a
later group can find it.

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:lazy:error` | the failed region | `ctx` (the request, when there was one), `error` |
| `ypy:lazy:group` | the first region of the batch, just before the request | `group`, `ids`, `url` |

Neither is cancelable. The normal request events (`ypy:before:request`, `ypy:after:swap`...) fire on the first
region for a group request.

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `lazy.margin` | `"200px"` | how far outside the window a region is loaded, a CSS margin |
| `lazy.window` | `50` | how long a group waits for the other regions, in milliseconds or a time (`"50ms"`) |

```html
<meta name="ypy-config" content='{"lazy": {"margin": "400px", "window": 100}}'>
```

## Pitfalls

- A grouped region needs an `id`. Without one it silently loads on its own.
- The answer to a group request is read as partials. A plain fragment with no `<ypy-partial>` leaves every region in
  the error state.
- `lazy.window` starts when the first region of the group is seen. A region that comes into view after the window has
  closed belongs to the next request.
- Give the placeholder a height (`min-h-*`). Placeholders that collapse to nothing sit close together, so many of them
  come into view at once and all of them load.
- Without `IntersectionObserver` the region loads as soon as it is processed.
- A lazy region is a normal `GET`. Do not use it for an action that changes state.

The tests are in `test/tests/ext/ypy-lazy.js`.
