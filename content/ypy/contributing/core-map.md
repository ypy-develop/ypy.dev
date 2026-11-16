---
title: "Core map"
description: "How src/ypy.js is divided, who owns which function, and the rules for adding to it. Read this before you change the core."
order: 20
---
# Core map

How `src/ypy.js` is divided, who owns which function, and the rules for adding to it.
Read this before you change the core.

## Rules

1. **One owner per function.** A function listed below is edited only by its owner task. Two
   tasks never edit the same function.
2. **Only add names.** About 45 test files call `ypy.__name(...)` directly. Do not rename or
   move an existing `__name`, and do not change its parameters; add a new one instead.
3. **A new attribute is read only through `__attributeValue(elt, 'ypy-<name>')`.** The resolver
   adds the `ypy-` and `data-ypy-` spellings (and `config.prefix`). Never call `getAttribute('ypy-...')`
   yourself.
4. **A new event is named `ypy:<noun>:<verb>` and fired with `__triggerNative(on, name, detail)`.**
   Events that already exist keep going through `__trigger`; both dispatch one event under one name.
5. **A new state of an element goes through `__setState(elt, state, on)`.** It sets the
   configured classes and the `data-ypy-state` attribute together.
6. **There is one behavior.** A default is the YPY default; do not add a switch that restores another
   library's behavior.
7. **The orchestrators are frozen.** `__issueRequest`, `swap`, `__insertContent` and
   `initialize` only order phases. Put your logic in the phase or hook you own. If you need a
   new call point, ask the orchestrator.
8. **`ypy.d.ts`**: edit only your `// #region TASK-nnn` block at the end of the file.
9. **ADR**: write `adr/NNNN-<slug>.md` in the control repository with the next free number (see its `adr/README.md`).
10. **Tests**: one new file per task in `test/tests/{unit,attributes,end2end,security}/`.
    Never edit another task's test. A test starts with `setupTest(this)`.
11. **Size**: every task has a budget in min+gzip bytes (see the ADR). Measure with
    `go run ./cmd/ypy build` and `gzip -c dist/ypy.min.js | wc -c`.

## Sections of `src/ypy.js`

Marked in the file with `// SECTION <name>`. A section appears again as `(cont.)` where the
file keeps related code apart; nothing was moved, to keep the history readable.

| Section | What it holds |
|---|---|
| `config` | construction, `__initConfig`, indicator CSS, `registerExtension` |
| `attributes` | attribute prefixes, `__attr`, `__attributeValue`, inheritance, `__getAttributeObject`, element queries (`__findAllExt`) |
| `processing` | `process`, element state, boosting, `__cleanup`, preserved elements, state classes |
| `request` | the request context, the request phases, indicators, form values, abort, status codes, `ajax` |
| `swap` | parsing a response, OOB and partials, the swap phases, view transitions |
| `morph` | the morphing algorithm |
| `history` | `__initHistory`, push, replace, restore, `__handleHistoryUpdate` |
| `events` | trigger specifications, `__trigger`, `trigger`, `on`, `ypy-on` |
| `security` | `__executeJavaScript`, `__parseHTML`, `__processScripts` |
| `a11y` | focus and scroll helpers, `__markBusy`, `__unmarkBusy`, `__announce`, `__initA11y` |
| `offline` | `__initOffline` |

## The request, phase by phase

`__issueRequest(ctx)` runs these phases. A phase that wants to stop the request returns `false`.
An exception in any phase goes to `__failRequest`. `__endRequest` always runs once the request
was admitted.

| Phase | Does | Events | Owner |
|---|---|---|---|
| `__admitRequest(ctx)` | asks the queue: run, wait or drop; sets `ctx.status = "issuing"`; returns a function that gives the slot back, or `null` | none | 008 |
| `__confirmRequest(ctx)` | only when `ctx.confirm` is set | `ypy:confirm` | 006 (frozen) |
| `__beginRequest(ctx)` | timeout, indicators, `ypy-disable`, then `__markBusy(ctx)` | none | 006 (frozen) |
| `__sendRequest(ctx)` | final headers, fetch, sets `ctx.response` | `ypy:before:request` | 010 |
| `__receiveResponse(ctx)` | reads the body, `YPY-*` headers, status rules; true when the response is to be swapped | `ypy:before:response`, `ypy:after:request`, `ypy:response:error` | 009 |
| `swap(ctx)` | the swap phases, below | `ypy:before:swap`, `ypy:after:swap`, `ypy:finally:swap` | 006 (frozen) |
| `__failRequest(ctx, error, elt)` | records the error, fires it unless it is a silent abort | `ypy:error` | 010 |
| `__endRequest(ctx, release, elt)` | `YPY-Trigger` header, hides indicators, `__unmarkBusy(ctx)`, `__announce(ctx)`, gives the queue slot back | `ypy:finally:request` | 006 (frozen) |

`elt` is the element the request started from. The swap may replace `ctx.sourceElement`, and the
request events keep firing on the original one.

`__unmarkBusy` and `__announce` run for every admitted request, including one that was denied
a confirmation, so they can run without a `__markBusy` before them.

### The swap

| Phase | Does | Owner |
|---|---|---|
| `__buildSwapTasks(ctx)` | parses `ctx.text`; returns `[tasks, mainSwap]`, `mainSwap` is also `tasks[0]` | 012 |
| (in `swap`) | `ypy:before:swap`, then `__handleHistoryUpdate(ctx)` | 006 (frozen), 016 |
| `__runSwapTasks(ctx, tasks, mainSwap)` | runs every task, in a view transition when asked | 012 |
| `__finishSwap(ctx, mainSwap)` | `ypy:after:swap`, title, anchor, then `__applyScrollPolicy(ctx, ctx.scroll)` | 013 then 014 |

### Inserting one task

`__insertContent(task, cssTransition)` calls, in this order:

| Phase | Does | Owner |
|---|---|---|
| `__resolveSwapTarget(task)` | the element to swap into | 011 |
| `__captureFocus(swapStyle, parentNode)` | remembers the focused element | 007 |
| `__applySwapStyle(target, swapStyle, swapSpec, fragment, parentNode)` | the DOM work of the style; returns `[target, newContent]` | 006 (frozen), morph branch 007 |
| `__restoreFocus(focusInfo, swapSpec)` | gives the focus back | 007 |
| `__afterInsert(task, target, swapSpec, newContent, settleTasks, settleDelay, cssTransition)` | settle events and classes, `process`, autofocus, scroll | 013 then 014 |

`__insertContent` itself also does the body strip, the `swapping` class, the swap delay, the
`delete` and `none` styles, the preserved elements (017) and the settle tasks.

## Hooks and utilities

Empty hooks, called by the orchestrators. The owner fills the body only.

| Hook | Called from | Owner |
|---|---|---|
| `__initHistory()` | `initialize` (has its body: the history listeners) | 016 |
| `__initOffline()` | `initialize` | 010 |
| `__initA11y()` | `initialize` | 013 |
| `__markBusy(ctx)` | `__beginRequest` | 013 |
| `__unmarkBusy(ctx)` | `__endRequest` | 013 |
| `__announce(ctx)` | `__endRequest` | 013 |
| `__applyScrollPolicy(ctx, policy)` | `__finishSwap`, and 016 for history restoration | 014 |

`initialize` can run more than once, so the three `__init*` bodies must be idempotent.

Shared utilities, owned by 006 (nobody else edits them):

- `__setState(elt, state, on)`: classes of `config.classes[state]` plus a token in
  `data-ypy-state`. Use it for every new state. When states start to be used on elements that
  morph, add `data-ypy-state` to `config.morphIgnore` (it belongs to task 002's `__initConfig`).
- `__triggerNative(on, name, detail)`: a `ypy:*` event, extension hooks
  first, `warn` and `error` details logged. Returns `false` when cancelled. Always bubbles.

## The request context (`ctx`)

The type is `YpyRequestCtx` in `src/ypy.d.ts`. Besides the fields that already existed
(`sourceElement`, `target`, `swap`, `request`, `response`, `text`, `status`...), these are
reserved. A name here is fixed; the exact values are decided by the task that owns it, which
tightens the type in its own `#region`.

| Field | Written by | Read by | Meaning |
|---|---|---|---|
| `ctx.errorKind` | 009, 010 | 013 | `validation`, `client`, `server`, `network` or `offline`; unset on success |
| `ctx.redirected` | 009 | 009, 016 | the response was reached through a redirect |
| `ctx.fallback` | 011 | 012, 009 | the swap target was missing and a fallback ran |
| `ctx.announce` | 009, 010, 011, attributes | 013 | message for the live region, or `false` for silence |
| `ctx.focus` | attributes | 014 | where the focus goes after the swap |
| `ctx.scroll` | attributes | 014, 016 | scroll policy, passed to `__applyScrollPolicy` |

`ctx.status` values, in order: `created`, `issuing`, `response received`, `swapped`, or
`error: <message>`. `ctx.keepIndicators` is true when an `YPY-Location` handed the indicators to
a follow-up request.

## Who owns which function

One owner per function. A function not listed belongs to the task that needs to change it first,
after asking the orchestrator.

| Function | Owner |
|---|---|
| `__issueRequest`, `swap`, `__insertContent`, `initialize`, `__confirmRequest`, `__beginRequest`, `__endRequest`, `__applySwapStyle` (but the morph branch is 007) | 006 (frozen) |
| `__setState`, `__triggerNative`, `test/lib/helpers.js` | 006 |
| `__initConfig` (but the `<meta>` block is 015), `__trigger` | 002 |
| `__attr`, `__attrName`, `__prefix*`, `__classNames`, `__addClass`, `__removeClass` | 001 |
| `RequestQueue`, `__admitRequest`, `__determineSyncStrategy`, `__getRequestQueue`, `__initializeAbortListener`, the queue part of `__cleanup` | 008 |
| `__receiveResponse`, `__handleStatusCodes` | 009 |
| `__initOffline`, `__sendRequest`, `__failRequest` | 010 |
| `__resolveTarget`, `__findOrWarn`, `__resolveSwapTarget`, the target in `__processMainSwap` | 011 |
| `__makeFragment`, `__createOOBTask`, `__processOOB`, `__processPartials`, `__insertNodes`, `__buildSwapTasks`, `__runSwapTasks` | 012 |
| `__markBusy`, `__unmarkBusy`, `__announce`, `__initA11y`, `__finishSwap`, `__afterInsert` (013 first, then 014) | 013 |
| `__setFocus`, `__handleAutoFocus`, `__handleScroll`, `__normalizeScrollTarget`, `__applyScrollPolicy` | 014 |
| the `<meta>` block of `__initConfig`, `registerExtension`, `__executeJavaScript`, `__parseHTML`, `__processScripts`, `src/ext/ypy-csp.js` | 015 |
| `__initHistory`, `__pushUrlIntoHistory`, `__replaceUrlInHistory`, `__restoreHistory`, `__resolveHistoryAction`, `__handleHistoryUpdate`, `ajax`, `src/ext/ypy-history-cache.js` | 016 |
| `__handlePreservedElements`, `__restorePreservedElements`, `onLoad`, the skip rules in `__morphNode` and `__copyAttributes` (after 007) | 017 |
| `__morph*`, `__copyAttributes`, `__captureFocus`, `__restoreFocus`, `__parseSwapSpec`, the morph branch of `__applySwapStyle` | 007 |
| `__attributeValue`, `__findThisElements` | 018 |

## Type definitions

`src/ypy.d.ts` ends with one region per task, `// #region TASK-007` to `TASK-018`. Declare your
additions by merging the interface in question inside your region:

```ts
// #region TASK-011 target fallback
export interface YpyNativeEventMap {
  'ypy:target:missing': { ctx: YpyRequestCtx; selector: string };
}
// #endregion
```

`YpyNativeEventMap` holds the events added after the first set.
