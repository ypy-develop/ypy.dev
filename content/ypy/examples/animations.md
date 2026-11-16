---
order: 120
title: Animations
description: Fade a swap in and out, animate a class that changed, or use view transitions. YPY adds classes during a swap and plain CSS does the motion.
---

# Animations

YPY does not animate anything by itself. It puts classes on the elements during a swap and leaves the motion to CSS.
Three ways to use them follow, from the simplest to the one that needs a browser feature.

The classes are `ypy-swapping` (on the swap target, until the content is inserted), `ypy-settling` (on the target,
during the settle delay) and `ypy-added` (on new content, before it settles). See
[CSS transitions](../index.md#css-transitions) and [CSS classes](../reference.md#css-classes).

## The page

### 1. Fade out and fade in

`ypy.css` has two opt-in classes. `ypy-fade` makes the swap target fade out while it is in the swapping state and fade
back in when the state ends. `ypy-fade-in` makes new content fade in, which suits content added next to what is already
there.

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.css">

<button ypy-get="/quote" ypy-target="#quote" ypy-swap="innerHTML swap:200ms">Another quote</button>
<blockquote id="quote" class="ypy-fade"><p>Start here.</p></blockquote>
```

`swap:200ms` delays the swap, so `ypy-swapping` stays on `#quote` for 200 ms and the fade-out has time to run. For a list
that grows, put `ypy-fade-in` on each new item and use `ypy-swap="beforeend"`.

### 2. Animate a class that changed

When the response holds an element with the same `id` and tag as one it replaces, YPY first inserts the new element with
the old `class`, `style`, `width` and `height`, and applies the new values after the settle delay. A CSS transition
then runs between the two. It works with `innerHTML` and `outerHTML`.

```css
.meter { height: 8px; background: var(--accent); transition: width 500ms ease-out; }
.meter.level-1 { width: 20%; }
.meter.level-3 { width: 60%; }
.meter.level-5 { width: 100%; }
```

```html
<div id="panel">
  <div id="meter" class="meter level-1"></div>
  <button ypy-get="/meter/3" ypy-target="#panel" ypy-swap="innerHTML settle:50ms">Raise</button>
</div>
```

### 3. View transitions

The `ypy-transitions` extension runs a swap inside `document.startViewTransition()`, so the browser animates between the
old and the new state of the page.

```html
<a href="/photos/2" ypy-get="/photos/2" ypy-target="#gallery" ypy-swap="innerHTML transition:true">Next photo</a>
```

```css
::view-transition-group(root) { animation-duration: 150ms; }
```

## What the server answers

Nothing here is special to the server: it answers an ordinary fragment with `Vary: YPY-Request`.

| Request | Status | Body |
|---|---|---|
| `GET /quote` | `200` | `<p>Make it work, then make it fast.</p>` |
| `GET /meter/3` | `200` | The same two elements with the new class: `<div id="meter" class="meter level-3"></div>` and the button |
| `GET /photos/2` | `200` | The next photo markup, for example `<img src="/photos/2.jpg" alt="Harbour at dawn" width="800" height="533">` |

For the class change, keep the `id` and the tag of `#meter` the same in every answer, and change only the class.

## How it works

- The core sets and removes the state classes. You write the CSS that reacts to them.
- `swap:` and `settle:` are modifiers of `ypy-swap`. They give a transition time to run.
- `ypy-added` lasts only until the settle step ends, which can be a single frame. `ypy-fade-in` does not depend on it:
  it starts from `@starting-style`.
- The time and the curve come from `--ypy-duration` and `--ypy-easing`. With `prefers-reduced-motion: reduce` the
  duration is zero.
- `transition:true` asks for a view transition on one swap, and `config.transitions` turns it on for all. A browser
  without the API, or a visitor who prefers reduced motion, gets the same swap with no animation.
  `ypy:before:viewTransition` and `ypy:after:viewTransition` fire around it.

## Variations

- Prefer a class to a `style` attribute for the new value: a strict CSP blocks inline styles. See
  [security](../index.md#security).
- Animate a removal: `ypy-swap="delete swap:300ms"` keeps `ypy-swapping` on the target for 300 ms before it goes.

## See also

- [`ypy-transitions`](../extensions/transitions.md)
- [`ypy-swap`](../reference.md#core-attributes) for every modifier
- [Delete row](delete-row.md)
