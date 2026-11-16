---
order: 120
title: Animations
description: Fade a swap in and out, animate a class that changed, or use view transitions. YPY adds classes during a swap and plain CSS does the motion.
---

# Animations

YPY does not animate anything by itself. It puts classes on the elements during a swap and leaves the motion to CSS.
This page shows three ways to use them, from the simplest to the one that needs a browser feature.

The classes are `ypy-swapping` (on the swap target, until the content is inserted), `ypy-settling` (on the target,
during the settle delay) and `ypy-added` (on new content, before it settles). Their rules are in
[styling](../reference/styling.md).

## 1. Fade out and fade in

`dist/ypy.css` has two opt-in classes. `ypy-fade` makes the swap target fade out while it is in the `swapping` state,
and fade back in when the state ends. `ypy-fade-in` makes new content fade in, which suits content that is added
next to what is already there.

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.css">
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<button ypy-get="/quote" ypy-target="#quote" ypy-swap="innerHTML swap:200ms">Another quote</button>
<blockquote id="quote" class="ypy-fade"><p>Start here.</p></blockquote>
```

`swap:200ms` delays the swap, so `ypy-swapping` stays on `#quote` for 200 ms and the fade-out has time to run. When
the content is inserted the class goes away and the target fades back in with the new text. The server only sends the
text:

```go
func Quote(w http.ResponseWriter, r *http.Request) {
	ypy.Fragment(w, `<p>`+html.EscapeString(nextQuote())+`</p>`)
}
```

For a list that grows, put `ypy-fade-in` on each new item instead: `ypy-swap="beforeend"` adds it and it fades in
from `@starting-style`.

The time and curve come from `--ypy-duration` and `--ypy-easing`. With `prefers-reduced-motion: reduce` the duration
is zero.

## 2. Animate a class that changed

When the response holds an element with the same `id` and tag as one it replaces, YPY first inserts the new element
with the old `class`, `style`, `width` and `height`, and sets the new values after the settle delay. A CSS transition
then runs between the two. It works with `innerHTML` and `outerHTML`.

```css
.meter { height: 8px; background: #0284c7; transition: width 500ms ease-out; }
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

The answer is the same two elements with the new class: `<div id="meter" class="meter level-3"></div>` and the
button. The `settle:50ms` modifier is the wait before the new class is applied. Prefer a class to a `style`
attribute; a strict CSP blocks inline styles (see [security](../reference/security.md)).

## 3. View transitions

The `ypy-transitions` extension runs a swap inside `document.startViewTransition()`, so the browser animates between
the old and the new page state.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-transitions.min.js"></script>

<a href="/photos/2" ypy-get="/photos/2" ypy-target="#gallery" ypy-swap="innerHTML transition:true">Next photo</a>
```

```css
::view-transition-group(root) { animation-duration: 150ms; }
```

`transition:true` asks for it on one swap; `config.transitions` turns it on for all. A browser without the API, or a
visitor who prefers reduced motion, gets the same swap with no animation. `ypy:before:viewTransition` and
`ypy:after:viewTransition` fire around it.

## How it works

- The core sets and removes the state classes. You write the CSS that reacts to them.
- `swap:` and `settle:` are modifiers of `ypy-swap`. They give a transition time to run.
- `ypy-added` lasts only until the settle step ends, which can be one frame. For a fade-in that does not depend on
  it, use `ypy-fade-in`, which starts from `@starting-style`.

## See also

- [Styling](../reference/styling.md) and [CSS classes](../reference/css-classes.md)
- [`ypy-swap`](../reference/attributes/index.md) for every modifier
