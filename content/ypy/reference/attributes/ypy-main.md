---
title: "ypy-main and ypy-fallback"
description: "When the selector of a swap target matches nothing, YPY can swap into the page's main target instead of dropping the response. This is on by default."
order: 60
---
# ypy-main and ypy-fallback

When the selector of a swap target matches nothing, YPY can swap into the page's main
target instead of dropping the response. This is on by default.

## The main target

Mark it with `ypy-main` (`data-ypy-main` works too). Without a marker, the
`<main>` element is the main target.

```html
<body class="min-h-screen bg-slate-50">
  <header class="p-4">...</header>
  <main ypy-main id="content" class="mx-auto max-w-3xl p-6">
    <a href="/orders" ypy-get="/orders" ypy-target="#panel">Orders</a>
  </main>
</body>
```

## The chain

For a selector that comes from `ypy-target`, the `target:` modifier of `ypy-swap`, or an
`YPY-Retarget` header, YPY tries in order:

1. the selector itself;
2. the first `[ypy-main]` of the current page;
3. the first `<main>` of the current page;
4. nothing: the swap is skipped, `ypy:target:missing` fires and a warning goes to the console.

When step 2 or 3 serves, `ypy:target:fallback` fires and `ctx.fallback` holds `[ypy-main]` or
`main`.

## When it is safe

Swapping a small fragment over the main would erase the page, so by default the fallback only
runs when the response is a whole page (`<html>` or `<body>`) or contains a main
(`[ypy-main]` or `<main>`). The response's own main then replaces the current one: its children for an
inner style such as `innerHTML` or `innerMorph`, the element itself for `outerHTML` and the
other outer styles. A whole page without a main goes into the current main as it is; with an
outer style that would drop the `<main>` element, so it is reported missing instead. A fragment
without a main is never swapped, and `ypy:target:missing` fires. `ypy-select` still picks the
content when it is set.

## ypy-fallback

| Value | Meaning |
|---|---|
| `true` (or no value) | the default described above |
| `always` | also swap a fragment into the main |
| `false` | no fallback; the target is reported missing |

Set it on the element that makes the request, or on an ancestor with `ypy-fallback:inherited`.
`always` is the way to also swap a fragment into the main target:

```html
<button ypy-get="/panel" ypy-target="#gone" ypy-fallback="always">Reload</button>
```

## Configuration

`ypy.config.targetFallback` is `true`, `false` or `"always"`. Unset, it is on. The attribute
overrides it.

## What never falls back

- Out-of-band elements and `<ypy-partial>` whose target is missing: it is ambiguous which
  region they meant. `ypy:target:missing` fires for each of them.
- `ypy.ajax()` with a target that does not exist: the promise rejects with "Target not found",
  because the check runs before the request.
