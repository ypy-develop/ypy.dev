---
title: "Core concepts"
description: "The model in one page: attributes on an element, a trigger, a request, an HTML answer, a swap into the target, and the events that fire at each step."
order: 20
id: "GUIDE-core-concepts"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en"]
---

# Core concepts

Everything YPY does follows one sequence. An element carries attributes. A trigger fires, YPY makes a request, the
server answers with HTML, and a swap puts that HTML into the page. An event fires at every step. There is also a
Portuguese version: [Conceitos centrais](../pt-BR/guides/core-concepts.md).

Contents: [Attributes](#attributes) · [Triggers](#triggers) · [Targets](#targets) · [Swaps and morph](#swaps-and-morph) ·
[Partials and out-of-band swaps](#partials-and-out-of-band-swaps) · [Requests and headers](#requests-and-headers) ·
[Inheritance](#inheritance) · [Defaults](#defaults) · [Events](#events) · [Styling the states](#styling-the-states) ·
[Progressive enhancement](#progressive-enhancement)

## Attributes

Every attribute is written `ypy-<name>`. The `data-ypy-<name>` spelling means exactly the same (use it if your
HTML validator complains). If one element has both, `ypy-*` wins.

| Group | Attributes | What they say |
|---|---|---|
| Request | `ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` | the verb and the URL |
| When | `ypy-trigger` | which event, and its modifiers |
| Where | `ypy-target`, `ypy-select`, `ypy-swap`, `ypy-swap-oob`, `ypy-select-oob` | where the answer goes and how |
| What is sent | `ypy-vals`, `ypy-include`, `ypy-headers`, `ypy-encoding`, `ypy-validate` | extra values, other elements' values, headers |
| History | `ypy-push-url`, `ypy-replace-url`, `ypy-boost` | the address bar and whole-page navigation |
| Feedback | `ypy-indicator`, `ypy-disable`, `ypy-confirm`, `ypy-busy`, `ypy-announce` | spinners, locking, asking, screen readers |
| Concurrency | `ypy-sync` | what to do when a second request arrives |
| Errors | `ypy-error-target`, `ypy-error-swap`, `ypy-error-focus`, `ypy-fallback`, `ypy-retry` | where failures go |
| Morph | `ypy-preserve`, `ypy-preserve-attrs`, `ypy-morph-force`, `ypy-focus`, `ypy-scroll`, `ypy-main` | what the merge keeps |
| Code | `ypy-on:<event>` | a handler (off by default, see [Security](security-checklist.md)) |
| Configuration | `<meta name="ypy-config" content='...'>`, `ypy-config` | settings of the page or of one request |

The less obvious ones have their own reference pages: [`ypy-error-*`](../reference/attributes/ypy-error.md),
[`ypy-sync`](../reference/attributes/ypy-sync.md), [`ypy-preserve`](../reference/attributes/ypy-preserve.md),
[`ypy-focus`](../reference/attributes/ypy-focus.md), [`ypy-scroll`](../reference/attributes/ypy-scroll.md),
[`ypy-main`](../reference/attributes/ypy-main.md), [`ypy-retry`](../extensions/retry.md),
[`ypy-announce`](../reference/attributes/ypy-announce.md), [`ypy-busy`](../reference/attributes/ypy-busy.md).

```html
<form ypy-post="/subscribe" ypy-target="#box" ypy-swap="outerMorph" ypy-indicator="#spin"
      class="flex items-center gap-2">
  <input name="email" type="email" required class="rounded border p-2">
  <button class="rounded bg-sky-600 px-3 py-2 text-white">Subscribe</button>
  <span id="spin" class="ypy-indicator text-sm text-slate-500">Sending...</span>
</form>
```

## Triggers

`ypy-trigger` names the event that starts the request. Without it, YPY picks the natural one: `submit` for a form,
`change` for an input, select or textarea, `click` for anything else.

```html
<!-- on load, then every 30 seconds -->
<div ypy-get="/stats" ypy-trigger="load, every 30s" class="rounded border p-4"></div>

<!-- active search: wait 300 ms after the last key, only when the value changed -->
<input name="q" ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms, search" class="rounded border p-2">

<!-- load the next page when this row scrolls into view -->
<tr ypy-get="/rows?page=2" ypy-trigger="revealed" ypy-swap="afterend"></tr>
```

| Trigger | Fires |
|---|---|
| any DOM event (`click`, `input`, `change`, `submit`, `keyup`, ...) | when that event happens on the element |
| `load` | once, when the element is processed |
| `revealed` | once, the first time it scrolls into view |
| `intersect` | every time it enters the viewport (`root:`, `threshold:`) |
| `every 5s` | on a timer, while the element is in the page |

| Modifier | Effect |
|---|---|
| `once` | only the first time |
| `changed` | only if the value changed since the last request |
| `delay:300ms` | wait; a new event restarts the wait (debounce) |
| `throttle:1s` | at most once per interval |
| `from:<selector>` | listen on another element (`from:document`, `from:closest form`) |
| `target:<selector>` | only for events that began on a matching element |
| `consume` | stop the event from bubbling to the ancestors |
| `[expression]` | a condition; it compiles code, so it needs `callbacks` on (see [Security](security-checklist.md#author-code)) |

## Targets

`ypy-target` says where the answer goes. It takes a CSS selector, or one of the relative forms:

| Value | Element |
|---|---|
| `#list`, `.panel`, `main` | the first match in the document |
| `this` | the element that sent the request |
| `closest form` | the nearest ancestor that matches |
| `find .row` | the first descendant that matches |
| `next .panel`, `previous .panel` | the next or previous sibling that matches |

No `ypy-target` means the element itself. When the selector matches nothing, YPY says so in the console and fires
[`ypy:target:missing`](../reference/events/target.md) instead of failing silently. `ypy-select` picks *part of the
response* by selector (`ypy-select="#results"` takes only that element from a whole page).

## Swaps and morph

`ypy-swap` says how the answer is put in. The default is a morph: the response is merged into the page, and elements
that exist on both sides stay the same DOM nodes. Focus, typed text, scroll position and open `<details>` survive.

| Style | What it does |
|---|---|
| `innerMorph` (default) | merge the response into the target's children |
| `outerMorph` | merge the response into the target itself |
| `innerHTML`, `outerHTML` | replace the children, or the target (raw, nothing is kept) |
| `beforeend`, `afterbegin`, `beforebegin`, `afterend` | insert next to or inside the target |
| `textContent`, `delete`, `none` | set text, remove the target, do nothing |

Modifiers go after the style: `ypy-swap="innerHTML swap:200ms settle:100ms show:top"`. Full rules, the order of
operations and what "the user wins" means are in [Swap styles and morph](../reference/swap-and-morph.md).

## Partials and out-of-band swaps

One response can update several places. Put each extra piece in a `<ypy-partial>`. The client swaps the main content
first, then each partial into its own target:

```html
<li class="rounded border p-2">New item</li>

<ypy-partial ypy-target="#count" ypy-swap="innerHTML">3 items</ypy-partial>
<ypy-partial id="toast"><p class="rounded bg-emerald-600 p-3 text-white">Saved</p></ypy-partial>
```

Without `ypy-target`, the partial replaces the element on the page with the same `id`. See
[`<ypy-partial>`](../reference/tags/ypy-partial.md), and [Multiple targets](multiple-targets.md) for the patterns.

## Requests and headers

YPY sends headers so the server can tell a fragment request from a page load, and the server answers with
headers that steer the client. Header names are `YPY-*`.

| Request header | Value |
|---|---|
| `YPY-Request` | `true` on every request the client makes |
| `YPY-Target` | the target as `tag#id`, for example `ul#list` |
| `YPY-Source` | the element that made the request, as `tag#id` |
| `YPY-Request-Type` | `partial` (a smaller target) or `full` (the page content is replaced) |
| `YPY-Current-URL` | the URL of the page |
| `YPY-Boosted` | `true` for a boosted link or form |
| `YPY-History-Restore-Request` | `true` when restoring a page after back or forward (`YPY-Request` is left out then, so a full page is answered) |
| `YPY-Prompt` | the answer of a prompt (extension) |

| Response header | Effect |
|---|---|
| `YPY-Redirect` | navigate with a full page load |
| `YPY-Location` | navigate with a YPY request, no full load |
| `YPY-Push-Url`, `YPY-Replace-Url` | change the address bar (`false` stops it) |
| `YPY-Refresh` | reload the page |
| `YPY-Retarget`, `YPY-Reswap`, `YPY-Reselect` | change where, how and which part is swapped |
| `YPY-Trigger` | fire client events when the response arrives (a name, a list, or a JSON object with details). To act after the swap, listen to `ypy:after:swap` |

Two rules every server follows:

- Answer a fragment only to a YPY request and a whole page to everything else, and add `Vary: YPY-Request` so a cache
  keeps the two apart.
- Use the status codes that already mean something: `422` for an invalid form, `303` after a `POST` that worked, `204`
  for "nothing to swap". [Forms](forms.md) has the details.

## Inheritance

An attribute reaches the elements below it only if you add `:inherited`:

```html
<section ypy-target:inherited="#results" ypy-swap:inherited="innerHTML">
  <button ypy-get="/people">People</button>
  <button ypy-get="/places">Places</button>
</section>
<div id="results"></div>
```

Nothing is inherited by accident. Projects written for the older library often expect the opposite, so this is the most
common surprise when moving one. The rules, `:append` and the table of attributes are in [Attribute
inheritance](../reference/inheritance.md).

## Defaults

YPY has a single behavior, and the defaults lean safe: the swap is a morph, inheritance is explicit, a `422` goes into
the form, author code (`ypy-on`, `js:` values) is off, and `aria-busy` is set while a request runs. If a project written
for the older library relies on other behavior, `ypy migrate` converts it and [Migration](../migration/index.md) lists
what changes.

## Events

Every step of the life of a request fires a DOM event named `ypy:<noun>:<verb>` that bubbles: `ypy:before:request`,
`ypy:after:swap`, `ypy:error` and so on. Listen with `addEventListener`, with `ypy-on::before:request` on an element, or in
an extension. The list, the order and the `detail` are in [Events](../reference/events/index.md).

## Styling the states

YPY injects no inline styles and ships no CSS you must use. It reports state through classes and attributes, so any CSS
system can style it:

| State | Class (configurable in [`config.classes`](../reference/config/classes.md)) | Attribute |
|---|---|---|
| request in flight | `ypy-request` on the element and its indicators | none |
| swapping, settling, new content | `ypy-swapping`, `ypy-settling`, `ypy-added` on the target | none |
| the last request failed | | `data-ypy-state="error"` on the form or element, until the next request |
| offline | `ypy-offline` on `<html>` | `data-ypy-state="offline"` on `<html>` |

With Tailwind CSS v4, turn those classes into variants once, in your stylesheet:

```css
@import "tailwindcss";

/* scan the server templates too, or the classes of a fragment are removed from the build */
@source "../templates";
@source "../**/*.go";

@custom-variant ypy-request (&.ypy-request, .ypy-request &);
@custom-variant ypy-error (&[data-ypy-state="error"], [data-ypy-state="error"] &);
@custom-variant ypy-offline (.ypy-offline &);
```

and write them in the markup:

```html
<button ypy-post="/save"
        class="rounded bg-sky-600 px-4 py-2 text-white ypy-request:opacity-50 ypy-request:cursor-wait">
  Save
</button>

<form ypy-post="/signup" class="rounded border p-4 ypy-error:border-red-500">...</form>

<p class="hidden ypy-offline:block text-amber-700">You are offline.</p>
```

A spinner that shows only during a request:

```html
<span id="spin" class="ypy-indicator hidden ypy-request:inline">Loading...</span>
```

For Bootstrap, Bulma and others, style the same classes. [Third-party components](third-party-components.md) covers
starting their scripts again after a swap.

## Progressive enhancement

Build the page so links and forms work with no script: a `<form method="post" action="/items">` and an
`<a href="/items/3">`. Then add `ypy-*` attributes. The server reads `YPY-Request` and answers YPY with a fragment and a
plain browser with a whole page or a `303` redirect. If the script fails to load, the site still works.

## Where next

[Events](../reference/events/index.md) · [Forms](forms.md) · [History and caching](history-and-caching.md) ·
[Accessibility](accessibility.md) · [Security checklist](security-checklist.md) ·
[Index of the documentation](../index.md)
