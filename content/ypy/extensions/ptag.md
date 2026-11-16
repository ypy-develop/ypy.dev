---
order: 340
title: "Skip unchanged polls: the ypy-ptag extension"
description: "The server tags each polled fragment and the client sends the tag back, so the server can answer 304 when nothing changed and the swap is skipped."
---

# Skip unchanged polls: the `ypy-ptag` extension

An element that polls gets the same HTML again and again, even when nothing changed. With `ypy-ptag` the server gives
each answer a tag. The client sends the tag back on the next request, and the server answers `304` when the content is
still the one that tag stands for. YPY does not swap on `304`, so the page, the focus and the scroll stay as they are.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-ptag.js"></script>
```

The extension is not part of `ypy.max`.

## Use it

```html
<div ypy-get="/news" ypy-trigger="every 3s" ypy-ptag="start">Latest news...</div>
```

1. The first request carries `YPY-PTag: start` (the value of the attribute). Without an initial value no header is
   sent the first time.
2. The server answers with `YPY-PTag: <tag>` and the HTML.
3. YPY stores the tag on the element and sends it with each next request.
4. When the content still has that tag, the server answers `304` with no body.

The tag belongs to the element, not to the page: two polling elements keep two tags.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-ptag` | string | the first tag the element sends |

## Headers

| Header | Direction | Value |
|---|---|---|
| `YPY-PTag` | request | the current tag of the element |
| `YPY-PTag` | response | a new tag, kept by the element and sent back next time |

The tag is opaque: a version number, a hash of the content, an update time. Only the server reads it.

## Server

The server keeps a tag that changes when the content changes. When the request carries `YPY-PTag` with the current tag,
answer `304` with no body. Otherwise send the fragment with `YPY-PTag` set to the current tag.

```http
GET /news
YPY-PTag: 42

HTTP/1.1 304 Not Modified
```

## Notes

- `304` and `204` are in the core's `noSwap` list, so nothing is swapped.
- It works with any request, not only polls, but polls are where the saving is.
- Pair it with [`ypy-poll`](poll.md), which slows polling in a hidden tab and lets the server end it.

## See also

- [Smart polling](poll.md)
- [Polling example](../examples/polling.md)
