---
order: 40
title: Delete row
description: Ask for confirmation, send a DELETE request for one table row, and remove that row from the page when the server says it is gone.
---

# Delete row

Each row of an invoice table has a **Delete** button. A click asks for confirmation, sends `DELETE` to the server, and
removes the row when the server answers.

## Try it

<!-- demo:delete-row -->

Delete a few rows, reload this page and they stay deleted, for you only. **Restore the rows** puts them back.

## The page

```html
<table>
  <thead><tr><th>Customer</th><th>Amount</th><th></th></tr></thead>
  <tbody>
    <tr>
      <td>Harbor Coffee Roasters</td>
      <td>$1,280.00</td>
      <td>
        <form ypy-delete="/invoices/1" ypy-headers='{"X-CSRF-Token": "…"}'
              ypy-confirm="Delete the invoice for Harbor Coffee Roasters?"
              ypy-target="closest tr" ypy-swap="delete">
          <button>Delete</button>
        </form>
      </td>
    </tr>
  </tbody>
</table>
```

## What the server answers

`DELETE /invoices/1`, with the header `X-CSRF-Token`, answers `200` with an empty body and `Vary: YPY-Request`.

| Case | Status | Body |
|---|---|---|
| Deleted | `200` | Empty |
| Already gone or unknown id | `404` | Empty |
| No valid CSRF token | `419` | The framework's refusal; the row stays |

Use `200`, not `204`: YPY does not swap a `204`, and a swap is what removes the row.

## How it works

- `ypy-confirm` asks before anything is sent. Cancel the dialog and no request leaves the page.
- `ypy-target="closest tr"` reaches the row from the button, so the rows need no ids.
- `ypy-swap="delete"` removes the target whatever the answer holds, so an empty `200` is enough.
- A `DELETE` has no body, so a hidden `_token` field cannot reach the server. `ypy-headers` sends the token as the
  `X-CSRF-Token` header instead; the [`ypy-csrf`](../extensions/csrf.md) extension does the same for every request
  from one `<meta>` tag. See [CSRF](../index.md#csrf).
- An answer of `419` or `404` is an error, not a success, so the row stays on the page.

## Variations

- Fade the row out first with a CSS transition: `ypy-swap="delete swap:300ms"` waits 300 ms while the target carries
  the `ypy-swapping` class. See [Animations](animations.md).
- Put the confirmation in a dialog with the [`ypy-prompt`](../extensions/prompt.md) extension, or see
  [Confirm and prompt](confirm-and-prompt.md).

## See also

- [Edit row](edit-row.md)
- [Bulk update](bulk-update.md)
- [Confirm and prompt](confirm-and-prompt.md)
