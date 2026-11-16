---
title: "Forms: validation errors and redirects"
description: "How YPY handles a form: a 422 puts the errors back in the form with the typed values kept, a 303 means done, and buttons, validation and sent fields."
order: 30
---
# Forms: validation errors and redirects

A form gets one of two answers from the server: "this is not valid, fix it" or "done, go on". HTTP already has statuses
for both, and YPY follows them.

| The server answers | It means | What the page does |
|---|---|---|
| `422 Unprocessable Content` | the form is not valid | renders the response inside the form, keeps what the user typed, focuses the first invalid field |
| `303 See Other` (after `POST`) | done | follows the redirect and swaps the final page like any success |
| any other 4xx or 5xx | something failed | goes where the request goes, or where `ypy-error-target` says |

The form convention is on by default. To send an error somewhere else, use `ypy-error-target`,
`ypy-error-swap` and `ypy-error-focus`. The attribute reference is [`ypy-error-*`](../reference/attributes/ypy-error.md).

## 1. The form

```html
<div id="page" class="mx-auto max-w-md p-6">
  <form ypy-post="/signup" ypy-target="#page" ypy-swap="innerHTML" class="space-y-4">
    <label class="block">Email
      <input id="email" name="email" type="email" required class="mt-1 w-full rounded border p-2">
    </label>
    <button class="rounded bg-sky-600 px-4 py-2 text-white">Create account</button>
  </form>
</div>
```

Without JavaScript the browser posts the form to `/signup` and the server answers with a whole page. With
YPY the same form updates `#page` and nothing else on the screen moves.

## 2. The server says "not valid": 422

Answer with the status 422 and the content of the form, with the invalid fields marked `aria-invalid="true"` and a
message tied to each one:

```html
<label class="block">Email
  <input id="email" name="email" value="ada@" aria-invalid="true" aria-describedby="email-error"
         class="mt-1 w-full rounded border p-2 aria-invalid:border-red-500">
</label>
<p id="email-error" role="alert" class="text-sm text-red-700">Enter a complete address.</p>
<button class="rounded bg-sky-600 px-4 py-2 text-white">Create account</button>
```

YPY puts this into the form (not into `#page`), morphs it over the current fields, and moves the focus to `#email`. The
form carries `data-ypy-state="error"` until the next request, so you can style it. Fields the user already filled in
stay filled in.

To answer with the whole `<form>` element, write `ypy-error-swap="outerMorph"` on the form. To put the
errors in a summary above the form, write `ypy-error-target="#summary" ypy-error-swap="innerHTML"`. A form
that already says `ypy-target="this" ypy-swap="outerHTML"` keeps doing it: the server answers with the
whole form and it replaces the form, with the focus on the first invalid field.

A 422 never changes the address bar, even when the form has `ypy-push-url`: an invalid form is not a new
page. A [boosted](../reference/history.md) form is different: it is a navigation, so a 422 there is a page
like any other.

## 3. The server says "done": 303

```go
http.Redirect(w, r, "/welcome", http.StatusSeeOther)
```

The browser's `fetch` follows the redirect, so YPY receives the page at `/welcome` and swaps it into
`#page` as a success. `ctx.redirected` is `true`. If the form is boosted, or has `ypy-push-url="true"`, the
address bar and the history get `/welcome`, the final URL, never `/signup`.

Use 303 and not 302 after a `POST`: 303 says "ask for the next page with a `GET`", which is what the
browser and `fetch` both do.

## 4. The whole round trip

```go
func signup(w http.ResponseWriter, r *http.Request) {
	form, problems := parseSignup(r)
	if len(problems) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)          // 422
		renderFields(w, r, form, problems)
		return
	}
	create(r.Context(), form)
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)         // 303
}
```

Add `Vary: YPY-Request` to every route that answers a fragment to YPY and a page to a browser, so a cache
does not mix them up.

## 5. The button that submitted

When the user presses a button, the browser tells the form which one it was. That button is the submitter, and YPY uses
it the way the browser does, with or without `ypy-boost`.

```html
<form ypy-post="/posts" ypy-target="#page" ypy-swap="innerHTML" class="space-y-4">
  <textarea name="body" required class="w-full rounded border p-2"></textarea>
  <button name="intent" value="publish" class="rounded bg-sky-600 px-4 py-2 text-white">Publish</button>
  <button name="intent" value="draft" class="rounded border px-4 py-2">Save draft</button>
  <button formaction="/posts/preview" formnovalidate class="rounded border px-4 py-2">Preview</button>
</form>
```

| On the button | What YPY does |
|---|---|
| `name` and `value` | only the button that was pressed is sent, as `intent=draft`. A button without a name (or with an empty one) sends nothing |
| `formaction` | the request goes to this URL, not to the `ypy-post` (or `action`) of the form. A button without it still goes where the form goes |
| `formmethod` | the request uses this method. It wins over `ypy-post`, `ypy-method` and the `method` of the form |
| `formmethod="dialog"` | no request is made |
| `formnovalidate` | the browser validation is skipped for this submit (see section 7) |

In the example, "Publish" and "Save draft" go to `/posts` with a different `intent`, and "Preview" goes to
`/posts/preview` without checking `required`.

Inside the handlers of `ypy:config:request` and the later events, the submitter is `ctx.request.submitter` and
the form is `ctx.request.form`.

### `ypy-*` attributes on the submitter

When the form carries the `ypy-post`, the form makes the request, and YPY reads its `ypy-*` attributes: `ypy-target`,
`ypy-swap`, `ypy-confirm`, `ypy-indicator`, `ypy-sync` and the rest. A `ypy-target` written on a plain submit button is
ignored. From the button YPY takes only the native attributes in the table above.

To give one button its own behavior, give it its own verb. The button then makes the request, and its
attributes apply; the fields of the form that holds it are still sent, with the name and value of the button:

```html
<form ypy-post="/posts" ypy-target="#page" ypy-swap="innerHTML">
  <textarea name="body"></textarea>
  <button>Publish</button>
  <button ypy-post="/posts/draft" name="intent" value="draft" ypy-target="#status" ypy-swap="innerHTML"
          ypy-confirm="Replace the saved draft?">Save draft</button>
</form>
```

The button that owns a verb does not also submit the form: the click does not reach the browser's own submit.
Because the button is now the element that makes the request, browser validation is off for it by default.
Section 7 shows how to turn it on.

### Boosted forms

A [boosted](../reference/history.md) form is a normal HTML form (`action` and `method`) that YPY sends with a
request. The `formaction` and `formmethod` of the button win over the form there too. YPY leaves the submit to
the browser when the browser does it better: a `target` or `formtarget` other than `_self`, a
`method="dialog"`, or an action on another origin.

If the page uses [`ypy-csp`](../reference/security.md#nonce-gating-the-ypy-csp-extension) with a nonce, a button that overrides `formaction` of a boosted
form must carry the nonce too. Without it the request is blocked and `ypy:security:violation` fires with the
reason `unnonced-submitter`. This keeps injected markup from sending a boosted form somewhere else.

## 6. What is sent

The body is built from the form, like the browser does it, plus what you add:

1. the fields of the form (a disabled field, an unchecked checkbox and a button that was not pressed are left
   out, as in the browser) and the submitter;
2. the elements named by `ypy-include` (selectors, `this`, `closest form`...). They are validated as well;
3. the values of `ypy-vals` (JSON or `js:`), which replace a field with the same name.

| Request | Where the fields go |
|---|---|
| `POST`, `PUT`, `PATCH` | in the body, `application/x-www-form-urlencoded`. The enclosing form of the element is always included |
| `GET`, `DELETE` | in the query string. A parameter already in the URL with the same name is replaced. Only an element that is itself a form sends its fields: a button inside a form does not |
| any of them with `enctype="multipart/form-data"` on the form, or `ypy-encoding="multipart/form-data"` | multipart, needed for files. For a progress bar see [`ypy-upload`](../extensions/upload.md) |

The `enctype` is read from the form or from `ypy-encoding`. A `formenctype` on the button is not read: write
`ypy-encoding` on the form.

A search form with `ypy-get` is a plain `GET`:

```html
<form ypy-get="/search" ypy-target="#results" ypy-push-url="true" class="flex gap-2">
  <input name="q" type="search" class="rounded border p-2">
  <button class="rounded bg-sky-600 px-4 py-2 text-white">Search</button>
</form>
```

It requests `/search?q=...`, and `ypy-push-url="true"` gives the page that same address.

## 7. Validation flow

A form goes through up to four checks, in this order. Each one stops the next when it fails.

| # | Who | Where | When it fails |
|---|---|---|---|
| 1 | the browser | `required`, `type=email`, `pattern`, `min`... | the request is not sent; the browser shows its own message on the first invalid field |
| 2 | the server, per field (optional) | [`ypy-validate`](../extensions/validate.md) extension | the message appears next to the field while the user fills the form; nothing is submitted |
| 3 | the server, the submit | `422` | the form is rendered again with the errors and the typed values (section 2) |
| 4 | the server, success | `303` or `200` | the page goes on (section 3) |

The browser check (1). A form is validated before the request by default, with the constraint API
(`reportValidity()`). It is off when:

- the form has `novalidate`;
- the button that submitted has `formnovalidate`;
- `ypy-validate="false"` is on the form (or inherited by it).

Validation is off by default for any element that is not a form. A button with its own `ypy-post`, or an input, sends
even when the fields next to it are invalid, unless you write `ypy-validate="true"` on it. With
`true` on a button, YPY validates the form that holds it and the elements added by `ypy-include`. A form with
`novalidate` can be validated again by a button with `ypy-validate="true"`.

```html
<form ypy-post="/signup">                       <!-- validated -->
<form ypy-post="/signup" novalidate>            <!-- not validated -->
<form ypy-post="/signup">
  <button>Create</button>                       <!-- validated -->
  <button formnovalidate>Skip checks</button>   <!-- not validated -->
</form>
```

The server check per field (2). The `ypy-validate` extension sends the form's values with the header
`YPY-Validate` when the user leaves a field. The handler must recognize it and change nothing:

```go
func signup(w http.ResponseWriter, r *http.Request) {
	q := ypy.FromRequest(r)
	form, problems := parseSignup(r)

	if q.IsValidation() {                       // only a check of one field
		if msg, bad := problems[q.Validate]; bad {
			ypy.ValidationError(w, `<span>`+html.EscapeString(msg)+`</span>`)
			return
		}
		ypy.NoContent(w)
		return
	}
	if len(problems) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		renderFields(w, r, form, problems)
		return
	}
	create(r.Context(), form)
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}
```

The server must still check everything at the submit (3). A field check is a convenience, and a user can skip it.

After a 422 (3), the response is rendered inside the form and the focus goes to the first
`aria-invalid="true"` field (or to the first field the browser calls invalid). `data-ypy-state="error"` stays on
the error target until the next request of that element. The browser check (1) runs before `ypy-confirm`: the
confirmation is asked only for a form that passed it.

### Pitfalls

- A button with its own `ypy-post` does not validate. This surprises people most often. Add `ypy-validate="true"`, or
  put the verb on the form and leave the button plain.
- `formnovalidate` on a button with its own `ypy-post` changes nothing, because that button is already not validated.
- With two buttons that share a `name` and differ in `value`, only the pressed one is sent, so the server should
  never read the value of the other.
- `formmethod="dialog"` makes no request at all. Use it only for a `<dialog>` that closes itself.

## See also

- [`ypy-error-*`](../reference/attributes/ypy-error.md) for the attributes, the precedence and `ctx.errorKind`.
- [`ypy-status:<code>`](../reference/inheritance.md) to handle one status in its own way.
- [History and caching](history-and-caching.md).
- [`ypy-validate`](../extensions/validate.md) for the per-field server check, and [`ypy-upload`](../extensions/upload.md) for upload progress.
- [Atomic swaps](../reference/atomic.md) when the answer to a form updates several places.
