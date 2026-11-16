---
order: 360
title: "Selectors by area: the ypy-scope extension"
description: "Three selectors, scope, layer and shadow, for ypy-target, ypy-indicator, ypy-include and ypy.find(), relative to the element that asks."
---

# Selectors by area: the `ypy-scope` extension

Reusable HTML often repeats: the same card, row or widget many times on a page. A fixed `#id` does not work for the
second copy. `ypy-scope` adds three selectors that find things relative to the element, so each copy talks to
its own parts.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-scope.js"></script>
```

The extension is part of `ypy.max`.

## The selectors

They work wherever a selector is accepted: `ypy-target`, `ypy-indicator`, `ypy-include`, `ypy-disable`, `ypy-sync`, the
`from:` of a trigger and `ypy.find()`.

| Selector | Finds |
|---|---|
| `scope` | the closest ancestor with `ypy-scope`, the element itself included; from inside a shadow root the search goes on from the host |
| `scope <selector>` | the elements that match the selector inside that scope |
| `layer` | the element of the topmost open layer of [`ypy-layers`](layers.md), or nothing when none is open |
| `layer <selector>` | the elements that match inside that layer |
| `shadow:<selector>` | the elements that match below the element, going into nested open shadow roots and through the slots; the element itself is not matched |

```html
<section ypy-scope>
  <button ypy-get="/cards/1" ypy-target="scope">Reload this card</button>
  <button ypy-get="/cards/1/hint" ypy-target="scope .hint">Show the hint</button>
  <p class="hint"></p>
</section>

<button ypy-get="/badges" ypy-target="shadow:.badge">Refresh the badges</button>
```

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-scope` | none | marks an element as a scope; the value is not used; not inherited |

## Notes

- Scopes nest: `scope` is the nearest one.
- `shadow:` does not enter closed shadow roots, and a light child that no slot shows is not found.
- With `layer`, a request made from inside a dialog can aim at the dialog, whatever the page behind it holds.
- A selector the extension does not know is left to the core, so `ypy-target="#id"` keeps working.

## See also

- [Layers](layers.md)
- [Multiple targets in one response](../index.md#out-of-band-swaps)
- [Attribute inheritance](../index.md#attribute-inheritance)
