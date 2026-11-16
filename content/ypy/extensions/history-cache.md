---
title: "Pages kept for back and forward: the ypy-history-cache extension"
description: "Saves a snapshot of each page in sessionStorage so back and forward show it at once with no request, keeping typed values and scroll position."
order: 150
---
# Pages kept for back and forward: the `ypy-history-cache` extension

By default, going back or forward asks the server for the page again (see [History](../reference/history.md)). `ypy-history-cache`
changes that: when the user leaves a page, the extension saves a snapshot of it in `sessionStorage`, and back or
forward puts the snapshot on the screen without a request. Typed values, checked boxes and scroll positions come
back with it.

A snapshot is a copy of the page as it was. It can be stale: a list another user changed, a session that ended, a form
with an old token. Use it for pages that do not change much, and read [History and caching](../guides/history-and-caching.md)
for the server side (`Vary: YPY-Request`), which still matters for every page the cache does not cover.

## Loading it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-history-cache.js"></script>

<main ypy-history-elt class="mx-auto max-w-3xl p-6">
  <a href="/items" ypy-boost class="text-sky-700 underline">Items</a>
</main>
```

`ypy.max` includes the extension but turns it off: it sets `historyCache.disable` to `true`, so a page that loads
`ypy.max` restores by request as before until you turn it on:

```js
ypy.config.historyCache.disable = false
```

Set the key on the existing object; replacing `ypy.config.historyCache` with a new object drops the defaults
(`size`). If you set `ypy.config.extensions`, list `ypy-history-cache` in it, or the
extension is refused. Without the extension, `ypy.config.historyCache` does nothing.

The snapshot covers the element marked `ypy-history-elt`, or `<body>` when there is none. On restore the same element is
replaced.

## When a page is saved and restored

- Saved: just before YPY pushes or replaces a URL, and when the user goes back or forward away from a page. The page
  that is being left is the one saved.
- Restored: on back or forward to an entry that has a snapshot. The core does not send its restore request.
- Not on a fresh page: an entry gets an id only when YPY wrote it (`history.state.ypy`). A page that came from a plain
  load or a POST result is left alone, so F5 on it still asks the browser to resubmit.

A page is not saved when:

- `ypy.config.historyCache.disable` is `true`, or `size` is `0`;
- any element of the page has `ypy-history="false"`;
- `sessionStorage` is not available;
- a `ypy:history:cache:before:save` listener cancels it.

If nothing is stored for the entry, `ypy:history:cache:miss` fires and the core restores the page by request as usual.

## What a snapshot holds

`{content, head, scroll, title}`: the outer HTML of the history element, the HTML of `<head>`, `window.scrollY` and
`document.title`. Before it is taken:

- the classes of `ypy.config.classes.request` are removed, and elements disabled by YPY during a request are enabled
  again;
- the value of each field, the checked state of checkboxes and radios, the selected options of a `<select multiple>`
  and the scroll position of scrolled elements are written to `data-ypy-history-value`, `data-ypy-history-checked` and
  `data-ypy-history-scroll`, and put back (and the attributes removed) after a restore;
- `type="file"` inputs are skipped.

Sensitive fields are skipped: a password input, and any field whose `autocomplete` is `off`, `one-time-code`,
`current-password`, `new-password` or starts with `cc-`. Neither what the user typed nor the value the server rendered
is kept. Hidden inputs are kept on purpose, so a restored form still has its token.

The list is an LRU: at most `size` pages, the oldest dropped first. A snapshot for an entry that already has one
replaces it. If the browser says the storage is full, the oldest snapshots are dropped until the new one fits.

## Configuration

`ypy.config.historyCache`:

| Key | Default | Meaning |
|---|---|---|
| `size` | `10` | pages kept in `sessionStorage`; `0` turns saving off |
| `refreshOnMiss` | `false` | on back or forward with no snapshot, reload the page |
| `disable` | `false` | turn the cache off without unloading the extension |
| `swapStyle` | `"outerSync"` | swap style used to put a snapshot back |
| `encrypt` | `false` | seal the snapshots: `true` or `"memory"` (see below) |

```js
ypy.config.historyCache = { size: 5, refreshOnMiss: true }
```

## Attribute

| Attribute | Meaning |
|---|---|
| `ypy-history="false"` | anywhere on the page: do not save this page when the user leaves it |

Use it on a page that shows something private or that is never right when old (a balance, an inbox):

```html
<body ypy-history="false">...</body>
```

It only stops the saving of that page. Going back to it then usually finds no snapshot and asks the server.

## Events

All fire on `document`.

| Event | Cancelable | `detail` |
|---|---|---|
| `ypy:history:cache:before:save` | yes: skips saving | `{target, head}`; change `head`, or the content of `target` |
| `ypy:history:cache:after:save` | no | `{content, head, scroll, title}` |
| `ypy:history:cache:hit` | yes: ignore the snapshot | `{path, item}`; `item` may be replaced |
| `ypy:history:cache:miss` | no | `{path, refreshOnMiss}`; set `refreshOnMiss` to reload |
| `ypy:history:cache:before:restore` | no | `{head, ready}`; set `ready` to a promise to delay the swap |
| `ypy:history:cache:after:restore` | no | `{head, ready, item}`, after the scroll was set |

`item` is `{content, head, scroll, title}`. Cancelling `hit` makes the core restore the page by request:

```js
document.addEventListener('ypy:history:cache:hit', (event) => {
  if (event.detail.path.startsWith('/inbox')) event.preventDefault()   // always ask the server
})
```

If `ypy-head` is loaded, it merges the saved `<head>` in `before:restore` and holds the swap until stylesheets have
loaded, which is what `ready` is for.

## Encrypted snapshots

A snapshot is the HTML the user saw. `historyCache.encrypt` seals it so it is not readable in the browser storage:
WebCrypto, AES-GCM, a 256-bit key made for the session, a new IV for every snapshot, the snapshot id as additional
data.

```js
ypy.config.historyCache = { encrypt: true }       // key in sessionStorage: snapshots survive a reload of the tab
ypy.config.historyCache = { encrypt: 'memory' }   // key in memory only: after a reload the old snapshots are gone
```

- With `"memory"`, after a reload the old snapshots cannot be opened, and those pages come from the server.
- Sealing is asynchronous, so `after:save` fires when the snapshot is stored. Restoring a sealed snapshot is
  asynchronous too: if a `hit` listener cancels it, or the snapshot cannot be opened, the page is reloaded.
- Without WebCrypto (a page that is not a secure context) nothing is stored, and a warning goes to the console. A
  snapshot is never written in the clear when encryption was asked for.
- A sealed snapshot is treated as a miss, and removed, when encryption is off or its key is gone.
- The key is as safe as the page: a script of the page can read it. This protects the storage at rest, not the page from
  its own scripts.

## Signing out: `ypy.history.clear()`

```js
ypy.history.clear()   // forgets every snapshot of this tab
```

Call it when the user signs out, so the next person on the computer cannot go back to the pages. With encryption it also
replaces the key, and a snapshot that was still being sealed is lost.

```html
<form id="logout" ypy-post="/logout">
  <button class="rounded bg-slate-800 px-3 py-1 text-white">Sign out</button>
</form>
<script>
  document.getElementById('logout').addEventListener('ypy:after:request', () => ypy.history.clear())
</script>
```

## Pitfalls

- A restored page is as old as the snapshot. After a change made on another tab, going back shows the old state. Mark
  such pages with `ypy-history="false"`, or turn the cache off.
- `refreshOnMiss` reloads with a full request; it is not a "refresh after restore".
- The request-restore safeguards of [History](../reference/history.md) (`historyRestore.cache`, the fragment check) do not apply to a
  page restored from a snapshot.
- Only the history element is replaced. State outside it (a header, a sidebar) is not part of the snapshot.
- Do not rely on `sessionStorage` surviving: the browser may refuse it (a private window, a full quota), and the page
  then works as without the extension.

See also: [History](../reference/history.md), [History and caching](../guides/history-and-caching.md), [Cache](cache.md) for the
response cache of requests.
