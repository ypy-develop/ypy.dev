---
order: 140
title: Keep file inputs after a form error
description: Return a 422 with validation errors without clearing the file the user chose, by keeping the file input in place with ypy-preserve.
---

# Keep file inputs after a form error

A form has a title and a file. The title is missing, so the server answers with an error. A script cannot put a file back
into a file field, so the user would have to choose it again unless the field stays the same element. `ypy-preserve`
keeps it, and the text the user typed stays too.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<form ypy-post="/documents" ypy-encoding="multipart/form-data"
      ypy-target="this" ypy-swap="innerHTML" ypy-error-swap="innerHTML">
  <label for="title">Title</label>
  <input id="title" name="title" required>

  <label for="doc">File</label>
  <input id="doc" name="doc" type="file" ypy-preserve>

  <button>Save</button>
</form>
```

## The server (Go)

The server answers with the fields of the form, not the `<form>` element, because `innerHTML` replaces what is
inside it.

```go
package documents

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

func fields(title, problem string) string {
	var b strings.Builder
	b.WriteString(`<label for="title">Title</label><input id="title" name="title" value="` + html.EscapeString(title) + `"`)
	if problem != "" {
		b.WriteString(` aria-invalid="true" aria-describedby="title-error"><p id="title-error" role="alert">` +
			html.EscapeString(problem) + `</p>`)
	} else {
		b.WriteString(`>`)
	}
	b.WriteString(`<label for="doc">File</label><input id="doc" name="doc" type="file" ypy-preserve>`)
	b.WriteString(`<button>Save</button>`)
	return b.String()
}

func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ypy.ValidationError(w, fields("", "The upload is too large."))
		return
	}
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		ypy.ValidationError(w, fields("", "Enter a title."))
		return
	}
	// ... store the file and the title ...
	ypy.Fragment(w, `<p>Saved `+html.EscapeString(title)+`.</p>`)
}
```

## How it works

- `ypy-preserve` with no value tells YPY to keep this DOM node, matched by its `id`, when a swap brings an element
  with the same id. The response holds a new `<input id="doc" type="file">`, and YPY leaves the old one in place, with
  the file the user chose. The id is required.
- `ypy-target="this"` and `ypy-swap="innerHTML"` replace what is inside the form after a success. `ypy-error-swap`
  does the same for an error. The `422` from `ypy.ValidationError` goes into the form of the element that sent
  the request, so the errors and the fields arrive in the same place. Without `ypy-preserve` the new file field would
  replace the old one and the choice would be lost.
- The default swap, `innerMorph`, keeps a file input without help, because a morph does not set `value` on
  `type="file"`. Use `ypy-preserve` when the swap is `innerHTML` or `outerHTML`, or when you want the guarantee
  written in the markup.
- The server writes the title back from what it received, escaped. The user sees what they typed and the message
  next to the field.

## Variations

- Keep what was typed without the server. The `ypy-preserve` extension has the modes `value`, `selection`,
  `focus` and `scroll`: `ypy-preserve="value"` on the form keeps typed text in fields that have an `id`. It skips
  file inputs, which need the plain `ypy-preserve` above.
- Let the form convention do it. Without `ypy-target` on the form, a `422` is morphed into the form and the fields
  stay as they are. See [forms](../guides/forms.md).

## See also

- [`ypy-preserve`](../reference/attributes/ypy-preserve.md)
- [File upload](file-upload.md) and [inline validation](inline-validation.md)
