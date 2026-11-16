---
order: 50
title: "Configuration"
description: "Every setting of the YPY core, where to set it (a meta tag, the ypy object or an attribute), and the settings that each extension adds."
---

# Configuration

YPY works without any setting. Change one when the default does not suit your page. This page lists every key of the
core with its type and default, and points to the settings that extensions add.

## Where to set it

There are three places, for three different jobs.

### The `ypy-config` meta tag

For the whole page, before the script runs. Put it in `<head>`, before the YPY script. Only the first
`<meta name="ypy-config">` of the `<head>` counts: a tag injected into the body is ignored.

```html
<head>
  <meta name="ypy-config" content='{"defaultTimeout": 20000, "history": "reload"}'>
  <script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
</head>
```

The content is JSON, or the shorter form of key and value pairs separated by spaces or commas. A bare dotted key
writes into a nested object, a bare key with no value is `true`, and a value is read as JSON when it can be:

```html
<meta name="ypy-config" content='callbacks:"all", defaultTimeout:20000, classes.request:"is-loading"'>
```

Use it for keys that are read once at start, such as `prefix` and `extensions`.

### The `ypy.config` object

In a script, at any time. A key that is read when a request runs takes effect at once.

```js
ypy.config.defaultSwap = 'innerHTML'
ypy.config.noSwap = [204, 304, 205]
ypy.config.classes.request = 'is-loading'
```

Keys that YPY reads only at load, such as `prefix`, do not change afterwards.

### The `ypy-config` attribute

For one element and what is inside it. It is not the page configuration: it sets options of the request of that
element, such as `timeout`, `credentials` or `cache`. It is inherited. It cannot set `mode`.

```html
<div ypy-config='{"timeout": 5000}'>
  <button ypy-get="/slow">Try</button>
</div>
```

## Keys of the core

A default written as "unset" means the key has no value until you set one, and the description says what YPY does
then.

### Requests

How requests are sent and when they time out.

| Key | Type | Default | Description |
|---|---|---|---|
| `mode` | `"same-origin"` \| `"cors"` \| `"no-cors"` | `"same-origin"` | The fetch `mode` of every request. An element cannot override it, and neither can `ypy-config`. |
| `selfRequestsOnly` | boolean | `true` | An alias of `mode`: `true` is `same-origin`, `false` is `cors`. Read it as a boolean. |
| `defaultTimeout` | integer | `60000` | Request timeout in milliseconds. `0` means no timeout. |
| `noSwap` | array of integers | `[204, 304]` | Status codes for which YPY does not swap content. |

### Swapping

What is inserted, where, and how long the settle phase waits.

| Key | Type | Default | Description |
|---|---|---|---|
| `defaultSwap` | string | `"innerMorph"` | The swap style used when an element has no `ypy-swap`. |
| `defaultSettleDelay` | integer | `1` | Milliseconds between the swap and the settle phase. |
| `allowEmptySwapAfterOOB` | boolean | `false` | Runs the main swap when the response held only out-of-band elements. Content of a `<ypy-partial>` always prevents the main swap. |
| `targetFallback` | `true` \| `false` \| `"always"` | unset (on) | What happens when a target selector matches nothing. `true` falls back to the main target for a whole page or a response with a main, `"always"` also for a fragment, `false` never. |
| `transitions` | boolean | `false` | Uses the View Transitions API for swap animations. |

### Morph

How a morph matches nodes and what it leaves alone. See [Swap styles and morph](../swap-and-morph.md).

| Key | Type | Default | Description |
|---|---|---|---|
| `morphIgnore` | array of strings | `["data-ypy-powered", "data-ypy-state"]` | Attribute name prefixes a morph leaves alone. |
| `morphScanLimit` | integer | `10` | How many siblings a morph scans when it matches nodes. |
| `morphSkip` | string | `"[ypy-morph-skip],[data-ypy-morph-skip]"` | Selector of the elements a morph skips entirely. |
| `morphSkipChildren` | string | `"[ypy-morph-skip-children],[data-ypy-morph-skip-children]"` | Selector of the elements whose children a morph freezes. Their attributes still update. |
| `morphThirdParty` | string | `"[data-reactroot],[data-v-app],[x-data]"` | Selector of the elements another script owns, besides every custom element. A morph keeps their children when the response gives them empty, unless they carry `ypy-preserve="false"`. |

### History

See [History](../history.md).

| Key | Type | Default | Description |
|---|---|---|---|
| `history` | `true` \| `false` \| `"reload"` | `true` | `true` pushes URLs and restores by request, `false` turns history off, `"reload"` pushes URLs and reloads the page on back and forward. |
| `historyRestore` | object | unset | Options of the request that restores a page from history. |
| `historyRestore.cache` | `"default"` \| `"no-store"` \| `"reload"` \| `"no-cache"` \| `"force-cache"` \| `"only-if-cached"` | unset (`no-cache`) | The fetch `cache` mode of that request. |
| `historyRestore.fragment` | boolean | unset (on) | Reloads the page when a restore gets something that is not a whole page, instead of swapping the fragment. |

### State classes and indicators

The classes YPY puts on elements while it works. See [`config.classes`](classes.md) and [Styling](../styling.md).

| Key | Type | Default | Description |
|---|---|---|---|
| `classes` | object | unset | One entry per state, each a space separated list. Set an entry to replace it. |
| `classes.request` | string | `"ypy-request"` | Applied to the indicators of a request while it runs, by default the requesting element. |
| `classes.swapping` | string | `"ypy-swapping"` | Applied to the swap target between the start of the swap and the swap itself. |
| `classes.settling` | string | `"ypy-settling"` | Applied to the swap target while it settles. |
| `classes.added` | string | `"ypy-added"` | Applied to fresh content until it settles. |
| `classes.indicator` | string | `"ypy-indicator"` | Applied to elements that show only during a request. |
| `classes.error` | string | none | Added to the error target after a response of 400 or more was swapped. `data-ypy-state="error"` is enough to style it without a class. |
| `classes.offline` | string | `"ypy-offline"` | Applied to `<html>` while the browser reports no connection. |
| `includeIndicatorCSS` | boolean | `true` | Includes the built-in CSS of the indicator class. |

### Focus and accessibility

See [Accessibility](../../guides/accessibility.md).

| Key | Type | Default | Description |
|---|---|---|---|
| `autofocus` | boolean | `false` | Focuses the `[autofocus]` element of swapped content. An `[autofocus]` element with `ypy-focus` asks for it by itself. |
| `defaultFocusScroll` | boolean | `false` | Scrolls the focused element into view after each swap. |
| `ariaBusy` | boolean | unset (on) | Marks the target `aria-busy="true"` while a request runs. |
| `announce` | boolean or object | unset (on) | Announcements through the managed live region. `false` turns them off; an object sets the messages. |
| `announce.clearAfter` | integer | `5000` | Milliseconds before the live region is emptied. |
| `announce.messages` | object | unset | Texts for the live region. The core carries none: a message that is not set says nothing. |
| `announce.messages.loading` | string | unset | Said when a request starts. |
| `announce.messages.done` | string | unset | Said when a request ends and its response was swapped. |
| `announce.messages.offline` | string | unset | Said when a request failed for lack of a connection. |
| `announce.messages.error` | string | unset | Said when a request failed for any other reason. |

### Security

See [Security](../security.md) and the [security checklist](../../guides/security-checklist.md).

| Key | Type | Default | Description |
|---|---|---|---|
| `callbacks` | `"none"` \| `"nonce"` \| `"all"` | unset (`"none"`) | What author code (`ypy-on`, `js:` values, `[filter]` triggers) may do. `none` never compiles code, `nonce` runs it only through `ypy-csp`, `all` always runs it. |
| `scripts` | `"auto"` \| `"nonce"` \| `"block"` | `"auto"` | What happens to the `<script>` elements of a response. `auto` lets them run, `nonce` runs only the ones with the page nonce, `block` runs none. A script that does not run is taken out of the response. Any other value counts as `block`. |
| `inlineScriptNonce` | string | unset | The `nonce` of the inline `<script>` elements YPY creates, for a strict CSP. |
| `extensions` | string | `""` | Comma separated list of the extensions allowed to register. Empty allows all. |

### Other

| Key | Type | Default | Description |
|---|---|---|---|
| `prefix` | string | `""` | Adds one more attribute prefix. `ypy-` and `data-ypy-` always work. Set it in the meta tag; changing it after load has no effect. |
| `metaCharacter` | string | unset | Replaces `:` in attribute modifiers, for template engines that treat `:` in a special way. |
| `logAll` | boolean | `false` | Logs every YPY event to the console. |

## Extension settings

Each extension reads its own keys from `ypy.config`. The page of the extension gives the type, the default and
the meaning of each key.

| Extension | Keys |
|---|---|
| [`ypy-browser-indicator`](../../extensions/browser-indicator.md) | `boostBrowserIndicator` |
| [`ypy-cache`](../../extensions/cache.md) | `cache`, `cache.all`, `cache.ttl`, `cache.fresh`, `cache.private`, `cache.offline`, `cache.size`, `cache.bytes` |
| [`ypy-commands`](../../extensions/commands.md) | `commands`, `commands.max` |
| [`ypy-csp`](../../extensions/csp.md) | `nonce`, `safeEval` |
| [`ypy-dedupe`](../../extensions/dedupe.md) | `dedupe`, `dedupe.ignore` |
| [`ypy-dev`](../../extensions/dev.md) | `dev`, `dev.latency`, `dev.offline`, `dev.reload`, `dev.lint`, `dev.log`, `dev.overlay`, `dev.docs` |
| [`ypy-disable`](../../extensions/disable.md) | `inert` |
| [`ypy-history-cache`](../../extensions/history-cache.md) | `historyCache`, `historyCache.size`, `historyCache.refreshOnMiss`, `historyCache.disable`, `historyCache.swapStyle`, `historyCache.encrypt` |
| [`ypy-inspector`](../../extensions/inspector.md) | `inspector`, `inspector.open`, `inspector.key`, `inspector.max` |
| [`ypy-layers`](../../extensions/layers.md) | `layers`, `layers.style`, `layers.scrollLock`, `layers.outside`, `layers.escape`, `layers.history`, `layers.classes` |
| [`ypy-lazy`](../../extensions/lazy.md) | `lazy`, `lazy.margin`, `lazy.window` |
| [`ypy-live`](../../extensions/live.md) | `live`, `live.inputDebounce`, `live.bindPrefix`, `live.useDollar` |
| [`ypy-lock`](../../extensions/lock.md) | `locked` |
| [`ypy-model`](../../extensions/model.md) | `announce.stream`, `announce.messages.answerStart`, `announce.messages.answerReady`, `announce.messages.answerStopped`, `announce.messages.answerFailed` |
| [`ypy-multipart`](../../extensions/multipart.md) | `multipart`, `multipart.reconnect`, `multipart.reconnectDelay`, `multipart.reconnectMaxDelay`, `multipart.reconnectMaxAttempts`, `multipart.reconnectJitter`, `multipart.connectTimeout`, `multipart.pauseOnBackground` |
| [`ypy-native`](../../extensions/native.md) | `native`, `native.path`, `native.web` |
| [`ypy-nav`](../../extensions/nav.md) | `nav`, `nav.boost`, `nav.prefetch`, `nav.prefetchDelay`, `nav.prefetchTtl`, `nav.prefetchMax`, `nav.prefetchParallel`, `nav.current`, `nav.settle`, `nav.confirmLeave`, `nav.root` |
| [`ypy-poll`](../../extensions/poll.md) | `poll`, `poll.hidden`, `poll.min` |
| [`ypy-preload`](../../extensions/preload.md) | `preload`, `preload.autoBoost`, `preload.boostTimeout`, `preload.boostEvent` |
| [`ypy-preserve`](../../extensions/preserve.md) | `compositionTimeout` |
| [`ypy-progress`](../../extensions/progress.md) | `progress`, `progress.delay`, `progress.all`, `progress.label` |
| [`ypy-retry`](../../extensions/retry.md) | `retry`, `retry.attempts`, `retry.on`, `retry.delay`, `retry.max`, `retry.backoff`, `retry.jitter`, `retry.methods`, `retry.unsafe` |
| [`ypy-signals`](../../extensions/signals.md) | `signals`, `signals.send`, `classes.dirty` |
| [`ypy-sort`](../../extensions/sort.md) | `sort`, `sort.messages`, `sort.messages.grab`, `sort.messages.move`, `sort.messages.drop`, `sort.messages.cancel`, `sort.messages.revert` |
| [`ypy-sse`](../../extensions/sse.md) | `sse`, `sse.reconnect`, `sse.reconnectDelay`, `sse.reconnectMaxDelay`, `sse.reconnectMaxAttempts`, `sse.reconnectJitter`, `sse.connectTimeout`, `sse.pauseOnBackground`, `sse.heartbeat`, `sse.releaseOn` |
| [`ypy-stream`](../../extensions/stream.md) | `stream`, `stream.pauseOnBackground`, `stream.merge`, `stream.frame` |
| [`ypy-test`](../../extensions/test.md) | `test`, `test.timeout`, `test.quiet` |
| [`ypy-validate`](../../extensions/validate.md) | `validate`, `validate.batch` |
| [`ypy-ws`](../../extensions/ws.md) | `ws`, `ws.reconnect`, `ws.reconnectCodes`, `ws.reconnectDelay`, `ws.reconnectMaxDelay`, `ws.reconnectMaxAttempts`, `ws.reconnectJitter`, `ws.pauseOnBackground`, `ws.heartbeat`, `ws.maxOutgoingMessagesQueueSize`, `ws.protocols` |

A key with a dot is a nested object. In JavaScript, `cache.ttl` is `ypy.config.cache.ttl`; in the meta tag it is
`cache.ttl:60000` or `{"cache": {"ttl": 60000}}`.

## See also

- [`config.classes`](classes.md): rename the state classes.
- [Attribute inheritance](../inheritance.md)
- [JavaScript API](../javascript-api.md)
- [Events list](../events/list.md)
