---
order: 20
title: "Alpine.js: the ypy-alpine-compat extension"
description: "Keeps Alpine.js state alive across a YPY swap or a restored page by moving it from the old node to the new one. Needed only when both run on one page."
---

# Alpine.js: the `ypy-alpine-compat` extension

Alpine.js keeps its state in the DOM nodes it has initialised. A swap that replaces those nodes loses it, and a morph
that matches an old node with a new one can leave Alpine pointing at the wrong data. `ypy-alpine-compat` makes the
two libraries agree: the response is morphed in, the Alpine state of the old node moves to the new one, and Alpine
sees the change once, after the swap.

You need it only when Alpine and YPY run on the same page. It does nothing when `window.Alpine` is missing.

## Load it

```html
<script src="ypy.min.js"></script>
<script defer src="alpine.js"></script>
<script src="ext/ypy-alpine-compat.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-alpine-compat` in it, or it is refused.
It is listed with the other official extensions in [Extensions](index.md).

## What it does

| Moment | What the extension does |
|---|---|
| Before a swap | Tells Alpine to hold its own mutation handling until the swap is finished. |
| Morphing a node | Gives the new node the Alpine data stack of the old one and lets Alpine clone the reactive state across. |
| Morphing a node with a reactive `:id` | Ignores the `id` mismatch when both nodes bind `id` through Alpine, so the node is matched by its tag. |
| After the swap, or when the request ends | Lets Alpine process what changed, once. |
| Saving a page into the history cache | Writes the `x-data` state of each component into a `data-alpine-state` attribute, then destroys the Alpine tree of the snapshot. |
| Restoring a page from the history cache | Reads `data-alpine-state` back into the components and removes the attribute. |

The last two rows need the [history cache](history-cache.md). Only values that survive `JSON.stringify` come back.

## Example

A counter that stays correct when its container is refreshed:

```html
<div id="panel" ypy-get="/panel" ypy-trigger="every 10s" ypy-swap="innerMorph">
  <div x-data="{ count: 0 }">
    <button x-on:click="count++">Add</button>
    <span x-text="count"></span>
  </div>
</div>
```

The server sends the same markup, with `count: 0` again. The morph keeps the existing nodes, and the extension keeps the
`count` the visitor reached. Without the extension, Alpine may start the component again from `0`.

## Notes

- The short bindings of [`ypy-live`](live.md) (`:text`, `:class`) use the same `:` prefix as Alpine. When Alpine is
  detected, YPY turns the short form off unless you set `ypy.config.live.bindPrefix`. Use the long form `ypy-live:text`
  on a page that has both.
- The extension changes how morphing matches nodes. A response that has to replace a component from scratch should use
  `ypy-swap="outerHTML"` or `ypy-morph-force`.

## See also

[Third-party components](../guides/third-party-components.md), [Swap styles and morph](../reference/swap-and-morph.md),
[`ypy-live`](live.md), [History cache](history-cache.md).
