---
title: "The inspector: the ypy-inspector extension"
description: "A floating panel for development that records requests, swaps, ypy events, signals and stream events as they happen. Open it with Alt+Shift+Y."
order: 160
---
# The inspector: the `ypy-inspector` extension

A floating panel that shows what YPY is doing as it happens: the requests, the swaps, the `ypy:*` events with where they
came from and where they act, the signals and the stream events. It needs no browser extension and no build step, and it is
meant for development.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-inspector.min.js"></script>
```

It is not part of `ypy.max`, so load it in the development layout only. For console messages and the simulators, see
[`ypy-dev`](dev.md). The two work together, and the `ypy-dev` messages show up as `ypy:dev:message` events in the Events tab.
Recording starts when the script loads, with the panel closed, so what happened before you opened it is there.

## The panel

A small `YPY` button sits in the bottom right corner. Click it, or press Alt+Shift+Y, and it becomes a dark panel about 680 px
wide and 40% of the window high. The title bar has one button per tab, each with its row count, then Pause, Clear, Export and
Close. A filter box sits under it and the list fills the rest.

- Each row starts with its number and the time since the page loaded. The newest row is on top.
- Rows are green when they went well, red when the status is 400 or more, the request failed or the event was cancelled,
  yellow when a request was dropped, aborted or is waiting, and grey while it is in flight.
- Click a row, or focus it and press Enter, to open its detail as formatted JSON.

| Tab | Row | Columns |
|---|---|---|
| Requests | one per request | method, URL, status (or `sent`, `queued`, `dropped`, `aborted`, `error`), time, size. The detail has the source, target, swap style, the `YPY-*` headers of both sides and, for a failure, its kind and message |
| Swaps | one per swap of a response, including each out-of-band element and partial | target, style, size, state; the detail links to the request |
| Events | one per event | name, origin and destination; the detail has the data and the elements that listen for it |
| Signals | the signals of the page now, then one row per `ypy:signals:*` event | scope, name and value; patches show the JSON patch |
| Stream | one per `ypy:stream:*`, `ypy:sse:*`, `ypy:ws:*` and `ypy:multipart:*` event | name, origin, destination, data |

The Signals tab appears when `ypy-signals` is loaded, and the Stream tab with the first stream event.

### Origin and destination

The element that fires an event is often not the element it acts on, so each Events row says both. The origin is the element
the event fired on (`button#save`, `document`). The destination is the swap target of the request, the selector that was
looked for (`ypy:target:missing`), or the targets of the swap tasks (`ypy:before:swap`). A dash means the event acts only on
its origin. Open the row to see the elements that carry a `ypy-on:<event>` attribute for it and the elements whose
`ypy-trigger` names it. Listeners added with `addEventListener` cannot be seen from the page. A row is red when a listener
called `preventDefault()`.

## Filter, pause, clear, export

- The filter keeps the rows whose text contains every word you type, in any order and case. `404 orders` shows the requests
  for `/orders` that ended in 404.
- Pause stops recording until you press it again, and Clear empties every list.
- Export downloads `ypy-inspector-<time>.json`, and `ypy.inspector.export()` returns the same text.

```json
{
  "version": 1,
  "url": "http://localhost:8080/orders",
  "at": "2026-11-15T14:03:22.114Z",
  "requests": [{"id": 7, "t": 2210, "method": "GET", "url": "/orders", "status": 200, "state": "done",
                "ms": 12, "size": 2210, "source": "button#reload", "target": "div#orders", "swap": "innerHTML",
                "requestHeaders": {"YPY-Request": "true"}, "responseHeaders": {"YPY-Trigger": "loaded"}}],
  "swaps": [], "events": [], "signals": [], "streams": [], "signalsNow": []
}
```

`t` is milliseconds since the script loaded. Each list holds at most `max` rows (500), and the oldest go first. Only `YPY-*`
headers are recorded, so cookies and `Authorization` are not. Response bodies are not recorded, only their size. Look at an
export before you share it, because URLs and signals may hold your data.

## Options and keys

`Alt+Shift+Y` opens and closes the panel, `Escape` closes it when focus is inside, and `Enter` on a row opens its detail.
Set options with `ypy.config.inspector`:

```html
<meta name="ypy-config" content='inspector.open:true,inspector.key:"Ctrl+Alt+K",inspector.max:1000'>
```

| Key | Value | Default |
|---|---|---|
| `open` | show the panel at start | `false` |
| `key` | the shortcut: modifiers (`Alt`, `Shift`, `Ctrl`, `Meta`) and a letter, joined with `+` | `Alt+Shift+Y` |
| `max` | rows kept in each list | `500` |

`inspector:false` leaves the page untouched. `ypy.inspector.disable()` removes the panel, the listeners and the shortcut, and
`enable()` starts it again. The Close button only hides the panel, and recording goes on.

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.inspector.open(tab?)`, `close()`, `toggle()` | Show, hide or toggle the panel; `tab` is `'requests'`, `'swaps'`, `'events'`, `'signals'` or `'streams'` |
| `ypy.inspector.pause(on = true)` | Stops or resumes the recording |
| `ypy.inspector.clear()`, `filter(text)` | Empties the lists; sets the filter |
| `ypy.inspector.data()`, `export()` | The lists as objects; the JSON text |
| `ypy.inspector.watch(...names)` | Records an event dispatched with `dispatchEvent` instead of `ypy.trigger` |
| `ypy.inspector.enable()`, `disable()` | Start and stop the inspector |

Every event YPY and its extensions fire goes through `ypy.trigger`, except a few the core dispatches directly, which the
inspector listens for by name. An event that an extension sends with `dispatchEvent` needs
`ypy.inspector.watch('ypy:the:name')`.

## Security and CSP

The panel is a Shadow DOM with an adopted style sheet, with no `style` attribute and no `<style>` element, so a strict
`style-src` is fine. It builds rows with `textContent`, so Trusted Types are fine too. It lives on `<html>` and carries
`ypy-ignore`, so a swap never removes or processes it. It records only what is on the page already, keeps it in memory and
sends it nowhere.
