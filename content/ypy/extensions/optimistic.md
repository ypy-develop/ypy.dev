---
title: "Optimistic UI: the ypy-optimistic extension"
description: "Shows the result of an action before the server answers and puts the page back if the request fails: a like, a new comment, a deleted row."
order: 260
---
# Optimistic UI: the `ypy-optimistic` extension

Waiting for the server makes a click feel slow. `ypy-optimistic` shows the result of an action at once and puts
everything back if the request fails. A like button turns to "Liked" on the click, a new comment appears in the list,
a deleted row disappears; if the request fails, each of them is back as it was.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-optimistic.js"></script>

<template id="liked"><button disabled>Liked</button></template>

<button ypy-post="/posts/7/like" ypy-target="this" ypy-swap="outerMorph"
        ypy-optimistic="#liked" ypy-status:5xx="swap:none">Like</button>
```

On the click the button is replaced on screen by the template. When the server answers `200` with the new button,
the morph merges it into the original button (the optimistic copy is gone by then). When the server answers
`500`, the original button is back, untouched, and `ypy:optimistic:rolledback` fires.

If you set `ypy.config.extensions`, list `ypy-optimistic` in it, or the extension is refused. It is listed with the
other official extensions in [Extensions](index.md). Without it the attributes do nothing.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-optimistic` | `#tpl` · `remove` · a registered name | what to show (see below) |
| `ypy-optimistic-target` | a selector (`#count`, `closest li`, `next .badge`...) | where, instead of the target of the request |
| `ypy-optimistic-swap` | a swap style (`innerHTML`, `outerMorph`, `beforeend`, `delete`...) | how, instead of the swap of the request |
| `ypy-optimistic="false"` | | off for this element |

They are read from the element that sends the request, and with `ypy-optimistic:inherited` from a container.

What to show. A selector names a `<template>` (its content is cloned) or any element (its children are cloned);
`remove` shows the target gone; any other name is a function registered with `ypy.optimistic.register`
([below](#functions)).

Where and how. By default the content goes where the response goes, with the swap style of the request:

| Swap style | The optimistic content |
|---|---|
| `innerHTML`, `innerMorph`, `morph` (and anything else that is not listed here, `none` included) | replaces the children of the target |
| `outerHTML`, `outerMorph`, `outerSync` | is shown before the target, and the target is hidden |
| `beforebegin`/`before`, `afterend`/`after`, `afterbegin`/`prepend`, `beforeend`/`append` | is inserted there |
| `delete` | the target is hidden; there is nothing to show |

An outer style hides the target (`display:none`) instead of taking it out: it stays in the page, so the request that
came from it and the swap that is aimed at it still find it. Its own `display` value is restored when the change is
undone. When the response is not swapped into the place (`swap:none`, a `204`, another target) use
`ypy-optimistic-target` to say where the change goes.

The optimistic nodes carry the class `ypy-optimistic` (`.ypy-optimistic{opacity:.6}` is a pending look). They are
not processed while temporary: they are a picture of the next state, not a live part of the page. Ids in them can
repeat ids that are in the page; interactions come back with the server's content.

## What ends the change

Every `ypy:optimistic:applied` is followed by exactly one `confirmed` or `rolledback`.

| What happens | The page | Event |
|---|---|---|
| status 4xx or 5xx | everything put back, before the error body is swapped (a `422` form re-render still works) | `rolledback` `{reason:'status', status}` |
| network failure or timeout | put back | `rolledback` `{reason:'error', kind}` (`kind` is `ctx.errorKind`) |
| the request is aborted (`ypy:abort`, `ypy-sync="replace"`) | put back | `rolledback` `{reason:'aborted'}` |
| a `ypy:before:request` listener cancels the request | put back | `rolledback` `{reason:'cancelled'}` |
| success, and the response is swapped into the same element | the original nodes come back, then the server's content goes in (a morph keeps the identity of the original nodes: focus, typed text, open `<details>`) | `confirmed` `{status}` |
| success, and nothing is swapped there (`swap:none`, `204`, another target) | the optimistic content stays as the final state, loses its class and is processed | `confirmed` `{status}` |

Detail of every event: `{ctx, target}`, plus the fields above. They fire on the element that sent the request and
bubble. `ypy:optimistic:applied` is cancelable; cancelling it undoes the change at once (`rolledback` with
`reason:'cancelled'`).

The change is made right after `ypy:before:request`, that is after the queue let the request in and after any
`ypy-confirm`. A request that is queued, dropped by `ypy-sync`, or declined by the user leaves nothing behind and
fires no event.

"Put back" is exact: the original nodes return, not copies, so an `<input>` keeps what the user typed. If something
else changed the place meanwhile (another swap replaced the content), the original is not forced back over it;
only the optimistic nodes that are still there are removed.

## Concurrency

There is one overlay per element. When two requests change the same element:

- the second change replaces the first one (`inner` styles) or adds to it (`beforeend` and the other insertions);
  two different shapes on the same element (an `innerHTML` and an `outerHTML`) are not mixed: the second is not
  applied and fires no event;
- the original is put back when the last request that wanted the overlay has failed. A request that fails while
  another one is still in flight leaves the page as it is: what the newer request shows supersedes it;
- a success ends the overlay for everybody. A request that finishes later only fires its own event.

So with `ypy-sync="this:replace"` the first request is aborted and rolls back (`reason:'aborted'`) while the page keeps
showing the change of the second one; and with the default `queue` strategy the second click applies its change when
its turn comes, after the first one ended.

## Functions

For content that depends on what the user entered, register a function:

```js
ypy.optimistic.register('new-comment', ({values, source, target, ctx}) => {
  let li = document.createElement('li');
  li.textContent = values.body;          // text, never markup
  return li;
});
```

```html
<form ypy-post="/posts/7/comments" ypy-target="#comments" ypy-swap="beforeend" ypy-optimistic="new-comment">
  <textarea name="body"></textarea><button>Send</button>
</form>
```

`values` are the values of the request as an object (the last one wins for a repeated name). The function returns:

- a node, a `DocumentFragment` or a list of nodes: they are shown as described above;
- a string of HTML: it is built inert (scripts in it never run) through the `ypy` Trusted Types policy; it is
  ignored when the page's CSP forbids that policy. Escape user data yourself;
- a function: the change is yours, and the returned function is the way back. It is called on a rollback; on
  `confirmed` nothing is undone;
- nothing: no change, no event.

## A Go handler

The extension decides on the status alone: any `4xx` or `5xx` is a failure, any other status is a success.

```go
package posts

import (
	"net/http"

	ypy "github.com/ypy-develop/go"
)

type Handler struct {
	Toggle func(r *http.Request, id string) (liked bool, err error)
}

// Like answers POST /posts/{id}/like with the button as it is now.
func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	liked, err := h.Toggle(r, r.PathValue("id"))
	if err != nil {
		// the page rolls back; the body is not swapped (ypy-status:5xx="swap:none")
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	label := "Like"
	if liked {
		label = "Unlike"
	}
	ypy.Fragment(w, `<button ypy-post="/posts/`+r.PathValue("id")+`/like" ypy-target="this" `+
		`ypy-swap="outerMorph" ypy-optimistic="#liked" ypy-status:5xx="swap:none">`+label+`</button>`)
}

func Wire(mux *http.ServeMux, h *Handler) http.Handler {
	mux.HandleFunc("POST /posts/{id}/like", h.Like)
	return ypy.Middleware(mux)
}
```

An error body is swapped by the core like any other response unless you say otherwise; for an optimistic request
write `ypy-status:5xx="swap:none"` (or `ypy-error-swap`) so the rolled-back page is not covered by the error text. A
`422` is different: the form re-renders in place as usual, after the optimistic change was undone.

## Limits

- The extension changes the DOM only. It does not queue requests, retry them (use `ypy-retry`), or keep the change
  across a reload.
- A swap that has `swap:` or `settle:` delays leaves the original on screen during the delay, because the optimistic
  content is replaced when the response arrives, before the delay starts.
- The optimistic content of an `outer` style hides the target; a rule such as `display:block !important` on the
  target wins over it.
- Size: under 2 kB min+gzip (`ypy build --sizes`).
