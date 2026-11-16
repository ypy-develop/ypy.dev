---
title: "ypy-sync"
description: "Decide what happens when an element asks for a request while another from the same place is still running: queue, drop, abort, replace and more."
order: 90
---
# ypy-sync

Decides what happens when an element asks for a request while another one from the same
place is still in flight. `ypy-sync` and `data-ypy-sync` mean the same thing and take the same value.

```html
<form ypy-post="/save" ypy-sync="this:replace">...</form>
```

## Syntax

```
ypy-sync="SELECTOR:STRATEGY"
```

`SELECTOR` names the element that owns the queue (default: the element itself). `STRATEGY`
is one of the values below (default: `queue first`). Either part may be left out:
`ypy-sync="drop"` is the element itself with `drop`, `ypy-sync="#panel"` is `#panel` with
`queue first`.

| Selector | Owner of the queue |
|---|---|
| none or `this` | the element that starts the request |
| a CSS selector | the first match, found like `ypy-target` finds its element |
| `closest form`, `find ...`, `next ...` | the extended selectors of the other attributes |
| `target` | the element the response is swapped into |

Two requests wait for each other only when they share the same owner. `target` is for the
case where different buttons update the same panel:

```html
<button ypy-get="/a" ypy-target="#panel" ypy-sync="target:replace">A</button>
<button ypy-get="/b" ypy-target="#panel" ypy-sync="target:replace">B</button>
<div id="panel"></div>
```

## Strategies

| Strategy | A request arrives while another is in flight |
|---|---|
| `drop` | the new request is refused |
| `abort` | the new request is refused; a request that was started with `abort` is itself aborted by any other kind of request |
| `replace` | the request in flight is aborted and the waiting ones are discarded; the new one runs |
| `queue first` (default) | the new request waits, unless one is already waiting, then it is refused |
| `queue last` | the new request waits and replaces the one that was waiting |
| `queue all` | the new request waits behind all the others |
| `latest` | the new request runs at once; the older ones keep running but their responses are discarded when they arrive |
| `parallel` | no queue: the request runs at once, waits for none and aborts none |

`queue first` is the default. Nothing is dropped or aborted without an
event: see [request events](../events/request-queue.md). `latest` and `parallel` refuse,
queue and abort nothing, so they fire none of those events.

## `latest`: the newest request wins

Use `latest` when the user can ask again before the answer comes, and only the last question matters: a
search box, a filter, a price that depends on a field.

```html
<input type="search" name="q" ypy-get="/search" ypy-trigger="input changed delay:200ms"
       ypy-target="#results" ypy-sync="this:latest">
<div id="results"></div>
```

With `replace` the request in flight is aborted. With `latest` it is not: the browser lets it finish, and
YPY throws its response away when it arrives. The difference matters when the server should not see
a cancelled request (an aborted `POST` may still have run on the server) and when the answer of the older
request may arrive after the newer one.

- The new request takes the slot at once. It does not wait and nothing is refused.
- The request that held the slot gets `ctx.superseded = true`. Its response is discarded whenever it
  arrives, before or after the response of the newest request: no swap, no history update, no
  `ypy:before:swap` or `ypy:after:swap`. Its own request events up to `ypy:after:request` still fire, and
  so does `ypy:finally:request`.
- Several requests can be in flight at once. Each new one supersedes the one that held the slot, and the ones
  superseded earlier stay superseded.
- Indicators, `ypy-disable` and `ypy-inert` stay on until the last request of the element has ended.
- Response headers that act on their own (`YPY-Redirect`, `YPY-Refresh`, `YPY-Location`) are read before the
  response is discarded, so a superseded response that carries one still runs it. Keep these headers out of
  endpoints that you call with `latest`.

## `parallel`: no queue at all

`parallel` takes no slot. Every request runs at once, in any number, and each response is swapped as it
arrives. Use it for requests that are independent of each other, such as several `like` buttons that
share one inherited `ypy-sync`, or content that is added to a list with `beforeend`.

```html
<ul id="log"></ul>
<button ypy-post="/log" ypy-target="#log" ypy-swap="beforeend" ypy-sync="this:parallel">Add entry</button>
```

Three quick clicks send three requests. Their answers are appended in the order they arrive. `parallel` does
not switch off the rule below: when two responses replace the same target, the older one never overwrites the
newer one.

## Request generations

Every request has an `id` (`ctx.id`), a number that grows with each request of the page. YPY uses it to stop
an old answer from overwriting a new one on the same target, even when the two requests do not share a queue:
two buttons with their own `ypy-sync`, a `parallel` element, or a `YPY-Retarget` that sends two answers to
one panel.

The rule is applied just before a response is inserted. The target remembers the `id` of the newest
request that overwrote it. If the request that is about to swap has a smaller `id`, its response is
discarded and the target is left alone.

```html
<button id="a" ypy-get="/a" ypy-target="#panel">A</button>
<button id="b" ypy-get="/b" ypy-target="#panel">B</button>
<div id="panel"></div>
```

Click A and then B. If the answer of B arrives first, `#panel` shows B, and the late answer of A is
discarded. If the answer of A arrives first, `#panel` shows A for a moment and then B.

- The rule holds for the swap styles that replace content: `innerHTML`, `outerHTML`, `textContent`, the
  morph styles and `outerSync`. Styles that only add content (`beforebegin`, `afterbegin`, `beforeend`,
  `afterend`) are never discarded, so a list built with `beforeend` keeps every answer.
- The rule works per target element. Two requests that go to different targets never discard each other.
- A request that was aborted is not applied either, also when the abort comes while its swap waits for the
  `swap:` delay.
- The rule needs no `ypy-sync`. To also stop the old request from finishing, use `ypy-sync="target:replace"`.

## Inheritance

Inheritance is explicit, like every other attribute: `ypy-sync:inherited` on a container
applies to its descendants. `this` then means the element that declares the attribute, so
all the buttons of the container share one queue.

```html
<div ypy-sync:inherited="this:drop">
  <button ypy-post="/like">Like</button>
  <button ypy-post="/share">Share</button>
</div>
```

## When the element goes away

The queue does not live on the element. When the element is removed, or morph rewrites one
of its `ypy-*` attributes and the element is restarted, the requests that
were waiting are discarded with reason `cleanup`, and the node starts with an empty queue for
whatever it holds next. The request already in flight is not aborted.
