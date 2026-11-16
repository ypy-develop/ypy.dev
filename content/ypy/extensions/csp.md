---
order: 70
title: "Content Security Policy: the ypy-csp extension"
description: "Strips ypy attributes from markup that carries no page nonce, enforces Trusted Types and runs author code without unsafe-eval, so a strict CSP stays strict."
---

# Content Security Policy: the `ypy-csp` extension

A page that follows a strict Content Security Policy (CSP) blocks injected scripts, but injected HTML can still carry
`ypy-*` attributes: a `ypy-post` on a link, an `ypy-on:click` on an image. `ypy-csp` closes that gap. Every element that
uses YPY must carry the page nonce, or its YPY attributes are removed before they run.

The whole model, including `config.callbacks`, is described in [Security](../index.md#security). This page is the
reference of the extension.

## Load it

```html
<meta name="ypy-config" content='extensions:"ypy-csp",nonce:true,callbacks:"nonce"'>
<script nonce="{{ nonce }}" src="ypy.min.js"></script>
<script nonce="{{ nonce }}" src="ext/ypy-csp.min.js"></script>
```

- `nonce:true` makes the core process nothing until `ypy-csp` has started. If the extension is missing, the page does
  not run YPY at all.
- `callbacks:"nonce"` lets author code (`ypy-on`, `js:` values, trigger filters) run only when its element has the nonce.
- The extension is not in `ypy.max`. List it in `config.extensions` and load it right after the core.

## How it works

1. At start-up the extension reads the page nonce from the first `<script nonce>`. With no nonce it logs an error and
   blocks YPY.
2. While YPY processes an element, the extension compares the element's own `ypy-nonce` with the page nonce. A
   nonce on a parent never counts. If they differ, every `ypy-*` and `data-ypy-*` attribute of the element is removed
   and `ypy:security:strip` fires.
3. In a response, the page nonce is removed from the text before it is parsed, so injected markup cannot copy it. A
   response that still carries the page nonce after parsing is refused whole.
4. When the response comes from the same origin and its `Content-Security-Policy` header has a nonce, that nonce is
   rewritten to the page nonce, so the server can render a fresh nonce per response.
5. Elements that act before they are initialised (`<ypy-partial>`, out-of-band elements) are checked in the same way.
6. Actions with a `js:` or `javascript:` address, and a boosted form whose submit button changes `formaction` without a
   nonce, are stopped.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-nonce` | the page nonce | Marks the element as markup the server rendered. It must equal the page nonce. |

An `ypy-ignore` marker only counts when it carries the nonce too.

## Events

| Event | Detail | When |
|---|---|---|
| `ypy:security:strip` | `reason` (`missing-nonce` or `nonce-mismatch`), `stripped` (the attribute names) | The attributes of an element were removed. |
| `ypy:security:violation` | `reason`, and `action`, `submitter` or `ctx` for some reasons | An action was stopped. Reasons: `nonce-mismatch-at-eval`, `javascript-url`, `unnonced-submitter`, `page-nonce-in-response`. |

## Settings

| Key | Default | Meaning |
|---|---|---|
| `nonce` | `false` | Requires `ypy-csp`: until it has started, YPY processes no element and sends no request. |
| `safeEval` | `false` | Replaces `Function` and `AsyncFunction` with scripts injected with the page nonce, so author code runs without `unsafe-eval` in the CSP. |

## Example

The server makes a fresh nonce for every response (128 bits from a cryptographic random source), sends it in the
`Content-Security-Policy` header, and stamps it on each element that uses YPY:

```http
Content-Security-Policy: script-src 'nonce-R4nd0m'; trusted-types ypy; require-trusted-types-for 'script'
```

```html
<button ypy-post="/save" ypy-target="#result" ypy-nonce="{{ .Nonce }}">Save</button>
<div id="result"></div>
```

A comment written by a visitor and rendered as `<a ypy-post="/delete">` has no nonce, so the link stays a plain link.

## Notes

- Add `trusted-types ypy` to the CSP to let only YPY touch the DOM sinks. If the browser refuses the `ypy` policy, the
  extension blocks YPY.
- The nonce must be new for every response and must never be written into content that comes from users.
- With `safeEval:true` each distinct expression is compiled once and cached. The names of the parameters must be plain
  identifiers.
- Lock these settings with [`ypy-lock`](lock.md) so a script cannot change them later.

## See also

[Security](../index.md#security) and [`ypy-lock`](lock.md).
