---
title: "ypy-preserve and ypy-preserve-attrs"
description: "Keep what another script owns across a swap or morph: a Bootstrap class, the children a chart drew, a playing video, text typed in a field."
order: 70
---
# `ypy-preserve` and `ypy-preserve-attrs`

Keep what a script other than YPY owns. A swap or a morph rewrites the page from the server's HTML,
so a class that Bootstrap added to a collapse, the children a chart drew into its element, or a `<video>`
that is playing would be lost. `ypy-preserve` says how much of an element the response must leave alone.

`ypy-preserve` and `data-ypy-preserve` are the same attribute. It is read on the element
itself and is never inherited. Put it on the element in the response, as the server renders it every time:
the morph reads it on the element that is already on the page, which got it from the first render.

```html
<nav id="menu" class="collapse navbar-collapse" ypy-preserve="attrs">...</nav>
```

## The forms

| Value | The node | Its attributes and classes | Its children | Needs an `id` |
|---|---|---|---|---|
| none, `true` | the same DOM node | kept | kept | yes |
| `all` | the same DOM node | kept | kept | in a raw swap; a morph matches by position |
| `attrs` | kept by the morph | kept: the response does not change them | morph as usual | no |
| `children` | kept by the morph | follow the response | kept | no |
| `false` | not preserved | follow the response | follow the response | no |

With no value (or `true`), the element and its state (a playing `<video>`, a loaded `<iframe>`, the text of a field, a
script's expando properties) move out of the page before the swap and take the place of the element with the same
`id` in the response afterwards. This works in every swap style, and the element needs an `id` on the page and in the
response.

`all` does the same, and in a morph it also skips the element when there is no `id`, the way `ypy-morph-skip` does. In a
raw swap such as `innerHTML`, an element can only be kept through its `id`.

`attrs` makes the morph leave every attribute of the element alone, `class` and `style` included, and its children are
merged as usual. `children` is the reverse: the attributes follow the response and the children are not touched, like
`ypy-morph-skip-children`. `false` switches off, for this element, the default for third-party elements (described
below).

`attrs`, `children` and `false` describe the morph (`innerMorph`, `outerMorph`, and the default swap). A raw swap such as `innerHTML` replaces the element, and only the forms that keep the node,
no value and `all`, can stop that.


## `ypy-preserve-attrs`: keep only what the script writes

A list of the attributes the morph must neither set nor remove. Everything else follows the response.

```html
<div id="settings" class="modal fade" tabindex="-1"
     ypy-preserve-attrs="class,style,aria-*,data-bs-*">
  <div class="modal-dialog">...</div>
</div>
```

Names are separated by commas or spaces. An entry is a prefix: `aria-` and `aria-*` keep every `aria-*`
attribute, `class` keeps `class`. The attribute may stand alone; it does not need `ypy-preserve`.
An attribute of the list that the response adds, changes or drops stays as it is on the page.

A prefix is a prefix: `data` keeps every `data-*` attribute, and `class` would also keep a hypothetical
`classic` attribute. Write the longest name you need.

## Third-party elements: the default

A morph also protects the elements a script manages, without any attribute:

- every custom element (a tag name with a hyphen, such as `<my-chart>`),
- the elements matched by `ypy.config.morphThirdParty`, by default
  `[data-reactroot],[data-v-app],[x-data]` (a React 17 root, a Vue 3 root, an Alpine component).

When the response gives such an element with no content (nothing or only whitespace, the markup of a
component that draws itself), the morph keeps the children the page has. Attributes still follow the response,
so the server can pass new values to the component. A response with content is merged as usual, so a Web
Component that only wraps server-rendered children keeps updating. `ypy-preserve="false"` on the element in the
page switches the default off.

Set `ypy.config.morphThirdParty` to your own selector to add roots; the value replaces the default.

## What is not touched

`ypy-preserve` does not stop the extension hooks `ypy:before:morph:node` and `ypy:before:morph:attr`, and a
`ypy-morph-skip` or `ypy-morph-skip-children` rule (or the `morphSkip`, `morphSkipChildren` and `morphIgnore`
config keys) keeps working next to it.

## Starting the script's components

`ypy.onLoad(selector, fn)` calls `fn(element)` once for each element that matches in the content YPY processes:
the page on load, every swap, every out-of-band element and partial. A node that a morph keeps is not called
again, so the component keeps its instance. See the [third-party components guide](../../guides/third-party-components.md).

```js
ypy.onLoad('[data-bs-toggle="tooltip"]', el => new bootstrap.Tooltip(el));
```

`ypy.onLoad(fn)` with one argument keeps its meaning: `fn(root)` for every processed root.
