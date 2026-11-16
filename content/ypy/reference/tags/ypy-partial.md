---
title: "<ypy-partial>"
description: "One piece of a response that goes to its own target. A response can carry many partials; each is swapped into the element it names."
order: 10
---
# `<ypy-partial>`

One piece of a response that goes to its own target. A response can carry any number of partials next to
the main content; each one is swapped into the element it names.

```html
<!-- response of POST /items -->
<li class="flex gap-2 p-2">New item</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">
  <span class="font-bold">12 items</span>
</ypy-partial>

<ypy-partial id="toast">
  <p class="rounded bg-green-100 p-3">Saved</p>
</ypy-partial>
```

The `<li>` goes to the target of the request (`ypy-target` on the element that made it). `#count` receives the
counter. The last partial has no target attribute, so the target is the element with its own `id`, `#toast`.

## Attributes

| Attribute | Meaning | Default |
|---|---|---|
| `ypy-target` | Selector of the target. It accepts the same extended selectors as the rest of YPY (`closest li`, `.row`, `find .name`). A selector that matches several elements gives each one a copy of the content. | the element with the `id` of the partial |
| `ypy-swap` | Swap style and modifiers (`innerHTML`, `beforeend`, `innerMorph`, `ignoreTitle:true`, `transition:true`...). | `ypy.config.defaultSwap` |
| `id` | Used as the target (`#id`) when there is no `ypy-target`. | none |

The `ypy-target` and `ypy-swap` attributes also work as `data-ypy-target` and `data-ypy-swap`. Tag names are not
case sensitive.

## Rules

| Topic | Rule |
|---|---|
| Order | The main content is swapped first, then the `ypy-select-oob` elements, then the `ypy-swap-oob` elements, then the partials, each group in document order. |
| Content | Rows, cells and options work as they do in a normal swap: a partial may hold `<tr>` for a `<tbody>` target or `<option>` for a `<select>`. A plain `<template>` inside the content stays in it. Scripts inside a partial run. |
| Title | A `<title>` in the response sets the document title. `ignoreTitle:true` in `ypy-swap` keeps it. |
| Only partials | A response with only partials does not swap the main target with an empty content. |
| Missing target | A partial whose target matches nothing (or that names no target) fires `ypy:target:missing` with `{ctx, selector, type: "partial"}` on the element that made the request, and the other items go on. |
| A failing item | It does not stop the others. See [Multiple targets](../../guides/multiple-targets.md). |

Any other `<ypy-name>` tag in a response (one word after the prefix) becomes a template of that type and goes
to the extensions as `ypy:process:name`. A name with a second hyphen (`<ypy-my-card>`) is an ordinary element.
