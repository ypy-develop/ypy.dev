---
order: 150
title: Reset the form after a save
description: Empty a comment form after a successful save and keep the typed text after a rejection, with a partial that swaps in a fresh form.
---

# Reset the form after a save

A comment form adds the comment to a list. After a successful save the fields must be empty again, and after a rejected
one they must keep what the visitor typed. The simplest way is to let the answer carry a fresh form.

## The page

```html
<ul id="comments"></ul>

<form id="new-comment" ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend"
      ypy-error-target="this" ypy-error-swap="innerMorph">
  <label for="body">Comment</label>
  <textarea id="body" name="body"></textarea>
  <button>Post</button>
</form>
```

## What the server answers

| Case | Status | Body |
|---|---|---|
| Saved | `200` | The new `<li>`, then a `<ypy-partial>` that holds an empty form |
| Empty comment | `422` | The fields again, with the typed text and a message |

On success the answer has two parts. The `<li>` goes to the target of the request. The partial has the `id` of the form
and the swap `outerHTML`, so an empty form replaces the old one:

```html
<li>Delivery was quick and the packaging was good.</li>

<ypy-partial id="new-comment" ypy-swap="outerHTML">
  <form id="new-comment" ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend"
        ypy-error-target="this" ypy-error-swap="innerMorph">
    <label for="body">Comment</label>
    <textarea id="body" name="body"></textarea>
    <button>Post</button>
  </form>
</ypy-partial>
```

On an error the answer is the inside of the form, with what was typed written back, escaped:

```html
<label for="body">Comment</label>
<p id="body-error" role="alert">Write something first.</p>
<textarea id="body" name="body" aria-describedby="body-error" aria-invalid="true">Delivery</textarea>
<button>Post</button>
```

Every answer is a fragment with `Vary: YPY-Request`.

## How it works

- `ypy-swap="beforeend"` appends the comment to the list. The main target is `#comments`, not the form.
- The partial has no `ypy-target`, so its `id` names the destination: the element with that id is replaced by the
  fresh form. The browser gets empty fields.
- On an error the server answers `422`. `ypy-error-target="this"` and `ypy-error-swap="innerMorph"` morph the fields
  into the form, so the message appears and the form itself is not replaced.
- The page and the answer draw the same form. Put it in one template that both use.

## Variations

- A listener in your own script file resets the form without a second copy of it. Add it to a file the page loads:

  ```js
  document.addEventListener('ypy:after:request', (event) => {
    const { ctx } = event.detail;
    if (ctx.sourceElement.matches('#new-comment') && ctx.response.status < 400) ctx.sourceElement.reset();
  });
  ```

- Only one field: answer with a `<ypy-partial id="body" ypy-swap="outerHTML">` that holds an empty `<textarea>`, and
  the rest of the form stays.

## See also

- [Updating other content](updating-other-content.md) and [`<ypy-partial>`](../index.md#out-of-band-swaps)
- [Validation](../index.md#validation)
- [Inline validation](inline-validation.md)
