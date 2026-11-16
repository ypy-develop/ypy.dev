---
title: "Atomic swaps and atomic navigation"
description: "Apply every part of a response or none of it with YPY-Group: atomic, and change the address bar and history only after a navigation succeeded."
order: 120
---
# Atomic swaps and atomic navigation

Two guarantees that keep the page from ending up half updated:

1. `YPY-Group: atomic`, a response header. The server asks that every part of the response (main content, out-of-band
   elements, partials) is applied, or none of it.
2. Atomic navigation, a fixed behavior of the core. The address bar and the history change only after the page
   changed, never before and never when the swap failed.

Both are in the core, so there is nothing to load.

## `YPY-Group: atomic`

A response can update several places at once: the main target, `ypy-swap-oob` elements and
[`<ypy-partial>`](tags/ypy-partial.md) elements. By default each one is applied if its target exists, and the
ones whose target is missing are reported and skipped. That is right for a toast that may be absent on some
pages. It is wrong when the parts only make sense together: a cart that shows 3 items in the list and 2 in the
counter is worse than a cart that did not change.

```http
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
YPY-Group: atomic

<li class="flex gap-2 p-2">Notebook</li>
<ypy-partial ypy-target="#cart-count" ypy-swap="innerHTML">3</ypy-partial>
<div id="cart-total" ypy-swap-oob="innerHTML">$ 42.00</div>
```

With the header, YPY looks for the target of every task before it inserts anything.

| Situation | Without the header | With `YPY-Group: atomic` |
|---|---|---|
| every target exists | all applied | all applied |
| one out-of-band element or partial has no target | the others are applied | none is applied |
| the main swap has no target | the others are applied | none is applied |
| a listener cancels `ypy:before:swap` | nothing is applied | nothing is applied |

When the swap is refused:

- the page stays as it was, content, title, focus and scroll included;
- the history is not updated: a `ypy-push-url`, `ypy-replace-url`, `YPY-Push-Url` or boosted push is skipped;
- `ypy:target:missing` fires for each missing target (`detail.selector`, `detail.ctx`, `detail.task`), so you can
  see which one stopped the group. See [target events](events/target.md);
- `ypy:after:swap` does not fire, and `ctx.timing.swapped` is not set. `ypy:finally:swap` and
  `ypy:finally:request` still fire. The request is not an error: `ypy:error` stays quiet.

The value is exactly `atomic`, in lower case. Any other value is ignored and the response is applied as usual.

### What "atomic" covers

The guarantee is about targets: YPY checks that every place exists, then inserts. It does not make the
insertions a database transaction. If a task throws while it inserts (a swap style that does not exist, a
failing script), the tasks that already ran are not undone. The failure is reported through
`ypy:swap:error`, and the history is not updated, as for any failed swap (see below).

Use the header for a missing or renamed element, which is what changes between pages and between versions of
a page. Do not use it as a way to roll back a bad insertion.

### The handler

The Go SDK has no helper for this header; write it before the first byte of the body, because `Fragment` sends
the status line.

```go
func addItem(w http.ResponseWriter, r *http.Request) {
	item := create(r.Context(), r.FormValue("name"))

	w.Header().Set("YPY-Group", "atomic")
	ypy.Fragment(w, renderItem(item))
	ypy.Partial(w, ypy.PartialSpec{Target: "#cart-count", Swap: "innerHTML"}, strconv.Itoa(count(r.Context())))
	ypy.Partial(w, ypy.PartialSpec{Target: "#cart-total"}, money(total(r.Context())))
}
```

```html
<ul id="cart" class="divide-y"></ul>
<span id="cart-count">0</span>
<span id="cart-total">$ 0.00</span>

<form ypy-post="/cart/items" ypy-target="#cart" ypy-swap="beforeend" class="flex gap-2">
  <input name="name" class="rounded border p-2">
  <button class="rounded bg-sky-600 px-4 py-2 text-white">Add</button>
</form>
```

If `#cart-total` is not on the page (a different layout, an old cached page), nothing changes and `#cart-count`
does not show a number that the list does not match.

## Atomic navigation

A request that changes the address (a boosted link or form, `ypy-push-url`, `ypy-replace-url`, the headers
`YPY-Push-Url` and `YPY-Replace-Url`) does it in this order:

1. the response is swapped into the page;
2. `ypy:after:process` fires for the new content;
3. `ypy:before:history:update` fires (cancelable), then the URL is pushed or replaced;
4. `ypy:after:history:push` or `ypy:after:history:replace`, then `ypy:after:history:update`;
5. `ypy:after:swap`.

So a listener of `ypy:before:history:update` sees the new content and the old URL, and after the
history events the two match. There is never a moment when the address names a page that is not on the screen.

The URL is not written when:

- a task of the swap failed (`ypy:swap:error` fired);
- the request was aborted during the swap;
- a `ypy:before:swap` listener cancelled the swap;
- the response was refused by `YPY-Group: atomic`;
- a 422 answered a form that is not boosted. An invalid form is not a new page.

Pressing Back after a failed navigation therefore goes where the user came from, not to a page that was never
shown. The details of an entry and of the restore are in [History](history.md).

### What a boost pushes

A boosted element pushes by default, but not for every method. The answer of a `POST`, `PUT`, `PATCH` or
`DELETE` is a result, not a page that can be loaded again.

| Boosted request | Pushes the URL? |
|---|---|
| link, or form with `GET` | yes, the final URL of the response |
| `POST`, `PUT`, `PATCH`, `DELETE` answered with a page | no, the URL stays |
| the same, answered with a redirect (`303` to a page) | yes, the page the redirect ended on |
| any of them with an explicit `ypy-push-url` | yes, as written |
| a request that is not boosted and has no `ypy-push-url` | no |

```html
<!-- boosted POST: after the 303 the address becomes /welcome, never /signup -->
<form action="/signup" method="post" ypy-boost="true" ypy-target="#page" ypy-swap="innerHTML">...</form>
```

### Scroll position and other scripts

- A push records where the page being left was scrolled to, measured before the swap. When the user comes
  back to it, `ypy-scroll` restores that position and not the one the new, shorter content left behind.
- If another script replaces the state of an entry that YPY wrote (`history.replaceState({router: 1}, ...)`),
  the next YPY push stamps the entry again, and going back to it restores the page and keeps the keys of the
  other script. YPY remembers the paths it wrote, so an entry whose state was emptied is still restored.
- An entry that another router owns, on a path YPY never wrote, is left to that router.

## Pitfalls

Keep optional parts out of the group. A toast that exists only on some pages, put in the same atomic group as the
list, blocks the list on the pages without it. Send the optional part in another request, or leave it out of the group.

A missing main target is also a refusal. When the main swap cannot find a target and cannot fall back (a `ypy-target`
that matches nothing and `ypy-fallback="false"`), the whole group is stopped.

The check looks at targets, not at the answer. A server that sends the header on a half-built response still gets that
response applied when every target exists.

## See also

- [`<ypy-partial>`](tags/ypy-partial.md) and [Multiple targets](../guides/multiple-targets.md)
- [Target events](events/target.md): `ypy:target:missing`
- [History](history.md) and [Navigation](../extensions/nav.md)
- [`ypy-sync`](attributes/ypy-sync.md): request generations, which keep an old response from overwriting a new one
