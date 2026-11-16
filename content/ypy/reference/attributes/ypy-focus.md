---
title: "ypy-focus"
description: "Choose where keyboard focus goes when a swap ends. After a boosted navigation YPY focuses the new page's heading so screen readers start there."
order: 40
---
# ypy-focus

Says where the keyboard focus goes when a swap ends. After a boosted navigation YPY moves it to
the heading of the new page, so a screen reader reads it and keyboard users do not start from the top.

```html
<nav ypy-boost="true" class="flex gap-4 p-4">
  <a href="/orders" class="text-blue-600 underline">Orders</a>
</nav>
<main class="mx-auto max-w-3xl p-6">
  <h1 class="text-2xl font-semibold">Dashboard</h1>
</main>
```

Click "Orders": when the new page is in, the focus is on its `<h1>` and the screen reader says
"Orders, heading level 1". No attribute is needed.

## Values

| Value | What gets the focus |
|---|---|
| `auto` | after a boosted request: the first `h1` inside `[ypy-main]` (else inside `<main>`), or that main itself when it has no `h1`. Any other request: nothing moves |
| `none` | nothing; the focus stays where the swap left it |
| `first-heading` | the first heading (`h1` to `h6`) of the swap target, for any request |
| a selector | the element it finds, for any request (see below) |

A selector uses the same syntax as `ypy-target`, relative to the element that sends the request:
`#message`, `closest form`, `next input`. A selector that matches nothing does nothing.

```html
<button ypy-post="/items" ypy-target="#list" ypy-focus="#flash"
        class="rounded bg-blue-600 px-3 py-1 text-white">Add</button>
<div id="list"></div>
<div id="flash" class="p-2"></div>
```

## Where it is on

Unset, the value is `auto`. `ypy-focus="none"` on a boosted link (or on an ancestor
with `ypy-focus:inherited="none"`) turns it off there.

## Inheritance

The attribute is read on the element that sends the request. To set it for a subtree write
`ypy-focus:inherited="none"` on the ancestor.

## What YPY does to the element

- A heading or a `<main>` cannot take the focus by itself. YPY gives it `tabindex="-1"` and removes the
  attribute when the focus leaves. If you wrote a `tabindex` yourself it is left alone, and a form field
  is never given one.
- The focus does not scroll the page. Use [`ypy-scroll`](ypy-scroll.md) or the `show:` modifier of
  `ypy-swap` for that.
- A boosted navigation does not take the focus from an `[autofocus]` field of the new page.
- Style the heading with `:focus-visible` (Tailwind: `focus-visible:outline`); a programmatic focus after a
  mouse click usually draws no ring, after a key press it does.

## Order

1. The swap gives the focus back to the element that had it, found again by `id`, with its text
   selection (`ypy-swap="morph"`, `innerHTML`, `outerHTML`).
2. `[autofocus]` of the new content.
3. `ypy-focus`.
4. [`ypy-scroll`](ypy-scroll.md).

## From a script

`ypy.ajax()` takes `focus` as an option, and a `ypy:before:swap` listener can set `ctx.focus`:

```js
ypy.ajax('GET', '/orders', {target: '#panel', focus: 'first-heading'});
```

`ctx.focus` is a selector or `false` and wins over the attribute. An error target
(`ypy-error-focus`) uses the same field.

## Related

[`ypy-scroll`](ypy-scroll.md), [`ypy-main`](ypy-main.md), [accessibility guide](../../guides/accessibility.md).
