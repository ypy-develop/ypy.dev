---
title: "Accessibility in YPY"
description: "What YPY does for screen reader and keyboard users after a swap: aria-busy, a live region, focus on the new heading, scroll, and what you still write."
order: 80
---
# Accessibility in YPY

A swap breaks four things for people who use a screen reader or a keyboard. YPY handles all four by default.

| What breaks | What YPY does | Reference |
|---|---|---|
| The target changes and nothing says it is loading | `aria-busy="true"` on the target while the request runs | [`ypy-busy`](../reference/attributes/ypy-busy.md) |
| New content appears and nobody hears it | text in a managed `aria-live` region | [`ypy-announce`](../reference/attributes/ypy-announce.md) |
| A navigation leaves the focus on `<body>` | the focus goes to the new page's heading | [`ypy-focus`](../reference/attributes/ypy-focus.md) |
| The page jumps, or starts at the top again | the scroll position is kept, reset or restored | [`ypy-scroll`](../reference/attributes/ypy-scroll.md) |

## A boosted link, end to end

```html
<body ypy-boost="true" class="bg-white">
  <nav class="flex gap-4 p-4"><a href="/orders" class="underline">Orders</a></nav>
  <main class="p-6">
    <h1 class="text-2xl focus-visible:outline">Dashboard</h1>
    <ul id="list" class="aria-busy:opacity-50">...</ul>
  </main>
</body>
```

Click "Orders":

1. `aria-busy="true"` is set on the target (`<body>` for a boosted link) until the response is in.
2. The new page is swapped in. The focus is given back to the element that had it, if it has an `id`.
3. The live region says the new title (a boosted navigation announces it by default).
4. The focus moves to the first `<h1>` of the new `<main>`. The heading gets `tabindex="-1"` until it loses
   the focus.
5. `show:top` of the boosted request puts the page at the top.

## What you still do

- Give every page one `<h1>` inside `<main>` (or inside the element marked `ypy-main`). That is what receives the
  focus.
- Write a visible focus style (`:focus-visible`). YPY adds no CSS.
- Put `ypy-announce` on actions that change the page without a navigation, or set `config.announce.messages`.
- On a page that polls, add `ypy-announce="none"` and `ypy-busy="false"` to the polling element.
- A sign-in page with `autofocus` keeps its field: the focus step does not take it away.

## Choosing, per element

```html
<a href="/report" ypy-focus="none">Download-style navigation: keep the focus</a>
<button ypy-post="/items" ypy-target="#list" ypy-focus="#flash" ypy-announce="Item added">Add</button>
<button ypy-get="/rows?page=2" ypy-target="#rows" ypy-swap="outerHTML" ypy-scroll="keep">More</button>
```

Everything accepts `:inherited` for a whole subtree, for example `ypy-focus:inherited="none"` on a toolbar.

## Order of events at the end of a swap

1. the swap gives the focus back (`ypy-swap` styles that replace in place)
2. `[autofocus]` of the new content
3. `ypy:after:swap`, the document title
4. `ypy-focus`
5. `ypy-scroll`
6. at the end of the request: `aria-busy` removed, then the announcement

## Testing

Test with the keyboard alone and with a screen reader (VoiceOver, NVDA). After each navigation the new heading should be
read, Tab should continue from it, and Shift+Tab should reach the navigation. The test suite covers the focus, the
`tabindex` clean-up and the scroll (`test/tests/attributes/ypy-focus.js`,
`ypy-scroll.js`, `end2end/a11y-aria.js`).
