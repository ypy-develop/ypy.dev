---
title: "Security"
description: "YPY is meant to be safe on a page that follows a strict Content Security Policy (CSP), and to stay safe when an attacker can put markup into a response."
order: 140
---
# Security

YPY is meant to be safe on a page that follows a strict Content Security Policy (CSP), and
to stay safe when an attacker can put markup into a response. This page lists what the core
does, the settings that control it, and the policy to deploy it with.

Contents: [Threat model](#threat-model) · [Settings](#settings) · [Author code: `callbacks`](#author-code-callbacks) ·
[The `ypy-config` meta tag and its lock](#the-ypy-config-meta-tag-and-its-lock) ·
[Same-origin requests](#same-origin-requests-selfrequestsonly) · [Nonce gating](#nonce-gating-the-ypy-csp-extension) ·
[Trusted Types](#trusted-types) · [Recommended CSP](#recommended-csp) · [Events](#events)

## Threat model

YPY turns attributes into behavior (`ypy-get`, `ypy-on`, `ypy-vals`...). An attacker who can write
markup into your pages, through stored content or a reflected parameter, does not need to run
a script: an element with `ypy-post` or `ypy-on:click` is enough. Three layers answer that.

| Layer | Stops | Where |
|---|---|---|
| Author code policy | `ypy-on`, `js:` values and `[filter]` triggers compiling code at all | core, `config.callbacks` |
| Nonce gate | attributes of elements that the server did not stamp with the page nonce | `ypy-csp` extension, `config.nonce` |
| Same-origin requests | an injected URL leaking data to another origin | core, `config.mode` |

What none of them does: protect against script that already runs on your page. Such a script
can call `ypy.config.lock` (from `ypy-lock`) or replace the core as easily as it can call `fetch`. The lock below
is for HTML injection and careless code, not for script injection. A strict script CSP is the
answer to that, and everything here works with one.

## Settings

Set them in a meta tag in `<head>` (see below) or on `ypy.config` before the first element is
processed.

| Setting | Values | Default |
|---|---|---|
| `callbacks` | `"none"`, `"nonce"`, `"all"` | `"none"` |
| `nonce` | `true` | unset |
| `locked` | `true` | unset |
| `mode` | `"same-origin"`, `"cors"`, `"no-cors"` | `"same-origin"` |
| `selfRequestsOnly` | `true` (same-origin) or `false` (cors); alias of `mode` | `true` |
| `inlineScriptNonce` | a nonce string for the scripts YPY re-creates | unset |
| `extensions` | comma separated list of the extensions allowed to register | `""` (any) |

## Author code: `callbacks`

Four places compile text from an attribute into code: `ypy-on` handlers, `js:` values (in
`ypy-vals`, `ypy-headers`, `ypy-confirm` and an action), and the `[condition]` filters of
`ypy-trigger`. One function, `__executeJavaScript`, does all of it, and `config.callbacks`
decides there.

| Value | Meaning |
|---|---|
| `"none"` | no author code is ever compiled |
| `"nonce"` | code runs only through the `ypy-csp` extension, which checks the nonce of the element when the code is about to run; without the extension nothing runs |
| `"all"` | code runs: every `ypy-on` handler and `js:` value is compiled |

Any other value, such as `"yes"` or `true`, counts as `"none"`: a typo cannot turn code on.

With the default (unset), the value is `"none"`. A page that has `ypy-on` handlers or `js:` values (for example
one converted by `ypy migrate`) sets `callbacks` on purpose, with the meaning in mind.

When code is blocked:

- nothing is compiled, and the `ypy:scope` extension hook is not called;
- the first time, one warning is logged: `ypy: author code is off (config.callbacks is "none")`;
- `ypy:security:blocked` fires on the element every time, with `{reason, code}`;
- the result is empty: a `[filter]` is false (the trigger does not fire), an `ypy-confirm="js:..."`
  denies the request, a `js:` value is dropped and the request goes out without it, an
  `ypy-on` handler does nothing.

To turn code on, write it down:

```html
<meta name="ypy-config" content='callbacks:"all"'>
```

Extensions that evaluate expressions (`ypy-live`) go through the same function, so they need
`"all"` too, or `"nonce"` with `ypy-csp` and a nonce on each element.

## The `ypy-config` meta tag and its lock

```html
<head>
  <meta name="ypy-config" content='extensions:"ypy-csp",nonce:true,callbacks:"nonce",locked:true'>
  <script src="ypy.min.js"></script>
</head>
```

- Only a tag inside `<head>` counts, and only the first one. A tag injected into `<body>`, or
  added after YPY started, is ignored.
- The lock is the [`ypy-lock` extension](../extensions/index.md#ypy-lock), not part of the core: load `ext/ypy-lock.js`
  right after the core (`ypy.max` already includes it). The core only merges `locked:true` into the config;
  nothing is frozen until the extension has started, so a page that sets `extensions` must list `ypy-lock` in it,
  or the extension is refused and the config stays writable.
- `locked:true` (or `ypy.config.lock()` at run time) freezes `mode`, `callbacks`, `extensions`,
  `nonce`, `inlineScriptNonce`, the approved extension list and the `ypy.config` object itself.
  A later assignment is ignored and a warning is logged, except when it assigns the value the
  setting already has. Everything else stays writable. The lock cannot be undone.
- Lock after you have set what you want. On a locked page `callbacks` keeps its
  default (`"none"`) when you did not write it down, so write `callbacks` explicitly.
- A tag placed in `<head>` before YPY's own script by an attacker is as powerful as any other
  injection into `<head>`; keep the head out of reach of user data.

## Same-origin requests: `selfRequestsOnly`

Every request is sent with fetch `mode: "same-origin"`, so the browser refuses to send it to
another origin and the promise rejects (YPY fires `ypy:error`). `selfRequestsOnly` is another
name for the same switch: `selfRequestsOnly:false` means `mode:"cors"`.

Nothing on an element changes the mode: `ypy-config`, an `ypy-boost` value and the options of
`ypy.ajax` are overridden before any page code sees the request. A listener you write on
`ypy:config:request` can still set `ctx.request.mode` on purpose.

## Nonce gating: the `ypy-csp` extension

Gating makes the server prove which elements are its own. Every element that carries YPY
attributes has to carry the page nonce in `ypy-nonce`; the others lose their
`ypy-*` and `data-ypy-*` attributes when they are processed.

```html
<meta name="ypy-config" content='extensions:"ypy-csp",nonce:true,callbacks:"nonce"'>
<script nonce="{{ nonce }}" src="ypy.min.js"></script>
<script nonce="{{ nonce }}" src="ext/ypy-csp.min.js"></script>

<button ypy-post="/save" ypy-nonce="{{ nonce }}">Save</button>
```

What the extension does:

- reads the page nonce from the first `script[nonce]` and compares it with the element's own
  nonce attribute. A nonce on a container, or an inheritable one, never vouches for the elements
  inside it;
- strips the attributes of the others and fires `ypy:security:strip` with
  `{reason: "missing-nonce" | "nonce-mismatch", stripped}`;
- before code runs (`callbacks:"nonce"`), checks the nonce again, so an element whose nonce was
  changed after it was processed does not run code (`ypy:security:violation`, reason
  `nonce-mismatch-at-eval`);
- removes the page nonce from every response, because the server cannot know it: its presence
  means someone copied it, and a response that still carries it after that (written with spaces
  around `=` or with character references) is refused whole
  (`ypy:security:violation`, reason `page-nonce-in-response`). A nonce the server itself put in a
  same-origin response is replaced by the page nonce. That nonce is read from the response's
  `Content-Security-Policy` header only, never from a `<meta>` inside the body: send the CSP
  as a header;
- vetoes an out-of-band element (`ypy-swap-oob`) and a `<ypy-partial>` without the page nonce
  (the core asks the extensions through `ypy:before:oob` before it takes them out of the response):
  the first stays ordinary content with its YPY attributes stripped, the second is dropped.
  Stamp the nonce on both;
- cancels an action URL that starts with `javascript:` or `js:`, and a boosted form whose
  submitter has a `formaction` but no nonce;
- honors an `ypy-ignore` marker only when it carries the nonce, so injected markup cannot switch
  processing off;
- with `config.safeEval:true`, compiles code by injecting a nonced script instead of
  `new Function`, so the CSP does not need `unsafe-eval`.

**Fail closed.** `nonce:true` is a promise that gating is on. Until the extension has started,
the core processes no element and sends no request, and logs
`ypy: config.nonce is set but ypy-csp did not start`. The extension does not start when it is not
in `extensions` (it is refused), when no `<script>` has a nonce, or when the CSP blocks the `ypy`
Trusted Types policy. Without `nonce:true`, an extension that is not approved or not loaded leaves
the page without gating; one that is loaded but cannot start (no nonce, blocked policy) blocks all
processing.

Gating is off by default because it needs the server to stamp the nonce on every element.

## Trusted Types

The core creates the Trusted Types policy `ypy` when the browser has Trusted Types. It passes
strings through: the trust comes from the same-origin rule and, when you use it, the nonce gate.
Every string YPY turns into DOM (parsing a response, re-creating a script) goes through it, so
a page can enforce the policy:

```
Content-Security-Policy: require-trusted-types-for 'script'; trusted-types ypy
```

If the CSP forbids the policy (`trusted-types` without `ypy`), YPY fails closed: a response is
not parsed (the request ends in `ypy:error` with `Trusted Types policy 'ypy' is blocked by the
CSP`) and scripts in responses are not re-created. A browser without Trusted Types gets a plain
pass-through.

Swaps (`innerHTML`, `outerHTML`, `beforeend`, morph) complete under `require-trusted-types-for`.

## Recommended CSP

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

- `connect-src 'self'` is the second lock behind `selfRequestsOnly`.
- No `unsafe-eval` and no `unsafe-inline`. Author code then needs `callbacks:"none"` (nothing
  uses it), or `callbacks:"nonce"` with `ypy-csp` and `safeEval:true`.
- Scripts that arrive in a response run only if they carry the nonce. YPY re-creates them
  so they execute; set `inlineScriptNonce` only if you trust the responses of your server
  to carry scripts, because it gives all of them the nonce.
- Add the names of any other Trusted Types policies your page creates to `trusted-types`.

Pair the CSP with server-side escaping. The gate is the safety net, not the first line.

## Events

| Event | Fired by | Detail |
|---|---|---|
| `ypy:security:blocked` | core, `ypy-multipart`, `ypy-ws` | `{reason: "callbacks-none" \| "callbacks-nonce", code}`, or `{reason: "redirect"}` (a redirect header or part), or `{reason: "websocket-origin", url}` |
| `ypy:security:strip` | `ypy-csp` | `{reason, stripped}` |
| `ypy:security:violation` | `ypy-csp` | `{reason, ...}`: `javascript-url`, `unnonced-submitter`, `nonce-mismatch-at-eval` |

## Redirects, scripts in responses and sockets

| Topic | What YPY does |
|---|---|
| Redirect header | `YPY-Redirect` (also in a multipart part) is resolved against the page and followed only when it is an http(s) URL. `javascript:`, `data:` and the like are dropped. It also stays on the page's own origin unless `config.mode` is `"cors"`. A dropped redirect fires `ypy:security:blocked` with reason `redirect`. |
| Scripts in responses | They run by default: YPY re-creates each `<script>` of a swapped response so that it runs. Set `allowScriptTags:false` to re-create them inert, and use the CSP (see above) to decide which ones may run. |
| Sockets | `ypy-ws` keeps sockets on the page's origin unless `config.mode` is `"cors"`. |
| Trusted Types | The `ypy` policy passes its input through (trust comes from same-origin and the gate), so a string handed to `ypy.swap({text})` gets no Trusted Types protection. Never pass untrusted text to it. |
| History snapshots | `ypy-history-cache` does not store password, `autocomplete="off"`, `cc-*` and `one-time-code` fields. Hidden fields are stored. Call `ypy.history.clear()` when the user signs out. |
