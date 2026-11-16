---
order: 140
title: "Page head: the ypy-head extension"
description: "When YPY swaps the body for a full page, merges that page's title, styles and scripts into the current head. A head inside a fragment is never merged."
---

# Page head: the `ypy-head` extension

When YPY swaps the body for a new page, the `<title>`, the stylesheets and the `<meta>` tags of that page stay the
old ones. `ypy-head` reads the `<head>` of a full-document response and merges it into the head of the page.

It looks only at a response that starts with a doctype or `<html>`. The `<head>` inside a fragment is treated as
content, never merged.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-head.min.js"></script>
```

The extension is not in `ypy.max`.

## Merge strategies

| Value of `ypy-head` | Effect |
|---|---|
| `merge` | The head becomes the head of the response: what the response lacks is removed, what is new is added. The default when the target is `body`. |
| `append` | New elements are added and old ones stay. A new `<title>` still replaces the old one. The default for any other target. |
| `re-eval` | On an element of the current head: remove it and add it again on every response, so a script runs again. |

Put `ypy-head` on the `<head>` of the response to choose the strategy for that response. Two elements are the same when
their `outerHTML` is the same, so an unchanged stylesheet is not loaded again.

```html
<head ypy-head="append">
  <title>Orders</title>
  <link rel="stylesheet" href="/assets/orders.css">
</head>
```

A head element with `ypy-preserve="true"` is kept even when the new head lacks it.

## What waits for what

- A new stylesheet is waited for before the swap, so the new page does not flash unstyled.
- A script without `async` or `defer` is loaded and run before the swap.
- A `defer` script is added after the swap, with a preload hint in the meantime.
- Other elements (`meta`, `title`, `link rel=icon`, async scripts) are added at once.
- Removals happen after the additions, so a stylesheet that stays is not fetched twice.

## Events

All four fire on the document, with `document.body` as the event target.

| Event | Cancelable | Detail | When |
|---|---|---|---|
| `ypy:head:before:merge` | yes | the detail of `ypy:before:response` | Before the merge. Cancel it to skip the merge. |
| `ypy:head:before:add` | yes | `headElement` | Before an element is added. Cancel it to leave it out. |
| `ypy:head:before:remove` | yes | `headElement` | Before an element is removed. Cancel it to keep it. |
| `ypy:head:after:merge` | no | `added`, `kept`, `removed` | After the merge. |

## Example

A boosted link to `/settings` brings a full page, `<!doctype html>` with a `<title>Settings</title>` and a
`<link rel="stylesheet" href="/assets/settings.css">` in its `<head>`. After the swap the tab title is "Settings" and `settings.css` is loaded. The target is `body`, so the strategy is
`merge`: a stylesheet of the old page that the new head lacks is removed.

## Notes

- The extension also runs when a page comes back from the [history cache](history-cache.md), using the saved head.
- Scripts added to the head execute. Only merge the head of responses you trust, and keep a strict
  [CSP](csp.md).

## See also

[Navigation](nav.md), [History and caching](../index.md#history), [`ypy-preserve`](../index.md#preserving-content).
