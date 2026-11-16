---
title: "ypy:target:fallback and ypy:target:missing"
description: "ypy:target:fallback fires when the main target takes over from a missing one, and ypy:target:missing when no target is found. Both bubble."
order: 30
---
# ypy:target:fallback and ypy:target:missing

Both bubble from the element that started the request. See [ypy-main](../attributes/ypy-main.md).

## ypy:target:fallback

The target selector matched nothing and the main target took over. Fires before the swap.

| `detail` | |
|---|---|
| `selector` | the selector that failed |
| `used` | `[ypy-main]` or `main` |
| `ctx` | the request context (`ctx.fallback` holds `used`) |

```js
document.addEventListener('ypy:target:fallback', e => {
  console.info(`${e.detail.selector} is gone, swapping into ${e.detail.used}`);
});
```

## ypy:target:missing

No target could be found: a main swap that could not fall back (the response was a fragment,
the page has no main, or `ypy-fallback="false"`), or an out-of-band element or partial whose
target is missing. Cancelable: `preventDefault()` silences the console warning.

| `detail` | |
|---|---|
| `selector` | the selector that failed |
| `ctx` | the request context |
| `task` | the swap task, for out-of-band elements and partials |

```js
document.addEventListener('ypy:target:missing', e => {
  e.preventDefault(); // this page handles it, no warning
  toast(`Nothing to update for ${e.detail.selector}`);
});
```
