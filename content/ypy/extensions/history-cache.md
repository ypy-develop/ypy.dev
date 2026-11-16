---
title: "Pages kept for back and forward: the ypy-history-cache extension"
description: "Saves a snapshot of each page in sessionStorage so back and forward show it at once with no request, keeping typed values and scroll position."
order: 150
---
# Pages kept for back and forward: the `ypy-history-cache` extension

By default, going back or forward asks the server for the page again (see [History](../index.md#history)).
`ypy-history-cache` changes that. When the user leaves a page, the extension saves a snapshot of it in `sessionStorage`,
and back or forward puts the snapshot on the screen without a request. Typed values, checked boxes and scroll positions
come back with it.

A snapshot is a copy of the page as it was, so it can be stale: a list another user changed, a session that ended, a form
with an old token. Use it for pages that do not change much. `Vary: YPY-Request` on the server still matters for every page
the cache does not cover.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-history-cache.min.js"></script>

<main ypy-history-elt>
  <a href="/items" ypy-boost>Items</a>
</main>
```

`ypy.max` includes the extension but turns it off with `historyCache.disable` set to `true`, so a page that loads `ypy.max`
restores by request until you turn it on. Set the key on the existing object, because replacing `ypy.config.historyCache`
drops the defaults.

```js
ypy.config.historyCache.disable = false
```

The snapshot covers the element marked `ypy-history-elt`, or `<body>` when there is none, and a restore replaces the same
element.

## When a page is saved and restored

A page is saved just before YPY pushes or replaces a URL, and when the user goes back or forward away from it. It is
restored on back or forward to an entry that has a snapshot, and the core does not send its restore request. An entry gets
an id only when YPY wrote it, so a page that came from a plain load or a POST result is left alone and F5 on it still asks
the browser to resubmit.

A page is not saved when `disable` is `true` or `size` is `0`, when any element has `ypy-history="false"`, when
`sessionStorage` is not available, or when a `ypy:history:cache:before:save` listener cancels it. If nothing is stored for
the entry, `ypy:history:cache:miss` fires and the core restores the page by request.

## What a snapshot holds

`{content, head, scroll, title}`: the outer HTML of the history element, the HTML of `<head>`, `window.scrollY` and
`document.title`. Before it is taken, the request classes are removed and elements YPY disabled during a request are enabled
again. The value of each field, the checked state of checkboxes and radios, the selected options of a `<select multiple>` and
the scroll of scrolled elements are written to `data-ypy-history-*` attributes and put back after a restore. `type="file"`
inputs are skipped.

Sensitive fields are skipped: a password input, and any field whose `autocomplete` is `off`, `one-time-code`,
`current-password`, `new-password` or starts with `cc-`. Hidden inputs are kept on purpose, so a restored form still has its
token. The list keeps at most `size` pages and drops the oldest first. When the storage is full, the oldest snapshots go
until the new one fits.

## Configuration

`ypy.config.historyCache`:

| Key | Default | Meaning |
|---|---|---|
| `size` | `10` | pages kept in `sessionStorage`; `0` turns saving off |
| `refreshOnMiss` | `false` | on back or forward with no snapshot, reload the page |
| `disable` | `false` | turn the cache off without unloading the extension |
| `swapStyle` | `"outerSync"` | swap style used to put a snapshot back |
| `encrypt` | `false` | seal the snapshots: `true` or `"memory"` |

`ypy-history="false"` anywhere on a page stops that page from being saved. Use it on a page that shows something private or
that is never right when old, such as a balance or an inbox.

## Events

All fire on `document`.

| Event | Cancelable | `detail` |
|---|---|---|
| `ypy:history:cache:before:save` | yes: skips saving | `{target, head}` |
| `ypy:history:cache:after:save` | no | `{content, head, scroll, title}` |
| `ypy:history:cache:hit` | yes: ignore the snapshot | `{path, item}`; `item` may be replaced |
| `ypy:history:cache:miss` | no | `{path, refreshOnMiss}`; set `refreshOnMiss` to reload |
| `ypy:history:cache:before:restore` | no | `{head, ready}`; set `ready` to a promise to delay the swap |
| `ypy:history:cache:after:restore` | no | `{head, ready, item}` |

Cancelling `hit` makes the core restore the page by request:

```js
document.addEventListener('ypy:history:cache:hit', (event) => {
  if (event.detail.path.startsWith('/inbox')) event.preventDefault()
})
```

If [`ypy-head`](head.md) is loaded, it merges the saved `<head>` in `before:restore` and holds the swap until stylesheets
have loaded, which is what `ready` is for.

## Encrypted snapshots

`historyCache.encrypt` seals each snapshot so it is not readable in the browser storage: WebCrypto, AES-GCM, a 256-bit key
made for the session and a new IV for every snapshot.

```js
ypy.config.historyCache = { size: 10, encrypt: true }       // key in sessionStorage: snapshots survive a reload
ypy.config.historyCache = { size: 10, encrypt: 'memory' }   // key in memory: after a reload the old ones are gone
```

Sealing is asynchronous, so `after:save` fires when the snapshot is stored. Without WebCrypto, which needs a secure context,
nothing is stored and a warning goes to the console. A snapshot is never written in the clear when encryption was asked for.
The key is as safe as the page, because a script of the page can read it. Encryption protects the storage at rest.

## Signing out

Call `ypy.history.clear()` when the user signs out, so the next person on the computer cannot go back to the pages. It
forgets every snapshot of the tab, and with encryption it also replaces the key.

```js
document.getElementById('logout').addEventListener('ypy:after:request', () => ypy.history.clear())
```

## Pitfalls

- A restored page is as old as the snapshot. After a change made in another tab, going back shows the old state, so mark such
  pages with `ypy-history="false"`.
- Only the history element is replaced. State outside it, such as a header or a sidebar, is not part of the snapshot.
- Do not rely on `sessionStorage` surviving. A private window or a full quota makes the page work as without the extension.
- The safeguards for restoring by request (`historyRestore.cache` and the fragment check) do not apply to a snapshot.

See [Cache](cache.md) for the response cache of requests.
