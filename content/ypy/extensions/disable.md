---
order: 110
title: "Disable while loading: the ypy-disable extension"
description: "Disables buttons and fields, or makes the target inert, while a request runs, so a visitor cannot send the same form twice."
---

# Disable while loading: the `ypy-disable` extension

A visitor who clicks "Save" and sees nothing happen clicks again. `ypy-disable` holds back the elements you name until
the request ends, so the second click does nothing. It also makes the target of the request `inert` on demand.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-disable.min.js"></script>
```

The extension is in `ypy.max`.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-disable` | a selector | The elements that get the `disabled` property until the request ends. Accepts extended selectors such as `find button` or `closest form`. Inherited. |
| `ypy-inert` | `true` or `false` | `true` makes the target of the request `inert` until the request ends. `false` turns it off for one element when `config.inert` is on. Inherited. |

The page [`ypy-inert`](../reference.md#core-attributes) has the full story of the second attribute.

## Settings

| Key | Default | Meaning |
|---|---|---|
| `inert` | `false` | Makes the target of every request inert while it runs. |

## Example

```html
<form ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend" ypy-disable="find button, find textarea">
  <textarea name="body" required></textarea>
  <button>Send</button>
</form>
<ul id="comments"></ul>
```

While the request runs, the button and the text box are disabled. When the response arrives, or the request fails, they
are enabled again.

## Notes

- Several requests may name the same element. It is enabled again when the last one ends.
- A request that hands its indicators to a follow-up (the `YPY-Location` header) hands these elements over too, and the
  follow-up releases them.
- The elements are disabled after the request is built, so disabling a field does not take its value out of the
  request that is already on its way.
- `disabled` only means something for form controls. For a region that holds links or custom widgets use `ypy-inert`.

## See also

[`ypy-inert`](../reference.md#core-attributes), [Styling](../index.md#animations) for the request classes,
[`ypy-pending`](pending.md).
