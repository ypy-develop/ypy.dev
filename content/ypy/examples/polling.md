---
order: 250
title: Polling
description: Poll a background job every second, let the server change the interval, and stop the polling with a 286 status when the job is done.
---

# Polling

A job runs in the background and the page shows its progress: the element asks for the status with `every 1s`, and the server decides when to slow down and when to stop.

## The page

The `ypy-poll` extension adds the rest: a hidden tab polls less, the server can change the interval, and a response
with status `286` stops the polling.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-poll.min.js"></script>

<div id="job" ypy-get="/jobs/42" ypy-trigger="every 1s">
  <progress value="0" max="100"></progress> Starting...
</div>
```

The default swap is a morph of the children, so the element itself stays in the page and keeps the interval the server
chose. With `outerHTML` the element would be replaced and start again from the `1s` written here.

## What the server answers

While the job runs, `GET /jobs/42` answers `200` with the new state and sets the pace:

```text
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
YPY-Poll-Interval: 2s
```

```html
<progress value="35" max="100"></progress> 35%
```

Near the end the server can ask for faster polls with `YPY-Poll-Interval: 500ms`. When the job is done it answers with
status `286` and the final state:

```text
HTTP/1.1 286
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
```

```html
<a href="/jobs/42/result">Download the result</a>
```

| Job state | Status | `YPY-Poll-Interval` | Body |
|---|---|---|---|
| Running, under 90% | `200` | `2s` | the progress bar |
| Running, over 90% | `200` | `500ms` | the progress bar |
| Done | `286` | none | the link to the result |
| Unknown job | `404` | none | an error message |

## How it works

- `every 1s` is part of the core: the element sends its request every second. The extension adds the rest.
- In a hidden tab the interval is 4 times longer. `ypy-poll="hidden:pause"` stops polling there, and
  `ypy-poll="hidden:10"` changes the factor. When the tab is shown again the element polls at once if a poll was due.
- `YPY-Poll-Interval` accepts milliseconds (`2000`) or a time (`2s`, `500ms`, `1m`). It replaces the interval from the
  next poll on and stays until another response changes it. The server cannot ask for less than `min` (`100ms` by
  default).
- `286` stops the polling of that element and fires `ypy:poll:stop`. The body is still swapped, so it carries the final
  state. Without the extension a `286` is an ordinary success and the polling goes on.
- A failed request does not stop the polling. Only `286` does, so send it on purpose.

## Variations

- Poll only while the user looks: `<div ypy-get="/feed" ypy-trigger="every 5s" ypy-poll="hidden:pause">`.
- Skip a swap whose content did not change with the [`ypy-ptag`](../extensions/ptag.md) extension.
- For many changes per second, or changes the server starts, use [live updates](live-updates.md) instead.

## See also

- [`ypy-poll`](../extensions/poll.md)
- [Progress bar](progress-bar.md)
- [Request headers and response headers](../reference.md#request-headers)
