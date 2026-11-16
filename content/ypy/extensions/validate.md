---
title: "Field validation on the server: the ypy-validate extension"
description: "Sends the form's values to the server when the user leaves a field and shows the answer next to it, without submitting the form."
order: 480
---
# Field validation on the server: the `ypy-validate` extension

Some rules only the server knows: this email is taken, this code has expired. `ypy-validate` asks the server about
them when the user leaves a field and shows the answer next to the field, before the user presses the button. The
form's values go to the server, but the form is not submitted.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-validate.js"></script>

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

Leave the email field after changing it: a `POST /signup/validate` goes out. If the server answers `422` with
`<span>This email is taken</span>`, that markup lands in `#email-error`, the field gets `aria-invalid="true"` and
`aria-describedby="email-error"`, and the focus stays where the user is.

If you set `ypy.config.extensions`, list `ypy-validate` in it, or the extension is refused. It is listed with the
other official extensions in [Extensions](index.md). Without it the attributes below do nothing.

## Attributes

| Attribute | On | Value | Meaning |
|---|---|---|---|
| `ypy-validate` | a field | a URL, or nothing | check this field; a bare attribute asks the action of the form |
| `ypy-validate-fields` | a form, or any element around fields | a URL, or nothing | check every field inside; a field with `ypy-validate="false"` is left out, and one with its own URL uses it |
| `ypy-validate-trigger` | a field, or the element with `ypy-validate-fields` | a trigger specification, as `ypy-trigger` takes | when to check; default `change delay:200ms` |
| `ypy-validate-batch` | a field, the element with `ypy-validate-fields`, or the form (the nearest wins) | a time, nothing, or `false` | send the fields asked within that window in one request; a bare attribute means `50ms`; off by default (see [Batches](#batches)) |
| `ypy-error-target` | the field | a selector | where the answer goes (see below) |
| `ypy-error-swap` | the field | a swap style | how it goes in; default `innerHTML` |
| `ypy-error-focus` | the field | a selector | where the focus goes after a `422`; default: nowhere, it stays |

The action of the form is its `ypy-post` (or `ypy-action`), else its `action` attribute, else the current path.

`ypy-validate` already has a meaning in the core: `ypy-validate="true"` and `="false"` switch the browser's own
constraint check (`required`, `pattern`...) before a request. Those two values are left to the core and this
extension ignores them. Any other value, including no value, is a server check. On a form write
`ypy-validate-fields`, not `ypy-validate`: a form whose `ypy-validate` is a URL would lose the browser check of its
own submit.

The default trigger is `change delay:200ms`: the field's value changed and the user left it, and 200 ms passed
without another event of the same field (a debounce). Other triggers use the same syntax as `ypy-trigger`:

```html
<input name="handle" ypy-validate="/handles/check" ypy-validate-trigger="input delay:400ms"> <!-- while typing -->
<input name="code"   ypy-validate="/codes/check"   ypy-validate-trigger="blur">               <!-- on every exit -->
```

Fields are wired when their subtree is processed, so fields that arrive in a swap work as soon as the swap is done.

## The request

A `POST` to the URL with every value of the form (so a rule can compare two fields), plus the header
`YPY-Validate` with the name of the field being checked (the `id` when the field has no `name`). It carries the
usual `YPY-*` request headers; its `YPY-Source` is the field.

> The server must look at `YPY-Validate` before it does anything else. The request goes to a real handler,
> often the one of the form itself. A handler that does not know the header would run the whole action: create the
> account, send the email. Check the header first, answer, and return.

## The answer

| Status | Meaning | The page |
|---|---|---|
| `422` | the field is invalid | the body is swapped into the place, the field gets `aria-invalid="true"` and `aria-describedby` |
| `204`, or an empty `200` | valid | the place is emptied, `aria-invalid` is removed |
| `200` with a body | valid, with something to say ("Available") | the body is swapped into the place, `aria-invalid` is removed |
| anything else (`5xx`, `403`...), or the network fails | no answer | nothing changes on the page; the core still fires `ypy:response:error` or `ypy:error`. The form validates again when it is submitted |

The place follows the core's [error convention](../reference/attributes/ypy-error.md): `ypy-error-target` on the field, with
every selector form (`#id`, `next .hint`, `closest .row`, `find .msg`). Without it, `#<id>-error` (the field's `id`,
else its `name`); when the page has no such element, the extension creates a `<div class="ypy-field-error"
aria-live="polite">` right after the field, once. `aria-describedby` keeps the ids that were there and adds the place
once.

A `YPY-Retarget` or `YPY-Reswap` header and `ypy-status:422` still win over all of this, as for any error. A
`422` for a *form* goes into the form; a `422` for a *field* never does, so the other fields,
what the user typed and the focus stay.

## Batches

By default every field is a request of its own. When a user pastes a whole form, or a script fills several fields
at once, that is one request per field. With `ypy-validate-batch`, the fields of one form that ask the same URL
within a short window go in a single request.

```html
<form ypy-post="/signup" ypy-validate-fields="/signup/validate" ypy-validate-batch="80ms">
  <input id="email" name="email" type="email"><small id="email-error"></small>
  <input id="username" name="username"><small id="username-error"></small>
  <button>Create account</button>
</form>
```

```js
ypy.config.validate = { batch: '80ms' };   // for every form of the page; ypy-validate-batch="false" turns one off
```

- The window starts at the first field that asks. Fields asked inside it, in the order of their first ask (a field
  that asks twice counts once), go together in one `POST` with every value of the form, as for one field.
- `YPY-Validate` then holds the names separated by commas: `YPY-Validate: email,username` (the `id` of a field with no
  `name`). A single field in a window is checked as without batching: one name and the usual answer.
- Fields of two forms, or of two different URLs, never share a request.
- If a field of a batch that is still in flight asks again, the request starts over with the whole batch.
- A field that left the page during the window is not sent.

It is off by default because the server must read the list. A handler that only knows one name would see
`email,username` and find no rule for it.

The answer. `422` when any field in the batch is invalid, any `2xx` when none is. The message of each invalid field
comes as an out-of-band element whose `id` is the `id` of that field's place (`#<id>-error`, or the element named by
`ypy-error-target`). A field with no such element is valid, and its place is emptied. `aria-invalid` and
`aria-describedby` follow what each place shows afterwards. Any other status changes nothing on the page.

```http
POST /signup/validate
YPY-Validate: email,username
Content-Type: application/x-www-form-urlencoded

email=ana%40example.com&username=ana
```

```http
HTTP/1.1 422 Unprocessable Content
Content-Type: text/html; charset=utf-8

<small id="email-error" ypy-swap-oob="innerHTML">This email is taken</small>
```

Here `username` is valid (no element for it), so `#username-error` is emptied and `username` loses `aria-invalid`.

```go
// ValidateBatch answers POST /signup/validate for one field or for a batch. It saves nothing.
func (h *Handler) ValidateBatch(w http.ResponseWriter, r *http.Request) {
	probs := h.problems(r) // field name -> message, as in the handler above
	var out strings.Builder
	for _, name := range strings.Split(r.Header.Get(ypy.HeaderValidate), ",") {
		if msg, bad := probs[name]; bad {
			out.WriteString(`<small id="` + html.EscapeString(name) + `-error" ypy-swap-oob="innerHTML">` +
				html.EscapeString(msg) + `</small>`)
		}
	}
	if out.Len() == 0 {
		ypy.NoContent(w)
		return
	}
	ypy.ValidationError(w, out.String())
}
```

The same code answers a single field (`YPY-Validate: email`), so a page can turn batching on and off without a change
on the server. This sketch assumes the place of each field is `#<name>-error`; when a field's `id` differs from its
`name`, use the `id` of the place.

## Concurrency

- A field has at most one check in flight: a new check cancels the previous one, so a slow answer never
  overwrites the fresh one.
- Different fields are checked independently and can be in flight together.
- A request sent by the form (its own submit) cancels the checks of its fields that are still running, so a late
  "invalid" does not appear after the form was accepted or re-rendered.
- A check is not sent if the field has left the page while the debounce was waiting.
- In a [batch](#batches) one request holds several fields: a new ask for any of them restarts the whole batch, and
  the submit of the form cancels it like a single check.
- The check changes nothing on the server and must not be counted as a write: with `ypy-cache`, a request that
  carries `YPY-Validate` does not expire the cache.

## A Go handler

```go
package signup

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

type Handler struct {
	EmailTaken func(email string) bool
	Register   func(email, name string) error
}

// problems returns the first message for each invalid field.
func (h *Handler) problems(r *http.Request) map[string]string {
	out := map[string]string{}
	email := strings.TrimSpace(r.PostFormValue("email"))
	switch {
	case !strings.Contains(email, "@"):
		out["email"] = "Enter an email address"
	case h.EmailTaken(email):
		out["email"] = "This email is taken"
	}
	if strings.TrimSpace(r.PostFormValue("name")) == "" {
		out["name"] = "Enter your name"
	}
	return out
}

// Validate answers POST /signup/validate. It saves nothing.
func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	field := r.Header.Get("YPY-Validate")
	if msg, bad := h.problems(r)[field]; bad {
		ypy.ValidationError(w, `<span class="text-red-700">`+html.EscapeString(msg)+`</span>`)
		return
	}
	ypy.NoContent(w)
}

// Create answers POST /signup, the real submit.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("YPY-Validate") != "" { // a field check that reached the action: never act on it
		h.Validate(w, r)
		return
	}
	if probs := h.problems(r); len(probs) > 0 {
		ypy.ValidationError(w, `<p>Fix the fields marked in red</p>`)
		return
	}
	if err := h.Register(r.PostFormValue("email"), r.PostFormValue("name")); err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	ypy.SeeOther(w, "/welcome")
}

func Wire(mux *http.ServeMux, h *Handler) http.Handler {
	mux.HandleFunc("POST /signup/validate", h.Validate)
	mux.HandleFunc("POST /signup", h.Create)
	return ypy.Middleware(mux)
}
```

When the validation URL is the form's own action (a bare `ypy-validate` or `ypy-validate-fields`), the first lines of
`Create` are what keep a field check from creating the account.

## Accessibility

- The answer region has `aria-live="polite"` when the extension creates it; give your own region the same attribute
  so a screen reader announces the message.
- `aria-invalid` and `aria-describedby` follow the check, and are removed or left in place by the next one.
- The focus is never moved by a check (the user is moving through the form). Use `ypy-error-focus` to change that
  deliberately, for one field.
- The check is a hint: the form must still validate on submit.

## Pitfalls

- A handler that does not read `YPY-Validate` runs the real action on every check. Test it with a request that carries
  the header.
- With batching on, `YPY-Validate` can hold commas. Split it; do not compare it with one field name.
- In a batch, a `422` without an out-of-band element for a field means that field is valid. A `422` whose body is plain
  text, with no element ids, empties every place and marks no field.
- The message of a batch needs the `id` of the place. With `ypy-error-target` pointing elsewhere, give that element the
  `id` the server writes.

## Limits

- Fields wired in a form after its container was processed are wired when that container is processed again.
- A form with `enctype="multipart/form-data"` sends its files with every check. Give the field
  `ypy-encoding="application/x-www-form-urlencoded"` to send the values as text only.
- Size: 1.53 kB min+gzip (`ypy build --sizes`).
