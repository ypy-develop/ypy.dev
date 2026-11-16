---
title: "ypy-announce"
description: "Tell screen readers what changed after a request through a live region YPY manages, with a message per element or defaults in the config."
order: 10
---
# ypy-announce

Tells screen readers what changed after a request, through a live region that YPY manages.

```html
<button ypy-post="/items" ypy-target="#list" ypy-announce="Item added"
        class="rounded bg-blue-600 px-3 py-1 text-white">Add</button>
```

When the response has been swapped, the text "Item added" is spoken. The region is a visually
hidden `<div role="status" aria-live="polite" data-ypy-live>` that YPY adds to `<body>`. It is
styled by an adopted style sheet, so there is no inline style and a strict CSP is fine.

## Values

| Value | What is announced |
|---|---|
| any text | the text, as it is (never parsed as HTML) |
| `title` | the title of the new document (`<title>` of the response) |
| `none` | nothing, even when a default message or `config.announce.messages.done` exists |

The attribute is read on the element that sends the request, then on the target. Use
`ypy-announce:inherited` to set it for a whole subtree.

## What is said without the attribute

- A boosted navigation (`ypy-boost`) announces the new title.
- A failed request is announced as `assertive` (`role="alert"`). The text comes from
  `config.announce.messages`: `offline` for a lost connection, `error` for anything else.
- The core carries no text of its own. A message you do not configure is not said.

```js
ypy.config.announce = {
  clearAfter: 5000, // ms until the region is emptied again (default 5000)
  messages: {
    loading: 'Loading',        // said when a request starts
    done: 'Done',              // said when a response was swapped
    offline: 'No connection',  // network failure or offline
    error: 'Something failed'  // any other failure
  }
}
```

A request that polls often should carry `ypy-announce="none"`, or `done` is said every time.

## The `ypy:announce` event

Fires on the source element (on `<body>` when it is gone) just before the text is written. It
bubbles and is cancelable. `event.detail` is `{ctx, message, politeness}`; a listener can change
`message` and `politeness` (`polite` or `assertive`) or call `preventDefault()` to say nothing.

## Where it is on

Automatic announcements are on by default. `config.announce = false` turns them off
everywhere; an object (as above) sets the messages.

The text goes in a moment after the region is emptied, so a screen reader says a repeated
message again.

## Related

[`ypy-busy`](ypy-busy.md) marks the target busy while the request runs.
