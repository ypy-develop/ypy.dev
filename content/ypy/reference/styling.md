---
title: "Styling"
description: "YPY works with any CSS: no inline styles, only ypy- classes and data-ypy-state on your elements, and no stylesheet you did not ask for."
order: 130
---
# Styling

YPY works with any CSS. It writes no inline style, puts only `ypy-` classes and the `data-ypy-state`
attribute on your elements, and never loads a stylesheet you did not ask for. This page says what YPY
writes, how to style it with Tailwind CSS v4 (first class), Bootstrap 5, Bulma 1, Pico CSS or plain CSS,
and how the matrix of tests keeps all of them working.

Contents: [What YPY writes](#what-ypy-writes) · [ypy.css](#ypycss-optional) ·
[Tailwind CSS v4](#tailwind-css-v4) · [Bootstrap and Bulma](#bootstrap-and-bulma) ·
[Pico, UIkit and plain CSS](#pico-uikit-and-plain-css) · [The tests](#the-visual-matrix) ·
[What the core does not do yet](#what-the-core-does-not-do-yet)

## What YPY writes

| What | Where | Notes |
|---|---|---|
| classes of `config.classes` | `ypy-request`, `ypy-swapping`, `ypy-settling`, `ypy-added`, `ypy-indicator`, `ypy-offline`, `ypy-dirty` | all start with `ypy-`; each one can be renamed or emptied ([`config.classes`](config/classes.md)) |
| `data-ypy-state` | a space separated list of the states an element is in | select with `~=`: `[data-ypy-state~="error"]` |
| `aria-busy="true"` | the target of a request ([`ypy-busy`](attributes/ypy-busy.md)) | |
| `data-ypy-powered`, `data-ypy-live` | elements YPY processes, the live region of announcements | |
| inline style | never, by the core | the extensions that write `element.style` do it only for an attribute you wrote (`ypy-pending`, `ypy-live`, `ypy-signals`' `ypy-show`); a Go test lists them |
| a stylesheet | one constructable sheet that hides `.ypy-indicator` until a request runs | off with `config.includeIndicatorCSS = false` |

A test (`go test ./internal/styling`) scans the core and the extensions for every class they add and fails
when one has no `ypy-` prefix or is also a class of Bootstrap 5.3.9, Bulma 1.0.4 or Pico 2.1.1, or one of
the names that styles use for state (`active`, `show`, `fade`, `collapse`, `modal`, `is-loading`,
`is-active` and more). So `ypy-request` can sit next to Bootstrap's `.active` and Bulma's `.is-loading`
without touching them.

### The states

| State | Class (default) | `data-ypy-state` | Put on | While |
|---|---|---|---|---|
| `request` | `ypy-request` | not yet (see [below](#what-the-core-does-not-do-yet)) | the element that sends the request, or each element of `ypy-indicator` | the request is in flight; counted, so two requests keep it until the last ends |
| `swapping` | `ypy-swapping` | not yet | the swap target | from the start of the swap to the swap itself; it is long enough to see only with a swap delay (`ypy-swap="innerHTML swap:200ms"`) |
| `settling` | `ypy-settling` | not yet | the swap target | the settle delay, when an element of the response has the id of an element it replaces (its attributes then settle) |
| `added` | `ypy-added` | not yet | each new element of the swap | from `ypy:before:settle` to the end of `ypy:after:settle`; see the [hook](#what-the-core-does-not-do-yet) before building a transition on it |
| `error` | none (set `config.classes.error`) | `error` | the target of an error response (status 400 or more, or a 422 into its form) | until the source element sends its next request |
| `offline` | `ypy-offline` | `offline` | `<html>` | the browser reports no connection |
| `dirty` | `ypy-dirty` | `dirty` | a [signals](../extensions/signals.md) scope | a client-owned signal differs from the saved value |
| `busy` | none | none: `aria-busy="true"` | the target of a request | the request is in flight |

Style the attribute where you can: it keeps working when a class is renamed to an empty string
(`ypy.config.classes.request = ""`).

## ypy.css (optional)

```html
<link rel="stylesheet" href="/vendor/ypy.css">
```

`dist/ypy.css` (0.6 kB min+gzip as `ypy.min.css`) adds only this:

| Piece | What it does |
|---|---|
| `.ypy-indicator` | Hidden until the element, or one of its ancestors, is in the `request` state. The hidden indicator keeps its room, so nothing moves when it appears. It fades in and out. |
| `<div class="ypy-progress" role="progressbar">` | A fixed bar at the top of the page, invisible until it is in the `request` state, then it runs to 90 % and waits. The navigation extension puts the state on it. With your own script, use `bar.classList.add('ypy-request')`. |
| `ypy-fade` | Opt-in. The swap target fades out while it is in the `swapping` state. |
| `ypy-fade-in` | Opt-in. New content fades in. It uses `@starting-style`, so it works in Chrome 117, Firefox 129 and Safari 17.5 or later. |

With `prefers-reduced-motion: reduce` the durations are zero and the bar does not run.

It styles only `.ypy-*` classes and `:root`, so no element of yours is touched. Theme it with custom
properties, on `:root` or on any element:

| Property | Default | What |
|---|---|---|
| `--ypy-duration` | `200ms` | every fade |
| `--ypy-easing` | `ease-in-out` | every fade |
| `--ypy-progress-color` | `AccentColor` | colour of the bar |
| `--ypy-progress-height` | `3px` | thickness of the bar |
| `--ypy-progress-duration` | `1.2s` | time the bar takes to reach 90 % |
| `--ypy-progress-z` | `9999` | stacking of the bar |

```css
:root { --ypy-progress-color: #0284c7; }
.checkout { --ypy-duration: 500ms; }
```

The selectors use the default class names. If you rename a class in `config.classes`, copy the rule you
use and change the name (the attribute selectors keep working for the states that have one).

## Tailwind CSS v4

`dist/ypy-tailwind.css` declares the variants. It holds only `@custom-variant` lines (1.9 kB), no utilities
and no theme, so it adds nothing to your CSS but the variants you use.

```css
/* app.css */
@import "tailwindcss";
@import "./ypy-tailwind.css";   /* copy dist/ypy-tailwind.css next to it */
```

| Variant | True when |
|---|---|
| `ypy-request:` | the element is in the `request` state (it sends the request, or it is the `ypy-indicator`) |
| `ypy-swapping:` | the element is the swap target, mid swap |
| `ypy-settling:` | the element is the swap target, settling |
| `ypy-added:` | the element is new content of a swap that has not settled |
| `ypy-error:` | the element is the target of an error response |
| `ypy-offline:` | the page is offline; it matches `<html>` and everything under it |
| `ypy-dirty:` | the element is a signals scope with unsaved changes |
| `ypy-busy:` | `aria-busy="true"`, the target of a running request |

Every one of them matches the class or `data-ypy-state`, so a class you renamed does not break it as
long as the state has an attribute. They compose like the built-in variants: `md:ypy-request:block`,
`not-ypy-error:flex`, `hover:ypy-error:underline`. To style a child from the state of its parent, use
`group` or `in-*`:

```html
<form ypy-post="/signup" class="group max-w-sm space-y-4">
  <input name="email" class="w-full rounded border px-3 py-2 aria-invalid:border-red-500">
  <button class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50">
    Create account
    <span class="hidden group-ypy-request:inline">...</span>
  </button>
</form>

<div id="panel" class="rounded border-2 border-gray-300 p-4 ypy-error:border-red-500">...</div>

<div class="hidden rounded bg-amber-100 p-3 ypy-offline:block">You are offline.</div>

<span class="size-4 animate-spin rounded-full border-2 opacity-0 transition-opacity ypy-request:opacity-100"
      id="spin"></span>
<button ypy-get="/slow" ypy-indicator="#spin">Load</button>
```

`ypy-indicator="#spin"` puts the `request` state on `#spin`, not on the button, so the variant is on the
spinner.

### `@source` for server templates

Tailwind keeps only the classes it finds. A fragment the server renders is not in your JavaScript, so tell
it where the templates are. Paths are relative to the CSS file.

```css
@import "tailwindcss";
@import "./ypy-tailwind.css";

/* Go html/template: .html, .gohtml, .tmpl */
@source "../templates";

/* templ: scan the .templ files, not the generated *_templ.go (git ignores them, so Tailwind does too) */
@source "../../components/**/*.templ";

/* classes written in Go code, for example the components of a Go UI library such as Kyse */
@source "../../internal/ui/**/*.go";

/* a class the server builds from pieces cannot be found: list it */
@source inline("ypy-error:bg-red-100 ypy-error:border-red-500");
```

- By default Tailwind scans from the directory you run it in, skipping what `.gitignore` lists. `.html`,
  `.gohtml`, `.templ` and `.go` files are all read. Write the `@source` lines anyway: they say what you
  mean, and they keep working when the build runs from another directory. With
  `@import "tailwindcss" source(none);` only the `@source` lines count.
- Write complete class names. `class="ypy-{{.State}}:opacity-50"` is not found; `class="ypy-request:opacity-50"`
  is. A Go library outside your project (a module in the module cache) is not scanned unless you add its
  directory: have your build script write the output of `go list -m -f '{{.Dir}}' <module>` into an
  `@source` line.
- A `<ypy-partial>` or any fragment returned by a handler is scanned like any other template file.

No Node is needed: Tailwind ships a standalone executable. The matrix of this repository is built with it
(`test/fixtures/css/README.md`).

## Bootstrap and Bulma

YPY and these frameworks write different things: they toggle `show`, `active`, `is-active`, `collapsing`,
`style`, `aria-*`; YPY writes `ypy-*` and `data-ypy-state`. Two things are left to you.

**1. Start components after a swap: `ypy.onLoad(selector, fn)`.** It calls `fn(element)` once for each
element that matches in the content YPY processes, on the first page load and on every fragment, and never
twice for the same node, so a morph that keeps an element does not create a second component on it.

```js
// Bootstrap 5
ypy.onLoad('.modal', el => bootstrap.Modal.getOrCreateInstance(el));
ypy.onLoad('[data-bs-toggle="tooltip"]', el => new bootstrap.Tooltip(el));
ypy.onLoad('[data-bs-toggle="dropdown"]', el => new bootstrap.Dropdown(el));
ypy.onLoad('.toast', el => bootstrap.Toast.getOrCreateInstance(el).show());

// Bulma 1 has no script of its own: the page toggles is-active
ypy.onLoad('.navbar-burger', el => el.addEventListener('click', () => {
  el.classList.toggle('is-active');
  document.getElementById(el.dataset.target)?.classList.toggle('is-active');
}));
ypy.onLoad('.modal .delete, .modal .modal-background', el =>
  el.addEventListener('click', () => el.closest('.modal').classList.remove('is-active')));
ypy.onLoad('.notification .delete', el =>
  el.addEventListener('click', () => el.parentNode.remove()));
```

Elements that leave the page can be cleaned on `ypy:before:cleanup` (for Bootstrap,
`bootstrap.Modal.getInstance(el)?.dispose()`).

**2. Keep what the component writes: `ypy-preserve`.** A morph that brings the server's version of an open
modal would remove the `show` class the framework added. List what the script owns
([`ypy-preserve`](attributes/ypy-preserve.md), [third-party components](../guides/third-party-components.md)):

```html
<!-- Bootstrap: Modal.show() writes class, style and aria-* -->
<div id="confirm" class="modal fade" tabindex="-1" ypy-preserve-attrs="class,style,aria-*,role">...</div>

<!-- Bulma: the page toggles is-active -->
<div id="confirm" class="modal" ypy-preserve-attrs="class">...</div>
```

**State styled with the framework's own tokens.** `data-ypy-state` is plain CSS:

```css
/* Bootstrap */
#panel[data-ypy-state~="error"] { border-color: var(--bs-danger) !important; }
[data-ypy-state~="offline"] .offline-banner { display: block; }

/* Bulma */
#panel[data-ypy-state~="error"] { border: 2px solid var(--bulma-danger); }
```

Bootstrap's `.is-invalid`, Bulma's `.is-danger` and `.is-loading` stay yours: send them in the HTML of the
response (the 422 convention swaps it into the form), as the matrix does.

## Pico, UIkit and plain CSS

**Pico CSS.** It is nearly classless: write the state in an attribute selector (`#panel[data-ypy-state~="error"]
{ border: 2px solid var(--pico-del-color); }`). Pico draws a spinner on every `[aria-busy="true"]` element,
and YPY sets that attribute on the target of a request. That is a ready-made loading spinner; to drop it, see
[`ypy-busy`](attributes/ypy-busy.md#pico-css). A `<dialog>` opened with `showModal()` stays open through a morph with
`ypy-preserve-attrs="open"`.

**UIkit and any other.** UIkit toggles `uk-open` and `uk-active`; Foundation, `is-active`; none of them uses
a `ypy-` name. Use `ypy.onLoad` to start the component and `ypy-preserve-attrs` for the classes it toggles.

**Plain CSS.** Everything above is a selector:

```css
[data-ypy-state~="request"] { cursor: progress; }
button[data-ypy-state~="request"], button.ypy-request { opacity: .6; }
html[data-ypy-state~="offline"] .offline-banner { display: block; }
.ypy-indicator { opacity: 0; }
.ypy-request > .ypy-indicator { opacity: 1; }
```

## The visual matrix

`test/tests/css/matrix.js` runs five scenarios against Tailwind v4, Bootstrap 5.3.9, Bulma 1.0.4, Pico 2.1.1
and plain CSS (`test/fixtures/css/<name>/fixture.js` holds the markup of each):

| Scenario | Checks |
|---|---|
| form | a 422 renders inside the form; the invalid field is drawn red by the framework; the message is visible and below the field; the field still fills the form |
| list | items appended with `beforeend` have the computed style of the old ones, stack and align, carry no class, attribute or inline style of YPY |
| modal | the closed modal is `display: none`; opened by the page script it is a fixed layer that covers the viewport; a morph that brings it again leaves it open (`ypy-preserve-attrs`); `ypy.onLoad` starts it once |
| indicator | it fades in during the request and out after it; `aria-busy` is on the target; the button, the indicator and their neighbours do not move; the swapping state is on the target during the swap delay |
| error | the target of a 500 gets the error state, styled in the framework's colour, cleared by the next request; the offline banner shows only while the page is offline |

Plus `tailwind-variants.js` (each variant, by class and by attribute, group, `in-`, `not-`, a breakpoint, and
a real request), `ypy-css.js` (the optional sheet and its properties) and `states.js` (where the core puts each
state). Run them in each browser:

```sh
go run ./cmd/ypy test --files 'test/tests/css/**/*.js'
go run ./cmd/ypy test --browsers firefox --files 'test/tests/css/**/*.js'
```

The sheets are the official files, kept in `test/fixtures/css` with their license, version and SHA-256; nothing
is loaded from a CDN. The runner measures layout and computed style; it does not take screenshots.

## What the core does not do yet

The tests of `states.js` whose title starts with `hook:` skip until the core has these, and then they assert.
Each hook is small and touches only `src/ypy.js`.

1. `data-ypy-state` for `request`, `swapping`, `settling` and `added`. `__addClass(elt, state)` and
   `__removeClass(elt, state)` put the class only. Make them keep the attribute too,
   with the same list logic as `__setState` (add the token on `__addClass`, remove it on `__removeClass`).
   `__showIndicators`, `__insertContent` and `__afterInsert` already call them, so nothing else changes. With
   it, `ypy-request:` and the other variants work after a class is renamed to an empty string.
2. Defaults for `error`, `dirty` and `offline` in `config.classes`. `__initConfig` lists `request`,
   `swapping`, `settling`, `added` and `indicator`. Add `error: "ypy-error"`, `offline: "ypy-offline"` and
   `dirty: "ypy-dirty"`: today `error` has no class, `offline` is set the first time the page goes offline,
   and `dirty` is a fallback in the signals extension. The variants and `ypy.css` already read the class.
3. `ypy-added` lasts a frame. `__afterInsert` removes the class in the same task it adds it unless the
   swap has settle tasks (an element with the same id), so no frame is ever painted with it, and a CSS
   transition cannot start from it. Wait for `settleDelay` (and one frame, or a default of about 20 ms
   instead of 1 ms) while `added` is set, for every swap with new content. Until then use `ypy-fade-in` of
   `ypy.css`, which starts from `@starting-style` and does not need the class.

The progress bar and the layers belong to the navigation and layers extensions. `ypy.css` styles a
`.ypy-progress` element by the same `request` state; a layer has no style here until that extension fixes
its class names.
