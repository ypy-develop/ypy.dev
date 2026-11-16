---
title: "Show or hide part of a form by a field: the ypy-switch extension"
description: "Shows or hides part of a form from the value of another field, with no request. Hidden fields are disabled, so they are not submitted."
order: 410
---
# Show or hide part of a form by a field: the `ypy-switch` extension

A form often has a part that only matters for some answers: a company registration number when the person chose
"Company", a card block when the payment is "Card". `ypy-switch` shows or hides such a part according to the value of
another field, with no request. A hidden part has its fields disabled, so they are not submitted and a `required`
one does not block the submit.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-switch.js"></script>

<form ypy-post="/accounts">
  <select name="kind" class="rounded border p-2">
    <option value="person">Person</option>
    <option value="company">Company</option>
    <option value="ngo">NGO</option>
  </select>

  <div ypy-switch="kind=company|ngo">
    <label>Registration number <input name="registration" required></label>
  </div>

  <label><input type="checkbox" name="newsletter"> Newsletter</label>
  <div ypy-switch="newsletter"><input name="email" type="email" required></div>

  <button>Create</button>
</form>
```

The registration block shows when the person picks "Company" or "NGO"; the e-mail block shows when the box is checked.
While they are hidden their fields are disabled: a person who never opens them can submit the form.

The extension is part of `ypy.max`. Without it `ypy-switch` does nothing and the part stays visible.

## The attribute

`ypy-switch` goes on the part that is shown or hidden. Its value is a rule about one field. It is not inherited.

| Rule | The part is shown when |
|---|---|
| `kind=company` | the field `kind` has the value `company` |
| `kind=company\|ngo` | the field has one of these values (separate them with `\|`) |
| `kind!=person` | the field has none of these values (`!=` accepts a list too) |
| `newsletter` | the field has a value; for a checkbox or a radio, it is checked |

Spaces around the name, the `=` and the values are ignored. A rule that cannot be read (an empty attribute) leaves the
part shown.

### How the field is found

- By name, inside the `<form>` that holds the part. A part that is in no form looks in the whole page. A field in
  another form is not read.
- A radio group, or checkboxes with the same name, count as one field: only the checked ones give values.
- A `<select multiple>` has several values; the rule holds if any of them matches.
- A checkbox without a `value` attribute has the value `on`, so `newsletter=on` and `newsletter` are the same for it.
- With `!=`, a field with no value at all (an unchecked box) is "none of these values", so the part is shown.

## What a hidden part gets

- the `hidden` attribute;
- every control inside it (`input`, `select`, `textarea`, `button`, `fieldset`; or the part itself when it is one) is
  disabled.

When the part is shown again the fields that the extension disabled are enabled again. A field that the page had
disabled itself stays disabled. Nested parts are evaluated from the outside in: while an outer part is hidden, the
fields of the inner ones stay disabled, whatever the inner rule says.

If your CSS gives the part a `display` value, it can win over `hidden`. Add `[hidden] { display: none }` for it.

## When it updates

The parts are evaluated:

- when the page is processed, so the first state is right (the first evaluation fires no event);
- on every `input` and `change` event of the page, for the parts of the form of that field (or all of them when the
  field is in no form);
- after a form `reset`;
- when content arrives in a swap: the new parts, and the parts of the form that content joined.

A script that sets a value without firing an event calls `ypy.switch.update()`:

```js
document.querySelector('#kind').value = 'company';
ypy.switch.update();        // the whole page
ypy.switch.update(form);    // or just a subtree
```

## Event

| Event | On | Detail |
|---|---|---|
| `ypy:switch` | the part | `shown` (`true` or `false`) |

It fires when a part goes from shown to hidden or back. Its first evaluation is silent, and it is not cancelable.

```js
document.addEventListener('ypy:switch', evt => {
  if (evt.detail.shown) evt.target.querySelector('input')?.focus();
});
```

## The server side

The browser does not send what the extension disabled, so the server sees the form as if the hidden fields did not
exist. Validate by the same rule on the server: when `kind` is `company`, `registration` is required, and a form
that breaks the rule is answered with `422` and the form again, which the client swaps in place.

Never rely on the form hiding a field: someone can send the request by hand.

## Pitfalls

- It hides, it does not remove. The fields stay in the DOM, disabled. Read them in a script only when the part is
  shown.
- A disabled field is not sent. If the server needs a default for a hidden part, set it on the server.
- The field is found by `name`, not by `id`. A field without a `name` cannot drive a part.
- `ypy-switch` does not validate. It does not stop a submit when the visible fields are wrong; that is the job of
  the browser (`required`, `pattern`) and the server. See [Validate](validate.md) for the extension that validates a
  field on the server while the person types.
- A script that changes a value without an event does not update the page until you call `ypy.switch.update()`.
- A part that sits inside another hidden part stays disabled even if its own rule holds.
- A rule that names a missing field: with `=` the part stays hidden (there is no value), with `!=` it is shown.
