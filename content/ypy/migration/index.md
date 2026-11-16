---
title: "Migration guide: from legacy markup to YPY"
description: "How to move a project written for the older library to YPY: the conversion command, the behavior review it leaves you, and the tables of what was renamed."
order: 70
id: "GUIDE-migration"
type: "guide"
status: "draft"
lang: "en"
updated: "2026-11-15"
tags: ["type/guide", "lang/en", "topic/migration"]
---

# Migration guide: from legacy markup to YPY

Moving a project from the older library to YPY has two parts. The first is mechanical: one command renames the
attributes, headers, events and classes, and you swap the old script file for `ypy.min.js`. The second is a review:
YPY behaves differently in a few places, and a renamed project does not behave the way it did. A worklist tells you
where. There is no compatibility layer, so a converted project runs on `ypy.min.js` from its first load.
Portuguese: [Guia de migração](../pt-BR/guides/migration.md).

This page has the path and the equivalence tables. The two tools have their own pages: [`ypy migrate`](migrate.md)
(every option and the JSON output) and [`ypy check`](upgrade-check.md) (the worklist, with one rule per behavior that
changed). There is no compatibility layer: YPY only has its native behavior, so a converted project runs on it from the first load.

## How names are written in this guide

The older markup spells attributes, headers, events and classes with names that YPY does not use. This guide writes
them as placeholders. `ypy migrate` learns the real ones from your project, or you pass them with a flag:

| Placeholder | Is | Example shape (a made-up one) |
|---|---|---|
| prefix `<p>-` | the legacy attribute prefix: `<p>-get`, `data-<p>-target`, `<p>-on::after-swap` | `acme-get` |
| prefix `<P>-` | the legacy header prefix: `<P>-Request`, `<P>-Target` | `ACME-Request` |
| `<id>` | the legacy identifier: its global object, its event prefix, its state classes | `acme.process(el)`, `acme:afterSwap`, `acme-request` |

If your project does not reveal them (a back end repository that only reads one header, say), pass them yourself:
`ypy migrate --prefix acme --header-prefix ACME --id acme .`

## The path

| Step | You do | Result |
|---|---|---|
| 1. Convert | `ypy migrate --write`, then load `ypy.min.js` instead of the old script | the names are YPY's and the project loads the YPY core |
| 2. Review the behavior | work through what `ypy check` lists: explicit inheritance, error responses, author code, events; page by page, at the pace the project allows | the project behaves the way YPY does, with nothing left to switch off later |

There is no bridge: no compatibility file, no profile, and no setting that brings the old behavior back. What changes
in behavior is listed [below](#what-changes-in-behavior), and `ypy check` finds it in your own files.

Your backend needs no change beyond the names. YPY sends and reads the same headers with the `YPY-` prefix, and a
server converted by `ypy migrate` reads `YPY-Request` where it read `<P>-Request`. The new headers (`YPY-Source`,
`YPY-Request-Type`, ...) come on top.

## Step by step

1. **Start from a clean tree.** Commit what you have. The conversion is an ordinary diff you can review and revert.
2. **Read the worklist.** `ypy check .` lists what breaks (`quebra`) or behaves differently (`muda`) besides the
   spelling, with the line, the fix and a link to the rule. It changes nothing. Each finding has a severity:

   | Severity | Meaning | Before you ship |
   |---|---|---|
   | `quebra` | stops working after the move to YPY | fix it |
   | `muda` | works, but behaves differently from what the project was written for | review it |
   | `aviso` | works; only the spelling is legacy | `ypy migrate` converts it |

3. **Dry run.** `ypy migrate .` prints the prefixes it learned and the diff, and writes nothing. Check that the
   prefixes and the identifier are the right ones.
4. **Write.** `ypy migrate --write .`, then `ypy migrate .` again: it must report `0 of N file(s) would change`.
5. **Swap the script.** Put `dist/ypy.min.js` where the old script was, delete the old file and its extension
   scripts, and load the official extensions you used from `dist/ext/` (see the [extension table](#extensions)
   below), or use `dist/ypy.max.min.js`, which carries them.
6. **Fix the behavior.** Go through the `ypy check` worklist (see [What changes in behavior](#what-changes-in-behavior)).
7. **Run your own gates**: formatting, build, vet, unit tests, browser tests. A test that asserts the old spelling of
   an attribute or header was converted with the rest.
8. **Look at the result** in a browser with the console open: one page of each kind (a list, a form, a search, a stream).

If several projects depend on each other (a framework, a UI kit, products built on them), convert in dependency
order: the libraries that write the attributes first, then the products. The longer note is in
[`ypy migrate`](migrate.md#step-by-step-for-an-arandu-or-kyse-project).

## What changes in behavior

The conversion renames. It cannot make a page behave the way it did. These are the differences, and each one is a
rule of `ypy check`:

| Behavior | YPY |
|---|---|
| Default swap | morph (`innerMorph`): live state such as focus, selection and scroll wins over the incoming HTML. Use `ypy-swap="innerHTML"` where you want a raw replacement ([Swap styles and morph](../reference/swap-and-morph.md)) |
| Attribute inheritance | explicit: an attribute on a container reaches its descendants only with `:inherited` (`ypy-target:inherited="#out"`). `ypy migrate` cannot know which container attributes your descendants rely on, so `ypy check` flags them ([Attribute inheritance](../reference/inheritance.md)) |
| `422` from a form | swapped into the form ([Forms](../guides/forms.md)) |
| Other 4xx and 5xx | decided by the core: error targets and `ypy-error-*` ([`ypy-error`](../reference/attributes/ypy-error.md)); `ypy-status:<code>` changes a status |
| Author code (`ypy-on`, `js:` values, event attributes) | `callbacks: "none"` by default; opt in with `callbacks:"all"` knowing what that means ([Security checklist](../guides/security-checklist.md#author-code)) |
| `aria-busy` while a request runs | automatic ([`ypy-busy`](../reference/attributes/ypy-busy.md)) |
| Default request timeout | 60 s (`defaultTimeout`); `0` is not "no timeout" |
| Cross-origin requests | same origin only (`mode`) |
| Event names and `detail` | `ypy:<noun>:<verb>`, with a `detail` of `{ctx}` ([Events](../reference/events/index.md#detailctx)) |
| Attributes that no longer exist | `<p>-vars`, `<p>-params`, `<p>-request`, `<p>-ext`, `<p>-inherit`, `<p>-disinherit`, `<p>-history`, the `queue:` trigger modifier, and the one-handler-per-line form of `<p>-on` (replacements below) |
| JavaScript API and configuration | the legacy functions and configuration names are gone; the replacements are in the tables below |

## Equivalence tables

The tables are what `ypy migrate` does (spelling) and what `ypy check` reports (behavior). Every row that is not a
plain rename is a rule in [the rule table](upgrade-check.md#reference-every-rule).

### Attributes

Every attribute keeps its name and loses the prefix: `<p>-get` is `ypy-get`, `data-<p>-target` is `data-ypy-target`,
`<p>-on::after-swap` is `ypy-on::after-swap`. These are the ones whose **meaning** or **availability** changed:

| Before | After |
|---|---|
| `<p>-get`, `-post`, `-put`, `-patch`, `-delete`, `-target`, `-swap`, `-trigger`, `-select`, `-include`, `-indicator`, `-confirm`, `-push-url`, `-boost`, `-headers`, `-vals`, `-sync`, `-on:<event>` | the same name with `ypy-` |
| `<<p>-partial>` (tag) | `<ypy-partial>` |
| `<p>-vars` | `ypy-vals` with the `js:` prefix: `ypy-vals="js:{a: 1}"` |
| `<p>-params` | removed: filter in a `ypy:config:request` listener (`detail.ctx.request.body` is a `FormData`) |
| `<p>-request` | `ypy-config` (same JSON: `timeout`, `credentials`, `noHeaders`) |
| `<p>-disabled-elt` | `ypy-disable` |
| `<p>-disable` (no value: skip processing) | `ypy-ignore` (do this **before** renaming the one above) |
| `<p>-ext` | deleted: an extension is active when its script is loaded; restrict with `config.extensions` |
| `<p>-inherit`, `<p>-disinherit` | deleted; write `:inherited` on the attributes that must reach descendants |
| implicit inheritance of `<p>-target`, `-swap`, ... | explicit: `ypy-target:inherited="#out"` |
| `<p>-history="false"` | deleted (snapshots are no longer kept in `localStorage`) |
| `<p>-prompt` | the `ypy-prompt` extension (same attribute, same `YPY-Prompt` header) |
| `queue:` modifier of the trigger | `ypy-sync="this:queue last"` (or `queue first`, `queue all`, `drop`, `abort`) |
| `show:<selector>:<position>`, `scroll:<selector>:<position>` | `show:top showTarget:#sel`, `scroll:bottom scrollTarget:#sel` |
| the dash spelling of `ypy-on` (`ypy-on-click`) | the colon spelling `ypy-on:click`, `ypy-on::after:swap` |
| `<p>-on="event: code"` (one handler per line) | one `ypy-on:<event>` attribute per handler, or `ypy-on="click -> code"` |
| the `inherit` keyword of `-include` and `-indicator` | gone; use `:inherited` |

### Events

The prefix `<id>:` becomes `ypy:`, and the names follow one pattern, `ypy:<noun>:<verb>`. The `detail` is `{ctx}`
([Events](../reference/events/index.md#detailctx)); the legacy one (`xhr`, `requestConfig`, `parameters`, ...) does not exist.
`ypy migrate` converts the prefix, and `ypy check` lists every event whose name or `detail` differ, with the name to
use.

| Before | YPY |
|---|---|
| `<id>:beforeRequest`, `beforeSend` | `ypy:before:request` |
| `<id>:configRequest` | `ypy:config:request` |
| `<id>:afterRequest` | `ypy:after:request` (`ypy:finally:request` also on failure) |
| `<id>:beforeSwap` | `ypy:before:swap` (the swap decision moved to `config.noSwap` and `ypy-status`) |
| `<id>:afterSwap`, `oobAfterSwap` | `ypy:after:swap` |
| `<id>:afterSettle` | `ypy:after:settle` |
| `<id>:load` | `ypy:after:process` |
| `<id>:beforeProcessNode`, `afterProcessNode` | `ypy:before:init`, `ypy:after:init` |
| `<id>:beforeOnLoad`, `afterOnLoad` | `ypy:before:response`, `ypy:after:request` |
| `<id>:beforeCleanupElement` | `ypy:before:cleanup` |
| `<id>:responseError` | `ypy:response:error` |
| `<id>:sendError`, `sendAbort`, `timeout`, `swapError`, `targetError` | `ypy:error` (`ypy:abort` to abort, `ctx.errorKind` for the kind) |
| `<id>:historyRestore`, `historyCacheMiss` | `ypy:before:history:restore` |
| `<id>:pushedIntoHistory`, `replacedInHistory` | `ypy:after:history:push`, `ypy:after:history:replace` |
| `<id>:beforeHistoryUpdate`, `beforeHistorySave` | `ypy:before:history:update` |
| `<id>:beforeTransition` | `ypy:before:viewTransition` |
| `<id>:sseOpen`, `sseMessage`, `sseError`, `sseClose` | `ypy:sse:after:connection`, `ypy:sse:after:message`, `ypy:sse:error`, `ypy:sse:close` |
| `<id>:wsOpen`, `wsClose`, `wsError` | `ypy:ws:after:connection`, `ypy:ws:close`, `ypy:ws:error` |
| `<id>:validation:*`, `<id>:xhr:*` | **gone**: validation uses the browser's API, requests use `fetch`; use the native `invalid` event and `ypy:before:request`, `ypy:finally:request`, `ypy:abort` (`quebra`) |
| `ypy-on::after-request` (kebab-case twin) | `ypy-on::after:request` |

The whole list (every event, one rule each) is `ypy check --list-rules`; the generated table is in
[the rule table](upgrade-check.md#reference-every-rule).

### Headers

`ypy migrate` converts the name wherever it appears in server code: `<P>-Request` is `YPY-Request`,
`Vary: <P>-Request` is `Vary: YPY-Request`, `HTTP_<P>_REQUEST` is `HTTP_YPY_REQUEST` (the CGI form that PHP, Python
WSGI and Ruby servers expose).

| Before | After | Notes |
|---|---|---|
| `<P>-Request` | `YPY-Request` | `true` on every YPY request |
| `<P>-Target` | `YPY-Target` | the value is `tag#id` (`div#feed`), not a bare id |
| `<P>-Trigger` (request) | `YPY-Source` | `tag#id` of the element that made the request |
| `<P>-Trigger-Name` (request) | `YPY-Source`, or a field in the body | |
| `<P>-Current-URL`, `<P>-Boosted`, `<P>-History-Restore-Request`, `<P>-Prompt` | `YPY-Current-URL`, `YPY-Boosted`, `YPY-History-Restore-Request`, `YPY-Prompt` | the prompt needs the `ypy-prompt` extension |
| `<P>-Redirect`, `-Location`, `-Push-Url`, `-Replace-Url`, `-Refresh`, `-Reswap`, `-Retarget`, `-Reselect` (response) | the same with `YPY-` | |
| `<P>-Trigger` (response) | `YPY-Trigger` | |
| `<P>-Trigger-After-Swap`, `-After-Settle` (response) | `YPY-Trigger`, and listen to `ypy:after:swap` or `ypy:after:settle` | |
| (new) | `YPY-Request-Type`: `partial` or `full` | a full request means the page content is replaced |
| `Vary: <P>-Request` | `Vary: YPY-Request` | keep it on every route that answers both a page and a fragment |

A **history restore** arrives without `YPY-Request` and with `YPY-History-Restore-Request: true`. Answer a whole page.

### Configuration

Set in `<meta name="<id>-config">` (now `ypy-config`) or on `<id>.config` (now `ypy.config`). A legacy key that is
not in this table is gone; `ypy check` lists the ones it knows.

| Before | After |
|---|---|
| `defaultSwapStyle` | `defaultSwap` |
| `globalViewTransitions` | `transitions` |
| `historyEnabled` | `history` |
| `includeIndicatorStyles` | `includeIndicatorCSS` |
| `timeout` (`0` was "none") | `defaultTimeout` (default 60000 ms) |
| `selfRequestsOnly` | `mode`: `true` is `"same-origin"`, `false` is `"cors"` |
| `disableInheritance` | removed: inheritance is always explicit |
| `requestClass`, `indicatorClass`, `addedClass`, `settlingClass`, `swappingClass` | `classes.request`, `classes.indicator`, `classes.added`, `classes.settling`, `classes.swapping` |
| `responseHandling` | `config.noSwap` (a list of statuses) or `ypy-status:<code>="swap:none"` |
| `withCredentials` | `ypy-config="credentials:include"` |
| `allowEval` | removed: the page's CSP decides, and `config.callbacks` |
| `allowScriptTags`, `allowNestedOobSwaps`, `attributesToSettle`, `triggerSpecsCache`, `refreshOnHistoryMiss`, `historyCacheSize`, `getCacheBusterParam`, `methodsThatUseUrlParams`, `scrollBehavior`, `scrollIntoViewOnBoost`, `defaultSwapDelay`, `disableSelector`, `ignoreTitle`, `reportValidityOfForms`, `wsBinaryType`, `wsReconnectDelay` | removed, each with a replacement in its rule (`ypy check --list-rules`) |

### Extensions

Extension files and attributes are renamed with the same prefix, and their registration is replaced (see the API
table below). Load the YPY file after the core.

| Legacy extension | YPY | Attributes |
|---|---|---|
| server-sent events (`<p>-sse.js`) | `ext/ypy-sse.min.js` | `sse-connect`, `sse-close` become `ypy-sse:connect`, `ypy-sse:close`. `sse-swap` becomes `ypy-sse-swap`, which the extension only reports. These are `quebra`: the swap-by-event-name model is gone; unnamed messages swap automatically and named events become DOM events |
| WebSocket | `ext/ypy-ws.min.js` | `ws-connect`, `ws-send` become `ypy-ws:connect`, `ypy-ws:send` |
| head support | `ext/ypy-head.min.js` | (none) |
| preload | `ext/ypy-preload.min.js` | `preload="mouseover"` becomes `ypy-preload="mouseover"` |
| response targets | the core: `ypy-status:<code>` | `ypy-status:404="target:#x"`, `ypy-status:4xx`, `ypy-status:5xx` |
| the morphing swap | the core: `innerMorph`, `outerMorph` | no extension needed |
| a script that registers an extension (`defineExtension`) | `ypy.registerExtension(name, hooks)`, event-based | rewrite it: [Extensions](../extensions/using-and-writing.md#write-one) |

`ypy.max.min.js` bundles SSE, WebSocket, preload and the other common extensions in one file.

### The JavaScript API

| Before | After |
|---|---|
| `<id>.process(el)`, `<id>.ajax(...)`, `<id>.on(...)`, `window.<id>` | `ypy.process(el)`, `ypy.ajax(...)`, `ypy.on(...)`, `window.ypy` (`ypy migrate` converts it) |
| `<id>.addClass`, `removeClass`, `toggleClass`, `takeClass`, `closest`, `remove`, `values`, `off`, `location`, `logger`, `logAll`, `logNone` | removed: `classList`, `closest()`, `remove()`, `new FormData(form)`, `removeEventListener()`, `ypy.ajax()`, `console`, `ypy.config.logAll` (the rule names the replacement) |
| `<id>.config.<key>` | `ypy.config.<key>`; renamed keys above |

### Classes

`<id>-request`, `<id>-indicator`, `<id>-swapping`, `<id>-settling`, `<id>-added` become `ypy-request`, `ypy-indicator`,
`ypy-swapping`, `ypy-settling`, `ypy-added` in markup, selectors and CSS. Change the names in
`ypy.config.classes` if your CSS uses others ([config](../reference/config/classes.md)). In Tailwind, see
[Styling the states](../guides/core-concepts.md#styling-the-states).

### Script tag and meta tag

| Before | After |
|---|---|
| `<script src=".../<id>.min.js">` | `<script src=".../ypy.min.js">` (a local file; an external URL is listed, not edited: use `--script-src`) |
| `<meta name="<id>-config" content='...'>` | `<meta name="ypy-config" content='...'>` |

## What `ypy migrate` leaves for you

- Comments and prose. `--comments` converts the recognized names in comments too, and the final report lists the
  lines that still carry an old name in words.
- External script tags, and any copy of an old script or one of its extensions in your repository. The tool lists
  and skips them; replace them with the YPY files. It also skips `node_modules`, `vendor`, `dist`, `build`, `.git`,
  minified and binary files and files over 2 MB.
- Names built at run time, such as `"<p>-" + name` or a header read with a name assembled from parts. The tool
  recognizes literal strings only, so search for these by hand.
- Third-party code that renders the old attributes (a UI kit, a server-side component library). Migrate or update it
  first, or YPY will not understand its output.

## Server-side notes by stack

After the conversion the server has four things to do: read `YPY-Request`, send `Vary: YPY-Request`, answer `422` for
an invalid form and `303` after a successful POST, and carry the CSRF token. Each recipe has working code:

| Stack | Recipe | The legacy header in server code looks like |
|---|---|---|
| Go (`net/http`) | [Go](../server/recipes/go.md) | `r.Header.Get("<P>-Request")` |
| Arandu (Hesape, Kyse) | [Arandu](../server/recipes/arandu.md) | the framework's helpers, converted with the library |
| Django, Flask, Jinja | [Django](../server/recipes/django.md) | `request.headers.get("<P>-Request")`, `request.META["HTTP_<P>_REQUEST"]` |
| Rails | [Rails](../server/recipes/rails.md) | `request.headers["<P>-Request"]` |
| Laravel, PHP | [Laravel](../server/recipes/laravel.md) | `$request->header('<P>-Request')`, `$_SERVER['HTTP_<P>_REQUEST']` |

A server that has to serve old and YPY pages during a transition can read both headers for a while. The old name then
appears in one place in your code, and you delete it with the last old page.

## Verify

```sh
ypy check .                      # the worklist: nothing "quebra" left
ypy migrate .                    # "0 of N file(s) would change": the conversion is complete and stable
ypy ci names                     # in the YPY repository: no legacy name anywhere (in your project, grep for the placeholders you learned)
```

Then run your own test suite in a browser, against the file you now serve.

## Roll back

The conversion is a normal commit. `git revert` it, put the previous script tag back, and you are where you started.
YPY does not touch your data or your server.

## Frequently asked

**Can I keep the old behavior for a while?** No. YPY has one behavior, so the `ypy check` review is part of the move.
You can do it page by page: convert, fix that page's worklist, ship.

**My `<p>-on` handlers stopped working.** Author code is off by default (`callbacks: "none"`). To turn it on:
`<meta name="ypy-config" content='callbacks:"all"'>`, knowing what that means
([Security checklist](../guides/security-checklist.md#author-code)).

**Can the old and the new spelling live together?** No. YPY does not read the old spelling at all, so every name has to
be converted. (The only precedence rule is between the two YPY spellings: on one element, `ypy-*` wins over
`data-ypy-*`.)

## See also

[`ypy migrate`](migrate.md) · [`ypy check`](upgrade-check.md) ·
[Events](../reference/events/index.md) · [Security checklist](../guides/security-checklist.md) · [Recipes](../server/recipes/go.md)
