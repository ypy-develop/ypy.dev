---
order: 60
title: Lazy loading
description: Draw the page at once with a placeholder, then load a slow report on its own request and swap it in when it is ready, with no script of your own.
---

# Lazy loading

A slow part of the page does not hold the rest back. The page ships a placeholder that asks for the real content as soon
as it loads, and the answer replaces the placeholder.

## Try it

<!-- demo:lazy-loading -->

The report takes about a second on the server. The skeleton is what the visitor sees until it arrives; **Load it
again** puts the placeholder back and starts the request again.

## The page

```html
<div id="report" ypy-get="/reports/weekly" ypy-trigger="load" ypy-swap="outerHTML">
  <p role="status">Loading the weekly report…</p>
  <div class="skeleton"></div>
</div>
```

## What the server answers

`GET /reports/weekly` answers `200` after the work is done, with a fragment that has the same `id`:

```html
<div id="report">
  <dl>
    <dt>Orders this week</dt><dd>128</dd>
    <dt>Revenue</dt><dd>$14,320.50</dd>
    <dt>Average order</dt><dd>$111.88</dd>
    <dt>Refunds</dt><dd>3</dd>
  </dl>
</div>
```

The answer carries `Vary: YPY-Request`. The server stops working if the visitor leaves the page before it is done.

## How it works

- `ypy-trigger="load"` fires the request when the element is processed, which is when the page, or the swap that
  brought it, is ready.
- While the request runs YPY sets `aria-busy="true"` on the element and removes it with the swap, so assistive
  technology knows the region is not final.
- The placeholder is sized like the result, so nothing below it jumps when the report arrives. Draw it with the same
  height as the real content.
- Without JavaScript the placeholder stays. For content that matters without script, render it in the page and use lazy
  loading only for extras.

## Variations

- Load only when the placeholder is seen: `ypy-trigger="revealed"`, or `ypy-trigger="intersect once"` to set a threshold.
  The [`ypy-lazy`](../extensions/lazy.md) extension loads a placeholder only when it nears the viewport.
- Load after a pause: `ypy-trigger="load delay:2s"`.
- Replace the placeholder with a spinner by adding `ypy-indicator` and the Basecoat `spinner` class.

## See also

- [Infinite scroll](infinite-scroll.md)
- [Click to load](click-to-load.md)
- [Progress bar](progress-bar.md)
