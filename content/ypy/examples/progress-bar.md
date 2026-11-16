---
order: 100
title: Progress bar
description: Start a job with a button, then let each answer ask for the next progress value after a short delay, until the server answers that the job is done.
---

# Progress bar

A button starts a job on the server. The page then asks for the job's progress every 600 milliseconds and shows a bar.
Each answer carries the next request itself, so the job ends the moment the server stops sending one.

## Try it

<!-- demo:progress-bar -->

The job takes about six seconds. **Start again** runs it from the beginning.

## The page

```html
<div id="job">
  <button ypy-post="/export/start" ypy-headers='{"X-CSRF-Token": "…"}'
          ypy-target="#job" ypy-swap="outerHTML">Start the export</button>
</div>
```

## What the server answers

`POST /export/start` answers `200` with the block that does the polling. It asks for `/export/progress` once, after
600 ms, and replaces itself with the answer:

```html
<div id="job" ypy-get="/export/progress" ypy-trigger="load delay:600ms" ypy-swap="outerHTML" role="status">
  <progress value="0" max="100" aria-label="Export progress"></progress>
  <span>0%</span>
</div>
```

`GET /export/progress` answers `200` with the same block and a new value, so the chain goes on:

```html
<div id="job" ypy-get="/export/progress" ypy-trigger="load delay:600ms" ypy-swap="outerHTML" role="status">
  <progress value="40" max="100" aria-label="Export progress"></progress>
  <span>40%</span>
</div>
```

When the job is done the block has no `ypy-get`, so nothing follows:

```html
<div id="job" role="status">
  <p>The export is ready: 1,284 rows.</p>
  <button ypy-post="/export/start" ypy-headers='{"X-CSRF-Token": "…"}'
          ypy-target="#job" ypy-swap="outerHTML">Start again</button>
</div>
```

All three are fragments with `Vary: YPY-Request`.

## How it works

- `ypy-trigger="load delay:600ms"` fires when the element is processed and waits 600 ms before the request. The block
  that comes back is processed in turn, so it asks again. No timer lives in the page.
- Each answer replaces the whole block (`outerHTML`). The last one has no `ypy-get`, which is how the polling stops.
- The server keeps the position. A page that is closed and opened again starts from the button; a job that keeps its
  progress under an id can resume it.
- `role="status"` makes the percentage audible without moving the focus.
- The `POST` changes state, so the button carries the CSRF token in a header with `ypy-headers`. The polling `GET`s
  need none.

## Variations

- Poll with a fixed rhythm instead: `ypy-trigger="every 1s"` on a block that is morphed, not replaced. The
  [`ypy-poll`](../extensions/poll.md) extension stops it when the server answers with the status `286`, slows it in a
  background tab and lets the server set the pace with `YPY-Poll-Interval`.
- A progress bar for a file upload needs no polling: see [file upload](file-upload.md).

## See also

- [`ypy-poll`](../extensions/poll.md) and [polling](polling.md)
- [Updating other content](updating-other-content.md), for a job that changes several parts of the page
- [Triggers](../index.md#triggers)
