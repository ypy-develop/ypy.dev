---
title: "Request queue events"
description: "Requests that do not run when asked, or that are cut short, always fire an event. The events bubble from the element the request started from."
order: 20
---
# Request queue events

A request that does not run when it is asked for, or that is cut short, is never silent. The
events bubble, and are dispatched on the element the request started
from. `ypy:error` is not fired for them; it keeps its meaning.

| Event | When | `detail.reason` |
|---|---|---|
| `ypy:request:queued` | the request waits for its turn | `queue` |
| `ypy:request:dropped` | the request never ran and will not | `drop`, `abort`, `replace`, `queue-overflow`, `cleanup` |
| `ypy:request:aborted` | the request in flight was aborted before its response | `replace`, `abort`, `timeout`, `cleanup`, `user` |

`detail.ctx` is the request the event is about: for `dropped` and `queued` the request that
was refused or is waiting, for `aborted` the one that was in flight, not the one that took its place.

Reasons:

- `drop`, `abort`: the strategy of the new request refused it.
- `replace`: a `replace` request took the slot; it aborts the one in flight and drops the waiting ones.
- `queue-overflow`: `queue first` already had a waiting request, or `queue last` replaced the waiting one.
- `timeout`: the request exceeded its timeout (`ypy:error` also fires, as before).
- `user`: `ypy:abort` was triggered on the element, or code called `ctx.request.abort()`.
- `cleanup`: the element was removed or restarted while the request waited.

```js
document.addEventListener('ypy:request:dropped', e => {
  console.log('refused', e.detail.ctx.request.action, e.detail.reason)
})
```

An abort that happens after the response arrived (during the swap) is not reported: nothing
was cut short.
