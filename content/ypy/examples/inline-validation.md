---
order: 70
title: Inline validation
description: Check one field on the server when the user leaves it with ypy-validate, and show the 422 message beside the field without submitting the form.
---

# Inline validation

Some rules only the server knows: this email is taken, this code has expired. With the `ypy-validate` extension, a field
is sent to the server when the user leaves it, and the answer appears next to the field. The form is not submitted.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-validate.min.js"></script>

<form ypy-post="/signup" ypy-target="#result" ypy-validate-fields="/signup">
  <label for="email">Email</label>
  <input id="email" name="email" type="email" required>
  <small id="email-error"></small>

  <label for="name">Name</label>
  <input id="name" name="name" required>
  <small id="name-error"></small>

  <button>Create account</button>
</form>
<div id="result"></div>
```

`ypy-validate-fields` on the form turns every field inside it into a server-checked field. Each field has a `<small>`
whose `id` is the field's `id` plus `-error`. That is where YPY puts the answer when you name nothing else.

## The server (Go)

The same route serves the field checks and the real submit, so the handler looks at the request first.

```go
package main

import (
	"html"
	"net/http"

	ypy "github.com/ypy-develop/go"
)

func signup(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		req := ypy.FromRequest(r)

		if req.IsValidation() {
			// A field check. Answer about the fields and do nothing else.
			// With one field req.Validate is its name, with a batch it is a list.
			if req.Validate == "email" && store.EmailTaken(r.PostFormValue("email")) {
				ypy.ValidationError(w, `This email is taken.`)
				return
			}
			ypy.NoContent(w) // valid: the message is cleared
			return
		}

		// The real submit.
		if store.EmailTaken(r.PostFormValue("email")) {
			ypy.ValidationError(w, `<p role="alert">This email is taken.</p>`)
			return
		}
		store.Create(r.PostFormValue("email"), r.PostFormValue("name"))
		ypy.Fragment(w, `<p>Welcome, `+html.EscapeString(r.PostFormValue("name"))+`.</p>`)
	}
}
```

Register it with `mux.Handle("POST /signup", ypy.Middleware(signup(store)))`.

## How it works

- When the user changes a field and leaves it, the extension waits 200 ms and sends `POST /signup` with every value of
  the form and the header `YPY-Validate: email`. The default trigger is `change delay:200ms`. Set
  `ypy-validate-trigger="input delay:400ms"` on a field to check while typing.
- `ypy.FromRequest(r).IsValidation()` is true when that header is present. The handler must answer and return before it
  creates anything, because the request goes to the real route.
- A `422` swaps its body into `#email-error`, sets `aria-invalid="true"` on the field and adds `aria-describedby`. A
  `204`, or an empty `200`, empties the message and removes `aria-invalid`.
- The field keeps the focus and the other fields keep what the user typed. A `422` for a field never goes into the form.
- The submit button runs the normal request. `required` and `type="email"` are checked by the browser first, and a
  `422` from the real submit goes into the form as in [Forms](../index.md#validation).
- Without the script, the form still posts to `/signup` and the server answers with a page.

## Variations

- Name the place yourself with `ypy-error-target="next .hint"` on the field. See
  [`ypy-validate`](../extensions/validate.md).
- A form that asks many fields at once can send them in one request with `ypy-validate-batch`. The handler then reads a
  comma separated `YPY-Validate` and answers each field with an out-of-band element.
- Check only one field with `ypy-validate="/emails/check"` on it, in place of `ypy-validate-fields` on the form.

## See also

- [`ypy-validate`](../extensions/validate.md)
- [Forms guide](../index.md#validation)
