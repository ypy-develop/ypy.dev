---
title: "Start-up and dependencies: the ypy-boot extension"
description: "Checks dependencies between extensions, fires ypy:ready once, makes ypy-ignore work on live elements and moves keyboard focus to the #fragment of a request."
order: 30
---
# Start-up and dependencies: the `ypy-boot` extension

The core stays small on purpose, so four pieces of start-up behavior live in `ypy-boot` instead:

1. a dependency check between extensions, and `ypy.supports()` to ask what is loaded;
2. the `ypy:ready` event, fired once when the page has been processed for the first time;
3. `ypy-ignore` that works on an element that is already live;
4. keyboard focus that moves to the `#fragment` of a request.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-boot.min.js"></script>
<script src="ext/ypy-signals.min.js"></script>
<script src="ext/ypy-expr.min.js"></script>
```

Load it first, right after `ypy.js` and before the other extensions. The order of the `<script>` tags is the order of
the checks: only extensions registered after `ypy-boot` are known to it. It is not part of `ypy.max`.

## Dependencies between extensions

`registerExtension()` reads three more keys of the object it is given. They matter when you write an extension; a page
that only loads extensions does not need them.

| Key | Example | Meaning |
|---|---|---|
| `version` | `"1.2.0"` | the extension is known as `name@1` (the major number; `1` when it says nothing) |
| `requires` | `["ypy-signals@1"]` | features that must already be registered |
| `provides` | `["expr@1"]` | features the extension adds under a name of its own |

When a required feature is missing, the extension is refused: the console gets a warning
(`ypy: needy needs base@1, which is not registered`) and `registerExtension()` returns `false`. The extension can
register again later, once the feature has arrived.

```js
ypy.registerExtension('my-badge', {
  version: '1.0.0',
  requires: ['ypy-signals@1'],
  init(api) { /* ... */ }
})
```

A requirement names a major version: `base@1` is missing when only `base@2` is registered.

### `ypy.supports(feature)`

Returns `true` when a feature is available. A feature is an extension by name (`"ypy-signals"`), by name and major
version (`"ypy-signals@1"`), or something an extension provides (`"expr@1"`).

```js
if (ypy.supports('expr@1')) {
  // author code runs without eval
}
```

`ypy-boot` itself registers `ypy-boot`, `ypy-boot@1` and `core@1`. Without it, `ypy.supports` does not exist.

## `ypy:ready`

Fires once on `document.body` (and bubbles to `document`) after the first `ypy.process()` of the page has finished.
Every element of the page is wired when it fires. It does not fire again on later swaps, and it does not fire without
the extension. `detail` is empty.

```html
<script>
  document.addEventListener('ypy:ready', () => {
    document.documentElement.dataset.ypy = 'ready'
  })
</script>
```

Use it where your code needs the page wired: reading `ypy-*` state, attaching a listener to an element YPY created,
or telling a test that the page is ready. A listener must be registered before the event fires, so put the
`<script>` in the page's `<head>`, or before the `ypy.js` tag.

## `ypy-ignore` on a live element

In the core, `ypy-ignore` keeps YPY from starting an element. With `ypy-boot` it also works afterwards:

- setting `ypy-ignore` (or `data-ypy-ignore`) on an element that was already processed tears it and its
  descendants down: no more requests, triggers, polls or handlers;
- removing the attribute processes the element again.

```html
<button id="save" ypy-post="/save" class="rounded bg-blue-600 px-3 py-1 text-white">Save</button>
<script>
  // while a dialog is open, the page below stops reacting
  document.querySelector('#save').setAttribute('ypy-ignore', '')
</script>
```

The attribute is watched on the whole document, with a custom `config.prefix` too. Without the extension, setting
`ypy-ignore` later does not undo what was processed.

## The focus of a `#fragment`

A request to `/page#section` already scrolls to the revealed element after the swap. `ypy-boot` also moves the focus
there, so a keyboard or screen reader user lands where the content is.

```html
<a href="/docs#install" ypy-boost="true" ypy-target="#main" ypy-swap="innerHTML">Install</a>
<main id="main"></main>
```

If the response has `<h2 id="install">`, the heading takes the focus.

- An element that cannot take the focus gets `tabindex="-1"`, and loses it again when it loses the focus.
- The focus never scrolls (`preventScroll`); the core has scrolled already.
- Nothing moves when the request carries its own [`ypy-focus`](../reference.md#core-attributes), when the response has no
  element with that `id`, or when the request has no fragment.

## Pitfalls

- Registered after other extensions, `ypy-boot` does not know them: `ypy.supports('ypy-signals')` is `false` for an
  extension that registered before it. Put its `<script>` first.
- `ypy:ready` is a one-time event. A listener added after it fired never runs; check the DOM state yourself in that case.
- A refused extension is not retried by itself. Fix the order of the `<script>` tags.

The tests are in `test/tests/ext/ypy-boot.js`; the changelog entry is under "Start-up and targets" in `CHANGELOG.md`.
