---
title: "Field validation on the server: the ypy-validate extension"
description: "Sends the form's values to the server when the user leaves a field and shows the answer next to it, without submitting the form."
order: 480
---
# Field validation on the server: the `ypy-validate` extension

Some rules only the server knows: this email is taken, this code has expired. `ypy-validate` asks the server about them
when the user leaves a field and shows the answer next to the field, before the user presses the button. The form's values
go to the server, but the form is not submitted.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-validate.min.js"></script>

<form ypy-post="/signup" ypy-target="#result" ypy-validate-fields="/signup/validate">
  <label for="email">Email</label>
  <input id="email" name="email" type="email">
  <small id="email-error"></small>

  <label for="name">Name</label>
  <input id="name" name="name">
  <small id="name-error"></small>

  <button>Create account</button>
</form>
```

Leave the email field after changing it and a `POST /signup/validate` goes out. If the server answers `422` with
`<span>This email is taken</span>`, that markup lands in `#email-error`, and the field gets `aria-invalid="true"` and
`aria-describedby="email-error"`. The focus stays where the user is. Without the extension the attributes below do
nothing.

## Attributes

| Attribute | On | Value | Meaning |
|---|---|---|---|
| `ypy-validate` | a field | a URL, or nothing | check this field; a bare attribute asks the action of the form |
| `ypy-validate-fields` | a form, or any element around fields | a URL, or nothing | check every field inside; a field with `ypy-validate="false"` is left out, and one with its own URL uses it |
| `ypy-validate-trigger` | a field, or the element with `ypy-validate-fields` | a trigger specification | when to check; default `change delay:200ms` |
| `ypy-validate-batch` | a field, the element with `ypy-validate-fields`, or the form (nearest wins) | a time, nothing, or `false` | send the fields asked within that window in one request; bare means `50ms`; off by default |
| `ypy-error-target` | the field | a selector | where the answer goes |
| `ypy-error-swap` | the field | a swap style | how it goes in; default `innerHTML` |
| `ypy-error-focus` | the field | a selector | where focus goes after a `422`; default: it stays |

The action of the form is its `ypy-post` (or `ypy-action`), else its `action` attribute, else the current path.

The core already uses `ypy-validate="true"` and `="false"` to switch the browser's own constraint check before a request, and
this extension ignores those two values. Any other value, including none, is a server check. On a form write
`ypy-validate-fields`, because a form whose `ypy-validate` is a URL would lose the browser check of its own submit.

Other triggers use the syntax of `ypy-trigger`. Fields are wired when their subtree is processed, so fields that arrive in a
swap work at once.

```html
<input name="handle" ypy-validate="/handles/check" ypy-validate-trigger="input delay:400ms">
<input name="code" ypy-validate="/codes/check" ypy-validate-trigger="blur">
```

## The request

A `POST` to the URL with every value of the form, so a rule can compare two fields, plus the header `YPY-Validate` with the
name of the field being checked (the `id` when the field has no `name`). It carries the usual `YPY-*` request headers, and its
`YPY-Source` is the field.

The server must look at `YPY-Validate` before it does anything else. The request often goes to the handler of the form
itself, and a handler that does not know the header would run the whole action: create the account, send the email.

```http
POST /signup/validate
YPY-Validate: email
Content-Type: application/x-www-form-urlencoded

email=ana%40example.com&name=Ana
```

## The answer

| Status | Meaning | The page |
|---|---|---|
| `422` | the field is invalid | the body is swapped into the place; the field gets `aria-invalid="true"` and `aria-describedby` |
| `204`, or an empty `200` | valid | the place is emptied and `aria-invalid` is removed |
| `200` with a body | valid, with something to say ("Available") | the body is swapped into the place and `aria-invalid` is removed |
| anything else, or the network fails | no answer | nothing changes; the form validates again when submitted |

The place follows the [error convention](../index.md#requests-and-responses): `ypy-error-target` on the field, with every
selector form. Without it, `#<id>-error` (the field's `id`, else its `name`). When the page has no such element, the
extension creates a `<div class="ypy-field-error" aria-live="polite">` right after the field. A `YPY-Retarget` or
`YPY-Reswap` header and `ypy-status:422` still win. A `422` for a field never goes into the form, so the other fields, what
the user typed and the focus stay.

## Batches

By default every field is a request of its own, so a paste into a whole form sends one request per field. With
`ypy-validate-batch`, the fields of one form that ask the same URL within a short window go in a single request.

```html
<form ypy-post="/signup" ypy-validate-fields="/signup/validate" ypy-validate-batch="80ms">
  <input id="email" name="email" type="email"><small id="email-error"></small>
  <input id="username" name="username"><small id="username-error"></small>
  <button>Create account</button>
</form>
```

`ypy.config.validate = { batch: '80ms' }` turns it on for every form, and `ypy-validate-batch="false"` turns one off. The
window starts at the first field that asks, and the fields go together in one `POST` with every value of the form. The header
then holds the names separated by commas, `YPY-Validate: email,username`, and a single field in a window is sent as without
batching. Fields of two forms or two URLs never share a request, and a field that left the page during the window is not
sent.

The answer is `422` when any field in the batch is invalid and any `2xx` when none is. The message of each invalid field
comes as an out-of-band element whose `id` is the `id` of that field's place. A field with no such element is valid, and its
place is emptied.

```http
HTTP/1.1 422 Unprocessable Content
Content-Type: text/html; charset=utf-8

<small id="email-error" ypy-swap-oob="innerHTML">This email is taken</small>
```

Batching is off by default because the server must read the list: a handler that only knows one name would see
`email,username` and find no rule for it. The same handler answers a single field, so a page can turn batching on and off
without a change on the server.

## Concurrency

A field has at most one check in flight, so a slow answer never overwrites a fresh one. Different fields are checked
independently. A request sent by the form cancels the checks of its fields that are still running, so a late "invalid" does
not appear after the form was accepted. A check is not sent if the field left the page while the debounce waited. A
request that carries `YPY-Validate` does not expire the [`ypy-cache`](cache.md).

## Accessibility

The answer region has `aria-live="polite"` when the extension creates it, so give your own region the same attribute.
`aria-invalid` and `aria-describedby` follow each check. A check never moves the focus, because the user is moving through the
form, and `ypy-error-focus` changes that deliberately for one field. The check is a hint, and the form must still validate on
submit.

## Pitfalls

- A handler that does not read `YPY-Validate` runs the real action on every check. Test it with a request that carries the
  header.
- With batching on, `YPY-Validate` can hold commas. Split it and do not compare it with one field name.
- In a batch, a `422` without an out-of-band element for a field means that field is valid. A `422` with a plain-text body
  empties every place and marks no field.
- A form with `enctype="multipart/form-data"` sends its files with every check. Give the field
  `ypy-encoding="application/x-www-form-urlencoded"` to send the values as text only.
