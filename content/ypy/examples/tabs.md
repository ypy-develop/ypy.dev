---
order: 180
title: Tabs
description: Tabs that ask the server for one panel at a time, with the strip and the panel updated by a single answer so the selected tab is never stored twice.
---

# Tabs

Each tab is a button. A click asks the server for its panel, and the answer also redraws the tab strip with the right tab
selected. The server is the only place that knows which tab is current.

## Try it

<!-- demo:tabs -->

Click a tab. The strip is redrawn with the answer, and the tab you clicked keeps the focus.

## The page

```html
<div role="tablist" id="tab-strip" data-variant="line" ypy-target:inherited="#tab-panel" ypy-swap:inherited="outerHTML">
  <button role="tab" id="tab-1" aria-selected="true"  aria-controls="tab-panel" ypy-get="/tabs/1">Profile</button>
  <button role="tab" id="tab-2" aria-selected="false" aria-controls="tab-panel" ypy-get="/tabs/2">Billing</button>
  <button role="tab" id="tab-3" aria-selected="false" aria-controls="tab-panel" ypy-get="/tabs/3">Team</button>
</div>
<div role="tabpanel" id="tab-panel" aria-labelledby="tab-1">
  <p>Marina Duarte, marina.duarte@example.com. Member since March 2024.</p>
</div>
```

The two attributes on the strip end in `:inherited`, so every tab below gets the same target and swap style.

## What the server answers

`GET /tabs/2` answers `200` with the new panel and, out of band, the strip with `aria-selected` moved:

```html
<div role="tabpanel" id="tab-panel" aria-labelledby="tab-2">
  <p>Plan: Team, 12 seats. Next invoice on 1 December, $240.00.</p>
</div>

<div role="tablist" id="tab-strip" data-variant="line" ypy-swap-oob="true" ypy-target:inherited="#tab-panel" ypy-swap:inherited="outerHTML">
  <button role="tab" id="tab-1" aria-selected="false" aria-controls="tab-panel" ypy-get="/tabs/1">Profile</button>
  <button role="tab" id="tab-2" aria-selected="true"  aria-controls="tab-panel" ypy-get="/tabs/2">Billing</button>
  <button role="tab" id="tab-3" aria-selected="false" aria-controls="tab-panel" ypy-get="/tabs/3">Team</button>
</div>
```

An unknown tab answers `404`. Every answer is a fragment with `Vary: YPY-Request`.

## How it works

- The request's own target is the panel (`ypy-target:inherited` on the strip). The first element of the answer replaces
  it.
- `ypy-swap-oob="true"` on the strip swaps it, as a morph, into the element with the same `id`, so the tab that was
  clicked keeps its focus and only `aria-selected` changes.
- The answer repeats the strip, which keeps the selection in one place. The page never edits `aria-selected` by itself.
- `aria-selected` and `aria-controls` give screen readers the state; the CSS styles `[aria-selected="true"]`.

## Variations

- Make each tab a real link to a full page, `<a href="/account?tab=billing" ypy-get="/account?tab=billing">`, and add
  `ypy-push-url:inherited="true"` on the strip. The address changes with the tab, reload and sharing work, and without
  JavaScript each link still opens its page. The Back button then restores the earlier tab; see
  [history](../index.md#history).
- Load a panel only the first time it is used and keep the others in the page, hidden.
- Add `ypy-indicator="#tab-panel"` to dim the panel while it loads.

## See also

- [Updating other content](updating-other-content.md)
- [Attribute inheritance](../index.md#attribute-inheritance)
- [History](../index.md#history)
