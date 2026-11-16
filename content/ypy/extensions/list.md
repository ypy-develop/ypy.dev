---
title: "Keyed and bounded lists: the ypy-list extension"
description: "Adds the key and limit modifiers to append and prepend swaps: an item that arrives twice replaces its copy, and the list stops growing past a set size."
order: 190
---
# Keyed and bounded lists: the `ypy-list` extension

A feed, a log or a chat adds items at one end of a list. Two problems come with that: the same item can arrive twice
(and appear twice), and the list only grows. `ypy-list` adds two modifiers to the `append` and `prepend` swap styles.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-list.js"></script>

<ul id="events" ypy-get="/events" ypy-trigger="every 3s"
    ypy-swap="append key:data-id limit:50">
  <li data-id="1">Started</li>
</ul>
```

Each poll appends the new events. An event that is already in the list replaces the old copy where it stands, and the
list never holds more than 50 items: the oldest go.

Without it the two modifiers are ignored and the swap works as
`append` or `prepend` always did.

## Modifiers

They go after the swap style: `ypy-swap="append key:id limit:20"`. They work for `append`, `prepend` and their
spellings `beforeend` and `afterbegin`, in the main swap, in a partial and in an out-of-band swap
(`ypy-swap-oob="append target:#feed limit:20"`; here the target is written with `target:`). On any other style they do
nothing.

| Modifier | Meaning |
|---|---|
| `key:<attribute>` | an incoming item whose attribute has the value of an item already in the list replaces it, in the same place. `key` alone means `id` |
| `limit:<n>` | after the swap the list has at most `n` items. For `append` the first (oldest) go, for `prepend` the last |

The list is the element the swap goes into, and every element child of it is an item. Text between items does not
count. Items that are not part of the feed (a header, an empty-state message) count too, so keep them outside the
list. Only the direct children of the list are compared by `key`: an element nested deeper in a response item is not.

`key`. An item without the attribute is always added. When one response holds the same key twice, the last one
wins. The old item is taken out the way the core takes out any removed content: its timers and listeners stop. The
new item is processed as any swapped content, so its `ypy-*` attributes work.

`limit`. It runs after `key`, so an update of an item already shown never pushes another one out. A value that is
not a number, or less than 1, is ignored. If a single response has more items than the limit, only the newest `n`
stay. A decimal is rounded down (`limit:2.5` keeps 2). Items taken out are cleaned up and, when there are any,
`ypy:list:pruned` fires.

## Events

| Event | On | Detail |
|---|---|---|
| `ypy:list:pruned` | the list | `removed` (the elements taken out, in document order), `limit`, `task` (the swap task) |

It fires after the elements are out of the page (`removed[0].isConnected` is `false`) and only when `limit` took
something out. An item that `key` replaced, or a duplicate inside one response, is not in `removed`, and the event is not
cancelable.

```js
document.addEventListener('ypy:list:pruned', evt => {
  console.log(evt.detail.removed.length, 'items dropped');
});
```

## A complete example: a polled feed

The page asks for news every 3 seconds. The server sends the latest items, whether or not the page already has them;
`key` makes the repeated ones replace the old copy, and `limit` keeps the list short.

```html
<ul id="events" class="divide-y"
    ypy-get="/events" ypy-trigger="load, every 3s"
    ypy-swap="append key:data-id limit:50">
</ul>
```

```http
HTTP/1.1 200 OK
Content-Type: text/html

<li data-id="41" class="py-2">Deploy finished (edited)</li>
<li data-id="43" class="py-2">New comment</li>
```

If `41` is on the page, it is replaced where it stands; `43` is added at the end; and if the list now has more than 50
items, the first ones go. The server only writes the items.

For a feed that is polled, see also [Smart polling](poll.md).

## The same rules in a stream

A [stream](stream.md) can send the same behavior without this extension. The event `ypy-patch-elements` takes `key` and
`limit` for the modes `append` and `prepend`, with the rules above: the key is applied before the limit, the same key
twice in one patch keeps the last, the oldest items go.

```text
event: ypy-patch-elements
data: mode append
data: selector #events
data: key data-id
data: limit 50
data: elements <li data-id="44">Another comment</li>
```

Differences from the swap modifiers:

- The stream client has the rules built in; `ypy-list` is not needed for it.
- `key` without a name means `id`, as in the swap. A `limit` that is not a whole number of 1 or more, or a `key` that is
  not an attribute name, is an error in a stream: the client reports it and applies nothing. In the swap modifier
  they are ignored.
- `key` and `limit` are ignored in any other mode.
- A stream does not fire `ypy:list:pruned`.

## How it relates to `ypy-upsert`

[`ypy-upsert`](upsert.md) is a swap style of its own that looks an item up by `id` anywhere in the document and
can keep the list sorted. `ypy-list` keeps `append` and `prepend` and works inside the list only, by any attribute.
Use `upsert` for a sorted list, `append`/`prepend` with `key` and `limit` for a feed in arrival order.

## Pitfalls

- Every child is an item. A "no results yet" row or a title inside the list counts toward `limit` and can be pruned.
  Keep them outside the list, or give the feed its own container.
- `limit` removes from the start for `append` and from the end for `prepend`. Choose the style so that what you want to
  lose is on the far side.
- An item without the key attribute is always added, so a response that repeats such an item repeats it on the page.
- `key` reads an attribute, not a property. The server must write `data-id="41"` (or the attribute you name) in the
  markup it sends.
- Other swap styles ignore both modifiers (`innerHTML`, `outerHTML`, ...): `innerHTML key:id` replaces the content as
  it always did.
- A pruned item is gone for good. If the user could still be reading it, raise the limit.
- The extension is needed for swaps. `append key:id limit:50` on a page that did not load `ypy-list` appends without
  keys or limits (a stream is the exception, see above).
