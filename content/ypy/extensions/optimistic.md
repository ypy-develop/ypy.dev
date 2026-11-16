---
title: "Optimistic UI: the ypy-optimistic extension"
description: "Shows the result of an action before the server answers and puts the page back if the request fails: a like, a new comment, a deleted row."
order: 260
---
# Optimistic UI: the `ypy-optimistic` extension

Waiting for the server makes a click feel slow. `ypy-optimistic` shows the result of an action at once and puts everything
back if the request fails. A like button turns to "Liked" on the click, a new comment appears in the list, a deleted row
disappears, and if the request fails each of them is back as it was.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-optimistic.min.js"></script>

<template id="liked"><button disabled>Liked</button></template>

<button ypy-post="/posts/7/like" ypy-target="this" ypy-swap="outerMorph"
        ypy-optimistic="#liked" ypy-status:5xx="swap:none">Like</button>
```

On the click the template replaces the button on screen. When the server answers `200` with the new button, the morph merges
it into the original button. When the server answers `500`, the original button is back, untouched, and
`ypy:optimistic:rolledback` fires. The extension decides on the status alone: any `4xx` or `5xx` is a failure, any other
status is a success. Without the extension the attributes do nothing.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-optimistic` | `#tpl`, `remove`, or a registered name | what to show |
| `ypy-optimistic-target` | a selector | where, instead of the target of the request |
| `ypy-optimistic-swap` | a swap style | how, instead of the swap of the request |
| `ypy-optimistic="false"` | | off for this element |

They are read from the element that sends the request, and with `ypy-optimistic:inherited` from a container. A selector names
a `<template>` (its content is cloned) or any element (its children are cloned). `remove` shows the target gone, and any other
name is a function registered with `ypy.optimistic.register`.

By default the content goes where the response goes, with the swap style of the request:

| Swap style | The optimistic content |
|---|---|
| `innerHTML`, `innerMorph`, `morph`, and anything not listed here | replaces the children of the target |
| `outerHTML`, `outerMorph`, `outerSync` | is shown before the target, and the target is hidden |
| `beforebegin`, `afterend`, `afterbegin`, `beforeend` (and `before`, `after`, `prepend`, `append`) | is inserted there |
| `delete` | the target is hidden; there is nothing to show |

An outer style hides the target with `display:none` instead of removing it, so the request that came from it and the swap
aimed at it still find it. When the response is not swapped into the place (`swap:none`, a `204`, another target) use
`ypy-optimistic-target` to say where the change goes. The optimistic nodes carry the class `ypy-optimistic`, so
`.ypy-optimistic{opacity:.6}` gives a pending look. They are not processed while temporary, and ids in them can repeat ids in
the page.

## What ends the change

Every `ypy:optimistic:applied` is followed by exactly one `confirmed` or `rolledback`.

| What happens | The page | Event |
|---|---|---|
| status 4xx or 5xx | everything put back before the error body is swapped, so a `422` form re-render still works | `rolledback` `{reason:'status', status}` |
| network failure or timeout | put back | `rolledback` `{reason:'error', kind}` |
| the request is aborted | put back | `rolledback` `{reason:'aborted'}` |
| a `ypy:before:request` listener cancels it | put back | `rolledback` `{reason:'cancelled'}` |
| success, swapped into the same element | the original nodes come back, then the server's content goes in | `confirmed` `{status}` |
| success, nothing swapped there | the optimistic content stays as the final state and is processed | `confirmed` `{status}` |

The detail of every event is `{ctx, target}` plus the fields above. They fire on the element that sent the request and
bubble. `ypy:optimistic:applied` is cancelable, and cancelling it undoes the change at once.

The change is made right after `ypy:before:request`, after the queue let the request in and after any `ypy-confirm`. A
request that is queued, dropped by `ypy-sync` or declined by the user leaves nothing behind. "Put back" is exact: the
original nodes return, not copies, so an `<input>` keeps what the user typed. If something else changed the place meanwhile,
the original is not forced back over it.

## Concurrency

There is one overlay per element. A second change replaces the first for `inner` styles and adds to it for insertions. Two
different shapes on the same element are not mixed, so the second is not applied. The original is put back when the last
request that wanted the overlay has failed, and a success ends the overlay for everybody. With `ypy-sync="this:replace"` the
first request is aborted and rolls back while the page keeps showing the change of the second one.

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

`values` are the values of the request as an object, and the last one wins for a repeated name. The function returns a node,
a `DocumentFragment` or a list of nodes; a string of HTML, built inert through the `ypy` Trusted Types policy and ignored
when the CSP forbids it, so escape user data yourself; a function, which is the way back and is called on a rollback; or
nothing, which means no change and no event.

An error body is swapped by the core like any other response, so write `ypy-status:5xx="swap:none"` (or `ypy-error-swap`)
and the rolled-back page is not covered by the error text. A `422` is different: the form re-renders in place, after the
change was undone.

## Limits

The extension changes the DOM only. It does not queue requests, retry them (use [`ypy-retry`](retry.md)) or keep the change
across a reload. A swap with `swap:` or `settle:` delays leaves the original on screen during the delay. A rule such as
`display:block !important` on the target wins over the hiding of an outer style.
