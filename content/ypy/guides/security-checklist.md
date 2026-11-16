---
title: "Security checklist"
description: "A pre-launch security checklist for YPY: escaping, author code, CSP, same-origin, CSRF, cache headers, config lock, nonce gating, streams and tests."
order: 90
id: "GUIDE-security-checklist"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en", "topic/security"]
---

# Security checklist

Go through this list before you ship. Every item is something to do or to verify. The reasons and the settings are in
[Security](../reference/security.md), and the findings of the last review are in the
[audit](../security/audit-2026-11-15.md). There is also a Portuguese version, [Checklist de
segurança](../pt-BR/guides/security-checklist.md).

The idea behind all of it is that YPY turns attributes into behavior. An attacker who can write markup into your pages
(stored content, a reflected parameter) does not need a script, because a `ypy-post` on a button is enough. Three layers
defend against that, and none of them replaces escaping on the server.

## 1. The server escapes everything

- [ ] Every value you put in HTML goes through the template engine's escaping (`html/template`, Jinja, ERB, Blade,
      Twig, templ). Fragments are HTML like any other.
- [ ] User text never lands inside a tag that YPY reads (`<div ` + user data) or in an attribute value that starts
      with `ypy-` (`ypy-get="` + user data). Escape attribute values, and never let a user choose an attribute name.
- [ ] User content (comments, profiles) is rendered in a region of the page you can reason about, sanitized on the
      server, and not allowed to carry `ypy-*` or `data-ypy-*` attributes. Strip them in your sanitizer.
- [ ] Responses to YPY are `Content-Type: text/html; charset=utf-8` and `X-Content-Type-Options: nosniff`.
- [ ] Values you send in `YPY-Trigger`, `YPY-Location` or `YPY-Redirect` come from your code, not from the request.
      A redirect target taken from a query parameter is an open redirect: accept only paths on your own site.

## 2. Author code stays off

- [ ] The page keeps the default `callbacks:"none"` and does not set `callbacks:"all"` (see
      [Author code](#author-code) below). Handlers of a project converted by `ypy migrate` do not run until you
      decide: move the logic to a script, or choose `"nonce"`.
- [ ] No `ypy-on:...` handler, no `js:` in `ypy-vals`/`ypy-confirm`, no `[condition]` in `ypy-trigger` unless you
      decided on one of the two safe ways:
  - `callbacks:"nonce"` with the `ypy-csp` extension and `safeEval:true`, or
  - move the logic to a script file that listens to events ([Events](../reference/events/index.md)).

## 3. A strict Content Security Policy

- [ ] Send a policy with no `unsafe-eval` and no `unsafe-inline`:

```
Content-Security-Policy:
  default-src 'self';
  script-src 'nonce-{{ nonce }}' 'strict-dynamic';
  object-src 'none';
  base-uri 'none';
  connect-src 'self';
  require-trusted-types-for 'script';
  trusted-types ypy
```

- [ ] `connect-src 'self'` is there: it is the second lock behind YPY's own same-origin rule.
- [ ] Add the names of other Trusted Types policies your page creates to `trusted-types`.
- [ ] No inline `style` attributes written by you either; YPY writes none. Tailwind and a stylesheet need none.

## 4. Same-origin requests

- [ ] `config.mode` is left at `"same-origin"` (the default). Change it only for a page
      that really talks to another origin, and set CORS on that server.
- [ ] No endpoint accepts a state change through `GET`. `ypy-get` is for reading; `ypy-post`, `ypy-put`,
      `ypy-patch`, `ypy-delete` change things.

## 5. CSRF

A site that authenticates with cookies needs protection against cross-site requests. Use the one your framework gives
you and check that YPY requests carry the token:

- [ ] Forms carry the token in a hidden field (the framework's form helper does it). That covers `<form ypy-post>`.
- [ ] Requests that no form backs (`ypy-delete` on a button, `ypy-patch` on a toggle) carry it in a header. Set it
      once on the layout, or add it to every request from one listener:

```html
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ csrfToken }}"}'>
```

```js
document.addEventListener('ypy:config:request', (event) => {
  event.detail.ctx.request.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]').content
})
```

- [ ] Cookies are `SameSite=Lax` (or `Strict`), `Secure` and `HttpOnly` for the session.
- [ ] An expired token answers a status the page can show (419 or 403), and the layout reloads or explains it.
      Each [recipe](../server/recipes/go.md) shows the header name its framework reads.

## 6. Caches and `Vary`

- [ ] Every route that answers a fragment to YPY and a page to a browser sends `Vary: YPY-Request`. A cache that
      ignores it can serve the fragment as a whole page, or one user's fragment to another.
- [ ] Private fragments send `Cache-Control: private, no-store` (or `private` with a short life), never a
      public cache header.
- [ ] A history restore request carries `YPY-History-Restore-Request: true` and no `YPY-Request`; the server
      answers a full page to it ([History](../reference/history.md)).

## 7. Lock the configuration

- [ ] Load `ypy-lock` and set `locked:true`, listing `ypy-lock` in `extensions`, after you have written the settings
      you want. The `<meta name="ypy-config">` tag only counts inside `<head>`, so keep user data out of the head.
- [ ] `extensions` lists only the names you ship ([Extensions](../extensions/using-and-writing.md#restrict)).

```html
<meta name="ypy-config" content='extensions:"ypy-lock,ypy-csp",nonce:true,callbacks:"nonce",locked:true'>
```

The lock guards against careless code and markup injected into the head. A script that already runs on your page gets
past it, and only the CSP stops that script.

## 8. Nonce gating, for pages that show user markup

- [ ] When users can put markup on a page, enable `ypy-csp` with `nonce:true`. The server stamps every YPY element
      it renders with the page nonce (`ypy-nonce="{{ nonce }}"`); elements without it lose their `ypy-*` attributes.
- [ ] A new nonce on every response, never a constant, never in a cached page.
- [ ] `inlineScriptNonce` is unset unless you trust the scripts in your own responses.

## 9. Streams and sockets

- [ ] The route behind `ypy-sse:connect` or `ypy-ws:connect` checks the session like any other route, and checks
      the `Origin` of a WebSocket upgrade.
- [ ] A broadcaster sends each user only what that user may see; filter per connection, not on the client.
- [ ] Do not put secrets in the stream URL or in `ypy-vals`: both appear in logs and in the page source.

## 10. Verify

- [ ] `ypy check` and the security tests of your own handlers pass ([Testing your app](../server/testing.md)): write one that
      posts a payload with `ypy-post` markup and asserts it comes back escaped.
- [ ] Open the page with the console on: `ypy: author code is off` and `ypy:security:blocked` events tell you
      something tried to run code; each one is a finding to explain.
- [ ] Run the tests against the built files: `go run ./cmd/ypy test --dist`.

## Author code

Four places compile text from an attribute into code: `ypy-on` handlers, `js:` values (`ypy-vals`, `ypy-headers`,
`ypy-confirm`), the `[condition]` of `ypy-trigger`, and expressions of `ypy-live`. One function decides, with
`config.callbacks`:

| Value | Meaning |
|---|---|
| `"none"` | nothing is compiled. The default |
| `"nonce"` | code runs only through `ypy-csp`, after it checks the element's nonce |
| `"all"` | code runs. Never the default |

Any other value counts as `"none"`: a typo cannot turn code on. When code is blocked, one warning is logged and
`ypy:security:blocked` fires on the element with `{reason, code}`; the filter is false, the confirm denies, the
`js:` value is dropped.

## See also

[Security reference](../reference/security.md) · [Audit 2026-11-15](../security/audit-2026-11-15.md) ·
[Extensions](../extensions/using-and-writing.md) · [Migration](../migration/index.md) · [Recipes](../server/recipes/go.md)
