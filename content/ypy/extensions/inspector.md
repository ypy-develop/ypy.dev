---
title: "The inspector: the ypy-inspector extension"
description: "A floating panel for development that records requests, swaps, ypy events, signals and stream events as they happen. Open it with Alt+Shift+Y."
order: 160
---
# The inspector: the `ypy-inspector` extension

A floating panel that shows what YPY is doing, as it happens: the requests, the swaps, the `ypy:*` events with where
they came from and where they act, the signals and the stream events. It needs no browser extension and no
build step, and it is meant for development.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-inspector.min.js"></script>
```

When the page sets `config.extensions`, list `ypy-inspector` in it, or it is refused. It is not part of `ypy.max`.
Load it in the development layout only; for the console messages and the simulators, see [`ypy-dev`](dev.md). The two
work together (the `ypy-dev` messages show up as `ypy:dev:message` events in the Events tab) and neither needs the
other. The extension is 4,749 bytes min+gzip.

Recording starts when the script loads, with the panel closed, so what happened before you opened it is there.

## What the panel looks like

A small `YPY` button sits in the bottom right corner of the page. Click it, or press Alt+Shift+Y, and it becomes a
dark panel about 680 px wide and 40% of the window high, in the same corner. From top to bottom:

```
+----------------------------------------------------------------------------+
| YPY inspector  [Requests 3] [Swaps 4] [Events 61] [Signals 2]   [Pause]    |
|                                                  [Clear] [Export] [Close]  |
| [ Filter: words that every row must contain ______________________ ]       |
|----------------------------------------------------------------------------|
| #14  +5120ms  GET   /orders?page=2          200   38 ms    1204 B          |  green
| #11  +4810ms  POST  /orders                 422   61 ms     318 B          |  red
| #7   +2210ms  GET   /orders                 200   12 ms    2210 B          |  green
|----------------------------------------------------------------------------|
```

- The title bar has one button per tab, each with the number of rows, then Pause, Clear, Export and Close.
- The filter box sits under it; the list fills the rest and scrolls.
- Each row starts with its number (`#14`) and the time since the page loaded (`+5120ms`); the newest row is on top.
- Rows are green when they went well, red when the status is 400 or more, the request failed or the event was
  cancelled, yellow when a request was dropped, aborted or is waiting, and grey while it is still in flight.
- Click a row (or focus it and press Enter) to open its detail under it as formatted JSON; click again to close it.
- The Signals tab appears when `ypy-signals` is loaded or a signal event happened; the Stream tab appears with the
  first stream event.

## The tabs

| Tab | Row | Columns |
|---|---|---|
| Requests | one per request | method, URL, status (or the state: `sent`, `queued`, `dropped`, `aborted`, `error`), time in ms, response size in bytes. The detail has the source element, the target, the swap style, the `YPY-*` headers of the request and of the response, and for a failure its kind (`network`, `offline`, `timeout`) and message |
| Swaps | one per swap of a response (the main one, each out-of-band element, each partial) | target, swap style, size of the inserted content, state (`swapping`, `done`, `error`); the detail links to the request (`request` is its row number) |
| Events | one per event | name, origin and destination (below); the detail has the data of the event and the elements that listen for it |
| Signals | the signals of the page now, then one row per `ypy:signals:*` event | scope, name and value of each signal now (read again every half second while the tab is open); patches show the JSON patch |
| Stream | one per `ypy:stream:*`, `ypy:sse:*`, `ypy:ws:*` and `ypy:multipart:*` event | name, origin, destination, data |

### Origin and destination of an event

Events that move through the page are hard to follow because the element that fires them is not the element they
act on. Each Events row says both:

- Origin is the element the event fired on (`button#save`, `document`).
- Destination is where it acts: the swap target of the request (`div#results`), the selector that was looked for
  (for `ypy:target:missing`, `ypy:target:fallback`), or the targets of the swap tasks (`ypy:before:swap`).
  A dash means the event acts only on its origin.

Open the row to see who listens: the elements that carry a `ypy-on:<event>` (or `ypy-on::<short event>`)
attribute for it, and the elements whose `ypy-trigger` names it. The kebab form (`after-swap`) and the colon form
(`after:swap`) are the same event. Listeners added with `addEventListener` cannot be seen from the page and are not
listed.

A row is marked cancelled (red) when a listener called `preventDefault()` on it.

## Filter, pause, clear, export

- Filter keeps the rows whose text contains every word you type, in any order and in any case:
  `404 orders` shows the requests for `/orders` that ended in 404. It applies to the tab you are on and survives
  new rows. The text of a row is what its columns show.
- Pause stops recording (rows that happen while it is on are not kept) and shows the list as it is. Press it
  again to go on.
- Clear empties every list.
- Export downloads `ypy-inspector-<time>.json`. `ypy.inspector.export()` returns the same text.

The JSON, schema 1:

```json
{
  "version": 1,
  "url": "http://localhost:8080/orders",
  "at": "2026-11-15T14:03:22.114Z",
  "requests": [{"id": 7, "t": 2210, "method": "GET", "url": "/orders", "status": 200, "state": "done",
                "ms": 12, "size": 2210, "source": "button#reload", "target": "div#orders", "swap": "innerHTML",
                "requestHeaders": {"YPY-Request": "true", "YPY-Request-Type": "partial"},
                "responseHeaders": {"YPY-Trigger": "loaded"}}],
  "swaps": [{"id": 8, "t": 2222, "request": 7, "type": "main", "target": "div#orders", "style": "innerHTML",
             "size": 2210, "state": "done", "ms": 3}],
  "events": [{"id": 5, "t": 2201, "name": "ypy:before:request", "from": "button#reload", "to": "div#orders",
              "info": {"ctx": "GET /orders"}}],
  "signals": [], "streams": [],
  "signalsNow": [{"scope": "div#filters", "name": "$q", "value": "open"}]
}
```

`t` is milliseconds since the script loaded and `id` is one counter for every list. The lists hold at most
`max` rows each (500); the oldest go first. Header values are only the `YPY-*` ones: cookies, `Authorization` and
every other header are not recorded. The body of a response is not recorded, only its size. The detail of an event
holds elements and contexts as short descriptions, never the objects.

## Keys and options

| Key | |
|---|---|
| `Alt+Shift+Y` | open and close the panel (also the `YPY` button) |
| `Escape` | close the panel, when the focus is inside it |
| `Enter` on a row | open or close its detail |

Set the options with `ypy.config.inspector`, from the configuration tag or a script before the extension loads:

```html
<meta name="ypy-config" content='inspector.open:true,inspector.key:"Ctrl+Alt+K",inspector.max:1000'>
```

| Key | Value | Default |
|---|---|---|
| `open` | show the panel at start | `false` |
| `key` | the shortcut, modifiers (`Alt`, `Shift`, `Ctrl`, `Meta`) and a letter, joined with `+` | `Alt+Shift+Y` |
| `max` | rows kept in each list | `500` |

`inspector:false` (the whole value) leaves the page untouched: no panel, no listener, no recording.

## Turn it off

- `inspector:false` in the configuration tag, or not loading the script.
- `ypy.inspector.disable()` removes the panel, the listeners and the shortcut and gives `ypy.trigger` back;
  `ypy.inspector.enable()` starts it again. The recorded rows stay until you clear them.
- The Close button only hides the panel; recording goes on.

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.inspector.open(tab?)` | Shows the panel, on a tab if given: `'requests'`, `'swaps'`, `'events'`, `'signals'`, `'streams'` |
| `ypy.inspector.close()`, `toggle()` | Hide or toggle the panel |
| `ypy.inspector.pause(on = true)` | Stops or resumes the recording |
| `ypy.inspector.clear()` | Empties every list |
| `ypy.inspector.filter(text)` | Sets the filter, and the text of the filter box |
| `ypy.inspector.data()` | The lists, as in the export, without serializing |
| `ypy.inspector.export()` | The JSON text |
| `ypy.inspector.watch(...names)` | Records an event that is dispatched with `dispatchEvent` instead of `ypy.trigger` |
| `ypy.inspector.enable()`, `disable()` | Start and stop the inspector |

## How it sees the events

Every event YPY and its extensions fire goes through `ypy.trigger`, except a few that the core dispatches directly
(`ypy:request:queued`, `ypy:request:dropped`, `ypy:request:aborted`, `ypy:target:missing`, `ypy:target:fallback`,
`ypy:swap:error`, `ypy:security:blocked`, `ypy:history:restore:fragment`, `ypy:offline`, `ypy:online`,
`ypy:announce`). The inspector listens for those by name and wraps `ypy.trigger` to listen, on first use, for every
other name, so an event of an extension it has never heard of (a stream extension, a layer extension) is recorded
from its first firing. An event that an extension sends with `dispatchEvent`, or that the core adds later and
dispatches directly, needs `ypy.inspector.watch('ypy:the:name')`. `config.metaCharacter` renames the events; the
inspector shows the names as they are fired.

## Security and CSP

- The panel is a Shadow DOM with an adopted style sheet: no `style` attribute and no `<style>` element, so a strict
  `style-src` is fine. It builds its rows with `textContent`; no markup from a response, an attribute or an event is
  ever parsed, so Trusted Types and `require-trusted-types-for 'script'` are fine too.
- It lives on `<html>`, outside `<body>`, and carries `ypy-ignore`, so a swap never removes or processes it.
- It records only what is on the page already, keeps it in memory and sends it nowhere. Export writes a file you
  choose to keep; look at it before you share it, the URLs and the signals may hold your data.
