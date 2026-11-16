---
order: 30
title: Progressive enhancement
description: Write links and forms that work with the script off, then add YPY attributes so the same elements update part of the page. Includes the Go handler.
---

# Progressive enhancement

Start with plain HTML. Links go to pages, forms post and redirect, and the server answers with a whole document. Then
add YPY, and the same links and forms update part of the page instead of loading a new one. If the script fails to
load, is blocked, or runs in an old browser, the user still gets the site.

## The pattern

Write the link or the form so it works alone, and put the YPY attributes next to what is already there.

```html
<form method="post" action="/items" ypy-post="/items" ypy-target="#app" ypy-swap="outerHTML">
  <input name="title" aria-label="New item">
  <button>Add</button>
</form>
```

Without the script, the browser sends the form using `method` and `action`. With YPY, the form sends the same request
through `ypy-post`, and the answer replaces `#app`.

Links work the same way if you keep the `href`:

```html
<a href="/items/7/edit" ypy-get="/items/7/edit" ypy-target="closest li" ypy-swap="outerHTML">Edit</a>
```

An element that is not a link or a form (a `<button>` with `ypy-delete`, a `<div>` with `ypy-get`) has no meaning in
plain HTML. Decide for each of those whether the action must work without the script. If it must, use a form with a
button.

## The server tells the two apart

A YPY request carries the header `YPY-Request: true`. The server uses it to choose between a fragment and a page, and
sends `Vary: YPY-Request` so a cache does not hand one out in place of the other. In Go:

```go
func (h *Items) Create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		h.renderForm(w, r, http.StatusUnprocessableEntity, title, "Write a title.")
		return
	}
	h.store.Add(title)
	// A YPY client gets YPY-Redirect, a browser gets 303 See Other.
	ypy.Redirect(w, r, "/items")
}

func (h *Items) renderForm(w http.ResponseWriter, r *http.Request, status int, title, msg string) {
	var buf bytes.Buffer
	if ypy.FromRequest(r).IsYPY {
		h.tmpl.ExecuteTemplate(&buf, "form", formView{Title: title, Error: msg}) // only the form
	} else {
		h.tmpl.ExecuteTemplate(&buf, "page", pageView{Form: formView{Title: title, Error: msg}}) // the whole page
	}
	ypy.FragmentStatus(w, status, buf.String())
}
```

Mount `ypy.Middleware` around the handler so every answer gets `Vary: YPY-Request`. The
[Go recipe](../server/recipes/go.md) shows the wiring.

## Post, redirect, get

The classic shape of a form still applies:

1. The form posts.
2. If the data is not valid, the server answers `422` with the form and the messages.
3. If it is valid, the server saves and answers with a redirect, so a reload does not post again.

YPY follows that shape. A `422` is swapped into the form and keeps what the user typed, and a `303 See Other` after a
`POST` is followed like any other success ([Forms](../guides/forms.md)). `ypy.Redirect` sends `303` to a browser and
the `YPY-Redirect` header to a YPY request.

## Boost: every link and form at once

If the links and forms already work, you can ask YPY to take all of them:

```html
<body ypy-boost="true">
```

`ypy-boost` makes the links and forms inside the element navigate with a request that swaps the body and updates the
address. The [`ypy-nav`](../extensions/nav.md) extension does it for the whole page with `config.nav.boost`. The
browser keeps handling the links it must: downloads, other origins, clicks with a modifier key and links to a
fragment of the same page.

A boosted request is an ordinary request, and the server decides whether to answer a fragment or the whole page. It
recognizes one by `YPY-Boosted: true`. A history restore (back or forward) arrives with
`YPY-History-Restore-Request: true` and without `YPY-Request`, and the server must answer it with a whole page.

## Test it

Turn the script off and use the site with the keyboard. Every action that matters must still work.

The browser tests of this project include a server-rendered site served under a policy that forbids scripts, which
shows the markup works alone. You can do the same for your own pages. In server tests, send a request with and
without `YPY-Request` and check both answers ([Testing your app](../server/testing.md)).

## Where there is no fallback

Some interactions mean nothing without script: a drag to reorder, a filter that never submits. Treat them as
enhancements, and give the user another way to the same result, such as an "Edit order" page with number fields or a
"Search" button inside `<noscript>`. The [examples](../examples/index.md) say where a fallback exists.

## See also

[Why hypermedia](why-hypermedia.md) · [Navigation](../extensions/nav.md) · [Forms](../guides/forms.md) ·
[Accessibility](../guides/accessibility.md)
