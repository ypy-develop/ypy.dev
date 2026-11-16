---
order: 450
title: "View transitions: the ypy-transitions extension"
description: "Runs each swap inside the browser's View Transitions API, queued one after another, with events you can hook into to prepare or clean up."
---

# View transitions: the `ypy-transitions` extension

The browser can animate between two states of a page with the View Transitions API. The core only decides whether a
swap asks for one. This extension does the animating: it runs the swap inside `document.startViewTransition()`.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-transitions.js"></script>
```

The extension is part of `ypy.max`.

## Turn a transition on

Add the `transition:true` modifier to one swap, or set `config.transitions` for all of them.

```html
<a ypy-get="/next" ypy-target="#page" ypy-swap="innerHTML transition:true">Next</a>
```

```js
ypy.config.transitions = true
```

Style the transition with the usual View Transitions CSS (`::view-transition-old(root)` and the others).

## What the extension does

- Runs each swap that asked for a transition inside `document.startViewTransition()`.
- Runs them one at a time: a second swap waits for the first one to finish.
- Fires `ypy:before:viewTransition` before each, and `ypy:after:viewTransition` after.

Without the extension the same swap runs at once, with no animation. The same happens in a browser without view
transitions and for a visitor who prefers reduced motion (`prefers-reduced-motion: reduce`), with or without the
extension. A transition that the browser skips is not an error.

## Events

| Event | Cancelable | `detail` | Fires |
|---|---|---|---|
| `ypy:before:viewTransition` | listed as yes | `task`, `ctx` | before the transition starts; `detail.task` is the function that does the swap and may be replaced |
| `ypy:after:viewTransition` | no | `task`, `ctx` | when the transition has finished |

The attribute file lists `ypy:before:viewTransition` as cancelable, but the extension in 0.1.1 does not read the
answer of its listeners. To change what runs inside the transition, replace `detail.task`.

## Settings

| Setting | Default | Meaning |
|---|---|---|
| `transitions` | `false` | ask for a view transition on every swap |

## Notes

- A transition holds the page for a moment while it runs: use it for navigation and large changes, not for a
  live counter.
- The swap modifier `ypy-swap="... transition:true"` is read by the core; the table of swap styles is in
  [Swap styles and morph](../index.md#swapping).
- For a simple fade without this API, use the classes in [CSS classes](../reference.md#css-classes) and the
  [animations example](../examples/animations.md).

## See also

- [Animations (example)](../examples/animations.md)
- [Events](../index.md#events-and-logging)
