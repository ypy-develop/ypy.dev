---
title: "Swap styles and morph"
description: "How ypy-swap puts a response into the page: the swap styles, what a morph keeps from the user, its cost, and how to protect components that another script owns."
order: 90
---
# Swap styles and morph

A swap puts the response into the page. `ypy-swap` (or `data-ypy-swap`) says how.

A swap that gives no style morphs: the response is merged into
the page, and the elements it shares with the page stay the same DOM nodes. Focus, the text
you typed, a selection, the scroll position of a block, a playing `<video>`, an open
`<details>` or `<dialog>` all survive. A page that wants the raw replacement writes `ypy-swap="innerHTML"`
or sets `ypy.config.defaultSwap`.

## Styles

| Style | What it does | Raw or morph |
|---|---|---|
| `innerMorph` | merges the response into the children of the target | morph |
| `outerMorph` | merges the response into the target and its siblings in place | morph |
| `morph` | reads as `innerMorph` | morph |
| `innerHTML` | replaces the children of the target | raw |
| `outerHTML` | replaces the target | raw |
| `outerSync` | copies the attributes and replaces the children of the target | raw |
| `beforebegin`, `afterbegin`, `beforeend`, `afterend`, `textContent`, `delete`, `none` | as before | raw |

A style you write keeps its meaning: `ypy-swap="innerHTML"` replaces the children. Only the style you did
not write is a morph:

| Where the style is not written | Style used |
|---|---|
| main swap of a request | `innerMorph` |
| `<ypy-partial>` without `ypy-swap` | `innerMorph` |
| out-of-band `ypy-swap-oob="true"` and `ypy-select-oob` | `outerMorph` |

```html
<!-- morphs: the input keeps what the visitor typed -->
<button ypy-get="/cart" ypy-target="#cart"
        class="rounded bg-sky-600 px-3 py-1.5 text-white ypy-request:opacity-50">Refresh</button>
<section id="cart">
  <input id="coupon" class="rounded border px-2 py-1">
  <details id="shipping"><summary>Shipping</summary>...</details>
</section>

<!-- the same swap, replacing the children -->
<button ypy-get="/cart" ypy-target="#cart" ypy-swap="innerHTML">Reload</button>
```

To change the default for the whole page, set `ypy.config.defaultSwap` (or
`<meta name="ypy-config" content='{"defaultSwap":"innerHTML"}'>`). A value that is set wins over the
built-in default; `ypy.config.defaultSwap = undefined` goes back to it.

## What the user did wins

A morph keeps what the user did to an element, because the node stays the same.

| What the user did | What the morph does |
|---|---|
| Opened a `<details>` or a `<dialog>` | The `open` attribute mirrors a state the user changes, so the response never opens or closes an element that is already on the page. A `<dialog>` opened with `showModal()` stays modal. |
| Typed in an `<input>` or a `<textarea>` | The value changes only if the response changes it (the `value` attribute, the text of the textarea) and the field does not have the focus. |
| Checked a box or picked an option | `checked`, `selected` and a select's choice are kept until the response changes them. |
| Put the focus or a selection somewhere | Kept. In a browser without `moveBefore`, a morph that has to move the focused node gives the focus back to the element with the same `id`. |
| Scrolled a block, or is playing a `<video>` | Kept. |

An element the response adds is new and is drawn as the response says, `open` included.

To let the response win on one element, mark it in the response with `ypy-morph-force`
(also written `data-ypy-morph-force`):

```html
<!-- the server closes this panel and puts its text in the field, whatever the user did -->
<details id="filters" ypy-morph-force><summary>Filters</summary>...</details>
<input id="q" value="" ypy-morph-force>
```

`ypy-morph-force` covers `open` and the value of a field that has the focus.

## A page change is not a refresh

A default style that meets `document.body` (a boosted link, `ypy-target="body"`) replaces the
page instead of morphing it: the text typed on the old page must not appear on the
new one. To refresh the same page by morphing, write the style:

```html
<form ypy-boost="true" ypy-swap="innerMorph" action="/search"> ... </form>
```

## Cost

A morph walks both trees, so it costs more than `innerHTML` on a large, very different
response. `ypy.config.morphScanLimit` (default 10) bounds how far it looks for the node that
matches; lower it for long lists, or write `ypy-swap="innerHTML"` for a region that is replaced
whole. Elements without an `id` match by position and tag, so give an `id` to what must keep its
identity (a form field, a panel).

## Third-party components

A script other than YPY may own part of the page: a custom element that draws into itself, a React or Vue
root, an Alpine component, the classes that Bootstrap or Bulma toggle. A morph would rewrite them from the
response, so there are three ways to protect them.

`ypy-preserve="attrs"` stops the morph from changing the attributes of the element, and its children are still
merged. You can also list the attributes in `ypy-preserve-attrs="class,style,aria-*"`. `ypy-preserve="children"`
keeps the children and lets the attributes follow the response, and `ypy-preserve="all"` keeps both.

Without any attribute, a custom element, or an element matched by `ypy.config.morphThirdParty` (default
`[data-reactroot],[data-v-app],[x-data]`), keeps its children when the response gives it with no content.
`ypy-preserve="false"` switches that off.

`ypy.onLoad(selector, fn)` starts a script's components once for each new element. The attribute is in
[`ypy-preserve`](attributes/ypy-preserve.md); examples for React, Web Components, Alpine, Bootstrap and Bulma are in
the [third-party components guide](../guides/third-party-components.md).
