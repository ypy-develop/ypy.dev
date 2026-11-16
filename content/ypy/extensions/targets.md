---
order: 420
title: "One response, many elements: the ypy-targets extension"
description: "Swaps one response into every element that matches a selector, such as a counter shown in both the header and the sidebar of a page."
---

# One response, many elements: the `ypy-targets` extension

`ypy-target` names one element. `ypy-targets` names a selector, and the response is swapped into every element that
matches. Use it when the same piece of HTML belongs in several places, for example a counter shown in a header and in a
sidebar.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-targets.js"></script>
```

The extension is part of `ypy.max`.

## Use it

```html
<span class="cart-count">2</span>
<aside><span class="cart-count">2</span> items</aside>

<button ypy-post="/cart/items" ypy-targets=".cart-count" ypy-swap="innerHTML">Add</button>
```

If the server answers `3`, both counters show `3`. Each element gets its own copy of the response, with the swap style
of the request.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-targets` | selector | every matching element receives the main content of the response; inherited |

## Notes

- When nothing matches, YPY writes a warning to the console and swaps nothing for this attribute.
- It replaces the main swap only. Out-of-band elements and `<ypy-partial>` parts keep their own targets.
- A response that is different for each place is not this extension's job: use `<ypy-partial>` with one target each, as
  in [Multiple targets in one response](../guides/multiple-targets.md).
- The selector can use [`scope`](scope.md) words when that extension is loaded.

## See also

- [Multiple targets in one response](../guides/multiple-targets.md)
- [`<ypy-partial>`](../reference/tags/ypy-partial.md)
- [Updating other content (example)](../examples/updating-other-content.md)
