---
order: 150
title: Reset the form after a save
description: Empty a comment form after a successful save and keep the typed text after a rejection, with a partial that swaps in a fresh form or one line of script.
---

# Reset the form after a save

A comment form adds the comment to a list. After a successful save the fields must be empty again, and after a
rejected one they must keep what the user typed. There are two ways: let the answer replace the form, or run a
line of script.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<ul id="comments"></ul>

<form id="new-comment" ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend"
      ypy-error-target="this" ypy-error-swap="innerMorph">
  <label for="body">Comment</label>
  <textarea id="body" name="body" required></textarea>
  <button>Post</button>
</form>
```

## The server (Go)

```go
package comments

import (
	"html"
	"net/http"
	"strings"

	ypy "github.com/ypy-develop/go"
)

// fields is the inside of the form; the typed text and the problem are optional.
func fields(body, problem string) string {
	msg := ""
	if problem != "" {
		msg = `<p id="body-error" role="alert">` + html.EscapeString(problem) + `</p>`
	}
	return `<label for="body">Comment</label>` + msg +
		`<textarea id="body" name="body" required>` + html.EscapeString(body) + `</textarea>` +
		`<button>Post</button>`
}

func blankForm() string {
	return `<form id="new-comment" ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend" ` +
		`ypy-error-target="this" ypy-error-swap="innerMorph">` + fields("", "") + `</form>`
}

func Create(w http.ResponseWriter, r *http.Request) {
	body := strings.TrimSpace(r.PostFormValue("body"))
	if body == "" {
		ypy.ValidationError(w, fields("", "Write something first."))
		return
	}
	// ... store the comment ...
	ypy.Fragment(w, `<li>`+html.EscapeString(body)+`</li>`) // goes to #comments
	ypy.Partial(w, ypy.PartialSpec{ID: "new-comment", Swap: "outerHTML"}, blankForm())
}
```

The answer has two parts. The `<li>` goes to the target of the request. The `<ypy-partial>` has the id of the form
and the swap `outerHTML`, so a new, empty form replaces the old one.

## How it works

- `ypy-swap="beforeend"` appends the comment to the list. The main target is `#comments`, not the form.
- `ypy.Partial` with `ID: "new-comment"` writes `<ypy-partial id="new-comment" ypy-swap="outerHTML">`. With no
  `ypy-target`, the element with that id is the target. The browser gets a fresh form with empty fields.
- On an error the server answers `422` with the fields. `ypy-error-target="this"` and `ypy-error-swap="innerMorph"`
  morph them into the form, so the message appears and the form itself is not replaced.
- The page and the handler build the same form. Put it in one template that both use.

## Variations

- A line of script instead of a second copy of the form: author code is off by default, so allow it first, then
  reset the form when the request succeeded:

  ```html
  <meta name="ypy-config" content='{"callbacks": "all"}'>

  <form ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend"
        ypy-on::after:request="ctx.response.status < 400 && this.reset()">
  ```

  `ypy-on::after:request` listens to `ypy:after:request` on the form. `ctx` is a name the handler receives, and `this`
  is the form. Read [security](../index.md#security) before you turn `callbacks` on.
- A listener in your script file does the same without `callbacks`: add a `ypy:after:request` listener on `document`
  and call `event.target.reset()` when `event.detail.ctx.response.status` is below 400.
- Only the field: give the answer a `<ypy-partial id="body" ypy-swap="outerHTML">` with an empty `<textarea>` to
  clear one field and keep the rest.

## See also

- [Updating other content](updating-other-content.md) and [`<ypy-partial>`](../index.md#out-of-band-swaps)
- [Forms](../index.md#validation)
