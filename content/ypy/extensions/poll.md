---
title: "Smart polling: the ypy-poll extension"
description: "Polling that slows down in a background tab, lets the server pick the next interval and lets the server end the polling."
order: 290
---
# Smart polling: the `ypy-poll` extension

A plain `every` trigger asks the server again and again at the same pace, whether anyone is looking at the page or
not. `ypy-poll` slows that down in a background tab, lets the server choose the interval and lets the server end the
polling.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-poll.js"></script>

<div ypy-get="/jobs/42/status" ypy-trigger="every 2s">Waiting...</div>
```

The element above asks every 2 seconds in a visible tab and every 8 seconds in a hidden one. When the job is done the
server answers `286` with the final markup, and the element stops asking.

Without the extension, `every` works as before: a fixed interval, and a
`286` is an ordinary success.

## Background tabs

While `document.hidden` is true the interval is longer: 4 times by default. The change is made when the tab is
hidden and when it is shown, not only at the next poll.

| `ypy-poll` / `ypy.config.poll` | Meaning |
|---|---|
| `hidden:<n>` | the interval is `n` times longer while the tab is hidden (default `4`; `1` keeps it) |
| `hidden:pause` | no polling while the tab is hidden |
| `min:<time>` | the shortest interval a server may ask for with `YPY-Poll-Interval` (default `100ms`) |

The options are written as in `ypy-trigger` and can be combined: `ypy-poll="hidden:pause min:1s"`. A `hidden` value
below `1` is read as the default `4`; `0` and `false` pause like `pause`.

```html
<div ypy-get="/feed" ypy-trigger="every 5s" ypy-poll="hidden:pause">...</div>
```

```js
ypy.config.poll = { hidden: 10 };        // for every poll of the page
```

The attribute wins over the configuration. When the tab is shown again, the element polls at once if a poll has
become due while it was hidden, and otherwise at the time it would have polled. Showing and hiding a tab again and
again never makes a poll come earlier than the interval allows.

`ypy-poll="false"` leaves the element to the core: no slowdown and no `YPY-Poll-Interval`. The status `286` still
stops it.

## The interval chosen by the server

The response header `YPY-Poll-Interval` sets the interval of the element that sent the request, from the next poll
on. The value is milliseconds, or a time with a unit.

```http
YPY-Poll-Interval: 5000
YPY-Poll-Interval: 5s
YPY-Poll-Interval: 500ms
YPY-Poll-Interval: 1m
```

- It replaces the interval of the `every` trigger, longer or shorter. A busy server answers `30s`; a job that is about
  to end answers `500ms`.
- It stays until another response changes it. A response without the header does not reset it.
- The slowdown for hidden tabs still applies to the new interval.
- A value that is not an interval is ignored. The server cannot ask for less than `min` (`100ms` by default, option
  `min:<time>` of `ypy.config.poll`), so a mistake cannot turn the page into a flood of requests.
- The interval belongs to the element. When a swap replaces the element with a new one, the new element starts again
  from its markup; with a morph swap the element stays and keeps the interval.

## Stopping with 286

A response with status `286` ends the polling of the element that sent it. The content of that response is
swapped as for any `2xx`, so it can carry the final state. Then `ypy:poll:stop` fires on the element.

```html
<div ypy-get="/export/7" ypy-trigger="every 1s">Preparing...</div>
```

```text
HTTP/1.1 286
Content-Type: text/html

<a href="/export/7.zip">Download</a>
```

The server decides when the work is over, so the page needs no timer of its own. While the job runs it answers `200` with
the progress and `YPY-Poll-Interval: 1s` to set the pace. When it is done it answers `286` with the link, and the body
is swapped like any 2xx.

Other elements keep polling. The polling restarts when the element is processed again (for example, a swap brings
it back).

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:poll:stop` | the polling element | `ctx` (the request), `status` (`286`) |

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `poll.hidden` | `4` | factor for a hidden tab, or `"pause"` |
| `poll.min` | `100` | shortest interval a server may set (ms, or a time) |

## Pitfalls

- A `286` is a success for the swap: the body is swapped. Send the final markup in it, not an error.
- A request that fails (the network is down) does not stop the polling; the element asks again at its pace. Only `286`
  stops it, so send it on purpose.
- A response without `YPY-Poll-Interval` does not reset the interval: a handler that sent `30s` once keeps that pace
  until it sends another value.
- An element that is swapped out (`outerHTML`) starts again from the interval written in its markup. Use a morph swap
  to keep the interval the server chose.
- A hidden tab with `hidden:pause` makes no requests, so the page can be stale when shown; the poll runs at once when
  the tab is visible and one was due.
- The extension takes over the `every` trigger of the elements it processes. `ypy-poll="false"` puts one back in the hands
  of the core.

## What it does not do

It drives the `every` trigger only. Other ways of polling (a `setInterval` of your own, `ypy-sse`) are not changed.
For skipping a swap whose content did not change, see [`ypy-ptag`](ptag.md).
