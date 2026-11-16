---
order: 20
title: Bulk update
description: Tick rows in a table, press Activate or Deactivate, and one request updates every ticked row and shows a message, with no page reload.
---

# Bulk update

A table of people with a checkbox on each row. Tick some rows, press **Activate** or **Deactivate**, and one request
updates the table and says what it did.

## Try it

<!-- demo:bulk-update -->

Press a button with nothing ticked: the server answers `422` and the form comes back with a message instead of changing
anything.

## The page

```html
<form id="team" ypy-put="/team/status" ypy-target="this" ypy-swap="outerHTML">
  <table>
    <thead>
      <tr><th></th><th>Name</th><th>Email</th><th>Status</th></tr>
    </thead>
    <tbody>
      <tr>
        <td><input type="checkbox" name="ids" value="1" aria-label="Select Marina Duarte"></td>
        <td>Marina Duarte</td><td>marina.duarte@example.com</td><td>Active</td>
      </tr>
      <tr>
        <td><input type="checkbox" name="ids" value="3" aria-label="Select Helena Prado"></td>
        <td>Helena Prado</td><td>helena.prado@example.com</td><td>Inactive</td>
      </tr>
    </tbody>
  </table>
  <button name="status" value="active">Activate</button>
  <button name="status" value="inactive">Deactivate</button>
</form>
```

## What the server answers

The request is `PUT /team/status` with the body `ids=1&ids=3&status=active`. The pressed button sends its own `name` and
`value`, so one form serves both actions.

| Case | Status | Body |
|---|---|---|
| At least one row ticked | `200` | The whole form again, rows in their new state, and `<p role="status">2 people activated.</p>` |
| Nothing ticked | `422` | The same form, with `<p role="alert">Tick at least one person first.</p>` |

Both are fragments with `Vary: YPY-Request`.

## How it works

- The form is the request: `ypy-put` sends every checked box and the pressed button. A form triggers on `submit`, so no
  `ypy-trigger` is needed.
- The answer replaces the form (`ypy-target="this"`, `outerHTML`), so the checkboxes come back unticked and the message
  sits where the buttons are.
- The server decides the new state of every row, so the table can never disagree with it.
- The message has `role="status"`, which screen readers announce without moving focus.

## Variations

- Morph instead of replace: leave out `ypy-swap` and the default `innerMorph` keeps the ticks the visitor made while the
  request ran.
- Send the ids as JSON with an [extension](../extensions/index.md) if your API expects it.
- Add `ypy-confirm="Deactivate the selected people?"` to the destructive button.

## See also

- [Delete row](delete-row.md)
- [Updating other content](updating-other-content.md)
- [Forms](../guides/forms.md)
