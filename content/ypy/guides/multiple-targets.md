---
title: "Multiple targets in one response"
description: "Update several places from one response with ypy-partial: the main target, a counter, a flash message. Covers out-of-band swaps, task order and failures."
order: 40
---
# Multiple targets in one response

One response can update several places on the page: the form's own target, a counter, a list, a flash message.

## Partials

Put the content for the request's own target first, then one `<ypy-partial>` for each other place.

```html
<tr id="row-7" class="border-b"><td class="p-2">Renamed</td></tr>

<ypy-partial ypy-target="#total" ypy-swap="innerHTML">
  <strong class="text-lg">41</strong>
</ypy-partial>
<ypy-partial ypy-target="#flash" ypy-swap="beforeend">
  <p class="rounded bg-green-100 p-3">Saved</p>
</ypy-partial>
```

The target of every piece is written next to the piece. See [`<ypy-partial>`](../reference/tags/ypy-partial.md).

## Out-of-band swaps

`ypy-swap-oob` on an element of the response and `ypy-select-oob` on the element that made the request swap
content somewhere else (`ypy-swap-oob="true"` is `outerHTML` on the element with the same `id`). They exist so that
responses written for the older library keep working. In new code, use a partial: the target and the swap style are
written out, and several targets need no repeated markup.

## What happens, in order

1. The response is parsed and split into a list of tasks: the main swap, the out-of-band elements, the partials.
2. `ypy:before:swap` fires with that list (`evt.detail.tasks`); a listener can cancel the swap.
3. The tasks run, in the order of the list, each through the same code as a normal swap (swap styles, morph,
   settle events, scripts).
4. `ypy:after:swap` fires, then the title is set and the scroll policy applies.

The list is ordered: the main swap first, then `ypy-select-oob` in the order of the attribute, then `ypy-swap-oob`
in document order, then the partials in document order. A task that does not wait finishes its DOM changes
before the next one starts; a task with `transition:true` runs later, inside the view transition.

### The task object

Each entry of `evt.detail.tasks` is an object with these fields. The shape is stable, because extensions that add tasks
(`ypy:process:<type>`) and the streaming patches of a later phase use the same pipeline.

| Field | Meaning |
|---|---|
| `type` | `"main"`, `"oob"`, `"partial"`, or the type an extension gave it |
| `fragment` | `DocumentFragment` with the content to insert |
| `target` | the element to swap into, or a selector resolved when the task runs; the final target after the swap |
| `swapSpec` | `{style, ...modifiers}` (or its string form) |
| `sourceElement` | the element the request started from |

The main task also has `transition`.

## Missing targets and failures

- A partial or out-of-band element with no element to swap into fires `ypy:target:missing`
  (`{ctx, selector, type}`) and is skipped. `ctx.missingTargets` lists them, and a response that only had
  such elements does not blank the main target.
- A task that throws (an unknown swap style, a failing script or morph) does not stop the others. Each failure
  fires `ypy:swap:error` with `{ctx, task, error}`, then `ypy:error` fires once with the first error, and the
  swap ends normally: `ypy:after:swap` and `ypy:finally:swap` still fire.

```js
document.addEventListener('ypy:swap:error', e => {
  console.warn('swap failed for', e.detail.task.target, e.detail.error)
})
```
