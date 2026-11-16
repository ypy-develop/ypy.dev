---
title: "ypy-scroll"
description: "Sets the scroll position once a swap ends: keep it, send it to the top, to a fixed offset, or put a saved position back."
order: 80
---
# ypy-scroll

Sets the scroll position once a swap ends: keep it, send it to the top, to a fixed offset, or put a saved
position back.

```html
<button ypy-get="/orders?page=2" ypy-target="#list" ypy-swap="outerHTML" ypy-scroll="keep"
        class="rounded border px-3 py-1">More</button>
<ul id="list" class="max-h-96 overflow-auto">...</ul>
```

## Values

| Value | Result |
|---|---|
| `keep` | the window and the swap target scroll back to where they were just before the swap, `scrollTop` and `scrollLeft` |
| `top` | the window and the swap target scroll to the top |
| `restore` | the window goes to the position saved for a history entry (see below) |
| a number: `120`, `120px` | the window scrolls to that vertical offset |

Anything else (a word that is not on the list, an empty value) does nothing.

`keep` matters when the swap replaces an element: after `outerHTML` the new `#list` would start at the top, and
a swap that shrinks the page lets the browser clamp the window. With `keep` the target starts where the old one
was. The position is taken just before the swap, so a visitor who scrolls while the request is in flight is not
thrown back.

## The `scroll` and `show` modifiers

`ypy-swap="innerHTML scroll:top show:top"` keeps working and means what it did. `ypy-scroll` is the short form
and wins when both are present: it runs after them.

A boosted request has `show:top` by default; `ypy-scroll="keep"` on the link cancels its effect.

## History

When a visitor goes back, YPY asks the server for the page again and then calls
`__applyScrollPolicy(ctx, "restore")` with `ctx.scrollPosition` (`[x, y]`, saved when the entry was left).
Without a saved position nothing moves. `ypy-scroll="restore"` on an ordinary request has no position and does
nothing.

## Reduced motion

`keep` and `restore` always move at once. `top` and an offset move at once when the visitor asks for less motion
(`prefers-reduced-motion: reduce`). Otherwise they follow the CSS `scroll-behavior` of the page, so
`html { scroll-behavior: smooth }` still animates them for everyone else.

## Inheritance

The attribute is read on the element that sends the request. `ypy-scroll:inherited="top"` on an ancestor sets it
for the subtree; a bare `ypy-scroll` on an ancestor is not inherited ([inheritance](../inheritance.md)).

## From a script

```js
ypy.ajax('GET', '/orders', {target: '#panel', scroll: 'top'});
```

`ctx.scroll` set by the caller (or by a `ypy:before:swap` listener) wins over the attribute.

## Related

[`ypy-focus`](ypy-focus.md), [history](../history.md), [accessibility guide](../../guides/accessibility.md).
