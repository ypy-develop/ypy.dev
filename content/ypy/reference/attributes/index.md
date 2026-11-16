---
order: 10
title: Attributes
description: "Every attribute of the YPY core, grouped by what it does, with the value each one takes and whether child elements inherit it."
---

# Attributes

This page lists the attributes of the core. Extensions add their own; those are in the page of each
[extension](../../extensions/index.md). Every attribute can be written with a `data-` prefix too, so
`data-ypy-get` works like `ypy-get` and passes strict HTML validators.

A child takes the value of an ancestor only when the ancestor writes the
attribute with the `:inherited` modifier, for example `ypy-target:inherited="#results"`. The `:inherited`
column says whether an attribute accepts that modifier (`no` means the attribute is read from the element
itself only). The rules are in [Attribute inheritance](../inheritance.md).

All examples load YPY from a script tag:

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<!-- or -->
<script src="https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js"></script>
```

## Request

An element with one of these attributes sends a request when its trigger fires and swaps the response into the
page.

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-get` | URL | no | Sends a `GET` request to the URL. |
| `ypy-post` | URL | no | Sends a `POST` request to the URL. |
| `ypy-put` | URL | no | Sends a `PUT` request to the URL. |
| `ypy-patch` | URL | no | Sends a `PATCH` request to the URL. |
| `ypy-delete` | URL | no | Sends a `DELETE` request to the URL. |
| `ypy-query` | URL | no | Sends a `QUERY` request to the URL. |
| `ypy-action` | URL | no | Sets the URL when the method comes from `ypy-method`. |
| `ypy-method` | `get`, `post`, `put`, `patch`, `delete` or `query` | no | Sets the HTTP method, together with `ypy-action`. |

```html
<button ypy-delete="/items/7" ypy-target="closest li" ypy-swap="outerHTML">Remove</button>
```

Use `ypy-action` and `ypy-method` when the method is chosen by data, for example from a template variable.

## Trigger and flow

These attributes decide when the request starts and what can stop or change it.

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-trigger` | `<event>[<filter>] [<modifier>...][, ...]` | yes | The events that start the request. Without it, an input or select uses `change`, a form uses `submit` and anything else uses `click`. |
| `ypy-sync` | `[<selector>:]<strategy>` | yes | What happens when the element asks for a request while another one from the same owner is in flight. The default is `queue first`. See [`ypy-sync`](ypy-sync.md). |
| `ypy-confirm` | text | yes | Asks the user before the request. The value is the message; `js:` runs an expression that may return a promise. |
| `ypy-validate` | string | yes | Checks the form with the browser's constraint API before the request. On by default for a form without `novalidate`. Set `true` for other elements, `false` to turn it off. |
| `ypy-indicator` | selector | yes | Elements that get the `ypy-request` class while the request runs. |
| `ypy-on` | JavaScript | no | Runs inline code when an event fires, as `ypy-on:<event>`. `ypy-on::<event>` is short for `ypy-on:ypy:<event>`. Runs only when `config.callbacks` allows it. |
| `ypy-ignore` | none | no | YPY does not process this element or its children. |

The trigger takes the values `load`, `revealed`, `intersect` and `every <time>` besides ordinary DOM events,
and these modifiers: `once`, `changed`, `delay:<time>`, `throttle:<time>`, `from:<selector>`,
`target:<selector>`, `prevent`, `stop`, `halt`, `capture`, `passive`, `root:<selector>` and
`threshold:<number>`.

```html
<input type="search" name="q"
       ypy-get="/search" ypy-target="#results"
       ypy-trigger="input changed delay:300ms">

<div ypy-get="/clock" ypy-trigger="load, every 10s"></div>
```

## Target and swap

These attributes say which part of the response goes where, and how it is inserted.

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-target` | selector, `this`, `closest <selector>`, `find <selector>`, `next [selector]`, `previous [selector]` | yes | Where the response goes. Defaults to the element itself. |
| `ypy-swap` | `<style> [<modifier>...]` | yes | How the response is inserted. Styles: `innerHTML`, `outerHTML`, `outerSync`, `innerMorph`, `outerMorph`, `textContent`, `beforebegin`, `afterbegin`, `beforeend`, `afterend`, `delete`, `none`, plus `download` and `upsert` when their extensions are loaded. |
| `ypy-select` | selector | yes | Takes only this part of the response. |
| `ypy-select-oob` | `<selector>[:<swap>], ...` | yes | Takes these elements of the response and swaps each one out of band, into the element with the same `id`. |
| `ypy-swap-oob` | `true`, `<style>` or `<style>:<selector>` | no | Written in the response, on an element that goes into the element with its `id` instead of the main target. |

Swap modifiers follow the style: `swap:<time>`, `settle:<time>`, `transition:true`, `ignoreTitle:true`,
`scroll:top|bottom`, `show:top|bottom`, `scrollTarget:<selector>`, `showTarget:<selector>`, `strip:true`,
`focusScroll:true`, `swapEmpty:true` and `target:<selector>`. The styles and the morph are described in
[Swap styles and morph](../swap-and-morph.md).

```html
<form ypy-post="/comments" ypy-target="#comments" ypy-swap="beforeend">...</form>
```

A response can update more than one place. This one fills the main target and refreshes a counter:

```html
<li>New comment</li>
<span id="comment-count" ypy-swap-oob="true">12</span>
```

For several places at once, see also [`<ypy-partial>`](../tags/ypy-partial.md) below and the guide on
[multiple targets](../../guides/multiple-targets.md).

## Parameters

These attributes change what the request carries.

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-vals` | JSON, or `js:<expression>` | yes | Adds values to the request parameters. |
| `ypy-include` | selector, `this`, or an extended selector | yes | Includes the values of other elements. |
| `ypy-headers` | JSON, or `js:<expression>` | yes | Adds request headers. |
| `ypy-encoding` | `multipart/form-data` | yes | Sets the encoding. Use it for file uploads. |
| `ypy-config` | JSON | yes | Sets fetch options of this element's requests: `timeout`, `credentials`, `cache` and others. It cannot override `mode`. |

```html
<button ypy-post="/vote" ypy-vals='{"choice": "b"}' ypy-include="#csrf-field">Vote B</button>
```

## History and navigation

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-boost` | boolean | yes | Turns links and forms inside the element into YPY requests that swap the body and push the URL. `false` opts a subtree out. |
| `ypy-push-url` | `true`, `false` or a URL | yes | Pushes a URL into the browser history after the swap. A failed swap pushes nothing. |
| `ypy-replace-url` | `true`, `false` or a URL | yes | Replaces the current history entry after the swap. |
| `ypy-history-elt` | none | no | Marks the element restored on back and forward, instead of `body`. |

A boosted request pushes the URL by default only when it is a `GET` or was redirected to a page. A `POST`,
`PUT`, `PATCH` or `DELETE` that answers with a page needs `ypy-push-url`.

```html
<nav ypy-boost="true">
  <a href="/docs">Docs</a>
  <a href="/download" ypy-boost="false">Download</a>
</nav>
```

More in [History](../history.md) and the [`nav`](../../extensions/nav.md) extension.

## Errors and status codes

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-status:<code>` | `<key>:<value> [<key>:<value>...]` | yes | Changes how one status code, or a range such as `5xx` or `50x`, is handled. Keys: `swap:`, `target:`, `select:`, `push:`, `replace:`, `transition:`. |
| `ypy-error-target` | selector | yes | Where a response of status 400 or more is swapped. Without it, a 422 goes into the form of the source element. |
| `ypy-error-swap` | swap spec | yes | The swap style for a response of status 400 or more. |
| `ypy-error-focus` | `first-invalid`, a selector or `false` | yes | Where the focus goes after an error response is swapped. |
| `ypy-main` | none | no | Marks the main target, the fallback when a target selector matches nothing. Without a marker `<main>` is used. |
| `ypy-fallback` | `true`, `always` or `false` | yes | Overrides `config.targetFallback` for the element. |

```html
<form ypy-post="/signup" ypy-error-target="#signup-errors" ypy-error-focus="first-invalid">...</form>
<div ypy-get="/panel" ypy-status:404="target:#not-found swap:innerHTML">...</div>
```

Pages: [`ypy-error-*`](ypy-error.md), [`ypy-main` and `ypy-fallback`](ypy-main.md). The events are
[`ypy:target:fallback` and `ypy:target:missing`](../events/target.md).

## Focus, scroll and accessibility

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-focus` | `auto`, `none`, `first-heading` | yes | Where keyboard focus goes when a swap ends. |
| `ypy-scroll` | `keep`, `top`, `restore` or an offset | yes | The scroll position when a swap ends. |
| `ypy-announce` | text, `title` or `none` | yes | What screen readers hear after the request, through a live region YPY manages. |
| `ypy-busy` | boolean | yes | `false` turns off the automatic `aria-busy` mark during a request. |

Pages: [`ypy-focus`](ypy-focus.md), [`ypy-scroll`](ypy-scroll.md), [`ypy-announce`](ypy-announce.md),
[`ypy-busy`](ypy-busy.md). The extension attribute `ypy-inert` is in [its own page](ypy-inert.md).

## Morph

These attributes control what a morph swap (`innerMorph`, `outerMorph`) leaves alone.

| Attribute | Value | `:inherited` | Description |
|---|---|---|---|
| `ypy-preserve` | `true`, `all`, `attrs`, `children`, `false`, `value`, `selection`, `focus` or `scroll` | no | How much of an element a swap or morph leaves alone. No value keeps the node by `id`. |
| `ypy-preserve-attrs` | `<name>[,<name>...]` | no | Attributes the morph neither sets nor removes. Entries may end with `*`. |
| `ypy-morph-skip` | none | no | Freezes the element: it is neither updated nor removed. |
| `ypy-morph-skip-children` | none | no | Updates the attributes of the element but freezes its children. |
| `ypy-morph-force` | none | no | Written in the response: the response wins over what the user did to the element, such as the `open` state of `<details>`. |

```html
<input id="note" name="note" ypy-preserve="value">
```

See [`ypy-preserve`](ypy-preserve.md) and [Swap styles and morph](../swap-and-morph.md).

## The `<ypy-partial>` element

| Element | Attributes | Description |
|---|---|---|
| `<ypy-partial>` | `ypy-target`, `ypy-swap` | Routes a piece of a response to its own target. Use `ypy-target` or `id` to choose the destination and `ypy-swap` to choose the swap style. |

```html
<ypy-partial ypy-target="#cart-count" ypy-swap="innerHTML">3 items</ypy-partial>
```

See [`<ypy-partial>`](../tags/ypy-partial.md).

## Attributes added by extensions

The core does not know `ypy-retry`, `ypy-sse:connect`, `ypy-signals`, `ypy-layer` and the other attributes of
the official extensions until you load the extension. Each page under [Extensions](../../extensions/index.md)
lists the attributes, events, headers and settings of one extension.
