---
order: 40
title: Template fragments
description: Define a row, a form or a list once as a named block of your template, then draw it inside the full page and alone as the answer to a YPY request.
---

# Template fragments

A YPY response is often a piece of a page: a row, a form, a list of results. If you write that piece twice, once in
the page template and once for the response, the two copies drift apart. A fragment is a named part of a template. The
page uses it, and the handler can also render it alone.

## The idea

Define the piece once, with a name. The page includes it, and the handler for a YPY request renders only that name.
Every template engine has a way to name a block; the syntax differs, the idea does not.

Take a list with a form that adds to it:

```html
<form method="post" action="/items" ypy-post="/items" ypy-target="#list" ypy-swap="beforeend">
  <input name="title" aria-label="New item"><button>Add</button>
</form>
<ul id="list">
  <!-- block "item", repeated for every item -->
  <li id="item-7">Milk</li>
</ul>
```

The page draws the block `item` for every row. The response to the form draws it once:

```http
POST /items HTTP/1.1
YPY-Request: true

HTTP/1.1 200 OK
Vary: YPY-Request
Content-Type: text/html

<li id="item-8">Bread</li>
```

The `<li>` in the answer is the same markup the page would have drawn for that row, because it comes from the same
block. A change to the row is made in one place.

## Several pieces in one answer

A response can fill more than one place. Put each extra piece in a `<ypy-partial>` with its own target:

```html
<li id="item-8">Bread</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">8 items</ypy-partial>
```

The first piece goes to the target of the request, and the partial goes to `#count`. See
[Out-of-band swaps](../index.md#out-of-band-swaps).

## Four habits that help

Give a swapped element an `id`. A morph and a partial match on it, and `ypy-target` can point to it.

Keep the fragment self-contained, so it takes its data from its arguments and not from variables of the page.

Leave the page shell out of fragments. If a piece must also work as a full page, the browser path renders the whole
template and the YPY path renders the block.

Send `Vary: YPY-Request` on a URL that answers both, and keep a fragment out of search engines. A fragment is half a
page and should not be what a result links to.

## See also

[Progressive enhancement](progressive-enhancement.md) · [Examples](../examples/index.md) ·
[Server side](../index.md#server-side)
