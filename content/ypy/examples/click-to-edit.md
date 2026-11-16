---
order: 10
title: Click to edit
description: Show a contact as text, swap in an edit form on click, and swap the text back after Save or Cancel, with no page reload and no script of your own.
---

# Click to edit

A contact is shown as plain text. A click on **Edit** swaps in a form, and **Save** swaps the updated text back. The page
does not reload and nothing outside the block changes.

## Try it

<!-- demo:click-to-edit -->

Clear the name, or type an address with no dot in the domain, and press **Save**: the server answers `422` and the form
comes back with the messages. Fix it and the text returns.

## The page

```html
<div id="contact">
  <dl>
    <dt>Name</dt><dd>Marina Duarte</dd>
    <dt>Email</dt><dd>marina.duarte@example.com</dd>
  </dl>
  <button ypy-get="/contacts/1/edit" ypy-target="#contact" ypy-swap="outerHTML">Edit</button>
</div>
```

## What the server answers

`GET /contacts/1/edit` answers `200` with a form that has the same `id` as the block it replaces:

```html
<form id="contact" ypy-put="/contacts/1" ypy-target="this" ypy-swap="outerHTML">
  <input type="hidden" name="_token" value="…">
  <label>Name <input name="name" value="Marina Duarte"></label>
  <label>Email <input type="email" name="email" value="marina.duarte@example.com"></label>
  <button>Save</button>
  <button type="button" ypy-get="/contacts/1" ypy-target="#contact" ypy-swap="outerHTML">Cancel</button>
</form>
```

`PUT /contacts/1` answers one of two things:

| Case | Status | Body |
|---|---|---|
| Valid | `200` | The read-only block again, with the saved values |
| Name empty or email malformed | `422` | The same form, with each field's message under it and `aria-invalid="true"` |

Every answer is a fragment, not a page, and carries `Vary: YPY-Request` so a cache never hands it to a visitor who
asked for the whole page.

## How it works

- `ypy-get` on the button asks for the form. `ypy-target="#contact"` names the block to replace and
  `ypy-swap="outerHTML"` replaces the element itself. Without `ypy-swap` the default is `innerMorph`, which keeps the
  block and morphs its children.
- The form carries `ypy-put`. A form sends its fields on `submit`, so no `ypy-trigger` is needed. `ypy-target="this"`
  points at the form, and `outerHTML` swaps it for the read-only block.
- The form has the same `id` as the block, so the next click on **Edit** finds it again.
- **Cancel** is `type="button"`, so it does not submit. It asks for the read-only block with `ypy-get`.
- A `422` is swapped like any other answer, so the form comes back with its errors. See
  [Forms](../index.md#validation) for the convention.
- `_token` is the CSRF field of a form that changes data; see [CSRF](../index.md#csrf).

## Variations

- Use `ypy-patch` instead of `ypy-put` when the edit is a partial update.
- To put an error next to a field instead of redrawing the form, use
  [`ypy-error-target`](../index.md#status-codes-and-errors).
- Add `ypy-confirm="Discard changes?"` to **Cancel** when the form is long.

## See also

- [Edit row](edit-row.md), the same idea inside a table row
- [Inline validation](inline-validation.md)
- [Attributes](../reference.md#core-attributes)
