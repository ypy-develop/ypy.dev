---
order: 60
title: CSS classes
description: "The classes and state attributes that YPY and its extensions put on elements, and the classes of the optional ypy.css stylesheet."
---

# CSS classes

YPY writes no inline style. It marks the states of an element with classes that start with `ypy-` and with the
`data-ypy-state` attribute, and you style those. This page is the list. How to style them with Tailwind,
Bootstrap and plain CSS is in [Styling](styling.md); how to rename them is in
[`config.classes`](config/classes.md).

## Classes of the core

These are the defaults of `config.classes`. Each entry can be renamed, or emptied with `""` to use
`data-ypy-state` only.

| Class | Put on | While |
|---|---|---|
| `ypy-request` | The element that sends the request and its indicators | A request is in flight. |
| `ypy-swapping` | The swap target | From the start of the swap to the swap itself. It is long enough to see only with a swap delay, for example `ypy-swap="innerHTML swap:200ms"`. |
| `ypy-settling` | The swap target | The settle phase, until it completes. |
| `ypy-added` | New content | After it is inserted and before it settles. |
| `ypy-indicator` | Elements that show only during a request | Always. The built-in CSS hides it until a containing or matching element has `ypy-request`. |
| `ypy-offline` | `<html>` | The browser reports no connection. `data-ypy-state="offline"` is set with it. |

`config.classes.error` has no default class. Set it to add a class to the target of an error response;
`data-ypy-state="error"` is there without it.

```html
<button ypy-post="/save" ypy-indicator="#spinner">Save</button>
<span id="spinner" class="ypy-indicator">Saving...</span>
```

```css
/* the button while its request runs */
button.ypy-request { opacity: .6; }

/* an error response was swapped into this panel */
#panel[data-ypy-state~="error"] { border-color: crimson; }

/* the whole page while offline */
html.ypy-offline body::before { content: "You are offline"; }
```

Select states with `~=` because an element can be in more than one: `[data-ypy-state~="error"]`.

## Classes added by extensions

| Class | Extension | Put on | While |
|---|---|---|---|
| `ypy-revalidating` | [`ypy-cache`](../extensions/cache.md) | The target of a request | The extension checks with the server a copy it already showed. |
| `ypy-layer-opener` | [`ypy-layers`](../extensions/layers.md) | The element that opened a popup | The popup is open, so the default style can anchor the popup below it. |
| `ypy-current` | [`ypy-nav`](../extensions/nav.md) | A link to the current page | Always while the URL matches (same path and query, no fragment). It comes with `aria-current="page"` and is renewed on every navigation. |
| `ypy-optimistic` | [`ypy-optimistic`](../extensions/optimistic.md) | The optimistic nodes | The request is in flight. |
| `ypy-pending` | [`ypy-pending`](../extensions/pending.md) | The placeholder that `ypy-pending` inserts | Until the request ends. |
| `ypy-dirty` | [`ypy-signals`](../extensions/signals.md) | A signals scope | A client-owned signal differs from its last saved value. The scope also gets `dirty` in `data-ypy-state`. |
| `ypy-field-error` | [`ypy-validate`](../extensions/validate.md) | The answer element created after a field | The page has no answer element for that field. |

## The optional `ypy.css`

YPY works without any stylesheet. `ypy.css` adds a small set of styles for the classes above and for a few
more, all themed by custom properties. Load it if you want them:

```html
<link rel="stylesheet" href="/vendor/ypy.css">
```

| Class | What it does |
|---|---|
| `ypy-indicator` | Hidden until the element, or one of its parents, has `ypy-request`. It keeps its room, so nothing moves when it appears. |
| `ypy-progress` | A fixed bar at the top of the page, invisible until it has the `request` state, then it runs. Use it as `<div class="ypy-progress" role="progressbar">`. |
| `ypy-fade` | Opt-in transition. The swap target fades out while it is `swapping`. Use a swap delay such as `swap:200ms`. |
| `ypy-fade-in` | Opt-in transition. New content fades in, starting from `@starting-style`. |

Theme with custom properties on `:root` or on any element:

| Property | Default | What it sets |
|---|---|---|
| `--ypy-duration` | `200ms` | Every fade. |
| `--ypy-easing` | `ease-in-out` | Every fade. |
| `--ypy-progress-color` | `AccentColor` | The colour of the bar. |
| `--ypy-progress-height` | `3px` | The thickness of the bar. |
| `--ypy-progress-duration` | `1.2s` | How long the bar takes to reach 90 percent. |
| `--ypy-progress-z` | `9999` | The stacking order of the bar. |

With `prefers-reduced-motion: reduce` the durations are zero and the bar does not run.

```html
<main ypy-get="/feed" ypy-trigger="load" ypy-swap="innerHTML swap:200ms" class="ypy-fade">...</main>
```

```css
:root { --ypy-progress-color: #0284c7; }
```

The progress bar of the [`ypy-progress`](../extensions/progress.md) extension is the element
`<ypy-progress-bar>`, appended to `<html>` while a navigation waits. It lives in a shadow root and takes the
same `--ypy-progress-*` variables.

## Turning the built-in indicator rule off

The core adds one rule of its own, the one that hides `.ypy-indicator` until a request runs. Turn it off with
`config.includeIndicatorCSS = false` when your own CSS does that, or use another class name through
`config.classes.indicator`.

## Markers YPY writes on its own

`data-ypy-powered` marks elements that YPY has processed, and `data-ypy-state` holds the state list. Both are
ignored by morph swaps (`config.morphIgnore`), so a swap does not remove them. They are not an interface to
build on: select by `data-ypy-state` and the classes in this page only.
