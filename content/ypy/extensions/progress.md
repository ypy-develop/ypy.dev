---
order: 320
title: "Progress bar: the ypy-progress extension"
description: "A thin bar along the top of the page while a boosted navigation or an element's request waits for the server. It needs no CSS."
---

# Progress bar: the `ypy-progress` extension

When a boosted link takes a while, the page looks frozen. `ypy-progress` draws a thin bar along the top of the page
while the server answers. It needs no CSS and no configuration.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-progress.js"></script>
```

The extension is part of `ypy.max`. The short version is also in [Navigation](nav.md).

## What shows the bar

- A boosted navigation and a history restore show it.
- An element or a region with `ypy-progress` shows it for every request that element (or something inside the region)
  makes.
- `ypy-progress="false"` on an element or a region turns it off for their requests.
- `progress.all` shows it for every request.

The bar waits `progress.delay` milliseconds before it appears, so a fast answer shows nothing. It runs to 90 % while the
request is open and fills up when the last request ends. Several requests at once share one bar.

```html
<button ypy-post="/reports/build" ypy-target="#report" ypy-progress>Build the report</button>
```

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-progress` | none | show the bar for the requests of this element or region |
| `ypy-progress` | `false` | do not show it for them |

## Settings

| Setting | Default | Meaning |
|---|---|---|
| `progress.delay` | `200` | milliseconds a request waits before the bar shows |
| `progress.all` | `false` | show the bar for every request |
| `progress.label` | `Loading` | the accessible name of the bar; translate it |

## The element and its style

The bar is one `ypy-progress-bar` element appended to `<html>`, outside `<body>`, so a swap never removes it. It is a
shadow root with its own stylesheet, so the CSS of the page cannot reach it and a strict CSP is fine (no `<style>`
element and no `style` attribute). Change it with variables on `:root`:

| Variable | Default | Use |
|---|---|---|
| `--ypy-progress-color` | `#2563eb` | colour |
| `--ypy-progress-height` | `3px` | thickness |
| `--ypy-progress-shadow` | a glow in the colour | any `box-shadow` |
| `--ypy-progress-z` | `2147483000` | `z-index` |
| `--ypy-progress-max` | `90%` | how far it goes before the answer |
| `--ypy-progress-duration` | `10s` | time to get there |

```css
:root { --ypy-progress-color: #16a34a; --ypy-progress-height: 2px; }
```

## Accessibility

The bar has `role="progressbar"` and an accessible name from `progress.label`. It is in the page only while it shows.
With `prefers-reduced-motion: reduce` its transitions are off. The element that makes a request still gets
`aria-busy` from the core ([`ypy-busy`](../reference.md#core-attributes)).

## See also

- [Navigation: boost, prefetch and progress](nav.md)
- [CSS classes](../reference.md#css-classes)
