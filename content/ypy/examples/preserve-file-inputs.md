---
order: 140
title: Keep file inputs after a form error
description: Return a 422 with validation errors without clearing the file the user chose, by keeping the file input in place with ypy-preserve.
---

# Keep file inputs after a form error

A form has a title and a file. The title is missing, so the server answers with an error. A script cannot put a file back
into a file field, so the visitor would have to choose it again unless the field stays the same element. `ypy-preserve`
keeps it, and the text the visitor typed stays too.

## The page

```html
<form ypy-post="/documents" ypy-encoding="multipart/form-data"
      ypy-target="this" ypy-swap="innerHTML" ypy-error-swap="innerHTML">
  <label for="title">Title</label>
  <input id="title" name="title">

  <label for="doc">File</label>
  <input id="doc" name="doc" type="file" ypy-preserve>

  <button>Save</button>
</form>
```

## What the server answers

The answer holds the fields of the form, not the `<form>` element, because `innerHTML` replaces what is inside it.

| Case | Status | Body |
|---|---|---|
| Title missing | `422` | The fields again, with the message next to the title |
| File too large | `422` | The fields again, with the message under the file input |
| Saved | `200` | `<p>Saved Quarterly report.</p>` |

A `422` carries the fields like this. The title is written back from what was sent, escaped, and the file input has the
same `id` and the same `ypy-preserve`:

```html
<label for="title">Title</label>
<input id="title" name="title" value="" aria-invalid="true" aria-describedby="title-error">
<p id="title-error" role="alert">Enter a title.</p>

<label for="doc">File</label>
<input id="doc" name="doc" type="file" ypy-preserve>

<button>Save</button>
```

Every answer is a fragment with `Vary: YPY-Request`.

## How it works

- `ypy-preserve` with no value tells YPY to keep this DOM node, matched by its `id`, when a swap brings an element with
  the same id. The answer holds a new `<input id="doc" type="file">`, and YPY leaves the old one in place, with the file
  the visitor chose. The `id` is required.
- `ypy-target="this"` and `ypy-swap="innerHTML"` replace what is inside the form after a success. `ypy-error-swap`
  does the same for an error: a `422` goes into the form of the element that sent the request, so the messages and the
  fields arrive in the same place. Without `ypy-preserve` the new file field would replace the old one and the choice
  would be lost.
- The default swap, `innerMorph`, keeps the existing nodes and updates them, so a file input survives it as well. Use
  `ypy-preserve` when the swap is `innerHTML` or `outerHTML`, or when you want the guarantee written in the markup.
- `ypy-encoding="multipart/form-data"` sends the file with the form.

## Variations

- Keep what was typed without the server writing it back: the [`ypy-preserve`](../extensions/preserve.md) extension has
  the modes `value`, `selection`, `focus` and `scroll`. `ypy-preserve="value"` on the form keeps the text of fields that
  have an `id`. It skips file inputs, which need the plain `ypy-preserve` above.
- Let the form convention do it: without `ypy-target` on the form, a `422` is morphed into the form and the fields stay
  as they are. See [validation](../index.md#validation).

## See also

- [Preserving content](../index.md#preserving-content)
- [File upload](file-upload.md) and [inline validation](inline-validation.md)
