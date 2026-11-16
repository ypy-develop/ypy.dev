---
title: "Events"
description: "How YPY reports every step of a request and a swap with ordinary DOM events, in the order they fire, with recipes for the common jobs."
order: 30
id: "GUIDE-events"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en"]
---

# Events

YPY tells the page about every step of a request and of a swap with ordinary DOM events. You can cancel some, change
the request in others, and use the rest to react: a toast, a re-initialized widget, a log line, an analytics ping.

Contents: [Naming](#naming) · [Listening](#listening) · [The life of a request](#the-life-of-a-request) ·
[Reference of events](#reference-of-events) · [`detail.ctx`](#detailctx) · [Recipes](#recipes) ·
[Server-sent events](#events-from-the-server)

## Naming

Events are named `ypy:<noun>:<verb>`: `ypy:before:request`, `ypy:after:swap`, `ypy:target:missing`. They bubble, so a
listener on `document` or `body` sees every event of the page. The event fires on the element that sent the request,
or on the element the event is about (the one processed, the one cleaned up).

Code that listened to events of a project written before YPY is converted by `ypy migrate` and `ypy check` lists the
new names ([Migration](../../migration/index.md)).

## Listening

Three ways, from the most common:

```js
// 1. addEventListener, anywhere (works with the strictest CSP, it is your own script file)
document.addEventListener('ypy:after:swap', (event) => {
  console.log('swapped', event.detail.ctx.target)
})

// 2. ypy.on, the same thing with the YPY object
ypy.on('ypy:response:error', (event) => showToast('The server said ' + event.detail.ctx.response.status))
```

```html
<!-- 3. an attribute on the element; it compiles code, so it needs callbacks on -->
<meta name="ypy-config" content='callbacks:"all"'>
<form ypy-post="/items" ypy-on::after:request="if (event.detail.ctx.response.status < 400) this.reset()">...</form>
```

`ypy-on::after:request` is `ypy-on:` followed by the event name written without the `ypy` prefix and with `::`: the
double colon is "the event of YPY itself". `ypy-on:click` listens to the browser's `click`.

Author code is off by default (`callbacks: "none"`): the third way does nothing until the page opts
in. Prefer the first two for anything that is not trivial; see [Security](../../guides/security-checklist.md#author-code).

Typed listeners in TypeScript come from `ypy.d.ts` (`CustomEvent<YpyEventMap['ypy:after:swap']>`).

## The life of a request

For a request that reaches the server and succeeds, the events fire in this order. Cancel an event with
`event.preventDefault()` where the table says *cancelable*.

| # | Event | Cancelable | Use it to |
|---|---|---|---|
| 1 | `ypy:before:process`, `ypy:after:process` | before | skip an element, or initialize something when new content is processed |
| 2 | `ypy:confirm` | handled by you | replace the browser's confirm box with your own (call `issueRequest()` or `dropRequest()`) |
| 3 | `ypy:config:request` | yes | change headers, timeout, credentials or the body; cancel the request |
| 4 | `ypy:before:request` | yes | last chance before `fetch` |
| 5 | `ypy:before:response` | yes | look at the status and headers before the body is read |
| 6 | `ypy:after:request` | yes | read `ctx.text`, the response as text |
| 7 | `ypy:before:swap` | yes | change what is swapped, or cancel the swap |
| 8 | `ypy:before:settle`, `ypy:after:settle` | no | once for each swapped target (the main swap, each out-of-band element, each partial): the content is in the DOM and its CSS transitions finished; the DOM is stable |
| 9 | `ypy:after:swap` | no | once for the whole swap, after every target settled and its new content was processed; re-initialize widgets |
| 10 | `ypy:after:history:push` or `ypy:after:history:replace` | no | only when the swap writes the address; fires after `ypy:after:swap`, so the address is the new one here |
| 11 | `ypy:finally:swap`, `ypy:finally:request` | no | run on success and failure; clean up |

The address is written after the swap, never before it: if the swap fails, the address does not change. A listener
of `ypy:after:swap` still sees the old address. To adjust the address after the swap (add a query parameter, read
`location`), listen to `ypy:after:history:push` or `ypy:after:history:replace`.

The order of 8 and 9 is the one above: `ypy:after:settle` fires before `ypy:after:swap`. Settling belongs to a
target (a response with a main swap and two out-of-band elements settles three times), while `ypy:after:swap` is the
end of the whole swap and fires once. The new content is processed (`ypy:before:process`, `ypy:after:process`, step 1
again, for the new nodes) between the two. So a listener that clears something in `ypy:after:settle` has done it by the
time `ypy:after:swap` fires, and a listener of `ypy:after:swap` finds the content already processed. A test fixes this
order.

Failures take another path:

| What happened | Event |
|---|---|
| the server answered 400 or higher | `ypy:response:error` (then the swap rules of [`ypy-error-*`](../attributes/ypy-error.md) apply) |
| the network failed, a timeout, or code threw | `ypy:error` with `detail.error` (and `ctx.errorKind`: `offline`, `network`, `timeout`) |
| the target does not exist | `ypy:target:missing` ([target events](target.md)) |
| the swap threw | `ypy:swap:error` |
| a second request met one in flight | `ypy:request:queued`, `ypy:request:dropped`, `ypy:request:aborted` ([queue events](request-queue.md)) |
| the browser went offline or came back | `ypy:offline`, `ypy:online` ([offline](ypy-offline.md)) |

`ypy:error` does not fire for a 404 or a 500: those are answers, not exceptions. Use `ypy:response:error`.

## Reference of events

| Event | `detail` | Notes |
|---|---|---|
| `ypy:before:init`, `ypy:after:init` | none | around the setup of one element |
| `ypy:before:process`, `ypy:after:process` | none | around the processing of a node and its subtree |
| `ypy:before:on:init` | none | cancel to stop `ypy-on` handlers being installed on the element |
| `ypy:before:cleanup`, `ypy:after:cleanup` | none | the element is leaving: remove your listeners and timers. Only elements YPY processed get it, that is, elements with a `ypy-*` or `data-ypy-*` attribute; see [Elements without a `ypy-*` attribute](#elements-without-a-ypy-attribute) |
| `ypy:confirm` | `ctx`, `issueRequest`, `dropRequest` | only for elements with `ypy-confirm` |
| `ypy:config:request` | `ctx` | the place to add a header for every request |
| `ypy:before:request` | `ctx` | |
| `ypy:before:response` | `ctx` | `ctx.response.raw` is the unread `Response` |
| `ypy:after:request` | `ctx` | |
| `ypy:response:error` | `ctx` | status 400 or higher |
| `ypy:error` | `ctx?`, `error` | exceptions; `ctx` can be missing |
| `ypy:finally:request` | `ctx` | success, failure or cancel |
| `ypy:before:swap` | `ctx`, `tasks` | `tasks` are the swaps about to happen |
| `ypy:after:swap` | `ctx` | once per swap, after `ypy:after:settle` of every target |
| `ypy:before:settle`, `ypy:after:settle` | `task`, `newContent`, `settleTasks` | once per swapped target, before `ypy:after:swap`; `newContent` are the new nodes |
| `ypy:finally:swap` | `ctx` | |
| `ypy:swap:error` | `ctx`, `task`, `error` | |
| `ypy:before:morph:node`, `ypy:before:morph:attr` | the node or attribute | cancel to keep what is on the page |
| `ypy:before:history:update`, `ypy:after:history:update` | `history`, `sourceElement`, `response` | see [History](../history.md) |
| `ypy:after:history:push`, `ypy:after:history:replace` | `path` | |
| `ypy:before:history:restore` | `path`, `cacheMiss` | cancel to restore by yourself |
| `ypy:before:viewTransition`, `ypy:after:viewTransition` | `task` | only with `config.transitions` |
| `ypy:announce` | `ctx`, `message`, `politeness` | change or cancel what a screen reader hears |
| `ypy:target:missing`, `ypy:target:fallback` | `selector`, ... | |
| `ypy:request:queued`, `ypy:request:dropped`, `ypy:request:aborted` | `ctx`, `reason` | |
| `ypy:request:retry` | `attempt`, `delay`, `ctx` | with the `ypy-retry` extension |
| `ypy:offline`, `ypy:online` | none | on `document` |
| `ypy:abort` | none | you fire it on an element to abort its request |
| `ypy:security:blocked`, `ypy:security:strip`, `ypy:security:violation` | `reason`, ... | see [Security](../security.md#events) |

Extensions add their own (`ypy:sse:*`, `ypy:ws:*`, the signals events); each page of
[Extensions](../../extensions/using-and-writing.md) names them.

## Elements without a `ypy-*` attribute

`ypy:before:cleanup` and `ypy:after:cleanup` reach only elements YPY processed: elements with a `ypy-*` or
`data-ypy-*` attribute (and the ones that carry `ypy-on` handlers). An element with none of them, for example one
that only has `data-widget="chart"`, is removed by a swap without any event. To release what such an element holds,
observe the document and read the removed nodes:

```js
new MutationObserver((records) => {
  for (const record of records) {
    for (const node of record.removedNodes) {
      if (node.nodeType === 1) node.querySelectorAll?.('[data-widget]').forEach(destroyWidget)
    }
  }
}).observe(document.body, { childList: true, subtree: true })
```

A node that moved is also reported as removed and added, so check `node.isConnected` before you destroy it.

## `detail.ctx`

Most events carry the request context. The fields you will use:

| Field | What it is |
|---|---|
| `ctx.sourceElement` | the element that sent the request |
| `ctx.target` | the element the response goes into |
| `ctx.request` | the outgoing request: `action` (URL), `method`, `headers`, `body` (`FormData`), `timeout`, `credentials` |
| `ctx.response` | after the fetch: `status`, `headers`, `raw` |
| `ctx.text` | the response body as text (after `ypy:after:request`) |
| `ctx.errorKind` | for failures: `offline`, `network` or `timeout` |
| `ctx.redirected` | `true` when the fetch followed a redirect (a `303`, for example) |

The authoritative list, with types, is `YpyRequestCtx` in `dist/ypy.d.ts`.

## Recipes

**Send a CSRF token with every request.** One listener, and no attribute on any element:

```js
document.addEventListener('ypy:config:request', (event) => {
  const token = document.querySelector('meta[name="csrf-token"]')?.content
  if (token) event.detail.ctx.request.headers['X-CSRF-Token'] = token
})
```

**A toast on any failed request:**

```js
document.addEventListener('ypy:response:error', (event) => {
  const status = event.detail.ctx.response.status
  document.querySelector('#toast').innerHTML =
    `<p class="rounded bg-red-600 p-3 text-white" role="alert">The server answered ${status}.</p>`
})
```

Start a widget in new content (Bootstrap, a date picker, a chart). `ypy.onLoad(selector, fn)` runs `fn` for every
matching element when it is processed, on the first load and after every swap:

```js
ypy.onLoad('[data-datepicker]', (el) => new Datepicker(el))
```

The same job with an event, if you need the whole swap: `ypy:after:swap` or `ypy:after:settle`, then look inside
`event.detail.ctx.target`. See [Third-party components](../../guides/third-party-components.md).

**Replace the confirm box:**

```js
document.addEventListener('ypy:confirm', (event) => {
  event.preventDefault() // we answer, not the browser
  const { issueRequest, dropRequest } = event.detail
  openDialog('Delete this?').then((yes) => (yes ? issueRequest() : dropRequest()))
})
```

Always call one of the two; the request waits for the answer.

Cancel a request from `ypy:config:request` or `ypy:before:request` with `event.preventDefault()`.

**Log everything while you develop:** `ypy.config.logAll = true`.

## Events from the server

The server can fire client events with a response header, no JavaScript on the page needed to ask:

```
YPY-Trigger: item:saved
YPY-Trigger: {"item:saved": {"count": 3}, "toast": "Saved"}
```

`YPY-Trigger` fires as soon as the response arrives. There is no "after the swap" header: to act once
the new content is in place, listen to `ypy:after:swap` or `ypy:after:settle` for the event you fired.

Listen to the event like any other: `document.addEventListener('item:saved', (e) => e.detail)`. In Go, `ypy.Trigger(w,
"item:saved", map[string]any{"count": 3})` writes the header, and
[`ypytest`](../../server/testing.md) checks it. A stream (`ypy.NewStream`) can also send named events (`ypy-event`); its wire format is the Go SDK's
`spec/PROTOCOL.md`, and [signals](../../extensions/signals.md) is the client state it patches.

## See also

[Core concepts](../../guides/core-concepts.md) · [Extensions](../../extensions/using-and-writing.md) · [Security checklist](../../guides/security-checklist.md) ·
[Migration](../../migration/index.md) for the event names of projects written before YPY and their replacements
