---
order: 310
title: "Keep what the visitor did: the ypy-preserve extension"
description: "Keeps typed text, caret, focus and scroll of fields and containers through a swap, and holds the swap while an input method is still composing."
---

# Keep what the visitor did: the `ypy-preserve` extension

A swap replaces HTML. The visitor, in the meantime, may have typed in a field, put the caret in the middle of a word or
scrolled a list. The core already keeps a node by `id` and knows `attrs` and `children`
([`ypy-preserve`](../reference/attributes/ypy-preserve.md)). This extension adds the words that name a state of the
visitor, and it makes a swap wait while someone is still composing text with an input method.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-preserve.js"></script>
```

The extension is part of `ypy.max`.

## The modes

Write the modes on the live element, separated by spaces. Each one keeps one state of the element and of its
descendants that have an `id`. They work in a morph and in a raw swap such as `innerHTML` or `outerHTML`.

| Mode | What it keeps |
|---|---|
| `value` | what the visitor typed or chose in a field, when it differs from the default of the HTML (a field they did not touch follows the response) |
| `selection` | the caret or the selected text of a field |
| `focus` | the focus, when the element had it before the swap |
| `scroll` | the scroll position of the element |

```html
<form id="search" ypy-preserve="value selection focus"
      ypy-get="/search" ypy-trigger="input changed delay:300ms" ypy-target="#results">
  <input id="q" name="q" type="search">
</form>
<div id="results" ypy-preserve="scroll">...</div>
```

The modes do not combine with `true`, `all`, `attrs` or `children`, which the core reads. File inputs are not
restored (the browser does not allow it).

When the server must win, mark the element in the response with `ypy-morph-force`. The saved state of that element is
then not put back.

## Input methods

While the visitor composes text (for example Japanese, Chinese or Korean input, where `compositionstart` has fired and
`compositionend` has not), a swap whose target contains that field waits for the end of the composition, so the
candidate text is not cut in half. The wait is at most `compositionTimeout` milliseconds.

| Setting | Default | Meaning |
|---|---|---|
| `compositionTimeout` | `3000` | the longest a swap waits for a composition to end, in milliseconds |

Without the extension a swap happens at once.

## Notes

- Give the fields an `id`. The state is matched to the new node by `id`.
- The state is put back right after the DOM work of each swap task, before the settle delay, so nothing flashes.
- For a field the visitor should be able to keep typing in while a poll or a stream refreshes the page, `value` and
  `focus` together are usually what you want.

## See also

- [`ypy-preserve` and `ypy-preserve-attrs`](../reference/attributes/ypy-preserve.md)
- [Swap styles and morph](../reference/swap-and-morph.md)
- [Third-party components](../guides/third-party-components.md)
