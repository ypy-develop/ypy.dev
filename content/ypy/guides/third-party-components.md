---
title: "Third-party components"
description: "Keep React, Web Components, Alpine, Bootstrap and Bulma working across swaps: what to mark so the morph skips them and how to start them with ypy.onLoad."
order: 70
---
# Third-party components

YPY keeps the page in step with the HTML the server sends. A script that writes into the page on its own (React, a Web
Component, Alpine, Bootstrap, Bulma) writes things the server knows nothing about. This guide shows what to mark so the
morph leaves them alone. The attribute is described in
[`ypy-preserve`](../reference/attributes/ypy-preserve.md); the rules of the morph are in
[Swap styles and morph](../reference/swap-and-morph.md).

The examples use Tailwind classes for the page, but nothing depends on it. To YPY, a class added by Tailwind, Bootstrap
or Bulma is a class like any other.

As a rule of thumb, protect the attributes of whatever the script toggles, protect the children of whatever it draws,
and start components with `ypy.onLoad`.

## Web Components

A custom element that draws into its own DOM needs no mark. When the response sends it empty, its children stay. Its
attributes still follow the response, so they act as the inputs of the component.

```html
<section id="stats" class="rounded border p-4" ypy-get="/stats" ypy-trigger="every 30s">
  <h2 class="text-lg font-semibold">Visits</h2>
  <visits-chart data-series="3,5,8,13"></visits-chart>
</section>
```

Each refresh updates the heading and `data-series`; the chart's own SVG stays and the component reacts to the
attribute. If the element wraps server-rendered children (a card with slotted text), send them: a response with
content is merged as usual. To opt one element out, write `ypy-preserve="false"`.

## React

A React root renders its own tree. Send the root empty and let React own what is inside it:

```html
<div id="cart-island" data-reactroot class="rounded border p-4"></div>
```

`[data-reactroot]` (and the Vue root `[data-v-app]`) are in `ypy.config.morphThirdParty`. React 18 does not
write `data-reactroot`, so put the attribute on your container yourself, or add a selector:

```js
ypy.config.morphThirdParty = '[data-reactroot],[data-v-app],[x-data],.react-island';
```

(the value replaces the default, so keep the entries you still need). A swap that replaces the root with
`outerHTML` unmounts React; use a morph, or `ypy-preserve` on the root so that it keeps its identity.

## Alpine

An element with `x-data` is covered by the same default when the response gives it empty. When it has content,
the morph merges it, and the `ypy-alpine-compat` extension keeps Alpine's state
and reactive nodes. For a component whose classes Alpine toggles (`:class`, `x-show`):

```html
<div x-data="{ open: false }" ypy-preserve-attrs="class,style,x-*">
  <button @click="open = !open" class="rounded bg-sky-600 px-3 py-1.5 text-white">Menu</button>
  <ul x-show="open" class="mt-2 space-y-1">...</ul>
</div>
```

## Bootstrap

Bootstrap writes `show`, `collapsing`, `aria-*`, `style` and `data-bs-*` into its components. List them and
start the components with `ypy.onLoad`:

```html
<div id="filters" class="collapse"
     ypy-preserve-attrs="class,style,aria-*">
  <form ypy-get="/orders" ypy-target="#orders">...</form>
</div>
<div id="confirm" class="modal fade" tabindex="-1"
     ypy-preserve-attrs="class,style,aria-*,role">...</div>

<script>
  ypy.onLoad('.modal', el => new bootstrap.Modal(el));
  ypy.onLoad('[data-bs-toggle="tooltip"]', el => new bootstrap.Tooltip(el));
</script>
```

`ypy.onLoad` runs once for each new element and never twice for the same node. A morph that keeps `#confirm` does not
create a second `bootstrap.Modal` on it, and a fragment that brings a new tooltip gets it started.

## Bulma

Bulma has no script of its own; the page toggles `is-active`. Mark the elements whose state the page toggles:

```html
<div class="dropdown" ypy-preserve="attrs">...</div>
<a class="navbar-burger" role="button" aria-expanded="false" ypy-preserve="attrs"></a>
<div class="modal" ypy-preserve-attrs="class">...</div>
```

```js
ypy.onLoad('.navbar-burger', el => el.addEventListener('click', () => {
  el.classList.toggle('is-active');
  el.setAttribute('aria-expanded', el.classList.contains('is-active'));
}));
```

## What YPY itself adds

YPY writes only classes with its own prefix (`ypy-request`, `ypy-swapping`, `ypy-settling`, `ypy-added`,
`ypy-indicator`) and the `data-ypy-state` attribute, so they cannot collide with `active`, `show`, `fade`, `modal`,
`is-active` or `is-loading`. You can rename them in `ypy.config.classes`.

## Checklist

1. Does the script write attributes or classes? Add `ypy-preserve="attrs"`, or `ypy-preserve-attrs` with the
   names it writes.
2. Does it draw children? Send the element empty (custom elements and the roots of `morphThirdParty` keep them),
   or add `ypy-preserve="children"`.
3. Does it keep a live node (`<video>`, `<iframe>`, a map)? Give it an `id` and `ypy-preserve`.
4. Does it need a start-up call? Use `ypy.onLoad(selector, fn)`.
