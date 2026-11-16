---
order: 80
title: Infinite scroll
description: Fetch the next page of rows when the last row scrolls into view. Each answer carries its own next trigger, so the server decides when to stop.
---

# Infinite scroll

The last row of a table asks for the next page as soon as it comes into view. The answer adds more rows after it, and
the new last row asks for the page after that. When no pages are left, the answer ends without a trigger.

## Try it

<!-- demo:infinite-scroll -->

Scroll inside the box. Each page holds ten rows, and the last page has no trigger, so the requests stop.

## The page

```html
<table>
  <thead><tr><th>Name</th><th>Email</th></tr></thead>
  <tbody>
    <tr><td>Marina Duarte</td><td>marina.duarte@example.com</td></tr>
    <!-- … eight more rows … -->
    <tr ypy-get="/contacts?page=2" ypy-trigger="revealed" ypy-swap="afterend">
      <td>Caio Nogueira</td><td>caio.nogueira@example.com</td>
    </tr>
  </tbody>
</table>
```

## What the server answers

`GET /contacts?page=2` answers `200` with the rows of that page. The last of them carries the next trigger:

```html
<tr><td>Beatriz Lemos</td><td>beatriz.lemos@example.com</td></tr>
<!-- … eight more rows … -->
<tr ypy-get="/contacts?page=3" ypy-trigger="revealed" ypy-swap="afterend">
  <td>Rafael Quintana</td><td>rafael.quintana@example.com</td>
</tr>
```

| Case | Status | Body |
|---|---|---|
| A page with more after it | `200` | The rows; the last row has `ypy-get` for the next page |
| The last page | `200` | The rows, none of them with `ypy-get` |
| A page past the end | `200` | Empty |

Every answer is a fragment of `<tr>` elements with `Vary: YPY-Request`.

## How it works

- `ypy-trigger="revealed"` fires once, when the row scrolls into the viewport. The observer then disconnects, so a fast
  scroll cannot ask for the same page twice.
- `ypy-swap="afterend"` puts the answer after the row that asked. The old last row stays where it is and keeps its
  attributes, but nothing watches it any more.
- The new last row is the one that carries `ypy-get`, so the chain continues for as long as the server keeps handing
  out a next address.
- The server decides when to stop: a last page whose rows have no `ypy-get` ends the chain.
- Rows are parsed inside a `<template>`, so `<tr>` elements arrive intact.

## Variations

- Start loading before the visitor reaches the end: put the trigger on a row some lines above the last one and use
  `ypy-trigger="intersect once"`, or `intersect root:#scroller threshold:0` when the rows scroll inside a box with that
  `id`.
- Visitors without JavaScript see only the first page. Close the table with a normal link to the next page, so the
  content stays reachable.
- The [`ypy-list`](../extensions/list.md) extension keeps a growing list bounded with `limit`, so a long session does
  not hold thousands of rows.

## See also

- [Click to load](click-to-load.md), the same pattern with a button
- [Lazy loading](lazy-loading.md)
- [Triggers](../index.md#triggers)
