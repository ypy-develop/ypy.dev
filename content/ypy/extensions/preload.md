---
order: 300
title: "Preload: the ypy-preload extension"
description: "Starts a GET on mousedown or touchstart, before the click finishes, and reuses the answer when the real request has the same URL."
---

# Preload: the `ypy-preload` extension

A request starts when the visitor clicks. `ypy-preload` starts the same `GET` a moment earlier, on `mousedown` or
`touchstart`, before the click is complete. When the real request has the same
URL and the answer is still fresh, YPY uses the answer that is already on its way instead of asking again.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-preload.js"></script>
```

The extension is part of `ypy.max`.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-preload` | none | an element with `ypy-get` preloads on `mousedown` and `touchstart` |
| `ypy-preload` | `<event> [<event> ...] [timeout:<time>]` | preloads on the events you list; `timeout:` is how long the answer stays valid (default `5s`) |

`ypy-preload` is inherited, so you can put it on a container. Only `GET` requests are preloaded: an element that
sends `POST`, `PUT`, `PATCH` or `DELETE` is left alone, whatever events you list.

```html
<button ypy-get="/orders/42" ypy-target="#detail" ypy-preload>Order 42</button>

<a href="/reports" ypy-get="/reports" ypy-target="#main" ypy-preload="mouseenter timeout:10s">Reports</a>
```

The preload sends what the real request would send: the form values and `ypy-vals` go into the query string, and the
`YPY-*` request headers are set. The element keeps one preload at a time. The preload is used by the next
request of that element and then dropped; an answer older than its timeout is ignored and the request goes out as
usual.

## Boosted links

A boosted link (see [Navigation](nav.md)) preloads without any attribute, on `mousedown` and `touchstart`.

## Settings

| Setting | Default | Meaning |
|---|---|---|
| `preload.autoBoost` | `true` | preload boosted links without writing `ypy-preload`; `false` turns it off |
| `preload.boostTimeout` | `5000` | how long an automatic preload stays valid, in milliseconds or as `"5s"` |
| `preload.boostEvent` | `mousedown` | the event that starts an automatic preload (`touchstart` always does too) |

## Notes

- A preload costs a request the visitor may never use. Do it for links that people follow most of the time, and
  prefer the events close to the click. `mouseenter` or `focus` trigger more often than the click that follows.
- The server sees an ordinary `GET`. Make sure it is safe to repeat: a preload that changes data is a bug.
- For links of a whole site, the prefetch of [`ypy-nav`](nav.md) (`ypy-prefetch`) is the better fit;
  `ypy-preload` is for one element.

## See also

- [Navigation: boost and prefetch](nav.md)
- [Cache with revalidation](cache.md)
