---
order: 20
id: "RECIPE-arandu"
type: "recipe"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/recipe", "lang/en", "stack/arandu"]
title: "Recipe: Arandu"
description: "Serve YPY from an Arandu application with Hesape and Kyse."
---

# Recipe: Arandu (Hesape and Kyse)

How a Go application built with Arandu serves YPY. The wire work is the Go SDK's, the same as in the
[Go recipe](go.md); what Arandu adds is the wiring in `bootstrap/app.go`, the CSRF middleware of Hesape, its redirect
and rejection helpers, and Kyse's components. Portuguese: [Receita: Arandu](../../pt-BR/recipes/arandu.md).

Hesape, the framework and Kyse are the first projects moved to YPY. The recipe describes the shape of the converted
code and uses the helpers that exist in Hesape today; a helper whose name is still being changed is described by what
it does, not named.

## Wire it where the application is assembled

An Arandu application wires everything explicitly in `bootstrap/app.go`; there is no container and no
auto-discovery. `ypy.Middleware` has the standard shape, so it fits the router's chain:

```go
// bootstrap/app.go
router.Get("/items", items.Index, ypy.Middleware)
router.Post("/items", items.Create, ypy.Middleware)
router.Get("/items/events", room.ServeHTTP, ypy.Middleware) // a stream: see the SDK README
```

`ypy.Middleware` reads the YPY headers into the request context (`ypy.FromContext(r.Context())`) and adds
`Vary: YPY-Request` to every response, so the page and the fragment of the same URL are never mixed by a cache.

Install the SDK from its repository, as every Arandu module: `go get github.com/ypy-develop/go@<tag>`, with the tag
you checked first; do not point `go.mod` at a local checkout with `replace`.

## The handler

Hesape handlers are plain `net/http` handlers. The YPY part is the same as anywhere; authorization and the tenant are
the Arandu part and stay as they are in every handler of your application (a Policy through `security.Authorize`, the
tenant from `data.Tenant(g)` of the Grant it returns, never from the path, the body or a header):

```go
type Items struct{ svc *itemsvc.Service }

func NewItems(svc *itemsvc.Service) *Items { return &Items{svc: svc} } // explicit wiring

func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	// 1. authorize: Policy -> security.Authorize -> Grant; the service takes the Grant, so the tenant is data.Tenant(g)
	// 2. validate and save (the service returns validation errors as values)
	item, errs, err := h.svc.Create(r.Context(), r.PostFormValue("title"))
	switch {
	case err != nil:
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	case len(errs) > 0:
		ypy.ValidationError(w, renderForm(r, errs)) // 422 and the form; see "Invalid forms" below
	case !ypy.FromRequest(r).IsYPY:
		ypy.SeeOther(w, "/items") // a browser without the script: post, redirect, get
	default:
		ypy.Fragment(w, renderRow(item)) // the new row, into #list
		ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, countText(r))
	}
}
```

`renderRow`, `renderForm` and `countText` are your Hesape views or Kyse components. Render the same component for the
page and for the fragment, so there is one place that draws a row.

## Redirects

`hesape/http.Redirect(w, r, to)` already does what `ypy.Redirect` does: a YPY request gets `YPY-Redirect` and `204`,
any other client gets `303`. Use whichever you have; do not hand-write the branch.

## CSRF

Hesape's `CSRFProtect` middleware (package `hesape/http/middleware`) refuses a state-changing request the browser
reports as cross-origin (`403`), then checks the token. It reads the `X-CSRF-Token` header first and the `_token` form
field after it, and answers `419` when the token is missing or expired.

- **Forms** carry the `_token` hidden field, so `<form ypy-post>` is covered whether it posts with the script or
  without it.
- **Requests that no form backs** (`ypy-delete` on a button, `ypy-patch` on a toggle) carry the token in a header.
  Put it once on the layout. An attribute reaches descendants only with `:inherited`:

```html
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}' class="bg-white text-slate-900">
```

- **An expired token** ends in `419`. Hesape answers a refusal to a YPY request with `YPY-Refresh: true`, so the page
  reloads and shows the refusal as an ordinary page (the reload is a `GET`, which is never checked, and it brings a
  fresh token). Nothing loops.

## Invalid forms: two ways, both work

1. **Flash and redirect (what Hesape's `Reject` does).** The answer to a rejected form is a redirect back, carrying
   the messages and what was typed in the flash. YPY follows the redirect (`YPY-Redirect` for a YPY request, `303`
   otherwise) and shows the page that comes back. It needs no change when you move to YPY and it also works for a
   browser with no script. The cost is a second request.
2. **`422` in place.** `ypy.ValidationError(w, renderForm(r, errs))` answers the form with its errors; YPY merges it
   over the form, keeps what was typed and focuses the first invalid field. Add
   `ypy-error-swap="outerMorph"` on the form when the answer is the whole `<form>`. One request, no flash. See
   [Forms](../../guides/forms.md).

Pick one per application and keep it. Do not mix them in one form.

## The layout and Tailwind

The layout loads the script, the CSRF header and the stylesheet. Serve `ypy.min.js` from the application's asset path:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="{{ asset "app.css" }}">
  <script src="{{ asset "ypy.min.js" }}"></script>
</head>
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}'>
  <main id="main" class="mx-auto max-w-3xl p-6">{{ template "content" . }}</main>
</body>
</html>
```

`asset` stands for however your application resolves an asset URL. For Tailwind, scan the Kyse and Go templates and
add the state variants once ([Styling the states](../../guides/core-concepts.md#styling-the-states)):

```css
@import "tailwindcss";
@source "../internal/**/*.go";
@source "../views/**/*";
@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

## Kyse

Kyse components that start requests render the YPY attributes. Use the component's own props for the verb, the target
and the swap instead of writing the attribute in a string, so the component keeps the escaping and the accessibility
it was built with.

An application written before YPY has `.kyse.go` sources and the compiled `.go` files beside them. `ypy migrate`
converts both the same way, and `aru view:build` regenerates the compiled files from the sources with the same result;
see [Migration](../../migration/migrate.md#step-by-step-for-an-arandu-or-kyse-project) for the order (hesape, framework,
kyse, then products).

## Tests

Test the handler without a browser, with `ypytest` in the application's `go test`:

```go
res := ypytest.Do(handler, ypytest.NewRequest(http.MethodPost, "/items",
	ypytest.Target("ul#list"),
	ypytest.Header("X-CSRF-Token", token),
	ypytest.Form(url.Values{"title": {"Buy milk"}})))
res.AssertStatus(t, http.StatusOK)
res.AssertVary(t, "YPY-Request")
```

[Testing your app](../testing.md) has the rest.

## Checklist

- [ ] `ypy.Middleware` in the chain of every route that answers a fragment and a page
- [ ] the Policy and the Grant decide, the tenant comes from `data.Tenant(g)`, the markup is the same component for page and fragment
- [ ] `CSRFProtect` is on, the layout carries the header, `419` reloads the page
- [ ] one way to answer an invalid form, flash redirect or `422`
- [ ] the activity log records the action as for any other request: YPY does not change it
- [ ] `ypy check` and `ypy migrate` ran over the `.kyse.go` sources and the compiled files
