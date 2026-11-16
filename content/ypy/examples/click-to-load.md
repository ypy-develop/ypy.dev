---
order: 30
title: Click to load
description: A table shows five orders and a Load more button. Each click adds the next five rows and moves the button down, until the list ends.
---

# Click to load

A table shows the first five orders. **Load more orders** asks for the next five, adds them to the end of the table and
moves itself below them. On the last page it is replaced by a line that says the list is complete.

## Try it

<!-- demo:click-to-load -->

## The page

```html
<table>
  <thead>
    <tr><th>Order</th><th>Customer</th><th>Total</th><th>Status</th></tr>
  </thead>
  <tbody>
    <tr><td>#2041</td><td>Harbor Coffee Roasters</td><td>$1,280.00</td><td>Paid</td></tr>
    <!-- … four more rows … -->
    <tr id="more-orders">
      <td colspan="4">
        <button ypy-get="/orders?page=2" ypy-target="#more-orders" ypy-swap="outerHTML">Load more orders</button>
        <span>5 of 14</span>
      </td>
    </tr>
  </tbody>
</table>
```

## What the server answers

`GET /orders?page=2` answers `200` with the next rows and a new button row, as one fragment:

```html
<tr><td>#2036</td><td>Meridian Cycles</td><td>$1,115.25</td><td>Paid</td></tr>
<!-- … four more rows … -->
<tr id="more-orders">
  <td colspan="4">
    <button ypy-get="/orders?page=3" ypy-target="#more-orders" ypy-swap="outerHTML">Load more orders</button>
    <span>10 of 14</span>
  </td>
</tr>
```

The last page ends with a row that has the same `id` and no button: `All 14 orders are shown.` The answer has
`Vary: YPY-Request`.

## How it works

- The button replaces its own row (`ypy-target="#more-orders"`, `outerHTML`) with several elements: the new rows and
  the next button row. Each answer carries the address of the page after it, so the server holds the position, not the
  page.
- The rows are appended because the replaced row is the last one in the table body; nothing else moves.
- Without JavaScript the button does nothing. For a version that works without script, make it a link to
  `/orders?page=2` and add `ypy-boost`, so the same address answers a whole page to a browser and a fragment to YPY.

## Variations

- Fire the request when the button scrolls into view: see [Infinite scroll](infinite-scroll.md).
- Show the count with `role="status"` so a screen reader hears that rows arrived.
- Keep the button where it is and append with `ypy-swap="beforeend"` on `tbody` if the list has no footer row.

## See also

- [Infinite scroll](infinite-scroll.md)
- [Active search](active-search.md)
- [Swap and morph](../index.md#swapping)
