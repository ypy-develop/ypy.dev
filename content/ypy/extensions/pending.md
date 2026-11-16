---
order: 270
title: "Placeholder while loading: the ypy-pending extension"
description: "Puts a skeleton or spinner from a template into the target while a request runs, and removes it just before the swap, so nothing is left behind."
---

# Placeholder while loading: the `ypy-pending` extension

A slow request leaves the old content on screen, or a blank space where the new content will be. `ypy-pending` puts a
placeholder in the target for as long as the request runs: a skeleton, a spinner, a line that repeats what the visitor
asked for.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-pending.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-pending` in it. It is listed with the other
official extensions in [Extensions](index.md).

## Attributes and classes

| Name | Value | Meaning |
|---|---|---|
| `ypy-pending` | a selector | The element whose content is the placeholder. A `<template>` is the usual choice. Inherited. |
| `ypy-pending` (class) | | On the inserted placeholder, until the request ends. |

## Where the placeholder goes

The placeholder is a `<div class="ypy-pending">` that holds a copy of the content of the element you named. Its place
follows the swap style of the request:

| Swap style | What happens |
|---|---|
| `innerHTML` | The children of the target are hidden and the placeholder is added inside it. |
| `beforebegin`, `afterbegin`, `beforeend`, `afterend` (and `before`, `prepend`, `append`, `after`) | The placeholder is inserted at that place. |
| Anything else (`outerHTML`, `innerMorph`, ...) | The target is hidden and the placeholder goes right after it. |

Before the swap, or when the request fails, the placeholder is removed and what was hidden is shown again. YPY processes
the placeholder, so `ypy-*` attributes inside it work.

## The request parameters

Each parameter of the request (form fields, `ypy-vals`) becomes a `data-*` attribute of the placeholder. A field named
`query` gives `data-query="shoes"`. Use it in CSS, or read it in a script, to say what is loading.

```css
.ypy-pending::before { content: "Searching for " attr(data-query) "..."; }
```

The values are set with the `dataset` API, so they are never read as markup.

## Example

```html
<template id="results-pending">
  <p aria-hidden="true">Searching...</p>
</template>

<form ypy-get="/search" ypy-target="#results" ypy-swap="innerHTML" ypy-pending="#results-pending">
  <input name="q" type="search">
  <button>Search</button>
</form>
<div id="results" aria-live="polite"></div>
```

```go
func search(w http.ResponseWriter, r *http.Request) {
	items := find(r.URL.Query().Get("q"))
	var b strings.Builder
	for _, it := range items {
		b.WriteString("<li>" + html.EscapeString(it) + "</li>")
	}
	ypy.Fragment(w, "<ul>"+b.String()+"</ul>")
}
```

## Notes

- The placeholder has reset styling (`all: initial`), so it does not inherit the look of the target. Style it through
  its class and the content you copy into it.
- Hidden elements use `style.display`. A target that is already `display:none` stays hidden afterwards.
- For a screen reader, do not rely on the placeholder alone. Mark the region with
  [`ypy-busy`](../reference/attributes/ypy-busy.md) and [`ypy-announce`](../reference/attributes/ypy-announce.md).
- To fill the target with something that looks like the answer and take it back on error, see
  [`ypy-optimistic`](optimistic.md).

## See also

[`ypy-disable`](disable.md), [Styling](../reference/styling.md), [`ypy-lazy`](lazy.md) for placeholders of regions that load
when seen.
