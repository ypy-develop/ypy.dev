---
title: "YPY core performance plan"
description: "Task TASK-033. The raw results, CPU profile summaries and throw-away prototype patches of that session were measured on the core before the repository was restarted and are not kept;"
order: 20
---
# YPY core performance plan

Task TASK-033. The raw results, CPU profile summaries and throw-away prototype patches of that
session were measured on the core before the repository was restarted and are not kept; the
numbers below are their summary. Run `ypy bench` to measure the current core. All numbers were taken on
one machine (Apple M4 Max, Chrome 157.0.8103.51 driven by `ypy bench` over the DevTools protocol, headless,
5 fresh pages per scenario, pooled). They are script time only: no layout or paint is forced, so
the numbers show the cost of YPY itself.

## 1. How to run

```sh
go run ./cmd/ypy bench                              # measures dist/ypy.min.js (or $YPY_BUILD, or --build PATH)
go run ./cmd/ypy bench --ref <commit>               # a reference commit, built with ypy build in a temporary directory
go run ./cmd/ypy bench compare bench/results/A.json bench/results/B.json   # % difference, REGRESSION above 5 %
go run ./cmd/ypy bench show bench/results/A.json [--md]                    # print a stored result
go run ./cmd/ypy bench profile swap.innerHTML.1000 --build dist/ypy.js     # CPU profile
```

Useful options: `--rounds n` (default 5), `--filter <regex>`, `--quick` (smoke run, not a result),
`--uam` (adds `measureUserAgentSpecificMemory`, about 18 s per call), `--list` (scenario names),
`--out <dir>`, `--label <text>`. Results go to `bench/results/<date>-<sha>.json`. Profiles use
the unminified build (`dist/ypy.js`) to show function names; line numbers are those of
`src/ypy.js`. `go run ./cmd/ypy bench --help` lists every option. The harness is Go (chromedp and the
system Chrome, the same pieces as `ypy test`); only the scenarios themselves run as JavaScript, in
the page (`bench/pages/`). A build made by `--ref` is made by today's `ypy build`, so its size can
differ by a few bytes from the file recorded when that commit was current.

Scenarios (see `bench/pages/scenarios.js`): `swap.*` (innerHTML/outerHTML, 10/100/1000 rows, rows
with and without YPY attributes and ids), `morph.*` (innerMorph/outerMorph, 10/100/1000 rows with a
focused input whose value and selection must survive; `preserve_failures` must stay 0), `process.*`
(1000 rich rows, 5000 plain nodes), `attr.*` (own, inherited and missing attribute reads at depth
20 and 100), `ctx.create.*`, `events.*`, `request.*` (a request with no work, a click-driven
request, event counts per request), `session.500` (500 requests of a fixed mix, then the page is
put back to its first state through YPY and heap, DOM nodes and listeners are compared after a
forced GC), and `size.*` (min, min+gzip, min+brotli of the build).

How to read a result:

- Every metric is "lower is better". `ypy bench compare` marks `REGRESSION` above 5 % and appends `~`
  when the round-to-round noise (interquartile range of the round medians, stored as
  `noisePct`) is as large as the difference.
- The machine drifts: the same build measured two hours apart differed by 10 to 15 % in every
  scenario. **Only compare runs made back to back** and repeat the reference run when in doubt.
  Quiet machine, no other work in parallel (a parallel build moved the noise from 2 % to 20 %).
- Timing is not a CI gate on shared runners. Gate on the deterministic metrics (section 7).

## 2. What the cleanup (TASK-030) already solved

Measured with the harness, `47900c9e2` (baseline) against `46e371775` (cleanup merged; the core
source of the new `develop` is byte-identical to it):

| metric | unit | 47900c9e2 median | p95 | 46e371775 median | change |
|---|---|---:|---:|---:|---:|
| swap.innerHTML.10 | ms | 0.416 | 0.498 | 0.286 | -31% |
| swap.innerHTML.100 | ms | 3.38 | 4.22 | 2.44 | -28% |
| swap.innerHTML.1000 | ms | 44.1 | 48.7 | 33.6 | -24% |
| swap.outerHTML.1000 | ms | 36.3 | 42.3 | 29.6 | -18% |
| morph.inner.100 | ms | 2.07 | 2.51 | 1.79 | -13% |
| morph.inner.1000 | ms | 21.6 | 24.0 | 18.6 | -14% |
| morph.outer.1000 | ms | 15.3 | 18.1 | 14.8 | -3% |
| process.rich (1000 rows) | ms | 30.5 | 36.1 | 24.0 | -21% |
| process.plain (5000 nodes) | ms | 3.11 | 3.25 | 2.44 | -22% |
| attr.inherit.hit.20 | us | 2.74 | 3.42 | 1.89 | -31% |
| ctx.create.flat | us | 18.4 | 19.9 | 10.3 | -44% |
| events.trigger.native | us | 0.681 | 1.02 | 0.356 | -48% |
| request.overhead | us | 77.2 | 85.9 | 52.2 | -32% |
| session.total (500 requests) | ms | 545 | 561 | 372 | -32% |
| session.request | us | 1,015 | 2,710 | 780 | -23% |
| size.min_gzip | B | 14,260 | 14,260 | 14,080 | -1% |

Mechanism: one event dispatch instead of two (events per request 90 -> 45 for a 5-row swap,
20 -> 10 for an empty one), the per-name caches of attribute spellings and prefix selectors, and
the `hasAttributes()` shortcut. Retained heap after 500 requests: 284 -> 270 KB; retained DOM
nodes and listeners after the session: 0 and 0 in both (no leak).

Second measurement requested by the Orchestrator: the `develop` of 18783ab16 built with
`go run ./cmd/ypy build`. The core source
is identical to `46e371775`, so it is the same code with a different minifier. Against a rerun of
the `46e371775` terser build made right after it: swap, attributes, requests and the session are
within 0 to 5 %, `process.*` +6 to +8 %, morph +11 to +18 % (several marked `~`), core
min+gzip +1.7 % (14,319 B against 14,080 B). Worth one look by whoever owns the Go build: either
the minifier output is a little slower in morph, or the morph scenarios are noisier than the rest.
Not a defect of the core source.

The rest of this document is about what is left in `develop`.

## 3. Hot spots of the current core (`46e371775`)

CPU profiles, 4 s per scenario after a warm-up, unminified build, harness setup excluded
(`bench/results/profile-<scenario>-46e371775.json`). Percentages are of the time spent running
JavaScript or GC. `(program)` is native browser work with no JS frame (DOM mutation, parsing
bookkeeping); it is 33 to 39 % of the 1000-row swaps and cannot be attacked from the core.

| # | Hot spot | Evidence (profile) | Why it costs |
|---|---|---|---|
| H1 | **Attribute lookup** `#attributeValue` (`ypy.js:356`), `#attr`, `#attrName`, `#attrNames`, `CSS.escape`, `closest` | `attr.inherit.hit.100`: `closest` 82 % self. `ctx.create.deep20`: `#attributeValue` 94 % inclusive, `closest` 52 %, `escape` 11 %, `#attrNames` 8 %. `request.overhead`: `escape` 12 %, `#attrNames` 10 %, `#prefixSelector` 3.8 %, attribute reads 53 % inclusive. `process.plain`: `#maybeBoost` 35 % inclusive | every call rebuilds the escaped selector strings (`CSS.escape`, string concatenation, a Map lookup keyed on the new string) and probes up to 8 spellings with `getAttribute`; `closest` tests a list of 8 selectors at every ancestor |
| H2 | **Per-element lifecycle events** `#trigger` (`:1912`), `CustomEvent`, `dispatchEvent` | `swap.innerHTML.1000`: `#trigger` 16 % inclusive, 7 % self, `dispatchEvent`+`CustomEvent` 9 % self. `process.rich`: `#trigger` 18.5 % inclusive. `session.500`: 14.7 % | `ypy:before:init`, `ypy:after:init`, `ypy:before:process`, `ypy:after:process`, `ypy:before:on:init`, cleanup events: a bubbling, composed, cancelable `CustomEvent` per element and phase (45 events for a 5-row swap) |
| H3 | **ypy-on discovery by XPath** `evaluate` (`#ypyOnQuery`, `process`) | `process.plain` 17 %, `process.rich` 11 %, morph 8 to 10 %, `swap.innerHTML.1000` 5.5 % self | `process()` evaluates an expression over every descendant with `starts-with(name(), ...)` for each prefix; after a swap it runs once per inserted top-level node |
| H4 | **`process()` called once per child** (`#afterInsert`, `:1880`) | `swap.innerHTML.1000`: `process` 38 % inclusive; `morph.inner.1000`: 24 %; `morph.inner.noop.1000`: 27 % (nothing changed) | each child repeats `closest` for `ypy-ignore` (`#ignore`), two `querySelectorAll` with a 28-selector and an 8-selector list (`#queryEltAndDescendants` 30 % in `process.plain`), the XPath, the boost query and two events; after a morph every child is processed again though only inserted or re-initialised nodes need it |
| H5 | **Trigger spec parsing** `#parseTriggerSpecs` (`:392`), `HCON.parse` (`:19`), `HCON.split` | `process.rich`: `#parseTriggerSpecs` 18.6 % inclusive, `HCON.parse` 13 % self | the same few strings (`click`, `input changed delay:300ms`) are split by a lookahead regex and parsed again for every element |
| H6 | **Boost checks** `#maybeBoost` (`:1171`), `#shouldBoost`, `#isSameOrigin` (`:1195`) | `process.rich`: `#maybeBoost` 22.5 %, `#shouldBoost` 9.5 %, `#isSameOrigin` 5.8 % (`new URL` per link) | one `ypy-boost` attribute read (with inheritance walk) per `a` and `form`, and a URL parse per boosted link |
| H7 | **morph `#copyAttributes`** (`:2713`) | `morph.inner.1000` 17.5 % self, 22 % inclusive; `morph.outer.100` 21 % self, 27.6 % inclusive; `#populateIdMapWithTree` 3 to 4 % | walks `source.attributes` (an `Attr` wrapper per attribute), `startsWith` against the ignore list per attribute, and a second reverse pass with `hasAttribute`; runs for every matched node |
| H8 | **Trigger filters** `spec.handler` (`:1006`), `#executeJavaScript` (`:1117`) | `events.filter.click` 12.9 us per click: `#executeJavaScript` 56 % self, `spec.handler` 22.6 % self | every event enumerates all properties of the event with a regex (`for (k in evt)`) and compiles the filter with `new Function` again; `ypy-on` handlers cost 3 us per event |
| H9 | **Request context** `#createRequestContext` (`:464`), `#determineMethodAndAction` (`:403`) | `ctx.create.deep20` 22 us: `#determineMethodAndAction` 17 % (up to 6 full attribute reads, one per verb); `request.overhead` 52 us for a request that swaps nothing | about 12 attribute reads per request, each a full H1 lookup |
| H10 | **GC pressure** | `session.500`: `(garbage collector)` 17.8 % self | per wired element: about 6 closures and 3 objects in `#onTrigger` (`inner`, `timed`, `spec.handler`, listener record, spec), plus one `CustomEvent` per event |
| H11 | **`#cleanup`** (`:1216`) | `swap.innerHTML.1000` 9.8 % inclusive; `swap.outerHTML.1000` 10.1 % | one `querySelectorAll('[data-ypy-powered]')` per child, array spread, two events per powered element, `removeEventListener` per listener (the events and listener removal are the real cost, see X1) |
| H12 | **`data-ypy-powered` defeats `isEqualNode`** | morph scenarios: rows with a wired descendant are never "equal" to the new row, so `__findBestMatch` falls back to soft matches and `__morphNode` always descends | the attribute is on the old DOM and not on the new fragment |

Not hot: HTML parsing (`parseHTMLUnsafe` 5 to 7 %, inherent), `#makeFragment` regexes, history
updates, indicators.

## 4. Optimization plan, ordered by gain over risk

"Measured" means a prototype was built on top of the `46e371775` core and run through the same
harness in the same session (throw-away prototypes, not kept, not reviewed code, and the test
suite was **not** run on them: the owning task must
re-implement and run the suites). The percentages are against the same-session reference.
"Estimated" means derived from the profile share only.

### Tier A: no behavior change, internal

Status after TASK-035: A2, A4 and A7 are done, A1 was already in the core, A3 and A5 were measured and removed for bytes, A6 and A8 are not done; see section 8 for what was measured.

| ID | Change | Where | Gain | Compatibility risk | Test that protects it |
|---|---|---|---|---|---|
| A1 | **Attribute record cache.** One record per `ypy-*` name (spellings per prefix for the plain, `:inherited`, `:append` and `:inherited:append` forms, plus the two inheritance selectors for explicit and implicit mode), invalidated when `config.prefix` or `config.metaCharacter` changes; `__attributeValue` reads it instead of rebuilding strings. Also cache the hook key of `__triggerExtensions` (`eventName.replace(/:/g,'_')`) per event name. Patch `A1-attribute-record-cache.patch` | `__attributeValue`, `__attrNames`, `__triggerExtensions` | **measured**: `ctx.create.flat` -50 %, `ctx.create.deep20` -31 %, `request.click` -25 %, `request.overhead` -30 %, `swap.*.10` -22 %, `swap.*.1000` -11 %, `process.rich` -10 %, `session.total` -24 % | low; semantics unchanged; the cache must be dropped when the prefix or the meta character changes at run time | `test/tests/unit/__attributeValue.js`, `attributes/inheritance*`, `unit/ypy.config.implicitInheritance.js`, prefix and metaCharacter config tests; add one that changes `config.prefix` after a first read |
| A2 | **Trigger spec cache.** Cache the parsed specs per trigger string and hand every element a copy (deep copy of nested objects, because `__onTrigger` writes `listeners`, `timeout` and so on into the spec). Do not cache a parse that throws. Patch `A2-trigger-spec-cache.patch` (apply on A1) | `__parseTriggerSpecs` | **measured on top of A1**: `process.rich` -33 % in total (-23 points beyond A1), `session.total` -31 % (-7 points), `swap.innerHTML.1000` -17 % (-5 points) | low; the risk is shared mutable state between elements, covered by the deep copy | `unit/__parseTriggerSpecs.js`, `ypy-trigger` tests (delay, throttle, from, filter, `every`), a test that two elements with the same trigger do not share a spec |
| A3 | **`ypy-ignore` once per `process()`.** One `querySelector` for `[ypy-ignore]` inside the root decides whether the per-element `closest` checks (three call sites: root, ypy-on nodes, `__shouldInitialize`) can be skipped for this call. Patch `A3-process-ignore-scope.patch` | `process`, `__ignore` | **measured on top of A1**: `process.rich` -14.5 % (-4 points), `swap.innerHTML.1000` -16 % (-4 points) | low to medium: an extension that adds `ypy-ignore` from a `ypy:before:process` listener while processing would not be seen; keep the check on elements added during the call | `attributes/ypy-ignore.js` plus a test that a `ypy:before:process` listener adds `ypy-ignore` to a descendant |
| A4 | **Trigger filters: compile once, read event properties lazily.** Compile the filter function when the trigger is set up (keyed by filter text), and instead of copying every lower-case property of the event, expose the event through `with(evt)` or a small Proxy so the expression reads what it names | `__onTrigger` (`spec.handler`), `__executeJavaScript` | **estimated**: `events.filter.click` 12.9 us -> about 2 to 3 us (the two costs are 79 % of it) | medium: today only properties matching `/^[a-z]+$/i` are in scope (for example `ctrlKey` yes, `_x` no); keep the same set | `ypy-trigger` filter tests, a test using `ctrlKey`, `key`, `target`, and a filter that references a missing name |
| A5 | **`__isSameOrigin` fast path.** No `new URL` for hrefs that cannot leave the origin (start with `/` but not `//`, `?`, `#`, or have no scheme), and a per-href cache for the rest | `__isSameOrigin` | **estimated**: `process.rich` 5.8 % inclusive -> about 1 % | low | `attributes/ypy-boost.js` (same-origin, cross-origin, protocol-relative, `javascript:`) |
| A6 | **`__determineMethodAndAction`**: one pass over the element's own attributes to find the verb attribute before the six full lookups (keep the lookup for `:inherited` forms) | `__determineMethodAndAction` | **estimated**: 17 % of `ctx.create.deep20`, about 4 us per request | low | `unit/__determineMethodAndAction.js`, verb tests for all six verbs and `ypy-action` |
| A7 | `__startCSSTransitions`: return before scanning the old target when the new fragment has no `[id]`; `Map` instead of an object keyed by id. Patch `A7-css-transitions-early-exit.patch` | `__startCSSTransitions` | **measured, small**: not visible on id-less swaps; `swap.ids.innerHTML.1000` improved with the other changes (2.7 % self in the profile) | low | `swap/settle*` tests |
| A8 | **GC: share handlers between elements.** `__onTrigger` allocates `inner`, `timed`, `spec.handler`, a listener record and a spec per element; build the common case (no delay, throttle, filter, `from`, `once`) as one small handler class with the spec as a field | `__onTrigger` | **estimated**: part of the 17.8 % GC of the session, 5 to 8 % of `session.total` | medium: the listener identity is used by `removeEventListener`; keep one handler per spec | `__cleanup` tests, `ypy-trigger` modifier tests |

Tried and **dropped** (measured, no gain, do not redo):

- `X1-cleanup-single-query`: one `querySelectorAll` for the whole target instead of one per child
  changes nothing measurable (`swap.innerHTML.1000` -9.7 % against -11.6 % for A1 alone, noise).
  The cost of `#cleanup` is the events and the listener removal, not the query.
- `X2-morph-micro`: hoisting the `ypy-` prefix list, caching `activeElement`, skipping the
  `querySelector('template')` per node, a loop instead of `[...oldSet].some`. Morph scenarios
  moved by -1 to +3 %. `#copyAttributes` cost is the DOM reads themselves.

### Tier B: needs a decision from the author (observable behavior or public contract)

| ID | Change | Gain | Risk and decision |
|---|---|---|---|
| B1 | **Morph processes only what it inserted or re-initialised.** Today `__afterInsert` calls `process()` on every child of the morph target. Collect nodes the morph inserted or whose `ypy-*` attributes changed (`needsReinit`) and process only those. Patch `B1-morph-process-touched-only.patch` (cut on top of X2; re-cut it on the clean core) | **measured**: morph scenarios -50 % to -70 % (`morph.inner.1000` 18.6 -> 6.4 ms, `morph.inner.noop.1000` -69 %), `session.request` median -42 % | **behavior change**: `ypy:before:process`, `ypy:after:process` (and so `onLoad` callbacks), the `added` class and `autofocus` no longer fire for matched, unchanged nodes. Arguably today's behavior re-runs author init code on every morph, so this is also a fix. Needs an ADR and tests that count process events after a morph. **Biggest single win in the plan** |
| B2 | **Lifecycle events cost.** `ypy:before/after:init` and `ypy:before/after:cleanup` are dispatched per element. Upper bound with the events removed: `swap.*.100` -19 % (outerHTML -35 %), `process.rich` -24 %, `session.total` -13 %. Patch `B2-no-lifecycle-events-UPPER-BOUND.patch` (not a proposal: removing them breaks the documented events) | measured upper bound | options: (a) keep as is; (b) dispatch only when a listener exists, which needs the app to register through a YPY API (breaks plain `addEventListener`); (c) one batched event per `process()` root next to the per-element ones, opt-in. Author decision, since the events are public API |
| B3 | **Two attribute spellings.** Applied with R-11: an attribute read tests `ypy-` and `data-ypy-` (and `config.prefix`, default `""`, when set). Patch `B3-two-prefixes-DECISION.patch` (the prototype, written against the earlier code that tested four spellings) | **measured** (prototype): `process.plain` -40 %, attribute reads -46 %, `ctx.create` -33 to -42 %, `request.click.deep20` -19 %, `swap` -10 %, `morph` -7 %, `session.total` -8 %. **measured** (TASK-040, on the real change): `process.plain` -36 %, `attr.inherit.*` -28 to -44 %, `ctx.create.flat` -31 %, `ctx.create.deep20` -41 %, `request.click.deep20` -24 %, core -11 B min and -49 B gzip | done | `test/tests/unit/ypy.config.prefix.js`, `perf-caches.js`, `inheritance-native.js` |
| B4 | **Move the "powered" marker out of the DOM** (H12). `data-ypy-powered` on old nodes makes `isEqualNode` false for every row with a wired descendant, so exact matches and the "equal subtree, skip" shortcut never apply. A registry instead of an attribute (`Set` plus `contains` in cleanup, or marking by `_ypy` only) would let `isEqualNode` work | **estimated**: large on morph of lists of wired rows, not prototyped | the attribute is selectable by authors' CSS and used by `__cleanup`, `morphIgnore` and tests; design task for Phase 1 |
| B5 | **ypy-on without XPath (H3).** Decide from the response text (does it contain an `-on` attribute for any prefix) whether the XPath scan is needed for the inserted content, and let `process` take that flag; a full document or a morph still scans | **estimated**: `evaluate` is 5.5 % of `swap.innerHTML.1000`, 11 % of `process.rich`, 17 % of `process.plain` | the flag must be exact, a false negative loses a handler; needs a test with `ypy-on` inside nested templates and partials |

### Tier C: later, structural

- Resolved-value cache for inherited attributes, valid for one swap or one request (H1: `closest`
  dominates deep inheritance even after A1).
- One scan per swap target instead of `process()` per child (H4): batch the action, boost and ypy-on
  queries over the target and keep the per-root events. Keep only if B1 and A3 leave it above 20 %.
- Reuse of `Attr`-free attribute comparison in morph (H7), if B1 does not make it irrelevant.

## 5. Order of work

1. **A1, A2, A3** (one task, `__attributeValue` / `__parseTriggerSpecs` / `process`): together about
   -25 to -35 % on swap and process, -50 % on request context creation. Owners of the inheritance
   task (TASK-018) touch `__attributeValue`: do A1 inside that task or right after it, to avoid
   two edits of the same function.
2. **B1** after an ADR (morph process events). Largest gain.
3. A4, A5, A6, A7 (small, independent).
4. B2, B3, B4, B5 as author decisions, each with the prototype patch and its measured number.
5. A8 and Tier C only if a re-profile after steps 1 to 3 still shows them.

After each step: run the suite, then `go run ./cmd/ypy bench compare <reference>.json <new>.json`, record
the result file in `bench/results/`, and re-profile the scenario the step targeted.

## 6. Limits of this work

- Script time only. Layout, paint and network are not in the numbers; an extra `requestAnimationFrame`
  timing mode is not implemented.
- Chrome only. Firefox and WebKit have
  different costs for `closest`, XPath and `CustomEvent`; the plan's ranking may change there.
- Fixtures are synthetic (list rows with two wired elements, an input row for morph). Pages with
  deep component trees or large forms weigh H1, H4 and H7 differently.
- The prototypes were not run against the test suite.
- `measureUserAgentSpecificMemory` is optional (`--uam`) because it takes about 18 s per call; the
  heap numbers come from forced GC through the DevTools protocol.
- An A/B mode that interleaves two builds round by round would remove the machine drift described
  in section 1; not implemented, run the two builds back to back.

## 7. CI step

`ypy ci bench-smoke` is the cheap gate: a quick run of the scenarios whose metrics are
deterministic (event counts, focus and input preserved, retained nodes and listeners, size),
compared with the most recent full result in `bench/results/` at threshold 0. It is off by default
(`ypy ci` does not run it) so the CI stays fast; name it to run it, and run it before a change
that touches the request path.

```sh
go run ./cmd/ypy ci bench-smoke
```

Timings only inform, because shared runners vary by far more than 5 %. Record a new reference
(`go run ./cmd/ypy bench`, commit the file in `bench/results/`) in the same commit as any intended
change in the deterministic metrics. `ypy ci sizes` stays the size gate.

## 8. Result of TASK-035 (Tier A on the core of `f6026d087`)

The prototypes of section 4 were cut on an older core, so every item was redone on the current
one (`__attributeValue` already rewritten by TASK-018, prefix spellings cached, aliases gone).
One commit per item on `work/TASK-035-perf-core`. Chrome 157, Apple M4 Max, `PW_CHANNEL=chrome`.

### Status per item

| ID | Status | What changed | Measured (same session, builds interleaved) |
|---|---|---|---|
| A1 | **already in the core** (TASK-018); two leftovers measured and **not applied** | names and selectors per attribute are built once and kept with the prefix selectors, so they go away with `config.prefix` or `config.metaCharacter`; the hook runs only when registered. Leftover 1: the lists of spellings (then four, now two) kept inside that record (no `__attrNames` lookup per read): `attr.inherit.*.20` **+17 %** in a repeated A/B, `ctx.create.*` inside the noise. Leftover 2: a `Map` for the hook key of `__triggerExtensions`: `events.trigger.*` unchanged, +54 B | none worth keeping |
| A2 | **done** | `__parseTriggerSpecs` keeps the parse of each trigger string (at most 500) and hands every caller its own copy; a parse that throws, or that holds a nested value, is never kept | `process.rich` -24 % (22.7 -> 17.4 ms), `swap.innerHTML.1000` -4 % |
| A3 | **tried, not kept** | one `closest` and one `querySelector` per `process()` call, lazily, so `__ignore` skips its per-element `closest` | `process.rich` -5 % (19.1 -> 18.1 ms), swaps unchanged in the noise; **+41 B**. Removed to fit the size budget (lowest gain per byte) |
| A4 | **done, differently from the plan** | `__executeJavaScript` passes only the API methods that the code names as parameters (a direct `eval` or an escaped identifier keeps all of them) and a trigger filter copies only the event properties its text names. The function cache is the one of the callbacks policy (TASK-015, 256 entries, keyed on the parameter list and the source, cleared by `initSecurity`). The plan's `with(evt)` or Proxy was not used: the constructor is still called for every call and what is cached is the function, never the decision (the nonce gate) | `events.filter.click` 13.1 -> 2.8 us (**-79 %**), `events.on.click` 3.1 -> 2.5 us (-20 %) |
| A5 | **tried, not kept** | `__isSameOrigin` answers at once for an absolute URL that starts with the page origin and a slash | `process.rich` -8 % (18.9 -> 17.4 ms); **+26 B**. Removed to fit the size budget |
| A6 | **not done** | `closest` is 68 % of `ctx.create.deep20` and each verb lookup is an ordinary `__attributeValue` read. Skipping the lookups that cannot match needs either the resolved-value cache of Tier C or a second copy of the lookup logic | no prototype worth measuring; Tier C item stays open |
| A7 | **done** | `__startCSSTransitions` returns before scanning the old content when the new one has no `[id]`; ids in a `Map` | `swap.*.1000` -1 to -2 %, as the plan expected (small) |
| A8 | **not done** | of the five allocations per wired element only one (`inner`, when no modifier is set) can go without changing the identity of the listener that `removeEventListener` needs; one closure in five is below what this machine can measure (+-4 %) | not measured |

### Behavior notes (none changes the documented behavior)

- A4: the `ypy:scope` hook, the callbacks policy and the constructor still run on every call. A hook that rewrites a filter so that it reads an event property the original text did not name would not see that property.
- Caches: trigger specs (500), compiled author JavaScript (500), per-name attribute records (unchanged). All of them are bounded and are dropped when full; none holds an element.
- `session.heap_retained` of the harness is the heap at the end minus the heap after a warm-up of 20 requests. The warm-up leaves about 15 KB less behind in the new build (2006 KB against 2020 KB), so the retained figure looks about 6 % higher, but the heap at the end of the session is the same (2309 to 2311 KB against 2308 to 2310 KB). Nodes and listeners retained: 0 and 0 in both.

### Size

On the `develop` that has the whole of phase 1 the core was 18,197 B and is now 18,304 B min+gzip (+107 B) of 18,432 B; the three items kept (A2, A4, A7) cost +107 B together. On that older core all five items cost +324 B (16,975 -> 17,299 B), too much for the budget after the merge, so A3 (+41 B, -5 %) and A5 (+26 B, -8 %) were removed, in that order. Savings tried to pay for them gave nothing: merging the four indicator and disabled-element counter functions into two helpers cost +5 B (gzip already folds the repeated code), `__trigger` and `__triggerNative` log in different formats that tests pin, and no function of the core is unused. The merge with the callbacks cache of TASK-015 recovered the bytes of the second cache and of the bound-method cache.

### Before and after

Method (measured before the merge with the phase 1 `develop`, with A3 and A5 still in the build; the figures of A2, A4 and A7 are the ones that stand): `f6026d087` (before) against the build of `9213941a3` (after), six full runs of 5 rounds each, three per build, alternated as before, after, after, before, before, after. Each cell is the median of the pooled round medians. The machine was not idle: another benchmark ran in parallel during part of the runs, so a difference of less than about 8 % is not a result (the first three runs alone put the swaps at -4 % to +4 %, all six at -2 % to -13 %, for code these changes barely touch). What the changes explain: `process.rich`, `events.on.click` and `events.filter.click`. The unit of `size.*` is the harness's own (gzip level 9); the Go build measures 16,975 B before and 17,299 B after (`go run ./cmd/ypy build --sizes`).

| metric | unit | before | after | change |
|---|---|---:|---:|---:|
| swap.innerHTML.10 | ms | 0.2879 | 0.2589 | -10% |
| swap.outerHTML.10 | ms | 0.2366 | 0.2087 | -12% |
| swap.innerHTML.100 | ms | 2.404 | 2.356 | -2% |
| swap.outerHTML.100 | ms | 1.956 | 1.771 | -9% |
| swap.innerHTML.1000 | ms | 36.36 | 33.53 | -8% |
| swap.outerHTML.1000 | ms | 31.38 | 28.91 | -8% |
| swap.plain.innerHTML.100 | ms | 1.034 | 0.9488 | -8% |
| swap.ids.innerHTML.100 | ms | 2.577 | 2.34 | -9% |
| swap.plain.innerHTML.1000 | ms | 22.66 | 21.86 | -4% |
| swap.ids.innerHTML.1000 | ms | 39.03 | 34.11 | -13% |
| morph.inner.10 | ms | 0.2527 | 0.231 | -9% |
| morph.inner.10.preserve_failures | count | 0 | 0 | 0 |
| morph.outer.10 | ms | 0.2374 | 0.2021 | -15% |
| morph.outer.10.preserve_failures | count | 0 | 0 | 0 |
| morph.inner.100 | ms | 1.956 | 1.929 | -1% |
| morph.inner.100.preserve_failures | count | 0 | 0 | 0 |
| morph.outer.100 | ms | 1.53 | 1.529 | -0% |
| morph.outer.100.preserve_failures | count | 0 | 0 | 0 |
| morph.inner.1000 | ms | 21.91 | 21.36 | -3% |
| morph.inner.1000.preserve_failures | count | 0 | 0 | 0 |
| morph.outer.1000 | ms | 16.76 | 16.39 | -2% |
| morph.outer.1000.preserve_failures | count | 0 | 0 | 0 |
| morph.inner.noop.100 | ms | 1.732 | 1.703 | -2% |
| morph.inner.noop.100.preserve_failures | count | 0 | 0 | 0 |
| morph.inner.noop.1000 | ms | 17.57 | 16.42 | -7% |
| morph.inner.noop.1000.preserve_failures | count | 0 | 0 | 0 |
| process.rich | ms | 24.53 | 16.91 | -31% |
| process.plain | ms | 2.123 | 2.057 | -3% |
| attr.local | us | 0.07875 | 0.07875 | +0% |
| attr.inherit.hit.20 | us | 1.446 | 1.389 | -4% |
| attr.inherit.miss.20 | us | 1.407 | 1.4 | -1% |
| attr.inherit.hit.100 | us | 5.409 | 5.399 | -0% |
| attr.inherit.miss.100 | us | 5.512 | 5.365 | -3% |
| attr.inherit.append.20 | us | 1.425 | 1.395 | -2% |
| attr.inherit.implicit.100 | us | 5.492 | 5.489 | -0% |
| attr.mixed.20 | us | 1.446 | 1.435 | -1% |
| ctx.create.flat | us | 5.91 | 5.86 | -1% |
| ctx.create.deep20 | us | 17.59 | 16.91 | -4% |
| events.trigger.plain | us | 0.36 | 0.3538 | -2% |
| events.trigger.native | us | 0.3613 | 0.375 | +4% |
| events.trigger.native.listeners5 | us | 0.92 | 0.8475 | -8% |
| events.on.click | us | 3.225 | 2.475 | -23% |
| events.filter.click | us | 14.11 | 2.788 | -80% |
| request.click | us | 49.45 | 47.7 | -4% |
| request.click.deep20 | us | 80.2 | 75.75 | -6% |
| request.overhead | us | 44.1 | 43.3 | -2% |
| request.overhead.listeners | us | 44.4 | 44.6 | +0% |
| request.events.empty.dispatched | count | 10 | 10 | +0% |
| request.events.empty.distinct | count | 10 | 10 | +0% |
| request.events.rows5.dispatched | count | 45 | 45 | +0% |
| request.events.rows5.distinct | count | 15 | 15 | +0% |
| session.total | ms | 391.7 | 375 | -4% |
| session.request | us | 795 | 762.5 | -4% |
| session.heap_peak | KB | 324 | 342.9 | +6% |
| session.heap_retained | KB | 287.8 | 305.3 | +6% |
| session.nodes_retained | count | 0 | 0 | 0 |
| session.listeners_retained | count | 0 | 0 | 0 |
| size.min | B | 46,802 | 47,502 | +1% |
| size.min_gzip (bench, gzip -9) | B | 16,919 | 17,227 | +2% |
| size.min_brotli | B | 15,209 | 15,494 | +2% |

