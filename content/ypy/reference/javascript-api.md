---
order: 40
title: "JavaScript API"
description: "The ypy object: how to send requests, listen to events, process new content and take YPY apart from your own script, plus the API that each extension adds."
---

# JavaScript API

Most pages never call the `ypy` object: attributes do the work. Use it when a script has to start a request, listen
to events, tell YPY about HTML you inserted yourself, or test the page.

The script puts `ypy` on `window` (and the ES module build exports the same object). TypeScript users get the types
from `ypy.d.ts`, including a typed `detail` for every event name.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<!-- or: https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js -->
<script>
  console.log(ypy.version) // "0.1.1"
</script>
```

## The core

| Member | Description |
|---|---|
| `ypy.version` | The version string of the script that is loaded. |
| `ypy.config` | The global [configuration](config/index.md). Read it and set keys at run time; some keys, such as `prefix`, only count when set before load. |
| `ypy.ajax(verb, path, options)` | Sends a request from a script and returns a promise that resolves after the response was swapped. |
| `ypy.find(selector)` / `ypy.find(elt, selector)` | The first element that matches, in the document or inside `elt`. Returns `null` when none does. |
| `ypy.findAll(selector)` / `ypy.findAll(elt, selector)` | An array of all the elements that match. |
| `ypy.on(event, handler)` / `ypy.on(target, event, handler)` | Adds a listener on `document` or on an element. Returns the handler. |
| `ypy.onLoad(callback)` / `ypy.onLoad(selector, callback)` | Calls `callback(elt)` for each element YPY processes, or for each one that matches `selector`. |
| `ypy.trigger(elt, event, detail, bubbles)` | Dispatches a `CustomEvent`. Returns `false` when a listener cancelled it. |
| `ypy.process(root, force)` | Makes YPY read the `ypy-*` attributes of `root` and everything under it. |
| `ypy.initialize()` | Sets up history and processes `document.body`. Runs by itself on `DOMContentLoaded`. |
| `ypy.dispose()` | Takes YPY apart. It cannot be undone. |
| `ypy.registerExtension(name, ext)` | Registers an extension. |
| `ypy.supports(feature)` | Whether an extension or feature is there. Available with the [`ypy-boot`](../extensions/boot.md) extension. |
| `ypy.timeout(time)` | A promise that resolves after the time. |
| `ypy.parseInterval(value)` | Turns `"2s"` or `"500ms"` into milliseconds. |
| `ypy.swap(ctx)` | Runs a swap. For extensions; prefer `ypy.ajax()`. |

### `ypy.ajax()`

```js
ypy.ajax('GET', '/items', '#list')                       // a selector for the target
ypy.ajax('POST', '/save', { target: '#result', swap: 'outerHTML', values: { id: 7 } })
```

The third argument is an element, a selector, or an options object. Without a target, the response goes where the
source element's `ypy-target` points, or into the source element itself.

| Option | Meaning |
|---|---|
| `source` | The element used as the source of the request: its headers, inherited attributes and `YPY-*` request values. An element or a selector. |
| `event` | The event that started the request. |
| `target` | An element or a selector to swap the response into. |
| `swap` | A swap style, such as `innerHTML` or `outerHTML`. |
| `values` | Extra values for the body (or the query string, for a `GET`). |
| `headers` | Extra request headers. |
| `select` | A selector that picks part of the response. |
| `selectOOB` | A selector for out-of-band elements. |
| `push` | A URL to push into the history after the swap. `true` uses the request URL. |
| `replace` | A URL to replace the current history entry with. `true` uses the request URL. |
| `request` | Fields merged into the request, such as `cache`, `signal` or any other `fetch` option. `mode`, `credentials` and the core headers keep the values YPY built. A header set to `null` is removed. |

The promise resolves when the swap is done. A request that is refused, cancelled or fails for another reason fires
the events of [the list](events/list.md) as any other request does.

### `ypy.find()` and `ypy.findAll()`

```js
const list = ypy.find('#list')
const rows = ypy.findAll(list, 'tr')
```

They accept the selectors that `ypy-target` accepts, including the ones an extension adds (see
[`ypy-scope`](../extensions/scope.md)).

### `ypy.on()`

```js
// on document
ypy.on('ypy:response:error', (event) => {
  console.warn('status', event.detail.ctx.response.status)
})

// on an element, by selector or by reference
ypy.on('#checkout', 'ypy:after:swap', () => focusFirstField())
```

For the names and the fields of each event, see the [event list](events/list.md). `ypy.on` is a short form of
`addEventListener`; both work.

### `ypy.onLoad()`

Runs for the elements YPY processes: the page on load, and every swap, out-of-band element and partial.

```js
ypy.onLoad('[data-tooltip]', (el) => new Tooltip(el))
```

With only a function, it runs for every processed element, which is the same as listening to `ypy:after:process`.

### `ypy.trigger()`

```js
const ok = ypy.trigger('#form', 'item:saved', { id: 7 })
// ok is false when a listener called preventDefault()
```

The event bubbles by default. A server can fire the same kind of event with the
[`YPY-Trigger`](headers.md) response header.

### `ypy.process()`

Insert HTML yourself and YPY does not know about it. Tell it:

```js
box.innerHTML = '<button ypy-get="/hello">Say hello</button>'
ypy.process(box)
```

`process(root, true)` tears down the elements that were already processed and sets them up again. Use it after you
change a `ypy-*` attribute on a live element. It accepts an element or a `ShadowRoot`.

### `ypy.initialize()` and `ypy.dispose()`

`initialize()` runs by itself. Call it by hand when you load the script after the page, or after a streamed response
delivered the whole page. It is safe to call more than once; the history listeners are added once.

`dispose()` is for tests and for applications that unmount YPY. It runs the `ypy_dispose` hook of the extensions,
aborts requests in flight, drops the waiting ones, removes the listeners the core installed and makes the object
inert: `process()`, `initialize()` and requests do nothing afterwards. Calling it again does nothing.

### `ypy.registerExtension()`

```js
ypy.registerExtension('app-flash', {
  version: '1.0.0',
  ypy_after_swap(elt, detail) { /* your hook */ }
})
```

The name must be in `ypy.config.extensions` unless that list is empty. A name registered twice is ignored. `version`
and `requires` are read by `ypy-boot`. The rest of the keys are hooks named after the events. See
[Using and writing extensions](../extensions/using-and-writing.md).

### `ypy.timeout()` and `ypy.parseInterval()`

```js
await ypy.timeout('2s')
ypy.parseInterval('500ms') // 500
ypy.parseInterval(1000)    // 1000
ypy.parseInterval('soon')  // undefined
```

`timeout()` returns `undefined` when the time is `0` or empty. `parseInterval()` accepts `ms`, `s` and `m`.

## The request context

Most events carry `detail.ctx`, the object that describes one request: its source element, target, request options,
response, swap tasks and `timing`. Its fields are in `ypy.d.ts` (`YpyRequestCtx`); the timing marks are in
[Request timing](timing.md).

## API added by extensions

An extension adds a member to `ypy` when it is loaded. Check that the member exists before you use it.

| Member | Extension | What it does |
|---|---|---|
| `ypy.signals` | [`ypy-signals`](../extensions/signals.md) | `register(name, fn)` makes a function callable from signal expressions; `get(path, element)` reads a value; `patch(patch, element, onlyIfMissing)` applies a merge patch like a server response; `stats()` counts live nodes. |
| `ypy.live` | [`ypy-live`](../extensions/live.md) | `q(selector)` and `$(selector)` return a query over elements; `debounce(ms)`, `refresh()`, `take()`, `toggle()`, `forEvent()` and `nextFrame()` for live expressions. |
| `ypy.stream` | [`ypy-stream`](../extensions/stream.md) | `close(element)` closes the stream of an element; `stats()` returns how many are open. |
| `ypy.layers` | [`ypy-layers`](../extensions/layers.md) | `open(spec, options)` opens a layer and returns a promise for its result; `stack`, `top()`, `accept(value)`, `dismiss(reason)` and `closeAll(reason)`. |
| `ypy.cache` | [`ypy-cache`](../extensions/cache.md) | `expire(patterns)`, `evict(patterns)`, `clear()` and `stats()`. |
| `ypy.history` | [`ypy-history-cache`](../extensions/history-cache.md) | `clear()` forgets every kept page and replaces the encryption key. Call it when the user signs out. |
| `ypy.optimistic` | [`ypy-optimistic`](../extensions/optimistic.md) | `register(name, fn)` makes `fn` usable as `ypy-optimistic="name"`. |
| `ypy.switch` | [`ypy-switch`](../extensions/switch.md) | `update(root)` evaluates the `ypy-switch` parts again. |
| `ypy.commands` | [`ypy-commands`](../extensions/commands.md) | `run(list, source)` runs a list of commands, as the `YPY-Commands` header does. |
| `ypy.native` | [`ypy-native`](../extensions/native.md) | `connect()`, `disconnect()`, `send()`, `visit()`, `load()`, `properties()`, and the read-only `ready`, `platform` and `caps`. |
| `ypy.idle(options)` | [`ypy-test`](../extensions/test.md) | A promise that resolves when no request or swap is pending, after a quiet period. For browser tests. |
| `ypy.onError(handler)` | [`ypy-telemetry`](../extensions/telemetry.md) | Calls `handler` for every failure of the page. Returns a function that cancels the subscription. |
| `ypy.dev` | [`ypy-dev`](../extensions/dev.md) | `latency()`, `offline()`, `reload()`, `known()`, `clear()` and the list of messages. |
| `ypy.inspector` | [`ypy-inspector`](../extensions/inspector.md) | `open()`, `close()`, `toggle()`, `pause()`, `clear()`, `filter()`, `data()`, `export()`, `watch()`, `enable()` and `disable()`. |

### Examples

```js
// open a layer from a script and wait for what the user decides
const result = await ypy.layers.open('/items/7/edit', { mode: 'drawer', side: 'end' })
if (result.accepted) refreshList()

// forget the cached copies when the user signs out
ypy.cache.clear()
ypy.history.clear()

// in a browser test
await page.evaluate(() => ypy.idle())

// send every failure to your own log
const stop = ypy.onError((error) => report(error.code, error.url))
```

`ypy.layers.open()` takes a string that starts with `<` as HTML and any other string as a URL.

## See also

- [Configuration](config/index.md)
- [Event list](events/list.md) and the [Events guide](events/index.md)
- [Request timing](timing.md)
- [Extensions](../extensions/index.md)
