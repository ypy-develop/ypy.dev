---
title: "ypy-error-target, ypy-error-swap and ypy-error-focus"
description: "Where an error response goes, how it is swapped, and where the focus lands. They apply to a response with status 400 or more; a success never reads them."
order: 30
---
# ypy-error-target, ypy-error-swap and ypy-error-focus

Where an error response goes, how it is swapped, and where the focus lands. They apply to a response
with status 400 or more; a success never reads them. `ypy-error-target` and
`data-ypy-error-target` are the same attribute (likewise for `swap` and `focus`).

```html
<form ypy-post="/signup" ypy-target="#page" ypy-swap="innerHTML"
      ypy-error-target="#errors" ypy-error-swap="innerHTML" ypy-error-focus="#email">
  <input id="email" name="email" class="rounded border p-2">
  <button class="rounded bg-sky-600 px-4 py-2 text-white">Create account</button>
</form>
<div id="errors" class="text-red-700"></div>
```

On a 201 the response replaces `#page`. On a 422, a 500 or any other error it goes into `#errors`, and
the focus moves to `#email` once the swap is done.

| Attribute | Value | Default |
|---|---|---|
| `ypy-error-target` | a selector, with the same syntax as `ypy-target` (`#id`, `closest form`, `find .errors`, `next .hint`) | none: the target of the request |
| `ypy-error-swap` | a swap spec, as in `ypy-swap` (`innerHTML`, `outerMorph`, `beforeend`...) | none: the swap of the request |
| `ypy-error-focus` | `first-invalid`, a selector, or `false` | none (a 422: `first-invalid`) |

All three are read from the element that sends the request. They are inherited only with `:inherited`,
like every YPY attribute (`ypy-error-target:inherited="#errors"` on a container).

## Precedence

From weakest to strongest, per key (target, swap, focus):

1. the default (the form convention below);
2. `ypy-error-*`;
3. the `YPY-Retarget` and `YPY-Reswap` response headers (the server decides);
4. `ypy-status:<code>`, `ypy-status:<NN>x` and `ypy-status:<N>xx`, which merge into the same context.

So `ypy-status:422="target:#form-errors"` beats `ypy-error-target`, and a key it does not set (the swap,
say) still comes from `ypy-error-swap`. A `ypy:response:error` listener runs after steps 1 and 2 and
before the headers, and may still change `ctx`.

A `target:` modifier written inside the `ypy-swap` of the request keeps priority over `ctx.target`
(that is how every swap works). To send errors somewhere else, give the error its own swap with
`ypy-error-swap`, which replaces the request's one.

A selector that matches nothing is treated like any other missing target: the console warns, the
swap is skipped, and the [main-target fallback](ypy-main.md) applies first.

## The form convention: 422 and 303

A 422 Unprocessable Content means "the form is not valid", and the response renders inside the form. A 303 See Other
after a `POST` means "done", and it is a success like any other.

A 422 with no `ypy-error-*` goes into the form of the element that sent the
request (`element.form`, else `closest form`; the element itself when it is the form), with the default
swap (`innerMorph`: the response is the content of the form), and the focus goes to the
first `[aria-invalid="true"]` field, or the first field the browser calls `:invalid`. The morph keeps
what the user typed, the open `<details>` and the focus of the other fields, so a form with three
errors does not lose the other twelve fields. The history is left alone (`ypy-push-url` is for the
success; the server can still send `YPY-Push-Url`).

Three cases keep the swap the author wrote, and only the focus is the default:

- the request already swaps into its own form (`ypy-target="this"` on the form, the classic "replace the
  form with the form the server renders" pattern): the server answers a 422 with the whole form, as
  it did before;
- the element is boosted: it is a navigation, so the 422 is a page like any other and goes where the boost sends it,
  the history entry included;
- a button outside any form: there is no form to render into.

A 303 needs nothing from YPY: `fetch` follows the redirect and YPY swaps the response of the final URL
as a success.

- `ctx.redirected` is `true` (it is `response.redirected`), and `ctx.errorKind` is unset.
- The history gets the final URL, not the URL of the `POST`, when the element is
  [boosted](../history.md) or has `ypy-push-url="true"`. An explicit `ypy-push-url="/path"` wins.
- Without a boost and without `ypy-push-url` the URL does not change, as for any swap.

The server side, in Go:

```go
func signup(w http.ResponseWriter, r *http.Request) {
	form, problems := parseSignup(r)
	if len(problems) > 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity) // 422: render the form again, with the errors
		tmpl.ExecuteTemplate(w, "signup-fields", view{Form: form, Problems: problems})
		return
	}
	create(r.Context(), form)
	http.Redirect(w, r, "/welcome", http.StatusSeeOther) // 303: done, go to /welcome
}
```

```html
{{define "signup-fields"}}
<label class="block">Email
  <input id="email" name="email" value="{{.Form.Email}}"
         {{if .Problems.Email}}aria-invalid="true" aria-describedby="email-error"{{end}}
         class="mt-1 w-full rounded border p-2 aria-invalid:border-red-500">
</label>
{{if .Problems.Email}}<p id="email-error" role="alert" class="text-sm text-red-700">{{.Problems.Email}}</p>{{end}}
<button class="rounded bg-sky-600 px-4 py-2 text-white">Create account</button>
{{end}}
```

Answer a request that came from YPY (`YPY-Request: true`) with the fields only and add
`Vary: YPY-Request`, so a cache never serves the fragment as a page. A visitor without JavaScript posts
the form as usual and gets the 422 as a whole page: render the same template inside the layout.

## What the core records

| What | Where | When |
|---|---|---|
| `ctx.errorKind` | the request context | for every response of 400 or more: `validation` (422), `client` (other 4xx), `server` (5xx). Network failures set `network`, `offline` and `timeout` (see [`ypy:offline`](../events/ypy-offline.md)). Unset on success |
| `ctx.focus` | the request context | what the attributes (or the 422 default) say: `"first-invalid"`, a selector string, or `false`. Applied by the focus step at the end of the swap (see [`ypy-focus`](ypy-focus.md)) |
| `ctx.redirected` | the request context | `true` when the response came through a redirect, `false` otherwise |
| `data-ypy-state="error"` | the error target | after the swap, until the element that sent the request sends its next one |

`data-ypy-state` holds a list of states, so `error` can sit beside others (`data-ypy-state="error offline"`).
It is set for every error. To also get a class, set `ypy.config.classes.error = "is-invalid"`
(see [`config.classes`](../config/classes.md)). The state survives a morph of the target: `data-ypy-state`
is on the `morphIgnore` list. A swap that replaces the target (`outerHTML`) puts the state on the new
element when the target is the element that sent the request, or when the selector finds it again.

### Styling with Tailwind CSS

The state is an attribute, so Tailwind v4 needs no plugin:

```html
<form ypy-post="/signup"
      class="rounded border p-4 data-[ypy-state~=error]:border-red-500 data-[ypy-state~=error]:bg-red-50">
```

### Focus

`ypy-error-focus="first-invalid"` is the first `[aria-invalid="true"]` or `input`, `select` or `textarea`
that matches `:invalid`, looked up inside the swap target. A selector uses the extended syntax, relative to the
element that sent the request (`#email`, `find input[name=email]`). `false` leaves the focus alone. The focus
is given by the same step as [`ypy-focus`](ypy-focus.md), after the swap, so the new fields exist, and like it
the focus does not scroll the page: scrolling is the job of [`ypy-scroll`](ypy-scroll.md). A selector that
finds nothing does nothing, and a swap that a listener cancelled moves no focus.

## Not an error

`204` and `304` (`config.noSwap`) swap nothing and are not errors. A response of 400 or more that a header
takes over (`YPY-Redirect`, `YPY-Location`, `YPY-Refresh`) is not swapped, and gets no state or focus.

## Related

- [`ypy-main` and `ypy-fallback`](ypy-main.md): what happens when the error target is missing.
- [`ypy-announce`](ypy-announce.md): errors are announced in the live region from `ctx.errorKind`.
- [Forms](../../guides/forms.md): the whole round trip, step by step.
