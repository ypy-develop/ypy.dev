---
order: 40
title: "Browser loading indicator: the ypy-browser-indicator extension"
description: "Shows the browser's own tab spinner and stop button while a YPY request runs, using the Navigation API. Does nothing where the API is missing."
---

# Browser loading indicator: the `ypy-browser-indicator` extension

A normal link click shows a spinner in the browser tab and a stop button in the toolbar. A YPY request does not: the
page stays still. `ypy-browser-indicator` brings those signs back for the requests you choose, so a visitor who looks
at the tab sees that something is loading and can press stop.

It uses the browser's Navigation API. In a browser that does not have it, the extension does nothing and the page works
as before.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-browser-indicator.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-browser-indicator` in it. It is listed with
the other official extensions in [Extensions](index.md).

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-browser-indicator` | `true` or `false` | Shows the indicator while the requests of this element run. Any value except `false` turns it on. `false` turns it off for an element, also when the page default is on. |

## Settings

| Key | Default | Meaning |
|---|---|---|
| `boostBrowserIndicator` | `false` | Shows the indicator for every boosted link and form, without the attribute on each one. |

## Example

```html
<meta name="ypy-config" content='{"boostBrowserIndicator": true}'>

<body ypy-boost="true">
  <a href="/reports">Reports</a>
  <button ypy-get="/export-status" ypy-target="#status" ypy-browser-indicator>Check export</button>
  <a href="/preview" ypy-boost="false">Preview</a>
</body>
```

The first link shows the indicator because it is boosted and the setting is on. The button shows it because of the
attribute. The last link is a plain link, so the browser shows its own.

## Notes

- Several requests at once share one indicator. It goes away when the last of them ends.
- When the visitor presses the browser's stop button, the extension aborts the requests that are showing the indicator.
- The indicator is cleared before YPY writes the address of a new page to the history, so the two do not fight.
- The indicator is for the eyes of the visitor. For screen readers, use [`ypy-announce`](../reference/attributes/ypy-announce.md)
  and [`ypy-busy`](../reference/attributes/ypy-busy.md).

## See also

[Navigation](nav.md) for boosted links and the page progress bar, [Styling](../reference/styling.md) for in-page
indicators.
