---
order: 470
title: "Update or insert by id: the ypy-upsert extension"
description: "A swap style for live lists: an item whose id is on the page replaces the old one, a new id is inserted, with an optional sort."
---

# Update or insert by id: the `ypy-upsert` extension

A live list gets a mix of changed items and new ones. The `upsert` swap style handles both: an item whose `id` is
already on the page replaces the old one, and any other item is inserted. No separate request for "add" and "update".

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-upsert.js"></script>
```

The extension is part of `ypy.max`.

## Use it

As a swap style:

```html
<ul id="tasks" ypy-get="/tasks/changes" ypy-trigger="every 5s" ypy-swap="upsert">
  <li id="task-1">Write the page</li>
</ul>
```

The server answers with the items that changed:

```html
<li id="task-1">Write the page (done)</li>
<li id="task-2">Check the links</li>
```

`task-1` is replaced where it is. `task-2` is new, so it is added at the end.

## Modifiers

Add them after `upsert` in `ypy-swap`.

| Modifier | Meaning |
|---|---|
| `sort` | keep the new items in ascending order of their key |
| `sort:desc` | descending order |
| `prepend` | an item with no key goes at the start instead of the end |
| `key:<attr>` | the attribute to sort by instead of `id` |

An item with an `id` that is on the page is always replaced in place. For a new item the key is its `id`, or the
attribute named by the `key` attribute of the response tag below. Keys compare as text with numbers understood, so
`task-2` sorts before `task-10`.

## As a tag in the response

The server can also send the parts with a tag, so one response can upsert into several places:

```html
<ypy-upsert ypy-target="#tasks" key="data-priority" sort>
  <li id="task-2" data-priority="2">Check the links</li>
</ypy-upsert>
```

| Attribute of the tag | Meaning |
|---|---|
| `ypy-target` | the element that receives the items |
| `key` | the attribute to sort by (default `id`) |
| `sort` | ascending, or `sort="desc"` for descending |
| `prepend` | items with no key go first |

## Server

The server writes one `<ypy-partial>` per changed item, each with `ypy-target="#tasks"` and `ypy-swap="upsert"`, and the
items are replaced by `id` or inserted.

## Notes

- Only direct children of the fragment are looked at. Give each item its own `id`.
- `upsert` replaces the element, so state inside it (a typed value, a focus) is lost; see
  [`ypy-preserve`](preserve.md).
- For a list that must not grow forever, see [`ypy-list`](list.md), which adds `key:` and `limit:` to `append` and
  `prepend`.

## See also

- [Keyed and bounded lists](list.md)
- [Swap styles and morph](../index.md#swapping)
- [Live updates (example)](../examples/live-updates.md)
