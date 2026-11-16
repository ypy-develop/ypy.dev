---
order: 30
title: Progressive enhancement
description: Write links and forms that work with the script off, then add YPY attributes so the same elements update part of the page instead of reloading it.
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

A YPY request carries the header `YPY-Request: true`. The server uses it to choose between a fragment and a page:

```http
POST /items HTTP/1.1
YPY-Request: true
YPY-Target: div#app

HTTP/1.1 200 OK
Vary: YPY-Request
Content-Type: text/html

<div id="app">...</div>
```

The same URL, asked by a browser without the script, gets the whole page. `Vary: YPY-Request` on both answers stops a
cache from handing out one in place of the other.

## Post, redirect, get

The classic shape of a form still applies. The form posts. If the data is not valid, the server answers `422` with the
form and the messages. If it is valid, the server saves and redirects, so a reload does not post again.

YPY follows that shape. A `422` is swapped into the form and keeps what the user typed, and a `303 See Other` after a
`POST` is followed like any other success. A server can send `YPY-Redirect` to a YPY request and a plain `303` to a
browser, so one handler serves both.

## Boost: every link and form at once

If the links and forms already work, you can ask YPY to take all of them:

```html
<body ypy-boost="true">
```

`ypy-boost` makes the links and forms inside the element navigate with a request that swaps the body and updates the
address. The browser keeps handling the links it must: downloads, other origins, clicks with a modifier key and links
to a fragment of the same page. See [Boosting](../index.md#boosting).

A boosted request is an ordinary request, and the server decides whether to answer a fragment or the whole page. It
recognizes one by `YPY-Boosted: true`. A history restore (back or forward) arrives with
`YPY-History-Restore-Request: true` and without `YPY-Request`, and the server must answer it with a whole page.

## Test it

Turn the script off and use the site with the keyboard. Every action that matters must still work. The browser tests of
the YPY repository include a server-rendered site served under a policy that forbids scripts, which shows the markup
works alone. You can do the same for your own pages, and check each route with and without `YPY-Request`.

## Where there is no fallback

Some interactions mean nothing without script: a drag to reorder, a filter that never submits. Treat them as
enhancements, and give the user another way to the same result, such as an "Edit order" page with number fields or a
"Search" button inside `<noscript>`. The [examples](../examples/index.md) say where a fallback exists.

## See also

[Why hypermedia](why-hypermedia.md) · [Boosting](../index.md#boosting) · [Validation](../index.md#validation) ·
[`ypy-nav`](../extensions/nav.md)
