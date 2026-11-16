---
order: 10
id: "RECIPE-go"
type: "recipe"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/recipe", "lang/en", "stack/go"]
title: "Recipe: Go"
description: "Serve YPY from a Go net/http application."
---

# Recipe: Go (`net/http`)

A complete server with the standard library and the YPY Go SDK: a page, a form that adds an item, a fragment for YPY,
a `303` for a browser without the script, a `422` for an invalid form, `Vary: YPY-Request`, CSRF protection, and
Tailwind classes. Portuguese: [Receita: Go](../../pt-BR/recipes/go.md). The test for it is in
[Testing your app](../testing.md).

Other stacks: [Arandu](arandu.md) · [Django](django.md) · [Rails](rails.md) · [Laravel](laravel.md).

## What the server has to do

| The server | How |
|---|---|
| tells a YPY request from a page load | `ypy.FromRequest(r).IsYPY`, which reads `YPY-Request` |
| answers a fragment to YPY and a page to everything else | one handler, two branches |
| adds `Vary: YPY-Request` | `ypy.Middleware` does it on every response |
| answers an invalid form with `422` and the form | `422`, body is the form with its errors; with `ypy-error-swap="outerMorph"` on the form the whole `<form>` is the answer |
| answers a valid form with `303` | `ypy.SeeOther`; the client follows it and swaps the final page |
| refuses cross-site requests | `http.CrossOriginProtection` (Go 1.25 and later) |
| updates more than one place | `ypy.Partial` once per target, after the main fragment |

## The program

```sh
go get github.com/ypy-develop/go    # Go 1.26 or later
```

`main.go`:

```go
package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	ypy "github.com/ypy-develop/go"
)

var tpl = template.Must(template.New("").Parse(`
{{define "layout"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Items</title>
  <link rel="stylesheet" href="/app.css">
  <script src="/ypy.min.js"></script>
</head>
<body class="mx-auto max-w-lg p-6">
  <h1 class="text-2xl font-semibold">Items <span id="count" class="text-sm text-slate-500">{{len .Items}} items</span></h1>
  {{template "form" .}}
  <ul id="list" class="mt-4 space-y-1">{{range .Items}}{{template "row" .}}{{end}}</ul>
</body>
</html>{{end}}

{{define "form"}}<form id="new" ypy-post="/items" ypy-target="#list" ypy-swap="beforeend" ypy-error-swap="outerMorph"
      method="post" action="/items" class="flex flex-wrap gap-2 ypy-error:ring-1 ypy-error:ring-red-500">
  <input id="title" name="title" value="{{.Title}}" required
         {{if .Error}}aria-invalid="true" aria-describedby="title-error"{{end}}
         class="flex-1 rounded border p-2 aria-invalid:border-red-500" aria-label="New item">
  <button class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50">Add</button>
  {{if .Error}}<p id="title-error" role="alert" class="w-full text-sm text-red-700">{{.Error}}</p>{{end}}
</form>{{end}}

{{define "row"}}<li class="rounded border p-2">{{.}}</li>{{end}}
{{define "count"}}{{.}} items{{end}}
`))

type view struct {
	Items []string
	Title string
	Error string
}

type app struct {
	mu    sync.Mutex
	items []string
}

func (a *app) snapshot() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.items...)
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.ExecuteTemplate(w, "layout", view{Items: a.snapshot()})
}

func (a *app) create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		tpl.ExecuteTemplate(w, "form", view{Title: title, Error: "Write something first."})
		return
	}
	a.mu.Lock()
	a.items = append(a.items, title)
	n := len(a.items)
	a.mu.Unlock()

	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/")
		return
	}
	var row, count strings.Builder
	tpl.ExecuteTemplate(&row, "row", title)
	tpl.ExecuteTemplate(&count, "count", n)
	ypy.Fragment(w, row.String())
	ypy.Partial(w, ypy.PartialSpec{Target: "#count", Swap: "innerHTML"}, count.String())
}

func main() {
	a := &app{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("POST /items", a.create)
	mux.Handle("GET /ypy.min.js", http.FileServer(http.Dir("dist")))

	csrf := http.NewCrossOriginProtection()
	log.Fatal(http.ListenAndServe(":8081", csrf.Handler(ypy.Middleware(mux))))
}
```

Put `ypy.min.js` (from `dist/`) in a `dist` folder next to the binary, and run it:

```sh
go run .          # then open http://localhost:8081
```

## How it works

**The page.** `index` renders the layout with the whole list. A link or a reload always gets this, with or without the
script: that is progressive enhancement.

**The form.** `ypy-post="/items"` posts it; `ypy-target="#list" ypy-swap="beforeend"` appends the answer to the list;
`method` and `action` keep it working with no script. Writing `ypy-error-swap="outerMorph"` tells YPY that when the
server answers `422`, the answer is the whole `<form>`, and it is merged over the current one, so what the user typed
stays and the focus goes to the first invalid field ([Forms](../../guides/forms.md)).

**The valid case.** `create` answers the new row (`ypy.Fragment`) and, in the same response, a `<ypy-partial>` that
updates the counter in the heading ([Multiple targets](../../guides/multiple-targets.md)). A browser without the script
gets a `303` to `/` instead and reloads the page with the item in it.

**The `422`.** The form template is rendered again with the error; `aria-invalid` and `aria-describedby` tie the
message to the field, `role="alert"` makes a screen reader say it.

**`Vary`.** `ypy.Middleware` adds `Vary: YPY-Request`. The same URL (`/`) can answer a page or a fragment depending on
that header, and a cache must not mix them.

## CSRF

`http.NewCrossOriginProtection()` (standard library, Go 1.25 and later) rejects a state-changing request that the browser
reports as cross-site, with `403`, by reading `Sec-Fetch-Site` and falling back to `Origin`. It lets through
same-origin requests and requests from clients that send neither header (command-line tools), so it is a good first
layer and costs no markup. Wrap the whole mux as in the program above.

If you want a token as well (an API used by other origins, defense in depth, or an auditor who asks for one), send it
from a header and check it in a middleware:

```go
func verifyToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get("X-CSRF-Token")
		if got == "" {
			got = r.PostFormValue("_token") // a plain form post carries it in a hidden field
		}
		want := tokenFor(r) // yours: derived from the session, never a constant
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

and put the token in the layout, once, so every YPY request carries it:

```html
<meta name="csrf-token" content="{{ .CSRFToken }}">
<body ypy-headers:inherited='{"X-CSRF-Token": "{{ .CSRFToken }}"}'>
```

(`html/template` escapes the value for the context it is in.) A form with no script adds
`<input type="hidden" name="_token" value="{{ .CSRFToken }}">`.

## Tailwind

The classes in the templates must be scanned by the build, or a class that only appears in a fragment is removed from
the CSS. With Tailwind v4, in `app.css`:

```css
@import "tailwindcss";
@source "./**/*.go";          /* the templates live in Go strings here; point to .html or .tmpl files if you have them */

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
```

The two variants are what make `ypy-request:opacity-50` and `ypy-error:ring-red-500` in the template work:
[Styling the states](../../guides/core-concepts.md#styling-the-states).

## Other things a Go handler does

```go
ypy.Redirect(w, r, "/login")                     // YPY: YPY-Redirect + 204; browser: 303
ypy.Location(w, ypy.LocationOptions{Path: "/items/7", Target: "#main"}) // navigate without a full load
ypy.PushURL(w, "/items?page=2")                  // change the address bar
ypy.Retarget(w, "#errors"); ypy.Reswap(w, "innerHTML") // change where and how this answer goes
ypy.Trigger(w, "item:saved", map[string]any{"count": n}) // fire a client event
ypy.Refresh(w)                                   // reload the page
```

A history restore (the back button) arrives **without** `YPY-Request` and with `YPY-History-Restore-Request: true`:
answer a whole page. `ypy.FromRequest(r).HistoryRestore` says so. Streaming (`ypy.NewStream`, `ypy.Broadcaster`) is in
the SDK README.

## Checklist

- [ ] `ypy.Middleware` wraps the routes that can answer both a fragment and a page
- [ ] `422` for an invalid form, with the form and `ypy-error-swap="outerMorph"` (or only the fields)
- [ ] `303` after a `POST` that worked, for the browser without the script
- [ ] every value in a template is escaped by `html/template`; no user text in an attribute name
- [ ] CSRF protection wraps the mux; requests with no form carry the token header
- [ ] a test per audience: YPY, plain browser, invalid form ([Testing](../testing.md))
