---
order: 70
title: Inline validation
description: Check an email address on the server when the visitor leaves the field, show the message under it, and check the whole form again on submit.
---

# Inline validation

The field asks the server whether its value is acceptable as soon as the visitor leaves it. The message appears under
the field, in place, without touching the rest of the form.

## Try it

<!-- demo:inline-validation -->

Type `admin@example.com` and leave the field: that address is taken, so the server answers `422` with a message. Fix it
and leave again to see the confirmation. The form is checked again when you press **Create account**.

## The page

```html
<form id="signup" ypy-post="/signup" ypy-target="this" ypy-swap="outerHTML" novalidate>
  <label for="signup-email">Email</label>
  <input id="signup-email" type="email" name="email" aria-describedby="email-msg"
         ypy-post="/signup/email" ypy-trigger="change"
         ypy-target="#email-msg" ypy-swap="outerHTML"
         ypy-error-target="#email-msg" ypy-error-swap="outerHTML">
  <p id="email-msg">Try admin@example.com, which is taken.</p>
  <button>Create account</button>
</form>
```

## What the server answers

`POST /signup/email` receives the whole form, so the field's value arrives as `email=…`, and answers with the message
element only:

| Case | Status | Body |
|---|---|---|
| Free and well formed | `200` | `<p id="email-msg">marina@company.example is available.</p>` |
| Empty | `422` | `<p id="email-msg">Enter your email address.</p>` |
| Malformed | `422` | `<p id="email-msg">Use an address like name@company.com.</p>` |
| Taken | `422` | `<p id="email-msg">This address already has an account.</p>` |

`POST /signup` runs the same checks. It answers `200` with the form and a confirmation, or `422` with the form, the
value kept, and the message in the same place. Both are fragments with `Vary: YPY-Request`.

## How it works

- `ypy-trigger="change"` on the field starts the request when the visitor leaves it with a new value.
- A `422` normally goes into the form of the source element. The field is in a form, so `ypy-error-target` and
  `ypy-error-swap` send the error to the message element instead.
- The same rule answers the field and the form, so what the field accepted is what the form accepts.
- `aria-describedby` points the field at the message, and `aria-invalid="true"` on the redrawn field marks the
  failure for screen readers.
- `novalidate` turns the browser's own check off so the server's message is the one the visitor reads. Leave it out to
  let the browser check first.

## Variations

- The [`ypy-validate`](../extensions/validate.md) extension does this from one attribute: `ypy-validate="/signup/email"`
  on the field.
- Check while typing with `ypy-trigger="input changed delay:400ms"`.

## See also

- [Click to edit](click-to-edit.md)
- [Reset user input](reset-user-input.md)
- [Validation](../index.md#validation)
