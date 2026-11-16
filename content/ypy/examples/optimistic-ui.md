---
order: 240
title: Optimistic UI
description: Show a like, a comment or a deleted row before the server answers, and restore the page if the request fails, with ypy-optimistic.
---

# Optimistic UI

Some actions almost always succeed: a like, a new comment, a deleted row. The `ypy-optimistic` extension shows the
result at once and undoes it if the server answers with an error or does not answer. This page has three small cases.

## The page

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-optimistic.min.js"></script>
<style>.ypy-optimistic { opacity: .6; }</style>

<!-- 1. A like button turns into its liked state -->
<template id="liked"><button disabled>Liked</button></template>
<button ypy-post="/posts/7/like" ypy-target="this" ypy-swap="outerMorph"
        ypy-optimistic="#liked" ypy-status:5xx="swap:none">Like</button>

<!-- 2. A delete makes the row disappear -->
<ul id="items">
  <li id="item-3">Buy milk
    <button ypy-delete="/items/3" ypy-target="#item-3" ypy-swap="delete"
            ypy-optimistic="remove" ypy-status:5xx="swap:none">Delete</button>
  </li>
</ul>

<!-- 3. A comment shows up before it is saved -->
<ul id="comments"></ul>
<form ypy-post="/posts/7/comments" ypy-target="#comments" ypy-swap="beforeend" ypy-optimistic="new-comment">
  <textarea name="body" required></textarea>
  <button>Send</button>
</form>
<script>
  ypy.optimistic.register('new-comment', ({values}) => {
    let li = document.createElement('li');
    li.textContent = values.body; // text, never markup
    return li;
  });
</script>
```

In case 1 the button on screen is replaced by the template on the click. In case 2 the row is hidden. In case 3 a new `<li>`
with the typed text appears at the end of the list. The optimistic nodes carry the class `ypy-optimistic`, which
the style above uses for a pending look.

## The server (Go)

```go
func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	liked, err := h.store.ToggleLike(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, "could not save", http.StatusInternalServerError) // the page rolls back
		return
	}
	label := "Like"
	if liked {
		label = "Unlike"
	}
	ypy.Fragment(w, `<button ypy-post="/posts/`+html.EscapeString(r.PathValue("id"))+`/like" ypy-target="this" `+
		`ypy-swap="outerMorph" ypy-optimistic="#liked" ypy-status:5xx="swap:none">`+label+`</button>`)
}

func (h *Handler) Comment(w http.ResponseWriter, r *http.Request) {
	c, err := h.store.AddComment(r.Context(), r.PathValue("id"), r.PostFormValue("body"))
	if err != nil {
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	ypy.Fragment(w, `<li>`+html.EscapeString(c.Body)+`</li>`) // replaces the optimistic copy
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, "could not delete", http.StatusInternalServerError)
		return
	}
	ypy.Fragment(w, "") // 200 with an empty body: the delete swap removes the row (a 204 is never swapped)
}
```

## How it works

- The change is made right after `ypy:before:request`, after the queue let the request in and after any `ypy-confirm`.
  A request that is dropped or declined changes nothing.
- A `4xx` or `5xx` status, a network failure, a timeout or an abort puts the original nodes back (the same nodes, so an
  `<input>` keeps what was typed). Then `ypy:optimistic:rolledback` fires with the `reason`.
- On success, a swap into the same place first restores the originals and then puts the server's content in, so a morph
  keeps focus and typed text. When nothing is swapped there (`swap:none`, a `204`), the optimistic content stays as the final state.
- Write `ypy-status:5xx="swap:none"` on requests like these, or the error body is swapped in on top of the restored page.
  A `422` still re-renders a form in place.
- Every `ypy:optimistic:applied` is followed by one `ypy:optimistic:confirmed` or `ypy:optimistic:rolledback`.

## Variations

- `ypy-optimistic-target` and `ypy-optimistic-swap` say where and how the change goes when the response goes elsewhere.
- Show a message when it fails:

```js
document.addEventListener('ypy:optimistic:rolledback', evt => {
  document.querySelector('#notice').textContent = 'That did not go through (' + evt.detail.reason + ').';
});
```

- Retries are another extension: add [`ypy-retry`](../extensions/retry.md) for flaky networks.

## See also

- [`ypy-optimistic`](../extensions/optimistic.md)
- [Delete row](delete-row.md)
- [Forms](../guides/forms.md)
