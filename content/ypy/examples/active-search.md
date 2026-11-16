---
order: 80
title: Active search
description: Filter a table while the visitor types, with a short delay between keystrokes, and show a message when nothing matches the search.
---

# Active search

A search box sends what was typed to the server a moment after the visitor stops typing, and the table body is replaced
by the matches.

## Try it

<!-- demo:active-search -->

Type `engineer`, `porto` or `designer`. Type something that matches no one to see the empty answer.

## The page

```html
<input type="search" name="q" placeholder="Name, role or city"
       ypy-get="/people" ypy-trigger="input changed delay:300ms, search"
       ypy-target="#search-results" ypy-indicator="#search-results" ypy-sync="replace">

<table>
  <thead><tr><th>Name</th><th>Role</th><th>City</th></tr></thead>
  <tbody id="search-results" aria-live="polite">
    <tr><td>Marina Duarte</td><td>Designer</td><td>Lisbon</td></tr>
    <!-- … -->
  </tbody>
</table>
```

## What the server answers

`GET /people?q=porto` answers `200` with the rows that match, as `<tr>` elements, and `Vary: YPY-Request`:

```html
<tr><td>Tomás Ribeiro</td><td>Engineer</td><td>Porto</td></tr>
<tr><td>Joaquim Teles</td><td>Sales manager</td><td>Porto Alegre</td></tr>
```

With nothing matching it is still `200`, with a row that says so:

```html
<tr><td colspan="3">No one matches “zzz”. Try a name, a role or a city.</td></tr>
```

An empty `q` answers every row.

## How it works

- An input sends its own `name` and `value`, so no form is needed.
- `ypy-trigger` lists two triggers: `input changed delay:300ms` waits 300 ms after the last keystroke and ignores a
  value that did not change; `search` fires when the visitor presses Enter or clears the field with the browser's
  button.
- `ypy-sync="replace"` aborts a request still running when a newer one starts, so a slow answer can never overwrite a
  faster, newer one.
- `ypy-indicator` puts the `ypy-request` class on the results while the request runs, so the stylesheet can dim them.
- `aria-live="polite"` makes the change audible; the empty row is a `200`, because "nothing found" is an answer.

## Variations

- Search on submit only: leave out `input` and keep `search`.
- Push the query into the address with `ypy-push-url="true"` so a search can be shared.
- Keep a feed of results bounded and free of duplicates with the [`ypy-list`](../extensions/list.md) extension.

## See also

- [Click to load](click-to-load.md)
- [Infinite scroll](infinite-scroll.md)
- [Value select](value-select.md)
