---
order: 50
title: Edit row
description: A table row turns into inputs on Edit and back into text on Save or Cancel, and the server checks the values before it keeps them.
---

# Edit row

A click on **Edit** turns one row of a team table into inputs. **Save** sends the row and swaps it back into text;
**Cancel** swaps it back untouched. The other rows never move.

## Try it

<!-- demo:edit-row -->

Empty the name and press **Save**: the server answers `422` and the row comes back with a message.

## The page

```html
<table>
  <thead><tr><th>Name</th><th>Role</th><th></th></tr></thead>
  <tbody>
    <tr id="member-1">
      <td>Marina Duarte</td>
      <td>Designer</td>
      <td>
        <button ypy-get="/team/1/edit" ypy-target="#member-1" ypy-swap="outerHTML">Edit</button>
      </td>
    </tr>
  </tbody>
</table>
```

## What the server answers

`GET /team/1/edit` answers `200` with the row as inputs, with the same `id`:

```html
<tr id="member-1">
  <td>
    <input type="hidden" name="_token" value="…">
    <input name="name" value="Marina Duarte" aria-label="Name">
  </td>
  <td><input name="role" value="Designer" aria-label="Role"></td>
  <td>
    <button ypy-put="/team/1" ypy-include="closest tr" ypy-target="#member-1" ypy-swap="outerHTML">Save</button>
    <button ypy-get="/team/1" ypy-target="#member-1" ypy-swap="outerHTML">Cancel</button>
  </td>
</tr>
```

`PUT /team/1` answers:

| Case | Status | Body |
|---|---|---|
| Valid | `200` | The row as text, with the saved values |
| Name or role empty | `422` | The edit row again, with `<p>Name and role cannot be empty.</p>` under the role field |

`GET /team/1` answers `200` with the row as text, which is what **Cancel** swaps in. Every answer is a `<tr>` fragment
with `Vary: YPY-Request`.

## How it works

- A table row cannot hold a `<form>` and still be a row, so the inputs sit directly in the cells and the **Save**
  button names them with `ypy-include="closest tr"`. YPY sends their values, including the hidden token, with the
  request.
- Both buttons replace the whole row (`ypy-target="#member-1"`, `outerHTML`). The `id` is the same in all three
  versions of the row, so the next click finds it again.
- Because the request is not a form, a `422` is swapped where the author pointed, which is the row itself.
- The response is parsed as markup that may start with a `<tr>`; YPY reads it inside a `<template>`, so the row keeps
  its cells.

## Variations

- Edit a whole table at once by wrapping it in a `<form>` and sending one `PUT`, as in [Bulk update](bulk-update.md).
- Let **Enter** save and **Escape** cancel with [Keyboard shortcuts](keyboard-shortcuts.md).

## See also

- [Click to edit](click-to-edit.md)
- [Delete row](delete-row.md)
- [Inline validation](inline-validation.md)
