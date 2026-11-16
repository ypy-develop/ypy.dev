---
title: "Getting started"
description: "Build a page where a button asks the server for HTML and the answer replaces part of the page, first with static files, then with a Go handler."
order: 10
id: "GUIDE-getting-started"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en"]
---

# Getting started

This guide builds a page where a button asks the server for HTML and the answer replaces part of the page. You can do
it two ways:

- [With plain HTML](#path-a-plain-html): a folder of static files and no server code.
- [With Go or Arandu](#path-b-go-and-arandu): a real handler that answers fragments and forms.

When you finish, [what to read next](#what-to-read-next) points to the next pages. There is also a Portuguese version,
[Primeiros passos](../pt-BR/guides/getting-started.md).

## What YPY is

With YPY, an HTML element says in its attributes which request to make, when to make it, and where to put the answer.
The server answers with HTML, not JSON. You add one `<script>` tag; there is nothing to build.

```html
<button ypy-get="/hello" ypy-target="#out"
        class="rounded bg-sky-600 px-4 py-2 text-white">Say hello</button>
<p id="out" class="mt-4 text-slate-700"></p>
```

Read it as a sentence: on click, `GET /hello`, put the answer into `#out`. Without the script the button does nothing,
so real navigation (links and forms) should work without it; see
[Progressive enhancement](core-concepts.md#progressive-enhancement).

## Get the file

The build writes the files you serve into `dist/`:

```sh
go run ./cmd/ypy build      # needs Go 1.26 or later; writes dist/
```

| File | Use it when |
|---|---|
| `dist/ypy.min.js` | a new project |
| `dist/ypy.max.min.js` | a new project that wants the official extensions (SSE, WebSocket, signals, preload, and more) |
| `dist/ext/*.min.js` | one extension at a time, after the core: see [Extensions](../extensions/using-and-writing.md) |
| `dist/ypy.d.ts` | TypeScript types for `window.ypy` |

Copy the file you need next to your pages, or load the core from a CDN while you try things out:

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
```

`npm install @ypy-dev/ypy@0.1.1` installs the same `dist/` files. See [Installing](../index.md#installing).

## Path A: plain HTML

You need only a folder and any static file server. Two files:

`index.html`

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>YPY, first page</title>
  <script src="/ypy.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script> <!-- any CSS works; Tailwind is only the example -->
</head>
<body class="mx-auto max-w-md p-8 font-sans">
  <h1 class="text-2xl font-semibold">Hello, YPY</h1>

  <button ypy-get="/hello.html" ypy-target="#out"
          class="mt-4 rounded bg-sky-600 px-4 py-2 text-white">Say hello</button>

  <div id="out" class="mt-4"></div>
</body>
</html>
```

`hello.html` is a fragment: no `<html>`, no `<head>`, just the piece.

```html
<p class="rounded border border-sky-200 bg-sky-50 p-3">Hello from a static file.</p>
```

Serve the folder and open it:

```sh
python3 -m http.server 8080     # or any static server; then open http://localhost:8080
```

Click the button. The paragraph appears in `#out` and the address bar does not change. A static file cannot tell a YPY
request from a normal one, so a link to `/hello.html` would open the bare fragment. That is fine for a first test. A
real server avoids it by reading the `YPY-Request` header, as in the next path.

> The Tailwind tag above is the Play CDN, which is meant for trying things. In a real project, build your CSS and add the
> template folders to `@source`; see [Core concepts](core-concepts.md#styling-the-states).

## Path B: Go and Arandu

The Go SDK (`github.com/ypy-develop/go`) is the server side of the same conversation. It needs Go 1.26 or later and
no other dependency.

```sh
go get github.com/ypy-develop/go
```

Here is a complete program that uses only the standard library and the SDK. It serves the page, the script, and one
endpoint that answers a YPY request and a plain browser differently:

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"

	ypy "github.com/ypy-develop/go"
)

const page = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Items</title>
  <script src="/ypy.min.js"></script>
</head>
<body class="mx-auto max-w-md p-8">
  <form ypy-post="/items" ypy-target="#list" ypy-swap="beforeend" method="post" action="/items" class="flex flex-wrap gap-2">%s</form>
  <ul id="list" class="mt-4 space-y-1">%s</ul>
</body>
</html>`

// fields is the inside of the form, with an error when there is one. The page and the 422 answer share it.
func fields(value, problem string) string {
	invalid, message := "", ""
	if problem != "" {
		invalid = ` aria-invalid="true" aria-describedby="title-error"`
		message = `<p id="title-error" role="alert" class="w-full text-sm text-red-700">` + html.EscapeString(problem) + `</p>`
	}
	return `<input id="title" name="title" value="` + html.EscapeString(value) + `" required` + invalid +
		` class="flex-1 rounded border p-2 aria-invalid:border-red-500" aria-label="New item">` + message +
		`<button class="rounded bg-sky-600 px-4 py-2 text-white">Add</button>`
}

type app struct {
	mu    sync.Mutex
	items []string
}

func (a *app) list() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var sb strings.Builder
	for _, it := range a.items {
		sb.WriteString(row(it))
	}
	return sb.String()
}

func row(title string) string {
	return `<li class="rounded border p-2">` + html.EscapeString(title) + `</li>`
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, page, fields("", ""), a.list())
}

func (a *app) create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		ypy.ValidationError(w, fields(title, "Write something first.")) // the inside of the form, with its error
		return
	}
	a.mu.Lock()
	a.items = append(a.items, title)
	a.mu.Unlock()

	if !ypy.FromRequest(r).IsYPY {
		ypy.SeeOther(w, "/") // no script: post, redirect, get
		return
	}
	ypy.Fragment(w, row(title)) // goes into #list, at the end
}

func main() {
	a := &app{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.index)
	mux.HandleFunc("POST /items", a.create)
	mux.Handle("GET /ypy.min.js", http.FileServer(http.Dir("dist")))

	log.Fatal(http.ListenAndServe(":8080", ypy.Middleware(mux)))
}
```

Three details in it apply well beyond this example.

1. One endpoint gives two answers. `ypy.FromRequest(r).IsYPY` is true when the YPY client made the request. A browser
   with no script, or a crawler, gets a redirect and a whole page, and YPY gets the fragment.
2. A form that is not valid gets a `422`, and one that worked gets a `303`. YPY puts the body of a 422 inside the form
   and follows a 303 like any success. See [Forms](forms.md).
3. `ypy.Middleware` adds `Vary: YPY-Request`. Without it a shared cache can hand the fragment to a page load.

The same handler in the other stacks is in the [recipes](../server/recipes/go.md). To see what the browser sends and
receives, open the developer tools: the request carries `YPY-Request: true`, `YPY-Target: ul#list` and
`YPY-Request-Type: partial`.

### In an Arandu application

An Arandu application wires the same handler where the application is assembled, usually `bootstrap/app.go`:

```go
router.Post("/items", items.Create, ypy.Middleware)
```

Hesape's router takes plain handlers and plain middleware, so nothing else changes. CSRF, sessions and the layout
come from Hesape; see the [Arandu recipe](../server/recipes/arandu.md).

## A project written before YPY

If you are moving a project that already has hypermedia markup from the older library, skip this guide. `ypy migrate`
converts the markup and points the script tag at `ypy.min.js`; see [Migration](../migration/index.md).

## What to read next

| You want to | Read |
|---|---|
| understand the vocabulary | [Core concepts](core-concepts.md) |
| move a project written before YPY | [Migration](../migration/index.md) |
| react to what happens in the page | [Events](../reference/events/index.md) |
| handle forms and validation | [Forms](forms.md) |
| make it safe | [Security checklist](security-checklist.md) |
| make it fast | [Performance](../essays/performance.md) |
| test the server side without a browser | [Testing your app](../server/testing.md) |
| copy a working server | [Recipes](../server/recipes/go.md): Go, Arandu, Django, Rails, Laravel |
| find a page for any task | [Documentation index](../index.md) |
