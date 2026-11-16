---
order: 210
title: "Configuration lock: the ypy-lock extension"
description: "Freezes the security settings of YPY once the page has chosen them, so a script that runs later, including an injected one, cannot change them."
---

# Configuration lock: the `ypy-lock` extension

`ypy.config` is a plain object. A script that runs after yours, including one injected into the page, could set
`callbacks` to `"all"` or add an extension. `ypy-lock` freezes the settings that matter for security once the page has
chosen them.

## Load it

Set `locked:true` in the first `ypy-config` meta tag and load the extension right after the core, before any other
script:

```html
<meta name="ypy-config" content='locked:true,callbacks:"none"'>
<script src="ypy.min.js"></script>
<script src="ext/ypy-lock.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-lock` in it too (for example
`extensions:"ypy-csp,ypy-lock"`). Without it in the list the extension is not registered and nothing is locked: YPY logs
an error that says so.

## What is frozen

| Frozen | Why |
|---|---|
| `mode` | decides whether requests may leave the origin |
| `callbacks` | decides whether author code runs |
| `extensions` | the list of extensions that may register |
| `nonce`, `inlineScriptNonce` | the nonce gate of [`ypy-csp`](csp.md) |
| `scripts` | whether `<script>` in a response runs |
| `locked`, and the `ypy.config` object itself | so it cannot be replaced |

After the lock, an assignment such as `ypy.config.callbacks = 'all'` is ignored and a warning is written to the
console. The properties cannot be redefined or deleted. The lock cannot be undone for the life of the page.

## Settings

| Key | Default | Meaning |
|---|---|---|
| `locked` | `false` | `true` in the meta tag closes the lock when the extension starts. |

You can also lock at run time with `ypy.config.lock()`, after your own code has set what it needs.

## Example

A page that allows only same-origin requests and no author code:

```html
<meta name="ypy-config" content='extensions:"ypy-lock",mode:"same-origin",callbacks:"none",locked:true'>
<script src="/assets/ypy.min.js"></script>
<script src="/assets/ext/ypy-lock.min.js"></script>
```

## Notes

- Other settings (`defaultSwap`, `history`, the classes) stay writable.
- Lock before any third-party script runs. A script that ran earlier may have changed the values already.

## See also

[`ypy-csp`](csp.md), [Security](../index.md#security), [Configuration](../reference.md#configuration).
