---
order: 10
title: What is different in YPY
description: For anyone who knows the older library: what the YPY core does on its own, what became an extension, which defaults changed and what was removed.
---

# What is different in YPY

YPY started from an older hypermedia library and now has its own names and its own behavior. This page is the summary
for someone who knows the old library: what the core does on its own, what became an extension, which defaults
changed and what is gone. The exact renames are in the [migration guide](index.md). [`ypy migrate`](migrate.md)
converts your files, and [`ypy check`](upgrade-check.md) lists what behaves differently in your project.

YPY has one behavior and one set of names. There is no compatibility layer, no switch that brings the old behavior
back, and no mixed spelling. A converted project runs on `ypy.min.js` from its first load.

## One set of names

| What | In YPY |
|---|---|
| Attributes | `ypy-*`, or `data-ypy-*` (when an element has both, `ypy-*` wins) |
| Tag | `<ypy-partial>` |
| HTTP headers | `YPY-*`, in requests and responses |
| Events | `ypy:<noun>:<verb>`, such as `ypy:before:request` and `ypy:after:swap` |
| Global object | `ypy` |
| State classes | `ypy-request`, `ypy-swapping`, `ypy-settling`, `ypy-added`, `ypy-indicator` |
| Extensions | `ypy-<name>`, in `config.extensions` and in `ypy.registerExtension` |

`ypy migrate` learns the old prefixes from your project and rewrites them. The [equivalence
tables](index.md#equivalence-tables) list every row.

## What the core does natively

These were separate pieces, or did not exist before, and now work with the core alone.

| Feature | In YPY |
|---|---|
| Morphing swap | `innerMorph` and `outerMorph` are core styles, and `innerMorph` is the default ([Swap styles and morph](../reference/swap-and-morph.md)) |
| Different target by status | `ypy-status:<code>` (also `4xx` and `5xx`), plus `ypy-error-target`, `ypy-error-swap` and `ypy-error-focus` ([`ypy-error-target`](../reference/attributes/ypy-error.md)) |
| Form validation errors | a `422` is swapped into the form, keeps what the user typed and focuses the first invalid field ([Forms](../guides/forms.md)) |
| Several targets from one response | `<ypy-partial>` ([reference](../reference/tags/ypy-partial.md)) and the out-of-band attributes `ypy-select-oob` and `ypy-swap-oob` |
| Request queue | `ypy-sync` with `queue`, `drop`, `abort` and `replace`, and the events `ypy:request:queued`, `ypy:request:dropped` and `ypy:request:aborted` ([`ypy-sync`](../reference/attributes/ypy-sync.md)) |
| Accessibility | `aria-busy` while a request runs ([`ypy-busy`](../reference/attributes/ypy-busy.md)), focus after a swap ([`ypy-focus`](../reference/attributes/ypy-focus.md), [`ypy-main`](../reference/attributes/ypy-main.md)), and a live-region announcement ([`ypy-announce`](../reference/attributes/ypy-announce.md)) |
| Missing target | `config.targetFallback` and the events `ypy:target:fallback` and `ypy:target:missing` ([events](../reference/events/target.md)) |
| Offline state | the events `ypy:offline` and `ypy:online` and the class `ypy-offline` ([offline](../reference/events/ypy-offline.md)) |
| Same-origin guard | `config.mode`, `"same-origin"` by default |
| Request timing | `ctx.timing` ([Timing](../reference/timing.md)) |
| Atomic updates | the response header `YPY-Group` ([Atomic swaps](../reference/atomic.md)) |
| Request and swap context | one `ctx` object in the `detail` of every event ([Events](../reference/events/index.md)) |
| Request transport | `fetch`, with an `AbortSignal`; there is no `XMLHttpRequest` object in the events |

## What is an extension

Some features that lived in the old core, or in one of its extensions, are now official extensions. Load them after
the core, or use `ypy.max.min.js`, which carries most of them.

| Feature | Extension |
|---|---|
| Server-sent events | [`ypy-sse`](../extensions/sse.md) |
| WebSocket | [`ypy-ws`](../extensions/ws.md) |
| Merging the `<head>` | [`ypy-head`](../extensions/head.md) |
| Preloading a link | [`ypy-preload`](../extensions/preload.md) |
| The prompt sent as a header | [`ypy-prompt`](../extensions/prompt.md) |
| Disabling elements during a request, and `ypy-inert` | [`ypy-disable`](../extensions/disable.md) |
| Snapshots for the back button | [`ypy-history-cache`](../extensions/history-cache.md) |

Watch `ypy-disable`: it used to be in the core and moved out to save bytes. If you use `ypy-disable` or `ypy-inert`
with `ypy.min.js`, load the extension.

## What YPY adds

These have no counterpart in the old library, so the converter has nothing to rename. All of them are in the
[extensions index](../extensions/index.md).

- requests: [`ypy-retry`](../extensions/retry.md), [`ypy-dedupe`](../extensions/dedupe.md),
  [`ypy-poll`](../extensions/poll.md), [`ypy-cache`](../extensions/cache.md),
  [`ypy-optimistic`](../extensions/optimistic.md);
- forms: [`ypy-validate`](../extensions/validate.md), [`ypy-switch`](../extensions/switch.md),
  [`ypy-upload`](../extensions/upload.md);
- navigation: [`ypy-nav`](../extensions/nav.md), [`ypy-progress`](../extensions/progress.md),
  [`ypy-persist`](../extensions/persist.md), [`ypy-lazy`](../extensions/lazy.md);
- real time: [`ypy-stream`](../extensions/stream.md), [`ypy-multipart`](../extensions/multipart.md),
  [`ypy-model`](../extensions/model.md) for streaming a language model answer;
- client state and code: [`ypy-signals`](../extensions/signals.md), [`ypy-live`](../extensions/live.md),
  [`ypy-expr`](../extensions/expr.md) (author code without `eval`),
  [`ypy-alpine-compat`](../extensions/alpine-compat.md);
- responses: [`ypy-download`](../extensions/download.md), [`ypy-upsert`](../extensions/upsert.md);
- interface: [`ypy-layers`](../extensions/layers.md), [`ypy-sort`](../extensions/sort.md),
  [`ypy-list`](../extensions/list.md), [`ypy-preserve`](../extensions/preserve.md);
- server to client: [`ypy-commands`](../extensions/commands.md);
- security: [`ypy-csp`](../extensions/csp.md), [`ypy-csrf`](../extensions/csrf.md), [`ypy-lock`](../extensions/lock.md);
- development and tests: [`ypy-dev`](../extensions/dev.md), [`ypy-inspector`](../extensions/inspector.md),
  [`ypy-telemetry`](../extensions/telemetry.md), [`ypy-test`](../extensions/test.md);
- native shells: [`ypy-native`](../extensions/native.md).

There is also a Go SDK for the server ([Server](../server/index.md)), with a test package that runs your handlers
without a browser.

## What behaves differently by default

Each row is a rule that `ypy check` reports on your files.

| Behavior | YPY | What to do |
|---|---|---|
| Default swap | `innerMorph`: focus, selection and scroll of live elements win over the incoming HTML | write `ypy-swap="innerHTML"` where you want a raw replacement |
| Attribute inheritance | explicit: an attribute reaches descendants only with `:inherited`, as in `ypy-target:inherited="#out"` | add `:inherited` where a container's attribute was meant for its children ([Inheritance](../reference/inheritance.md)) |
| Author code | off: `config.callbacks` is `"none"`, so `ypy-on`, `js:` values and `[filter]` triggers do not run | enable `"all"` or `"nonce"` on purpose ([Security](../reference/security.md)) |
| Cross-origin requests | blocked: `config.mode` is `"same-origin"` | change `mode` only for a page that needs it |
| Timeout | 60 000 ms by default, and `0` is not "none" | set `defaultTimeout` |
| Statuses that skip the swap | `204` and `304` (`config.noSwap`); `422` swaps into the form; other errors follow the error rules | use `ypy-status:<code>` for a different choice |
| History | a back or forward request asks the server again, with `YPY-History-Restore-Request`, and the server answers a full page; snapshots are not kept in `localStorage` | add `ypy-history-cache` if you want snapshots ([History](../reference/history.md)) |
| Request headers | `YPY-Target` is `tag#id`, not a bare id, and `YPY-Source` names the element that made the request | update server code that parsed the old value |
| `ypy-on` syntax | the colon form `ypy-on:click` and `ypy-on::after:swap`, one attribute per handler | rewrite the handlers that used one attribute for many events |
| Validation | the browser's own API | listen to the native `invalid` event |

## What was removed

| Gone | Instead |
|---|---|
| the attributes that set request values and filter parameters from script | `ypy-vals` with the `js:` prefix, or a `ypy:config:request` listener |
| the attribute with per-element request settings | `ypy-config` |
| the attribute that lists extensions on an element | load the extension script; restrict with `config.extensions` |
| the attributes that turn inheritance on and off | `:inherited` on the attribute |
| the `queue:` trigger modifier | `ypy-sync` |
| the old extension registration function | `ypy.registerExtension(name, hooks)` |
| helper functions of the JavaScript API (class helpers, `closest`, `remove`, `values`, `off`, logging) | the DOM and `console`; see [`ypy check --list-rules`](upgrade-check.md#reference-every-rule) |
| most of the old configuration keys | each has a named replacement in its rule; the renames are in [the migration guide](index.md#configuration) |
| events of the request object and of the validation layer | `ypy:before:request`, `ypy:finally:request`, `ypy:abort` and the native `invalid` event |
| the swap-by-event-name model of server-sent events | unnamed messages swap, and named events become DOM events ([`ypy-sse`](../extensions/sse.md)) |

`ypy check` flags a page that still carries a removed attribute, and the [`ypy-dev`](../extensions/dev.md) extension
warns when it meets one in the browser. The old names and their replacements are in
[the rule table](upgrade-check.md#reference-every-rule).

## Moving over

1. `ypy check .` to see the worklist.
2. `ypy migrate .` to see the diff, then `ypy migrate --write .` to apply it.
3. Swap the script tag for `ypy.min.js`, with the extensions you use.
4. Work through the `ypy check` list: inheritance, errors, author code, events.

The steps in full are in the [migration guide](index.md#the-path).
