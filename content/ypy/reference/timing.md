---
title: "Request timing: ctx.timing"
description: "Every request context carries ctx.timing, an object with the moments (in milliseconds) at which the request reached each phase."
order: 110
---
# Request timing: `ctx.timing`

Every request context carries `ctx.timing`, an object with the moments (in milliseconds) at which the request
reached each phase. It costs nothing when nobody reads it: the core only writes five numbers.

Use it to answer "why was this click slow?": the request waited in a queue, the server was slow, or the swap
took long. There is nothing to load and no configuration.

```js
document.addEventListener('ypy:finally:request', ({detail: {ctx}}) => {
  const t = ctx.timing
  console.table({
    id: ctx.id,
    action: ctx.request.action,
    waited: t.sent - t.start,                        // queue, confirmation, ypy:before:request
    server: t.response - t.sent,                     // until the response headers arrived
    swap: t.swapped != null ? t.swapped - t.response : null,   // body, parse and insertion
    total: t.settled - t.start,
  })
})
```

## The fields

All values come from `performance.now()`: milliseconds, with decimals, counted from the page load. Only the
differences mean something. A phase that did not happen is absent (`undefined`), never `0`.

| Field | When it is set |
|---|---|
| `start` | the request context was created: the trigger fired and YPY built the request. It is set before the request waits for its turn |
| `sent` | just before `fetch` is called, after the queue (`ypy-sync`), the `ypy-confirm` answer and the `ypy:before:request` listeners |
| `response` | the response headers arrived. The body is read after this |
| `swapped` | every swap task of the response was inserted. Absent when nothing was swapped |
| `settled` | the request ended. It is the last thing set before `ypy:finally:request` |

The fields are always in this order: `start <= sent <= response <= swapped <= settled`.

`ctx.id` is next to it. It is a number that grows with every request of the page and is the same one used to
discard an old response, see [request generations](attributes/ypy-sync.md).

## Where you can read it

`ctx` is in the `detail` of every request and swap event, so a listener sees the fields that were set so far.

| Event | What `ctx.timing` already has |
|---|---|
| `ypy:before:request` | `start` |
| `ypy:after:request`, `ypy:before:swap` | `start`, `sent`, `response` |
| `ypy:before:history:update`, `ypy:after:swap`, `ypy:finally:swap` | also `swapped` |
| `ypy:finally:request` | all of them that apply, `settled` included |

`ypy:finally:request` is the one to use for a report: it fires once per request, also when the request failed
or was aborted. A request that the queue refused never ran and has no such event; see
[request queue events](events/request-queue.md).

## What the numbers say

| Interval | Meaning | When it is large |
|---|---|---|
| `sent - start` | time before the request left the browser | it waited behind another request (`queue` strategies of [`ypy-sync`](attributes/ypy-sync.md)), a `ypy-confirm` dialog was open, or a listener of `ypy:before:request` was slow |
| `response - sent` | request, server and network, until the headers | the server or the network is slow. With [`ypy-retry`](../extensions/retry.md) it also includes every retry and its wait, because the retry happens inside that one call |
| `swapped - response` | download of the body, parsing, inserting and settling | a big response, a slow swap listener, a `swap:` or `settle:` delay, or a view transition |
| `settled - swapped` | what is left after the swap | normally close to zero. It grows when an extension that delivers its own response is still working |
| `settled - start` | everything | |

A request without a response (the network failed, the timeout fired, the user aborted) has `start` and perhaps
`sent`, and `settled`. Read `ctx.errorKind` to know why.

## Examples

**Show only slow requests.**

```js
document.addEventListener('ypy:finally:request', ({detail: {ctx}}) => {
  const t = ctx.timing, total = t.settled - t.start
  if (total > 1000) console.warn(`slow request: ${ctx.request.method} ${ctx.request.action}`, Math.round(total), 'ms')
})
```

**Find a queue problem.** If `sent - start` is high for a button that the user presses twice, the second
press waited for the first. Look at the `ypy-sync` strategy of the element and at the
[request queue events](events/request-queue.md).

**Join the server's own timing.** `ctx.response.headers` is the `Headers` object of the response. A server
that sends the standard `Server-Timing` header can be read next to the client numbers, to separate the time
spent in the handler from the time spent on the network:

```go
w.Header().Set("Server-Timing", "db;dur=12, render;dur=4")
```

```js
document.addEventListener('ypy:after:request', ({detail: {ctx}}) => {
  console.log(ctx.id, ctx.response.headers.get('Server-Timing'), ctx.timing.response - ctx.timing.sent)
})
```

## TypeScript

```ts
interface YpyRequestTiming {
  start: number
  sent: number
  response: number
  swapped: number
  settled: number
}
```

`ctx.timing` is typed as complete. In a listener that runs early, the later fields are not set yet, so test with
`!= null` before you subtract.

## Pitfalls

The numbers are not wall-clock time, so do not compare them with `Date.now()` or send them to a server as dates. Send
the differences.

A context that code passes to `ypy.swap()` does not come from a request, and it may have no `ctx.timing` until the swap
sets `swapped`. Check that it exists.

`swapped` is absent when no swap ran: the response was cancelled in `ypy:before:swap`, refused by
[`YPY-Group: atomic`](atomic.md), discarded by `ypy-sync="latest"`, or answered with a header such as `YPY-Redirect`.
Subtracting from it gives `NaN`.

A queued request starts counting early. Its `start` is when the user acted, not when the request was sent, so
`sent - start` is how long the user waited.

Copy the numbers you need and let `ctx` go. It holds the elements and the response, and keeping it alive keeps them
too.
