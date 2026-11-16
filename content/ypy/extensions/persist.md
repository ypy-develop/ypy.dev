---
order: 280
title: "Persistent elements and deploy detection: the ypy-persist extension"
description: "Keeps a video, player or widget alive across boosted navigations, and reloads the page when a new deploy changes the assets."
---

# Persistent elements and deploy detection: the `ypy-persist` extension

The extension has two jobs that both concern a page that stays open across many navigations. `ypy-persist` keeps an
element alive (a video, an audio player, a widget) when the visitor moves to another page. `ypy-track="reload"` notices
that the site was deployed in the meantime and loads the new page instead of swapping it into old code.

Both are described with the navigation features, in [Navigation](nav.md#ypy-persist). This page is the reference.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-nav.min.js"></script>
<script src="ext/ypy-persist.min.js"></script>
```

The extension is in `ypy.max`.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-persist` | none, or `false` | On an element that has an `id`: it stays in the page across a navigation whose response has an element with the same `id`. `false` turns it off. |
| `ypy-track` | `reload` | On an element of the page, usually in `<head>`: its value (`content`, `href`, `src`, `value` or text) is compared with the same element of a page response. A difference means a new deploy. |

Neither is inherited.

## Events

| Event | Cancelable | Detail | When |
|---|---|---|---|
| `ypy:track:reload` | yes | `ctx`, `url`, `from`, `to` | A whole-page response carries other tracked values than the page. Cancel it to swap the response instead of loading `url`. |

## `ypy-persist`

A navigation is a boosted link or form, or back and forward. When its response has an element with the same `id`, the
live element takes the place of the new one, with its state. The attribute goes on the live element only, so the server
does not mark the next page. A response without that `id` removes the element as usual, and a swap that is not a
navigation never keeps anything. It uses the same mechanism as [`ypy-preserve`](../index.md#preserving-content).

```html
<audio id="player" ypy-persist src="/radio.mp3" controls></audio>
```

## `ypy-track="reload"`

```html
<head>
  <meta name="app-version" content="2026.11.15-a1b2c3" ypy-track="reload">
  <link rel="stylesheet" href="/assets/app-4f9d2c.css" ypy-track="reload">
</head>
```

The check runs on a `GET` whose response is a whole page and succeeded. If the tracked elements differ (a changed value,
an element added or gone), the extension fires `ypy:track:reload` and then loads the address of the response with a
normal page load. Fragments, errors and answers to a `POST` never trigger it, so the visitor lands on the new code at
the next navigation.

```js
document.addEventListener('ypy:track:reload', e => {
  // e.detail.url is the page about to load; e.preventDefault() swaps the response instead
  if (formHasUnsavedChanges()) e.preventDefault();
})
```

## Notes

- A page that tracks nothing pays one attribute lookup per response.
- A reload loses the state of the old page, including any `ypy-persist` element. Cancel the event on screens where that
  would hurt.

## See also

[Navigation](nav.md), [`ypy-preserve`](../index.md#preserving-content), [History and caching](../index.md#history).
