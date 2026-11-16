---
title: "ypy-inert"
description: "Make the target of a request inert while it runs: no clicks, no focus, no typing, so nobody edits a form that is about to be replaced."
order: 50
---
# ypy-inert

Makes the target of a request `inert` while the request runs. An inert element takes no click, no focus and no
typing, so a visitor cannot change a form or press a button of the region that is about to be replaced. When the
request ends the element is active again. `ypy-inert` and `data-ypy-inert` mean the same thing.

It is part of the core. No extension is needed.

```html
<form ypy-post="/comments" ypy-target="this" ypy-swap="innerHTML" ypy-inert="true" class="space-y-2">
  <textarea name="body" class="w-full rounded border p-2"></textarea>
  <button class="rounded bg-blue-600 px-3 py-1 text-white">Send</button>
</form>
```

The form targets itself. While `POST /comments` is in flight the whole form is `inert`: a second click on "Send" does
nothing and the visitor cannot edit the text that was already sent. The server answers with the new content of the
form; the response is swapped in and the form works again.

```go
func comments(w http.ResponseWriter, r *http.Request) {
	save(r)
	ypy.Fragment(w, `<textarea name="body" class="w-full rounded border p-2"></textarea>
<button class="rounded bg-blue-600 px-3 py-1 text-white">Send</button>`)
}
```

## Values

| Value | Meaning |
|---|---|
| `true` | the target is `inert` until the request ends |
| `false` | the target stays active; use it to turn off a `config.inert` that is on |

It is off by default. Any other value counts as off.

## Turning it on for the whole page

`config.inert` makes the target of every request inert, and `ypy-inert="false"` on an element opts that element out:

```html
<meta name="ypy-config" content='{"inert": true}'>

<button ypy-get="/slow" ypy-target="#panel" class="rounded border px-3 py-1">Load</button>
<button ypy-get="/ping" ypy-target="#status" ypy-inert="false" class="rounded border px-3 py-1">Ping</button>
<div id="panel"></div>
<div id="status"></div>
```

Only the first button makes its target inert.

## Inheritance

`ypy-inert` follows the same explicit rule as the other inherited attributes. Written on an element, it applies to
the requests that element sends. Written as `ypy-inert:inherited` on a container, it applies to the requests of every
descendant:

```html
<section ypy-inert:inherited="true">
  <button ypy-post="/a" ypy-target="#a">A</button>
  <button ypy-post="/b" ypy-target="#b">B</button>
</section>
```

A descendant with its own `ypy-inert="false"` wins.

## What it holds, and when it lets go

- The element held is the target of the request (`ypy-target`, or the element itself, or the body for a boosted
  link). It is the same element for the swap. It is released when the request ends, whether it ends in a swap or in an
  error status (both are covered by the tests).
- The release is counted. When two requests hold the same element, it stays inert until both have ended, and
  `ypy-disable` shares the same count.
- A request that has no target element holds nothing.
- The core sets the `inert` property. The attribute is not written to the markup, so there is nothing to clean up if
  the swap replaces the element.

## `ypy-inert` or `ypy-disable`

| | `ypy-disable="<selector>"` | `ypy-inert` |
|---|---|---|
| What it holds | the elements the selector finds | the target of the request |
| How | `disabled` on form controls | `inert` on any element, with its whole subtree |
| Screen readers | the controls are announced as disabled | the region is skipped, as if not there |

Use `ypy-disable` to dim one button. Use `ypy-inert` to freeze a region that is about to be replaced.

## Pitfalls

- An inert element and its children are hidden from assistive technology, and an element inside it that had the
  focus loses it. Pair it with [`ypy-focus`](ypy-focus.md) when the focus matters after the swap, and with an
  indicator (`ypy-indicator`) so a sighted visitor sees that something is happening.
- With `ypy-target="body"`, a boosted link makes the whole page inert for the wait. That is usually too much; prefer
  `ypy-progress` (see [Navigation](../../extensions/nav.md#progress-bar)) for navigations.
- It guards the target, not the source. A button outside the target can still be pressed; guard it with
  [`ypy-sync`](ypy-sync.md) so a second request is dropped or queued.
- `inert` is not security. It only stops a visitor's input; the server still has to refuse duplicate requests.

The tests are in `test/tests/attributes/ypy-form-submitter.js`, in the `inert` group.
