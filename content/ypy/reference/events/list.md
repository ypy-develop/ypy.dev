---
order: 10
title: "Event list"
description: "Every event of the YPY core, by phase, with whether it can be cancelled, where it fires and what it carries, plus the events of each extension."
---

# Event list

This is the full list of the events of the core, in the order of a page's life. How to listen, how to read
`detail.ctx` and what to do in each event is in the [Events guide](index.md). The names, the targets and the
`detail` fields come from the same file the editor data is built from, so this page and your editor agree.

All events are named `ypy:<noun>:<verb>`, bubble, and fire on the element that sent the request or the element the
event is about. The events marked `document` fire on `document`. A table cell of `none` means the event carries
no `detail` fields, or has no DOM target.

## Typical order of a request

For a request that succeeds and swaps one target:

1. `ypy:before:process` and `ypy:after:process` (only when the element is new to the page).
2. `ypy:confirm`, when the element has `ypy-confirm`.
3. `ypy:config:request`, then `ypy:before:request`.
4. `ypy:before:response`, when the headers arrive.
5. `ypy:after:request`, when the body was read.
6. `ypy:before:swap`.
7. For each swapped target: `ypy:before:settle`, then `ypy:after:settle`, then `ypy:after:process` for its new content.
8. `ypy:after:swap`, once.
9. `ypy:after:history:push` or `ypy:after:history:replace`, only when the swap writes the address.
10. `ypy:finally:swap` and `ypy:finally:request`.

A status of 400 or more adds `ypy:response:error` after `ypy:after:request`. A network failure or an exception adds
`ypy:error` and skips the swap. A request that waits in the queue fires `ypy:request:queued` first; one that is
refused fires `ypy:request:dropped` and none of the rest. The same steps with the reasons for each are in
[The life of a request](index.md#the-life-of-a-request).

## Processing and initialization

YPY processes the page on load and every piece of HTML it swaps in. These events let you skip an element or start your own code on it.

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:before:process` | yes | element | Before YPY processes a node or a subtree. Cancel it to leave the subtree unprocessed. | none |
| `ypy:after:process` | no | element | After a node or a subtree was processed. `ypy.onLoad()` is built on it. | none |
| `ypy:before:init` | yes | element | Before YPY attaches its data to an element. Cancel it to leave the element alone. | none |
| `ypy:after:init` | no | element | After the element has all its YPY behaviors and listeners. | none |
| `ypy:before:on:init` | yes | element | Before the `ypy-on` attributes of an element are bound. Cancel it to leave them unbound. | none |
| `ypy:before:cleanup` | no | element | Before YPY removes the listeners and data of an element, on removal in a swap or before reprocessing. Only elements YPY processed get it. | none |
| `ypy:after:cleanup` | no | element | After the listeners and data of an element were removed. Same reach as `ypy:before:cleanup`. | none |

## Request

Events of the request, from the confirmation to the end. The `ctx` in `detail` is the request context; its fields are in [Request timing](../timing.md) and in `ypy.d.ts`.

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:confirm` | yes | element | An element with `ypy-confirm` was triggered, before the request. Call `preventDefault()` to replace the dialog, then call `issueRequest()` or `dropRequest()`. | `ctx`, `issueRequest`, `dropRequest` |
| `ypy:config:request` | yes | element | After the values were collected and validated, before encoding and sending. Change `ctx.request` for headers, timeout or credentials. | `ctx` |
| `ypy:before:request` | yes | element | Right before `fetch()`. | `ctx` |
| `ypy:before:response` | yes | element | When the response headers arrive, before the body is read. `ctx.response.raw` is the unread `Response`. Cancel it to skip the body and the swap. | `ctx` |
| `ypy:after:request` | yes | element | After the body was read, before the swap. Cancel it to skip the swap. | `ctx` |
| `ypy:response:error` | no | element | The server answered with status 400 or more. Network errors do not fire it; use `ypy:error`. | `ctx` |
| `ypy:error` | no | element | An exception ended the request or the swap. HTTP 4xx and 5xx do not fire it. `ctx` is absent when the error comes before the context exists. | `ctx`, `error` |
| `ypy:finally:request` | no | element | The request completed, failed or was cancelled. It does not fire when processing stops before the request starts. | `ctx` |

## Swap and settle

One response can swap several targets (the main one, out-of-band elements, partials). `before:settle` and `after:settle` fire once for each target, and `after:swap` once for the whole swap. See [The life of a request](index.md#the-life-of-a-request).

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:before:swap` | yes | element | After the response was parsed, before it is inserted. `tasks` lists the swap tasks. Cancel it to prevent the swap. | `ctx`, `tasks` |
| `ypy:before:settle` | no | element | After new content was inserted, before CSS transitions are applied. Change `newContent` here. | `task`, `newContent`, `settleTasks` |
| `ypy:after:settle` | no | element | After the settle phase of one task, including every CSS transition. The DOM is stable. It fires before `ypy:after:swap`. | `task`, `newContent`, `settleTasks` |
| `ypy:after:swap` | no | element | Once per swap, after every task was inserted, settled and processed. Focus, title and scroll are applied right after. | `ctx` |
| `ypy:swap:error` | no | element | A swap task failed while the others went on. Once per failed task, before `ypy:after:swap`. | `ctx`, `task`, `error` |
| `ypy:finally:swap` | no | element | At the end of the swap, whether it worked or failed. | `ctx` |
| `ypy:before:viewTransition` | yes | element | Before a View Transition starts (`config.transitions` on, browser support). Cancel it to skip the transition for this swap. | `task` |
| `ypy:after:viewTransition` | no | element | After the View Transition finished. | `task` |

## History

The address is written after the swap. These events fire on `document`.

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:before:history:update` | yes | document | Before `history.pushState()` or `history.replaceState()`. Cancel it to leave the history alone. | `history`, `sourceElement`, `response` |
| `ypy:after:history:update` | no | document | After one of them completed. | `history`, `sourceElement`, `response` |
| `ypy:after:history:push` | no | document | After a new history entry was created. | `path` |
| `ypy:after:history:replace` | no | document | After the current entry was replaced. | `path` |
| `ypy:before:history:restore` | yes | document | The user went back or forward, before the page is restored. Cancel it to restore by hand. An extension such as `ypy-history-cache` can answer a cache hit. | `path`, `cacheMiss` |
| `ypy:history:restore:fragment` | no | document | A restore got something that is not a whole page. The page reloads right after. Turn this off with `config.historyRestore.fragment`. | `ctx` |

## Queue and abort

What happens to requests that wait, are refused or are stopped. See [`ypy-sync`](../attributes/ypy-sync.md) and [Request queue events](request-queue.md).

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:abort` | no | element | Control event: dispatch it on an element to abort its request in flight. YPY listens to it and never fires it. | none |
| `ypy:request:queued` | no | element | A request waits behind the one in flight (`queue first`, `queue last`, `queue all`). | `ctx`, `reason` |
| `ypy:request:dropped` | no | element | A request that never ran was discarded. | `ctx`, `reason` |
| `ypy:request:aborted` | no | element | The request in flight was aborted before its response. | `ctx`, `reason` |

## Offline and announcements

The connection state and the live region. See [Offline state](ypy-offline.md) and [`ypy-announce`](../attributes/ypy-announce.md).

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:offline` | no | document | The browser went offline. `<html>` carries the offline class. | none |
| `ypy:online` | no | document | The connection came back, or a response proved it. | none |
| `ypy:announce` | yes | element | On the source element, just before a message goes into the live region. Change `message` and `politeness`, or cancel to say nothing. | `ctx`, `message`, `politeness` |

## Targets

What YPY does when a selector matches nothing. See [Target events](target.md).

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:target:fallback` | no | element | The main target matched nothing and the chain `ypy-main`, `<main>` took over. | `selector`, `used`, `ctx` |
| `ypy:target:missing` | yes | element | No target was found for a main swap, an out-of-band element or a partial. Cancel it to silence the console warning. | `selector`, `ctx`, `task` |

## Security

See [Security](../security.md).

| Event | Cancelable | Target | When it fires | `detail` |
|---|---|---|---|---|
| `ypy:security:blocked` | no | element | Author code was not compiled because of `config.callbacks`, or a redirect or WebSocket address was refused. `reason` says which. | `reason`, `code` |

## Hooks for extensions

These names are in the list of events, but they are not DOM events. They are hooks an extension implements
(see [Using and writing extensions](../../extensions/using-and-writing.md)); they are never dispatched on the DOM, so
`addEventListener` does not see them.

| Hook | What it is |
|---|---|
| `ypy:before:oob` | Extension hook (`ypy_before_oob`): runs for each out-of-band element and each partial of a response, before the core takes it out of the response. Return `false` to veto it: an out-of-band element stays ordinary content, a partial is dropped. `ypy-csp` uses it to ignore the ones without the page nonce. Never dispatched on the DOM. |
| `ypy:scope` | Extension hook: runs before author code is compiled, so an extension can add names to its scope or rewrite `code`. Never dispatched on the DOM. |
| `ypy:before:morph:node` | Extension hook: runs for each pair of nodes a morph matches. Return `false` to skip the node. Never dispatched on the DOM. |
| `ypy:before:morph:attr` | Extension hook: runs for each attribute a morph is about to set or remove. Return `false` to keep the attribute as it is. Never dispatched on the DOM. |
| `ypy:resolve:target` | Extension hook (`ypy_resolve_target(elt, selector)`): runs for every part of a selector (`ypy-target`, `ypy.find()`...) the core does not know, before it falls to `querySelectorAll`. Return an element, an array or a NodeList of them, `null` for no element, or `undefined` when the selector is not yours. `ypy-scope` uses it. Never dispatched on the DOM. |
| `ypy:dispose` | Extension hook (`ypy_dispose`): runs once when `ypy.dispose()` starts, before anything is taken down, so an extension removes the listeners, timers and observers it installed. Never dispatched on the DOM. |
| `ypy:process:*` | Extension hook family: `ypy:process:<type>` runs for a `<ypy-partial>` whose `type` is not built in, so an extension can add swap tasks. Never dispatched on the DOM. |

## Events added by extensions

Each extension adds its own events when it is loaded. The page of the extension says what each one carries.

### [`ypy-boot`](../../extensions/boot.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:ready` | no | element |

### [`ypy-cache`](../../extensions/cache.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:cache:hit` | no | element |
| `ypy:cache:revalidated` | no | element |
| `ypy:cache:expire` | no | document |
| `ypy:cache:evict` | no | document |

### [`ypy-commands`](../../extensions/commands.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:before:command` | yes | element |
| `ypy:commands:ignored` | no | element |

### [`ypy-csp`](../../extensions/csp.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:security:strip` | no | element |
| `ypy:security:violation` | no | element |

### [`ypy-csrf`](../../extensions/csrf.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:csrf:missing` | yes | element |

### [`ypy-dedupe`](../../extensions/dedupe.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:dedupe:joined` | no | element |

### [`ypy-dev`](../../extensions/dev.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:dev:message` | no | element |

### [`ypy-download`](../../extensions/download.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:download:start` | no | element |
| `ypy:download:progress` | no | element |
| `ypy:download:complete` | no | element |

### [`ypy-head`](../../extensions/head.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:head:before:merge` | yes | document |
| `ypy:head:before:add` | yes | document |
| `ypy:head:before:remove` | yes | document |
| `ypy:head:after:merge` | no | document |

### [`ypy-history-cache`](../../extensions/history-cache.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:history:cache:before:save` | yes | document |
| `ypy:history:cache:after:save` | no | document |
| `ypy:history:cache:before:restore` | no | document |
| `ypy:history:cache:after:restore` | no | document |
| `ypy:history:cache:hit` | yes | document |
| `ypy:history:cache:miss` | no | document |

### [`ypy-layers`](../../extensions/layers.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:layer:before:open` | yes | element |
| `ypy:layer:opened` | no | element |
| `ypy:layer:before:close` | yes | element |
| `ypy:layer:accepted` | no | element |
| `ypy:layer:dismissed` | no | element |

### [`ypy-lazy`](../../extensions/lazy.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:lazy:error` | no | element |
| `ypy:lazy:group` | no | element |

### [`ypy-list`](../../extensions/list.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:list:pruned` | no | element |

### [`ypy-multipart`](../../extensions/multipart.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:multipart:before:connection` | yes | element |
| `ypy:multipart:after:connection` | no | element |
| `ypy:multipart:before:part` | yes | element |
| `ypy:multipart:after:part` | no | element |
| `ypy:multipart:error` | no | element |
| `ypy:multipart:close` | no | element |

### [`ypy-native`](../../extensions/native.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:native:event` | yes | element |
| `ypy:native:before:visit` | yes | element |
| `ypy:native:navigate` | yes | element |

### [`ypy-optimistic`](../../extensions/optimistic.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:optimistic:applied` | yes | element |
| `ypy:optimistic:confirmed` | no | element |
| `ypy:optimistic:rolledback` | no | element |

### [`ypy-persist`](../../extensions/persist.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:track:reload` | yes | element |

### [`ypy-poll`](../../extensions/poll.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:poll:stop` | no | element |

### [`ypy-prompt`](../../extensions/prompt.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:prompt` | yes | element |

### [`ypy-retry`](../../extensions/retry.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:request:retry` | yes | element |

### [`ypy-signals`](../../extensions/signals.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:signals:patch` | yes | element |
| `ypy:signals:readonly` | no | element |
| `ypy:signals:error` | no | element |

### [`ypy-sort`](../../extensions/sort.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:sort:change` | yes | element |
| `ypy:sort:revert` | no | element |

### [`ypy-sse`](../../extensions/sse.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:sse:before:connection` | yes | element |
| `ypy:sse:after:connection` | no | element |
| `ypy:sse:before:message` | yes | element |
| `ypy:sse:after:message` | no | element |
| `ypy:sse:error` | no | element |
| `ypy:sse:close` | no | element |
| `ypy:release` | no | element |

### [`ypy-stream`](../../extensions/stream.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:stream:open` | no | element |
| `ypy:stream:reconnect` | no | element |
| `ypy:stream:close` | no | element |
| `ypy:stream:error` | no | element |
| `ypy:stream:signals` | no | element |
| `ypy:stream:refresh` | yes | element |
| `ypy:stream:pause` | no | element |
| `ypy:stream:resume` | no | element |

### [`ypy-switch`](../../extensions/switch.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:switch` | no | element |

### [`ypy-telemetry`](../../extensions/telemetry.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:telemetry:error` | no | element |

### [`ypy-test`](../../extensions/test.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:test:idle` | no | document |

### [`ypy-upload`](../../extensions/upload.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:upload:progress` | no | element |

### [`ypy-ws`](../../extensions/ws.md)

| Event | Cancelable | Target |
|---|---|---|
| `ypy:ws:before:connection` | yes | element |
| `ypy:ws:after:connection` | no | element |
| `ypy:ws:close` | no | element |
| `ypy:ws:error` | no | element |
| `ypy:ws:before:message:outgoing` | yes | element |
| `ypy:ws:after:message:outgoing` | no | element |
| `ypy:ws:before:message:incoming` | yes | element |
| `ypy:ws:after:message:incoming` | no | element |
