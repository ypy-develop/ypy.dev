---
title: "aria-busy and ypy-busy"
description: "During a request YPY sets aria-busy=\"true\" on the target and removes it when the request ends, including when it fails, times out or is aborted."
order: 20
---
# aria-busy and ypy-busy

During a request YPY sets `aria-busy="true"` on the target and removes it when the request ends,
including when it fails, times out or is aborted.

```html
<form ypy-post="/search" ypy-target="#results">...</form>
<ul id="results" class="aria-busy:opacity-50" aria-live="polite"></ul>
```

In Tailwind, `aria-busy:` styles the target while it loads. With plain CSS use
`[aria-busy="true"]`.

- Several requests to one target are counted; the attribute stays until the last one ends.
- A value you put there yourself (`aria-busy="false"`) is restored afterwards.
- A request whose confirmation was denied never marks anything.

## Opt out

```html
<button ypy-get="/poll" ypy-busy="false">Refresh</button>
```

`ypy-busy="false"` on the element that sends the request, or on the target, skips the mark.
`ypy-busy:inherited="false"` on an ancestor skips it for the subtree. `ypy.config.ariaBusy = false`
turns it off everywhere.

## Pico CSS

Pico draws a loading spinner on every `[aria-busy="true"]` element. With the automatic mark, a
target gets that spinner while it loads. If you do not want it, set `ypy-busy="false"` where it
is not wanted, set `ypy.config.ariaBusy = false`, or override the style (adjust the selector to your Pico version):

```css
[aria-busy="true"]:not(button, input, a)::before { content: none; }
```

## Related

[`ypy-announce`](ypy-announce.md) says what changed once the request ends.
