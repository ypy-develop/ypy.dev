---
title: "config.classes"
description: "The classes YPY puts on elements, one entry per state, each a space-separated list. Replace an entry, or set it empty to use data-ypy-state only."
order: 10
---
# `config.classes`

The classes YPY puts on elements, one entry per state, each a space-separated list. Set an entry to replace it,
or to `""` to use the `data-ypy-state` attribute only. See `YpyClasses` in `src/ypy.d.ts`.

| Key | Default | Applied to | While |
|---|---|---|---|
| `request` | `ypy-request` | the requesting element and its indicators | a request is in flight |
| `swapping` | `ypy-swapping` | the swap target | between the start of the swap and the swap itself |
| `settling` | `ypy-settling` | the swap target | the settle delay |
| `added` | `ypy-added` | freshly swapped content | until it settles |
| `indicator` | `ypy-indicator` | elements shown only during a request | always (hidden until a request runs) |
| `offline` | `ypy-offline` | `<html>` | the browser reports no connection |

`offline` also sets `data-ypy-state="offline"` on `<html>`; see [`ypy:offline`](../events/ypy-offline.md).

```js
ypy.config.classes.offline = 'is-offline'
```

```html
<meta name="ypy-config" content='{"classes": {"offline": "is-offline"}}'>
```
