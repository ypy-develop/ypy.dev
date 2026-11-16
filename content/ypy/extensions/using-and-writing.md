---
title: "Extensions"
description: "How to choose, load, restrict and write YPY extensions: the core handles requests and swaps, and everything else is a small script you add."
order: 10
id: "GUIDE-extensions"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en"]
---

# Extensions

The core does what every page needs: requests, swaps, morph, history, errors, accessibility. Everything else is an
extension: a small script you load after the core, or get all at once with `ypy.max`. The core has a budget of
18,432 bytes min+gzip and each extension 6,144 bytes; `ypy build --sizes` checks both in CI.

This guide covers how to choose, load, restrict and write extensions. The reference with the extension API and the
hooks is [Official extensions](index.md).

## Choose

| You want | Extension | Attribute or setting |
|---|---|---|
| a request to repeat when the network fails | `ypy-retry` | `ypy-retry`, `config.retry` |
| live updates from the server | `ypy-sse` (Server-Sent Events) | `ypy-sse:connect` |
| a stream of HTML, signals, redirects and events over one response | `ypy-stream` ([reference](stream.md)) | `ypy-stream` |
| two-way messages | `ypy-ws` (WebSocket) | `ypy-ws:connect`, `ypy-ws:send` |
| client state without a build step | `ypy-signals` ([guide](signals.md)) | `ypy-signals`, `ypy-bind`, `ypy-show`, `ypy-text` |
| reactive expressions on elements | `ypy-live` | needs `callbacks` on |
| one response into several elements | `ypy-targets` | `ypy-targets` (see also [Multiple targets](../guides/multiple-targets.md)) |
| load a page's data on hover or touch | `ypy-preload` | `ypy-preload` |
| a placeholder while the request runs | `ypy-pending` | `ypy-pending` |
| a download from a response | `ypy-download` | `ypy-swap="download"` |
| insert or update by `id` | `ypy-upsert` | `ypy-swap="upsert"` |
| the browser's own loading indicator | `ypy-browser-indicator` | |
| snapshot cache for the back button | `ypy-history-cache` | [History and caching](../guides/history-and-caching.md) |
| merge `<head>` content (title, styles) | `ypy-head` | |
| skip unchanged polling swaps | `ypy-ptag` | |
| multipart responses | `ypy-multipart` | |
| the answer of a prompt sent as a header | `ypy-prompt` | `ypy-prompt`, `YPY-Prompt` |
| Alpine.js on the same page | `ypy-alpine-compat` | |
| nonce gate and Trusted Types for strict CSP | `ypy-csp` | `config.nonce` ([Security](../reference/security.md)) |
| freeze the configuration | `ypy-lock` | `locked:true` |

The full table, with the file of each one and whether `ypy.max` includes it, is in the
[reference](index.md#official-extensions). Pages being written by other tasks (layers, the
navigation extras, the cache and optimistic UI, the inspector) are listed in the [index](../index.md#coming-in-the-reference).

## Load

After the core, one `<script>` per extension:

```html
<script src="/ypy.min.js"></script>
<script src="/ext/ypy-sse.min.js"></script>
<script src="/ext/ypy-retry.min.js"></script>
```

Or one file with the common ones: `ypy.max.min.js`. It bundles every extension marked "yes" in the reference table
and leaves out `ypy-csp`, `ypy-head`, `ypy-multipart`, `ypy-prompt` and `ypy-ptag`, which you load yourself when
you need them.

An extension that is not in the page costs nothing and its attributes do nothing and break nothing. That is the
reason to prefer `ypy.min.js` plus the two or three extensions you use over `ypy.max.min.js` on a page where bytes
count ([Performance](../essays/performance.md)).

## Restrict

By default any script that calls `ypy.registerExtension` is accepted. On a page where that is too open (user content
near the `<head>`, third-party scripts), allow only the names you ship:

```html
<meta name="ypy-config" content='extensions:"ypy-lock,ypy-csp,ypy-sse,ypy-retry",locked:true'>
```

- The list matches the name an extension registers with, and that name is the file name without `.js`:
  `ypy-sse.js` registers as `ypy-sse` and `ypy-preload.js` as `ypy-preload`. Write the full name, with the
  `ypy-` prefix; the short names (`sse`, `preload`) are not accepted.
- An extension that is not on the list is refused: it logs a warning and its hooks never run.
- List `ypy-lock` too when you set `locked:true`, or the lock is refused and the configuration stays writable.

## Write one

An extension is a script that calls `ypy.registerExtension(name, hooks)`. It uses only the public `ypy.*` API and
the object that `init` receives; nothing private (`ypy.__x`), which only exists in the source build.

A hook is named after an event: the event name with every colon replaced by an underscore and a `ypy_` prefix
already in it. `ypy:before:request` is `ypy_before_request`; `ypy:after:swap` is `ypy_after_swap`. A hook receives
`(elt, detail)`; returning `false` cancels a cancelable event.

An extension that adds an attribute, `ypy-flash`, which makes the target of a swap flash:

```js
// ypy-flash.js
(() => {
  let api

  ypy.registerExtension('ypy-flash', {
    init(internalApi) {
      api = internalApi
    },

    // after every swap, look at the element that sent the request
    ypy_after_swap(elt, detail) {
      const color = api.attributeValue(detail.ctx.sourceElement, 'ypy-flash')
      if (!color) return
      const target = detail.ctx.target
      if (!(target instanceof Element)) return

      target.setAttribute('data-flash', color) // style it: [data-flash="amber"] { ... } or a Tailwind variant
      setTimeout(() => target.removeAttribute('data-flash'), 600)
    },

    // forget what you set up when an element leaves
    ypy_before_cleanup(elt) {
      elt.removeAttribute('data-flash')
    },
  })
})()
```

```html
<script src="/ypy.min.js"></script>
<script src="/ypy-flash.js"></script>

<button ypy-post="/save" ypy-target="#panel" ypy-flash="amber"
        class="rounded bg-sky-600 px-3 py-2 text-white">Save</button>
<div id="panel" class="rounded border p-4 data-[flash=amber]:bg-amber-100 transition-colors"></div>
```

Rules of the road:

1. Read attributes with `api.attributeValue(elt, 'ypy-name')`, never `getAttribute('ypy-...')`: the resolver
   knows the `data-ypy-` spelling and `:inherited`.
2. Do one job in one hook. Setup in `ypy_after_init`, work in the request or swap hooks, teardown in
   `ypy_before_cleanup`. Nothing you start survives the element.
3. Do not compile text into code. If you need author code, go through `api.executeJavaScript`, which obeys
   `config.callbacks`; this is the only door.
4. Stay small: 6 kB min+gzip is the budget of an official extension, and a good target for yours.
5. Fire events named `ypy:<noun>:<verb>` (`api.triggerEvent(elt, 'ypy:flash:done', detail)`), so other extensions
   and pages can hook them the same way.

### Test it

The browser suite tests an extension in a fresh page: a test file loads the script from `src/ext/` the way a page
does and exercises it. Model yours on `test/tests/ext/ypy-download.js` (it shows the set-up and tear-down helpers),
put it in `test/tests/ext/`, and run only it:

```sh
go run ./cmd/ypy test --files 'test/tests/ext/ypy-flash.js'
go run ./cmd/ypy test --dist     # again against the built, minified files
```

`--dist` matters: it fails an extension that reaches for a private member, which works from `src/` and breaks in
`dist/`. For the server side of an extension that talks to a Go handler, see [Testing your app](../server/testing.md).

## See also

[Official extensions](index.md) · [Events](../reference/events/index.md) · [Security checklist](../guides/security-checklist.md) ·
[Performance](../essays/performance.md)
