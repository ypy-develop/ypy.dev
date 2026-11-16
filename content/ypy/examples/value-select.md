---
order: 110
title: Dependent selects
description: Fill a model select from the make chosen in the first one: the server returns the options and owns the rule that links the two.
---

# Dependent selects

Choose a make and the list of models fills in. The page asks the server for the options, and the server owns the rule
that says which model belongs to which make.

## Try it

<!-- demo:value-select -->

## The page

```html
<label for="make">Make</label>
<select id="make" name="make" ypy-get="/models" ypy-target="#model" ypy-swap="outerHTML">
  <option value="">Choose a make</option>
  <option value="audi">Audi</option>
  <option value="toyota">Toyota</option>
  <option value="volkswagen">Volkswagen</option>
</select>

<label for="model">Model</label>
<select id="model" name="model" disabled>
  <option value="">Choose a make first</option>
</select>
```

## What the server answers

`GET /models?make=toyota` answers `200` with the second select, whole, so the `disabled` attribute goes away with it:

```html
<select id="model" name="model">
  <option value="">Choose a model</option>
  <option>Corolla</option>
  <option>Yaris</option>
  <option>Hilux</option>
</select>
```

| Request | Status | Body |
|---|---|---|
| `?make=audi` | `200` | The select with A3, A4 and Q5 |
| `?make=` or an unknown make | `200` | The disabled select with `Choose a make first` |

Both are fragments with `Vary: YPY-Request`.

## How it works

- A `<select>` triggers on `change`, so `ypy-get` needs no `ypy-trigger`. It sends its own value as `make`, in the
  query string.
- `ypy-target="#model"` with `ypy-swap="outerHTML"` replaces the second select. A change of make always brings back
  a select with the empty option chosen, so a model that belongs to the old make cannot stay selected.
- The server owns the map from makes to models. When the form is saved, check the pair again: a client can send any
  `model` it likes.
- The label stays on the page and points at the new select by `id`.

## Variations

- Send more than the select's own value: `ypy-include="closest form"` adds the other fields of the form to the
  request.
- Update a third place with the same answer, such as a price, with a `<ypy-partial>` next to the select. See
  [updating other content](updating-other-content.md).
- Fill the first select from the server too: give it `ypy-trigger="load"` on a wrapper of its own, or render it in the
  page.

## See also

- [`ypy-include` and `ypy-vals`](../reference.md#core-attributes)
- [Inline validation](inline-validation.md)
- [Parameters](../index.md#parameters)
