---
title: "Every failure in one place: the ypy-telemetry extension"
description: "Reports every client failure (no answer, a 5xx, a missing target, a refused swap, a lost connection) in one shape for an error tracker or a log."
order: 430
---
# Every failure in one place: the `ypy-telemetry` extension

A page fails in many ways. A request gets no answer, a response is a 500, a target matches nothing, a swap is refused, an
extension loses its connection. Each one has its own event, and some of them only reach the console. `ypy-telemetry`
reports all of them in one shape, so you can send them to an error tracker or a log with one handler.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-telemetry.js"></script>
<script>
  ypy.onError(report => navigator.sendBeacon('/client-errors', JSON.stringify(report)))
</script>
```

The extension is part of `ypy.max`. The extension reads
nothing from the page and adds no attribute and no configuration key.

## The report

Every failure becomes `{code, url, method, status, target, ctxId}`:

| Field | Value |
|---|---|
| `code` | a stable name of what failed (table below) |
| `url` | the request address without query, fragment or credentials; `null` when there is none |
| `method` | the HTTP method in upper case; `null` when there is none |
| `status` | the HTTP status of the answer; `null` when there was none |
| `target` | a short selector such as `"div#results"` (tag and id, never content), or the selector that matched nothing; `null` when unknown |
| `ctxId` | a number shared by every report of one request; `null` outside a request |

Codes:

| `code` | What failed |
|---|---|
| `network` | the request got no answer |
| `offline` | the browser said there was no connection |
| `timeout` | the request ran out of time |
| `status:<n>` | the response has status 400 or more: `status:404`, `status:500` |
| `target-missing` | no element matches the target; `target` is the selector |
| `swap-error` | putting the response into the page failed (reported once) |
| `blocked:<reason>` | a swap, redirect or script was refused, for example `blocked:redirect` |
| `extension:<name>` | an extension reported an error: `sse`, `ws`, `stream`, `multipart`, `lazy` or `signals` |
| `error` | any other `ypy:error`, for example a listener that threw |

## Two ways to listen

`ypy.onError(handler)` calls `handler(report)` for every failure and returns a function that cancels the
subscription. It throws a `TypeError` when the argument is not a function. A handler that throws is logged to the
console and does not stop the others. The same function registered twice is two subscriptions, cancelled one by one.

```js
const stop = ypy.onError(report => {
  if (report.code.startsWith('status:5')) tracker.capture('server error', report)
})
// later: stop()
```

`ypy:telemetry:error` is the same report as an event. It fires on the element the failure belongs to (on `document`
when that element is gone), bubbles and is not cancelable. `event.detail` is the report.

```js
document.addEventListener('ypy:telemetry:error', (event) => {
  console.warn(event.detail.code, event.detail.url)
})
```

Use the event when the code that reports lives in a part of the page that does not own `ypy`, or when you want to
listen on one subtree.

## What the report never holds

No body, no field value, no header, no cookie and no error message: those can carry what the user typed or a token.
The address loses its query, fragment, user and password. A path can still hold an identifier (`/orders/8841`), so a
tracker that must not see it should cut the `url` on its side.

## Example: a 500 from a form

```html
<form ypy-post="/orders" ypy-target="#result" ypy-swap="innerHTML" class="space-y-2">
  <input name="sku" class="rounded border px-2 py-1">
  <button class="rounded bg-sky-600 px-3 py-1 text-white">Order</button>
</form>
<div id="result"></div>
```

The server answers `500`. The handler receives:

```json
{"code": "status:500", "url": "https://shop.example/orders", "method": "POST",
 "status": 500, "target": "div#result", "ctxId": 3}
```

If the server was unreachable instead, the code is `network` and `status` is `null`. Reports of the same request (for
instance a `status:422` followed by a `target-missing`) share the `ctxId`, so a tracker can group them.

## The server

The server has no part in the report: it is made in the browser, from what the browser saw, and nothing is sent unless
your handler sends it. The `ctxId` is local to the page. An endpoint that receives the reports (the `sendBeacon` above) is an ordinary POST handler. Cap the size of the body it accepts, read the fields you want
(`code`, `url`, `method`, `status`, `target`, `ctxId`) and log them.

## Pitfalls

- Each report is one failure, not one request. A request that fails can give two reports (`status:500` and a
  `target-missing`). Group by `ctxId`.
- A swap error is reported once, as `swap-error`, not also as `error`.
- A refused redirect has no `ctxId`: `blocked:*` reports are about a URL, not a request.
- A 4xx you handle yourself is still a report. A `422` with an error target is reported as `status:422`. Filter by
  code in your handler if validation answers are not errors for you.
- Streams: a lost SSE or WebSocket connection is `extension:sse` or `extension:ws`, usually with no `status`.
- Without the extension, `ypy.onError` does not exist: guard calls from code that may load before it
  (`ypy.onError?.(...)`).

See also: [Events](../index.md#events-and-logging), [Testing](test.md), [`ypy-dev`](dev.md) for the same information while you
develop.
