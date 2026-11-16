---
title: "History"
description: "How YPY keeps the address bar and the back button honest, and how a page is restored when the user goes back or forward."
order: 100
---
# History

How YPY keeps the address bar and the back button honest, and how a page is restored when the
user goes back or forward.

## What an entry is

A swap that pushes or replaces a URL (`ypy-push-url`, `ypy-replace-url`, a boosted link) writes
an entry whose state is `{ ypy: true }`. The `ypy` key marks the entry as one YPY can restore.
Other keys of the state are kept.

Entries that YPY did not write (another router, a plain page load, a fragment jump) are not
restored by YPY. The browser handles them. The first swap after a page load stamps the entry
the user is on, so going back to it works.

## Restoring a page

On back or forward to an entry with `ypy: true`, YPY does not use a copy of the page. It asks
the server for the page again:

```
GET /the/url
Accept: text/html
YPY-Current-URL: https://example.com/the/url
YPY-Source: body
YPY-Request-Type: full
YPY-History-Restore-Request: true
```

- `YPY-Request` is left out on purpose, so a server that answers `YPY-Request: true` with a
  fragment answers this request with the whole page.
- The request has the same `mode` and `credentials` as any YPY request, a timeout, and an abort
  controller. A second back or forward aborts the first restore.
- The response is swapped with `outerSync` into the element marked `ypy-history-elt` (or
  `ypy-history-elt`), or into `<body>` when there is none. Only that element changes.

`ypy.config.history` controls it:

| Value | Meaning |
|---|---|
| `true` (default) | push URLs, restore by request |
| `"reload"` | push URLs, and reload the page on back and forward |
| `false` | YPY does not touch history |

## The browser cache

A server that returns a fragment for `YPY-Request: true` and a page for a plain request, on the
same URL, must tell caches so with `Vary: YPY-Request`. See
[History and caching](../guides/history-and-caching.md). Two defenses cover a server that does
not.

### `config.historyRestore.cache`

The `cache` mode of the restore request. By default it is `"no-cache"`: the
browser sends the request to the server, with the validators it has (`ETag`), and uses its
copy only when the server answers `304`. Any value of the `fetch` `cache` option works:

```js
ypy.config.historyRestore = { cache: 'reload' }   // never use a cached copy
ypy.config.historyRestore = { cache: 'default' }  // browser default
```

### `config.historyRestore.fragment`

If the response of a restore is not a whole page (it has no doctype, `<html>` or `<body>`
tag), swapping it would put a fragment where a page belongs. By default YPY
fires `ypy:history:restore:fragment` on `document` and reloads the page, so the browser asks
for the page itself. Nothing is swapped.

```js
document.addEventListener('ypy:history:restore:fragment', (event) => {
  console.warn('history restore got a fragment from', event.detail.ctx.request.action)
})
```

Set `ypy.config.historyRestore = { fragment: false }` to swap whatever comes back.

## Scroll

A push records `[scrollX, scrollY]` in the state of the entry being
left: `{ ypy: true, scroll: [0, 420] }`. When that entry is restored, YPY passes
`scroll: "restore"` and `scrollPosition: [0, 420]` in the request context, so
`__applyScrollPolicy(ctx, "restore")` puts the page back after the swap (`ypy-scroll`).

An entry left by going back or forward carries no position of its own. The browser's scroll
restoration still applies to it where it works (the Navigation API, in Chromium).
Where the browser restores the scroll before the new content arrives (the `popstate` path,
used by Firefox) such an entry comes back at the top until it is left by a push again.

## Events

| Event | Fires | Detail |
|---|---|---|
| `ypy:before:history:update` | before a push or replace (cancelable) | `{history, sourceElement, response}` |
| `ypy:after:history:push`, `ypy:after:history:replace` | after it | `{path}` |
| `ypy:after:history:update` | after it | `{history, sourceElement, response}` |
| `ypy:before:history:restore` | before a restore (cancelable) | `{path, cacheMiss}` |
| `ypy:history:restore:fragment` | the restore got a fragment | `{ctx}` |

## `ypy.ajax()` and `options.request`

The restore uses `ypy.ajax()`, and so can your code. `options.request` is merged into the
request YPY builds:

```js
ypy.ajax('GET', '/report', {
  target: '#report',
  request: {
    cache: 'no-store',                 // any fetch option
    signal: controller.signal,         // combined with YPY's own, either one aborts
    headers: { 'X-Mode': 'print', 'YPY-Request': null }   // null removes a header
  }
})
```

Before, passing `request` replaced the whole object and dropped the `Accept` and
`YPY-Current-URL` headers, `mode`, `credentials`, `abort` (so the timeout no longer aborted) and
`validate`. `options.headers` also removes a header set to `null`.

## The `ypy-history-cache` extension

`ypy-history-cache` keeps up to 10 pages in `sessionStorage` and restores them without a
request. That is a copy of the page as it was, so it can be stale: a list another user changed,
a session that ended, a form with an old token. It is opt in, and the restore request and the
safety net above do not apply to a page it restores. Use it for pages that do not change, and
see its own `refreshOnMiss` option.
