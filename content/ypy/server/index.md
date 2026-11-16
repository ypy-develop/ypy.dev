---
order: 80
title: "Server"
description: "The official Go SDK for YPY servers: install it, wire the middleware, and find the page for requests, streaming, signals, CSRF, assets and tests."
---

# Server

A YPY page asks for HTML and the server answers with HTML. Any server can do that, since the requests are plain HTTP
with a few extra headers. This section covers Go and the official SDK, `github.com/ypy-develop/go`. The SDK reads the
YPY headers, writes the answers, streams changes to the page and lets you test all of it without a browser. It
depends only on the standard library and `net/http`.

## Install

```sh
go get github.com/ypy-develop/go
```

It needs Go 1.26 or later. Nothing is registered globally and there is no container. You wire the pieces where you
build the application.

## The smallest server

```go
package main

import (
	"net/http"

	ypy "github.com/ypy-develop/go"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		ypy.Fragment(w, `<p>Hello from the server</p>`)
	})

	// Middleware reads the YPY headers once and adds "Vary: YPY-Request" to every response.
	http.ListenAndServe(":8080", ypy.Middleware(mux))
}
```

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<!-- or: https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js -->

<button ypy-get="/hello" ypy-target="#out">Say hello</button>
<div id="out"></div>
```

`ypy.Middleware` has the standard shape (`func(http.Handler) http.Handler`), so it fits any router. Inside a handler,
`ypy.FromRequest(r)` or `ypy.FromContext(r.Context())` tells you whether the YPY client made the request and what it
expects back.

## What the SDK covers

| Page | What it is about |
|---|---|
| [Requests and responses](requests-and-responses.md) | Reading the request, answering fragments and partials, and the `YPY-*` response headers |
| [Streaming](streaming.md) | Server-sent events to one client or to many, with `Stream` and `Broadcaster` |
| [Signals](signals.md) | Reading and patching signals, and signing the ones the page must not change |
| [CSRF](csrf.md) | A token bound to the session, sent by the page on every unsafe request |
| [Assets and CSP](assets.md) | CSS and scripts next to templates, hashed, with `integrity` and the page nonce |
| [`ypytest`](ypytest.md) | The test package: requests, assertions and streams, with no browser |
| [Testing your app](testing.md) | A walk through the tests of a real handler, forms and streams |
| [Streaming a model answer](model-streaming.md) | Markdown from a language model, written into the page as it grows |
| [Layers with the Go SDK](layers.md) | Answer a layer, open one from a response, close one with an event |

## Recipes by stack

The same handlers written for other servers. The headers are identical, so what you learn on one carries over to the
others.

| Recipe | Stack |
|---|---|
| [Go](recipes/go.md) | `net/http` |
| [Arandu](recipes/arandu.md) | Hesape and Kyse |
| [Django](recipes/django.md) | Python |
| [Rails](recipes/rails.md) | Ruby |
| [Laravel](recipes/laravel.md) | PHP |

## More

- The [SDK guide](../../go/guide.md) and the [SDK documentation index](../../go/README.md) in the Go section of this
  site, with the helpers listed one by one.
- The API reference is the package documentation: `go doc github.com/ypy-develop/go`.
- The wire format of streaming and signed signals is in `spec/PROTOCOL.md` of the SDK repository.
- The headers the client sends and reads are listed in [Headers](../reference/headers.md).
