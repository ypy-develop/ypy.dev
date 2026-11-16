---
order: 20
title: When to use YPY
description: Which projects suit YPY, which need care, which should use something else, and a four-question test to decide before you write the first screen.
---

# When to use YPY

YPY suits a site where the server already knows how to draw the page and the browser only needs to update parts of it.
For some projects it is the wrong tool, and it is cheaper to find out before the first screen.

## It fits

- Server-rendered applications: admin panels, dashboards, back offices, CRUD screens, stores, content sites. The data
  and the rules are on the server, and most interactions are "do this, show me the new state".
- Sites that have to work without the script. Links and forms come first and YPY improves them ([Progressive
  enhancement](progressive-enhancement.md)).
- Teams that write the server in one language. The templates are the interface, so one person can change a screen from
  the data down to the markup.
- Pages with some life: search as you type, infinite lists, inline editing, validation, live counters, a stream of
  updates. The [examples](../examples/index.md) show each one.
- Projects that cannot add a build pipeline for the front end, or do not want to maintain one ([No build
  step](no-build-step.md)).

## It fits with care

- Real-time collaboration on a shared document. Server-sent events and streams work for notifications, progress and
  feeds ([`ypy-sse`](../extensions/sse.md), [`ypy-stream`](../extensions/stream.md)). Several people editing the same
  text need merge logic, and no swap provides it.
- Offline use. The [cache](../extensions/cache.md) and the [history cache](../extensions/history-cache.md) keep pages
  around, and [`ypy-retry`](../extensions/retry.md) repeats a request that failed on the network. A page that must work
  with no connection at all needs a client that holds the data.
- A native app. The [`ypy-native`](../extensions/native.md) extension lets a shell show the same pages in a web view.
  That works when the screens are mostly content and forms.

## It does not fit

- Applications drawn mainly on the client: drawing tools, spreadsheets with live formulas, map editors, games, audio and
  video editors. The state is large, it changes many times a second, and the server plays no part in each change. Use a
  client framework and send data to the server on save.
- Interactions where the round trip is the problem: dragging an element under the pointer, resizing a panel, scrubbing a
  timeline. Keep those in a script and send the result when the gesture ends. [`ypy-sort`](../extensions/sort.md) does
  this for lists.
- A backend that only serves JSON to many different clients. If the server returns data and cannot draw HTML, YPY has
  nothing to swap. You can put a thin HTML layer in front, but at that point ask whether you are building the right
  thing.
- Pages that keep rich state across screens, such as a long wizard whose steps depend on each other in ways the server
  cannot see. Often the server can hold the wizard's state. If it cannot, think twice.

## A quick test

| Question | If yes |
|---|---|
| Can the server draw the answer to this interaction in one request? | YPY fits |
| Does the interaction need to feel instant while the network is slow? | add [optimistic UI](../extensions/optimistic.md), or keep this one in a script |
| Is the state only in the browser, and does it change continuously? | a client framework fits better |
| Do you need the page to work with no script? | YPY fits, and it is the case it was built for |

## Mixing

You choose per page, not per project. A server-rendered site with YPY can have one screen that is a client application,
mounted in an element YPY leaves alone (`ypy-ignore`), or a [third-party component](../index.md#third-party-javascript-and-web-components)
inside a page. The two meet through events and a few HTTP calls.

## See also

[Why hypermedia](why-hypermedia.md) · [Docs](../index.md) · [Migration](../migration.md)
