---
order: 70
title: Migration from the older library
description: Move a project written for the older library to YPY: the path, what is different, the ypy migrate conversion and the ypy check worklist, rule by rule.
---

# Migration from the older library

YPY started from an older hypermedia library and now has its own names and its own behavior. Moving a project has two
parts. The first is mechanical: one command renames the attributes, headers, events and classes, and you swap the old
script file for `ypy.min.js`. The second is a review. YPY behaves differently in a few places, so a renamed project does
not behave the way it did, and a worklist tells you where.

There is no compatibility layer, no profile and no switch that brings the old behavior back. A converted project runs
on `ypy.min.js` from its first load. Your server needs no change beyond the names: YPY sends and reads the same headers
with the `YPY-` prefix.

The guide writes the old names as placeholders, because the tools learn the real ones from your project:

| Placeholder | Is | Made-up example |
|---|---|---|
| `<p>-` | the old attribute prefix: `<p>-get`, `data-<p>-target`, `<p>-on::after-swap` | `acme-get` |
| `<P>-` | the old header prefix: `<P>-Request`, `<P>-Target` | `ACME-Request` |
| `<id>` | the old identifier: its global object, event prefix and state classes | `acme.process(el)`, `acme:afterSwap`, `acme-request` |

If your project does not reveal them (a back end repository that only reads one header, say), pass them yourself:
`ypy migrate --prefix acme --header-prefix ACME --id acme .`

## Before you start

| Step | You do | Result |
|---|---|---|
| 1. Read the worklist | `ypy check .` | what breaks, what behaves differently and what is only spelled the old way |
| 2. Convert | `ypy migrate --write`, then load `ypy.min.js` in place of the old script | the names are YPY's and the project loads the YPY core |
| 3. Review the behavior | work through the `ypy check` findings: explicit inheritance, error responses, author code, events; page by page, at the pace the project allows | the project behaves the way YPY does |

In steps:

1. **Start from a clean tree.** Commit what you have. The conversion is an ordinary diff you can review and revert.
2. **Read the worklist.** `ypy check .` changes nothing; it prints each finding with its line, the fix and its severity.
3. **Dry run.** `ypy migrate .` prints the prefixes it learned and the diff, and writes nothing. Check that the prefixes
   and the identifier are the right ones.
4. **Write.** `ypy migrate --write .`, then `ypy migrate .` again: it must report `0 of N file(s) would change`.
5. **Swap the script.** Put `dist/ypy.min.js` where the old script was, delete the old file and its extension scripts,
   and load the official extensions you used from `dist/ext/`, or use `dist/ypy.max.min.js`, which carries them
   ([Installing](index.md#installing)).
6. **Fix the behavior.** Go through the `ypy check` findings until the list is quiet.
7. **Run your own gates**: formatting, build, vet, unit tests, browser tests. A test that asserts the old spelling of an
   attribute or header was converted with the rest.
8. **Look at the result** in a browser with the console open: one page of each kind (a list, a form, a search, a
   stream).

If several projects depend on each other (a framework, a UI kit, products built on them), convert in dependency order.
The libraries that write and sanitize attributes go first and the products after them, or a product would write
attributes its dependencies still refuse. Bump the dependency versions together.

## Native differences

This section is for someone who knows the older library: what the core does on its own, what became an extension,
which defaults changed and what is gone.

### One set of names

| What | In YPY |
|---|---|
| Attributes | `ypy-*`, or `data-ypy-*` (when an element has both, `ypy-*` wins) |
| Tag | `<ypy-partial>` |
| HTTP headers | `YPY-*`, in requests and responses |
| Events | `ypy:<noun>:<verb>`, such as `ypy:before:request` and `ypy:after:swap` |
| Global object | `ypy` |
| State classes | `ypy-request`, `ypy-swapping`, `ypy-settling`, `ypy-added`, `ypy-indicator` |
| Extensions | `ypy-<name>`, in `config.extensions` and in `ypy.registerExtension` |

The old spelling is not read at all, so old and new names cannot live together.

### What the core does natively

Several things that were separate pieces, or did not exist, now work with the core alone: the morphing swap
(`innerMorph` and `outerMorph`, with `innerMorph` as the default), a different target by status (`ypy-status:<code>`,
`ypy-error-target`, `ypy-error-swap`, `ypy-error-focus`), a `422` swapped into the form with the focus on the first
invalid field, several targets from one response (`<ypy-partial>`, `ypy-select-oob`, `ypy-swap-oob`), a request queue
(`ypy-sync`), accessibility (`ypy-busy`, `ypy-focus`, `ypy-main`, `ypy-announce`), the same-origin guard (`config.mode`),
atomic updates (`YPY-Group`) and one `ctx` object in the `detail` of every event. Requests use `fetch` with an
`AbortSignal`, so no `XMLHttpRequest` object reaches the events. The [Docs](index.md) cover each one.

### What is an extension

Some features that lived in the old core, or in one of its extensions, are now official extensions. Load them after the
core, or use `ypy.max.min.js`, which carries most of them.

| Feature | Extension |
|---|---|
| Server-sent events | [`ypy-sse`](extensions/sse.md): `sse-connect` and `sse-close` become `ypy-sse:connect` and `ypy-sse:close`; `sse-swap` is gone (unnamed messages swap, named events become DOM events) |
| WebSocket | [`ypy-ws`](extensions/ws.md): `ws-connect` and `ws-send` become `ypy-ws:connect` and `ypy-ws:send` |
| Merging the `<head>` | [`ypy-head`](extensions/head.md) |
| Preloading a link | [`ypy-preload`](extensions/preload.md): `preload="mouseover"` becomes `ypy-preload="mouseover"` |
| The prompt sent as a header | [`ypy-prompt`](extensions/prompt.md) |
| Disabling elements during a request, and `ypy-inert` | [`ypy-disable`](extensions/disable.md); it moved out of the core, so load it with `ypy.min.js` |
| Snapshots for the back button | [`ypy-history-cache`](extensions/history-cache.md) |

Extensions written against the old registration function must be rewritten for `ypy.registerExtension(name, hooks)`,
which is event based ([Extensions](index.md#extensions)).

YPY also has extensions with no counterpart in the old library, so the converter has nothing to rename: retry,
dedupe, polling, caching, optimistic updates, validation, uploads, navigation, streaming, signals, layers, sorting and
the security extensions. They are all in the [extensions index](extensions/index.md).

### What behaves differently by default

Each row is a rule that `ypy check` reports on your files.

| Behavior | YPY | What to do |
|---|---|---|
| Default swap | `innerMorph`: focus, selection and scroll of live elements win over the incoming HTML | write `ypy-swap="innerHTML"` where you want a raw replacement |
| Attribute inheritance | explicit: an attribute reaches descendants only with `:inherited`, as in `ypy-target:inherited="#out"` | add `:inherited` where a container's attribute was meant for its children; `ypy migrate` cannot know which ones ([Attribute inheritance](index.md#attribute-inheritance)) |
| Author code | off: `config.callbacks` is `"none"`, so `ypy-on`, `js:` values and `[filter]` triggers do not run | enable `"all"` or `"nonce"` on purpose ([Security](index.md#security)) |
| Cross-origin requests | blocked: `config.mode` is `"same-origin"` | change `mode` only for a page that needs it |
| Timeout | 60 000 ms by default, and `0` is not "none" | set `defaultTimeout` |
| Statuses that skip the swap | `204` and `304` (`config.noSwap`); `422` swaps into the form; other errors follow the error rules | use `ypy-status:<code>` for a different choice |
| History | a back or forward request asks the server again, with `YPY-History-Restore-Request`, and the server answers a full page; snapshots are not kept in `localStorage` | add `ypy-history-cache` if you want snapshots ([History](index.md#history)) |
| Request headers | `YPY-Target` is `tag#id`, not a bare id, and `YPY-Source` names the element that made the request | update server code that parsed the old value |
| `ypy-on` syntax | the colon form `ypy-on:click` and `ypy-on::after:swap`, one attribute per handler | rewrite the handlers that used one attribute for many events |
| Validation | the browser's own API | listen to the native `invalid` event |
| `aria-busy` | set on the target while a request runs | opt out with `ypy-busy="false"` |

### What was removed

The attributes that set request values from script, the per-element request settings, the list of extensions on an
element, the switches for inheritance, the `queue:` trigger modifier, the old extension registration function, the
helper functions of the JavaScript API (class helpers, `closest`, `remove`, `values`, `off`, logging), most of the old
configuration keys, and the events of the request object and of the validation layer. Each one has a named replacement
in a `ypy check` rule below.

### Attributes whose meaning changed

Every attribute keeps its name and loses the prefix: `<p>-get` is `ypy-get`, `data-<p>-target` is `data-ypy-target`,
`<p>-on::after-swap` is `ypy-on::after-swap`, and the tag `<<p>-partial>` is `<ypy-partial>`. These are the ones whose
meaning or availability changed:

| Before | After |
|---|---|
| `<p>-vars` | `ypy-vals` with the `js:` prefix: `ypy-vals="js:{a: 1}"` |
| `<p>-params` | removed: filter in a `ypy:config:request` listener (`detail.ctx.request.body` is a `FormData`) |
| `<p>-request` | `ypy-config` (same JSON: `timeout`, `credentials`, `noHeaders`) |
| `<p>-disabled-elt` | `ypy-disable` |
| `<p>-disable` with no value (skip processing) | `ypy-ignore`; do this before renaming the one above |
| `<p>-ext` | deleted: an extension is active when its script is loaded |
| `<p>-inherit`, `<p>-disinherit` | deleted; write `:inherited` on the attributes that must reach descendants |
| `<p>-history="false"` | deleted (snapshots are no longer kept in `localStorage`) |
| `queue:` trigger modifier | `ypy-sync="this:queue last"` (or `queue first`, `queue all`, `drop`, `abort`) |
| `show:<selector>:<position>`, `scroll:<selector>:<position>` | `show:top showTarget:#sel`, `scroll:bottom scrollTarget:#sel` |
| `<p>-on="event: code"` (one handler per line) | one `ypy-on:<event>` attribute per handler, or `ypy-on="click -> code"` |
| the `inherit` keyword of `-include` and `-indicator` | gone; use `:inherited` |

### Headers

`ypy migrate` converts the name wherever it appears in server code: `<P>-Request` is `YPY-Request`,
`Vary: <P>-Request` is `Vary: YPY-Request`, and the CGI form `HTTP_<P>_REQUEST` is `HTTP_YPY_REQUEST`.

| Before | After | Notes |
|---|---|---|
| `<P>-Request` | `YPY-Request` | `true` on every YPY request |
| `<P>-Target` | `YPY-Target` | the value is `tag#id` (`div#feed`), not a bare id |
| `<P>-Trigger` (request) | `YPY-Source` | `tag#id` of the element that made the request |
| `<P>-Trigger-Name` (request) | `YPY-Source`, or a field in the body | |
| `<P>-Current-URL`, `<P>-Boosted`, `<P>-History-Restore-Request`, `<P>-Prompt` | the same with `YPY-` | the prompt needs the `ypy-prompt` extension |
| `<P>-Redirect`, `-Location`, `-Push-Url`, `-Replace-Url`, `-Refresh`, `-Reswap`, `-Retarget`, `-Reselect`, `-Trigger` (response) | the same with `YPY-` | |
| `<P>-Trigger-After-Swap`, `-After-Settle` (response) | `YPY-Trigger`, and listen to `ypy:after:swap` or `ypy:after:settle` | |
| (new) | `YPY-Request-Type`: `partial` or `full` | a full request means the page content is replaced |
| `Vary: <P>-Request` | `Vary: YPY-Request` | keep it on every route that answers both a page and a fragment |

A history restore arrives without `YPY-Request` and with `YPY-History-Restore-Request: true`. Answer a whole page.

### Classes, script tag and meta tag

`<id>-request`, `<id>-indicator`, `<id>-swapping`, `<id>-settling` and `<id>-added` become `ypy-request`,
`ypy-indicator`, `ypy-swapping`, `ypy-settling` and `ypy-added` in markup, selectors and CSS. If your CSS uses other
names, set them in `config.classes`.

| Before | After |
|---|---|
| `<script src=".../<id>.min.js">` | `<script src=".../ypy.min.js">` (a local file; an external URL is listed, not edited: use `--script-src`) |
| `<meta name="<id>-config" content='...'>` | `<meta name="ypy-config" content='...'>` |

Events, configuration keys and JavaScript functions have one row each in the `ypy check` tables below. The rename of
the prefix and the new `ypy:<noun>:<verb>` names are in [Events & logging](index.md#events-and-logging) and the
[event table of the reference](reference.md#events).

## `ypy migrate`

`ypy migrate` rewrites a project written for the older library so that it uses YPY names. The result loads
`ypy.min.js`, with no compatibility layer behind it. The tool reads files as text and edits only where it recognizes
the construct. It is a subcommand of the `ypy` binary, with no Node and no other dependency.

```sh
ypy migrate .            # report and unified diff, nothing is written
ypy migrate --write .    # apply; run it twice and the second run changes nothing
```

### What it converts

The tool contains none of the old prefixes. It learns them from your project (see
[How the prefixes are recognized](#how-the-prefixes-are-recognized)), or you give them with `--prefix`,
`--header-prefix` and `--id`.

| Before | After | Where |
|---|---|---|
| `<p>-get`, `data-<p>-target`, `<p>-on::after-swap` | `ypy-get`, `data-ypy-target`, `ypy-on::after-swap` | attributes in templates (HTML, Go `html/template`, templ, `.kyse.go`, Jinja, ERB, Blade, Twig, PHP, JSX, Vue), strings and selectors in JavaScript, TypeScript and CSS |
| `<<p>-partial>` | `<ypy-partial>` | markup |
| `<P>-Request`, `Vary: <P>-Request`, `HTTP_<P>_REQUEST` | `YPY-Request`, `Vary: YPY-Request`, `HTTP_YPY_REQUEST` | strings in server code and templates |
| `<id>:afterSwap`, `<id>:afterSettle`, `<id>:beforeRequest`, `<id>:afterRequest`, `<id>:load` | `ypy:after:swap`, `ypy:after:settle`, `ypy:before:request`, `ypy:finally:request`, `ypy:after:process` | strings and attribute values, `ypy-on::` and `ypy-on:` attributes. An event with no YPY equivalent (`<id>:xhr:loadstart`) and a read of `detail.elt` (use `evt.target`) are left as they are and listed |
| `<id>.process(el)`, `window.<id>` | `ypy.process(el)`, `window.ypy` | JavaScript, inline scripts, inline handlers |
| `<id>-request`, `<id>-indicator`, `<id>-swapping`, ... | `ypy-request`, `ypy-indicator`, `ypy-swapping`, ... | class attributes, selectors, CSS |
| `<p>-sse.js`, `sse-connect`, `sse-swap`, `ws-connect`, `ws-send`, `preload="mouseover"` | `ypy-sse.js`, `ypy-sse:connect`, `ypy-sse-swap`, `ypy-ws:connect`, `ypy-ws:send`, `ypy-preload` | extension files and attributes; `ypy-sse-swap` is only reported by the extension, see `V2-EXT-SSE-SWAP` below |
| `<meta name="<id>-config">` | `<meta name="ypy-config">` | markup |
| `<script src=".../<id>.min.js">` | `<script src=".../ypy.min.js">` | markup, and the file name in strings such as `view.Asset("<id>.min.js")` |
| `<P>Target`, `<p>On`, `<p>_request` (the prefix as a word of an identifier) | `YpyTarget`, `ypyOn`, `ypy_request` | all code |
| `Is<P>`, `Has<P>`, `With<P>` | `IsYPY`, `HasYPY`, `WithYPY` | Go files only |
| `extensions:"ypy-lock,sse,ws"`, `config.extensions = 'preload'` (the short names of the official extensions) | `extensions:"ypy-lock,ypy-sse,ypy-ws"`, `config.extensions = 'ypy-preload'` | the `extensions` setting in markup, scripts and server code; the names `sse`, `ws`, `preload`, `ptag`, `upsert`, `download`, `history-cache`, `browser-indicator`, `alpine-compat`, and no other |

### What it leaves for you

- Comments and prose. `--comments` converts the recognized names in comments too; identifiers such as `YpyTarget` are
  renamed everywhere, comments included.
- External script tags, and any copy of an old script or one of its extensions in your repository. The tool lists and
  skips them; replace them with the YPY files. `--script-src URL` rewrites an external tag. A file that imports the old
  script as a module keeps its `<id>` global; the import binding is yours to change.
- Files in `node_modules`, `vendor`, `dist`, `build` and `.git`, minified and binary files, and files over 2 MB.
- Names built at run time, such as `"<p>-" + name`. The tool recognizes literal strings only, so search for these by
  hand.
- Third-party code that renders the old attributes (a UI kit, a server-side component library). Migrate or update it
  first, or YPY will not understand its output.
- Identifiers are renamed across the whole project, so another module that uses your exported field `<P>Target` has to
  be migrated too. `--no-identifiers` skips them.

It ends with a list of what it left alone: the places that still carry an old name and the number of lines that still
name it in words.

### Options

| Option | Meaning |
|---|---|
| `--write` | Apply the changes. Without it nothing is written. |
| `--prefix P` | The attribute prefix, when the project does not reveal it (`acme`, as in `acme-get`); also the header prefix. |
| `--header-prefix P` | The header prefix, when it differs from the attribute prefix. |
| `--id NAME` | The old identifier (global, events, classes), when the code does not reveal it. Repeatable. |
| `--min-evidence N` | Distinct names that make a prefix learned (default 2). |
| `--comments` | Also convert the recognized names inside comments. |
| `--no-identifiers` | Leave identifiers (`<P>Target`, `<p>_request`) alone. |
| `--script-src URL` | Replace the URL of an external script tag of an old script. |
| `--ext .a,.b` | Convert more extensions as templates. |
| `--json` | One JSON document on stdout (schema 1). |
| `--no-diff` | Text output without the diff. |

Exit status `0` when it ran, `2` on a usage error or an unreadable path. The JSON document has the fields `schema`,
`tool` (`ypy-migrate`), `rulesVersion`, `mode` (`dry-run` or `write`), `legacy` (the learned prefixes and identifiers),
`scanned`, `summary`, `files`, `unconverted`, `skipped` and `warnings`. The kinds are `attribute`, `header`, `event`,
`class`, `global`, `identifier`, `script`, `extension`, `partial` and `config`.

## `ypy check`

`ypy check` scans a project written for the older library and finds every use of a removed or changed behavior. For each
one it says whether it will break in YPY, behave differently, or only need renaming. It reads files as text and never
runs your code, and it needs no Python, no Node and no browser.

```sh
ypy check ./templates          # a directory, several paths or single files
ypy check .                    # the whole project (node_modules, vendor, dist and .git are skipped)
ypy check ./app --json > check.json
ypy check . --id acme          # name the old identifier when the code does not reveal it
```

Typical output:

```text
templates/items.html:1:6  muda   V2-INHERIT-IMPLICIT  Attribute inherited implicitly by descendants; inheritance is explicit now.
    fix:   Add :inherited to the attribute on the container, for example ypy-target:inherited="#out".
templates/items.html:2:29  muda   V2-ATTR-VARS  acme-vars is removed.
    fix:   Use ypy-vals with the js: prefix: ypy-vals="js:{a: 1}".

Legacy markup learned: attribute prefix acme, from acme-target (templates/items.html:1) and acme-vars (templates/items.html:2).
1 file(s) scanned, 1 with findings: 0 quebra, 2 muda, 2 aviso.
```

Each finding gives `file:line:column`, the severity, the rule id, what changes and the fix.

### Severities

The labels are the tool's own: `quebra` (breaks), `muda` (changes) and `aviso` (warning).

| Severity | Meaning | Exit status |
|---|---|---|
| `quebra` | Stops working after the move to YPY. Fix before shipping. | 1 |
| `muda` | Keeps working, but behaves differently from what the project was written for. Review it. | 0 (1 with `--fail-on muda`) |
| `aviso` | Works. Only the spelling is old and a YPY spelling exists. `ypy migrate` converts it. | 0 (1 with `--fail-on aviso`) |

Exit status `2` is a usage error: an unknown option, an unreadable path or an unknown rule id.

### Options

| Option | Meaning |
|---|---|
| `--json` | One JSON document on stdout (schema 1). |
| `--min-severity aviso\|muda\|quebra` | Hide the findings below that level. Default `aviso`. |
| `--fail-on quebra\|muda\|aviso\|never` | Lowest severity that makes the exit status 1. Default `quebra`. |
| `--prefix P`, `--header-prefix P` | The attribute and header prefixes, when the code does not reveal them. |
| `--min-evidence N` | Distinct names that make a prefix learned (default 2). |
| `--id NAME[,NAME]` | The old identifier (its global, event prefix and class prefix), when the code does not reveal it. |
| `--min-confidence medium\|high` | `high` hides the findings whose identifier rests on weak signs. Default `medium`. |
| `--ignore ID[,ID]` | Skip rules (repeatable). Use it for a rule you decided to live with. |
| `--only ID[,ID]` | Report only these rules. |
| `--ext .a,.b` | Scan more extensions as templates. `.md`, `.markdown` and `.json` are read by default. |
| `--list-rules` | Print the rule table (honors `--json`). |
| `--no-color` | Plain output. Colors are also off when stdout is not a terminal or `NO_COLOR` is set. |

For continuous integration, `ypy check ./app` fails only on what breaks, and `ypy check ./app --min-severity muda
--fail-on muda` fails on anything that needs a human decision. `ypy check ./app --json` gives one object per finding to
turn into annotations.

### How legacy markup is recognized

The rules never contain the name of the library the markup came from. They know its shape: the names of its events
(`afterSwap`), configuration keys (`historyCacheSize`), API methods (`defineExtension`) and state classes (`request`,
`indicator`). The identifier and the attribute and header prefixes in front of them are learned from your project in a
first pass over all the files.

- Strong evidence names an identifier alone: an event only the old API spells, a call of a method only it has, a
  configuration key under `<id>.config` or in the `<id>-config` meta tag, two different state classes together, or the
  name given with `--id`.
- Weak evidence is ambiguous alone: a script called `<id>.min.js`, a single state class, an `<id>:load` event. Two
  different kinds count, and the findings that rest on them get medium confidence.
- A prefix is learned from at least two distinct names of the request vocabulary (`get`, `post`, `target`, `swap`,
  `trigger`, `vals`, `include`, `push-url`, and so on), so a `x-on` of another framework stays quiet. Headers are
  learned from the word in front of `Request`, `Trigger`, `Target`, `Redirect` or `Location`, or from the CGI spelling.
- `--prefix P` and `--header-prefix P` (written without the hyphen) force a prefix the project does not reveal.

Code that only looks like it (`Object.values(x)`, `el.remove()`, `socket.on('chat:load')`) stays quiet until something
in the project reveals the identifier. A project that uses the old script without calling its API is not reported until
you pass `--id`.

### What is scanned, and what it cannot see

Templates are read as text: HTML, Go `html/template` and templ, `.kyse.go`, Jinja, ERB, PHP and Blade, Twig and the
common JavaScript template formats, plus scripts and styles. The template syntax is blanked out keeping every line and
column, then the markup is tokenized with a lenient parser. Server code is read only for the header rules.

It cannot see server responses (a 4xx or 5xx swap change shows only where the code names it), values built at run time,
inheritance across files (a base template with a `target` attribute on `<body>` is flagged because the requests live in
other files), default values that changed when the project never sets them, and what an extension you wrote yourself
does. A hit can be a false positive when the same text means something else, such as a `preload` attribute on a custom
element; use `--ignore` for those.

### Every rule

The tables list each rule the tool emits, by id, as `ypy check --list-rules` prints them. `<id>` stands for the old
identifier, `<p>-` for its attribute prefix and `<P>-` for its header prefix. Use the id with `--ignore` and `--only`.

#### Attributes and extensions

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V1-ATTR-SSE` | quebra | `<p>-sse` | `ypy-sse:connect="/url"`; swap targets and events follow the native model |
| `V1-ATTR-WS` | quebra | `<p>-ws` | `ypy-ws:connect="/url"` and `ypy-ws:send` |
| `V1-ATTR-ON-LEGACY` | quebra | `<p>-on="event: code"` | one `ypy-on:<event>` attribute per handler, or `ypy-on="click -> code"` |
| `V2-ATTR-VARS` | muda | `<p>-vars` | `ypy-vals="js:{a: 1}"` |
| `V2-ATTR-PARAMS` | muda | `<p>-params` | filter in a `ypy:config:request` listener (`detail.ctx.request.body` is a `FormData`) |
| `V2-ATTR-PROMPT` | aviso | `<p>-prompt` | load the `ypy-prompt` extension; the attribute and the `YPY-Prompt` header stay |
| `V2-ATTR-EXT` | muda | `<p>-ext` | delete it; an extension is active when its script is loaded; restrict with `config.extensions` |
| `V2-ATTR-DISINHERIT` | muda | `<p>-disinherit` | delete it; add `:inherited` where a container's attribute must reach descendants |
| `V2-ATTR-INHERIT` | muda | `<p>-inherit` | delete it; add `:inherited` where a container's attribute must reach descendants |
| `V2-ATTR-REQUEST` | muda | `<p>-request` | `ypy-config`, same JSON (`timeout`, `credentials`, `noHeaders`) |
| `V2-ATTR-HISTORY` | aviso | `<p>-history` | delete it |
| `V2-ATTR-DISABLE` | muda | `<p>-disable` with no value | `ypy-ignore`; do this before renaming the next row |
| `V2-ATTR-DISABLED-ELT` | aviso | `<p>-disabled-elt` | `ypy-disable` (the `ypy-disable` extension) |
| `V2-ATTR-ON-DASH` | muda | `<p>-on-click`, `<p>-on--after-swap` | colon spelling: `ypy-on:click`, `ypy-on::after:swap` |
| `V2-EXT-SSE-CONNECT` | aviso | `sse-connect` | `ypy-sse:connect` |
| `V2-EXT-SSE-CLOSE` | aviso | `sse-close` | `ypy-sse:close` |
| `V2-EXT-SSE-SWAP` | quebra | `sse-swap` | drop it; unnamed messages swap, named ones become DOM events (listen with `ypy-on`) |
| `V2-EXT-WS-CONNECT` | aviso | `ws-connect` | `ypy-ws:connect` |
| `V2-EXT-WS-SEND` | aviso | `ws-send` | `ypy-ws:send` |
| `V2-EXT-PRELOAD` | aviso | `preload` | `ypy-preload`; values such as `mouseover` stay |
| `V2-EXT-RESPONSE-TARGETS` | aviso | `<p>-target-404`, `-4*`, `-*`, `-error` | `ypy-status:404="target:#x"`; `ypy-status:4xx`, `ypy-status:5xx` for ranges |
| `V2-SWAP-MORPH` | aviso | `<p>-swap="morph"` | `innerMorph`, `outerMorph`; no extension needed |

#### Syntax and inheritance

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V1-DELETE-BODY` | muda | `<p>-delete` | check the handler: read DELETE data from the query string, or add `ypy-include` |
| `V1-CROSS-ORIGIN` | muda | `<p>-get="https://other.example/..."` | another origin needs `config.mode = "cors"` |
| `V2-ATTR-OOB` | aviso | `<p>-swap-oob` | review code that depends on order; `config.allowEmptySwapAfterOOB` or `swapEmpty:true` restores the empty swap |
| `V2-SWAP-SHOW-SYNTAX` | muda | `show:<selector>:top`, `scroll:<selector>:bottom` | `show:top showTarget:#sel`, `scroll:bottom scrollTarget:#sel` |
| `V2-TRIGGER-QUEUE` | muda | `queue:first\|last\|all\|none` in a trigger | `ypy-sync="this:queue last"` (or `queue first`, `queue all`, `drop`, `abort`) |
| `V2-INCLUDE-INHERIT` | muda | the `inherit` keyword in `<p>-include`, `<p>-indicator` | put the value on the parent: `ypy-include:inherited:append="#shared"` |
| `V2-INHERIT-IMPLICIT` | muda | `<p>-target`, `<p>-swap`, ... on a container whose descendants make requests | add `:inherited` (`ypy-target:inherited="#out"`) only where a descendant relies on it |
| `V2-DELETE-IN-FORM` | muda | a `<p>-delete` element inside a form | add `ypy-include="closest form"` |

#### Names, prefixes and the global

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `ANY-ATTR-PREFIX` | aviso | any attribute with the `<p>-` or `data-<p>-` prefix | `ypy-*`, `data-ypy-*`; `ypy migrate --write` converts it |
| `ANY-META-CONFIG` | aviso | `<meta name="<id>-config">` | `<meta name="ypy-config">`; when both exist, `ypy-config` wins |
| `ANY-CLASS-PREFIX` | aviso | `<id>-request`, `<id>-indicator`, `<id>-swapping`, `<id>-settling`, `<id>-added` | the `ypy-` classes or `data-ypy-state`; drop the legacy ones |
| `ANY-API-GLOBAL` | aviso | `<id>.<method>`, `window.<id>` | `ypy.<method>`, `window.ypy` |
| `ANY-EVENT-PREFIX` | aviso | `<id>:error`, `<id>:abort`, `<id>:confirm` | `ypy:error`, `ypy:abort`, `ypy:confirm` |
| `V4-EVENT-PREFIX` | aviso | `<id>:before:request`, `<id>:after:swap`, ... | same event name with the `ypy:` prefix |
| `ANY-IDENT-PREFIX` | aviso | `<P>Target`, `<p>_request`, `Is<P>` | `YpyTarget`, `ypy_request`, `IsYPY`; `--no-identifiers` keeps them |

#### Events

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V2-EVENT-AFTER-ON-LOAD` | aviso | `<id>:afterOnLoad` | `ypy:after:request` |
| `V2-EVENT-AFTER-PROCESS-NODE` | aviso | `<id>:afterProcessNode` | `ypy:after:init` |
| `V2-EVENT-AFTER-REQUEST` | aviso | `<id>:afterRequest` | `ypy:finally:request` (fires on failure too); `ypy:after:request` fires before the swap |
| `V2-EVENT-AFTER-SETTLE` | aviso | `<id>:afterSettle` | `ypy:after:settle` |
| `V2-EVENT-AFTER-SWAP` | aviso | `<id>:afterSwap` | `ypy:after:swap` |
| `V2-EVENT-BEFORE-CLEANUP-ELEMENT` | aviso | `<id>:beforeCleanupElement` | `ypy:before:cleanup` |
| `V2-EVENT-BEFORE-HISTORY-SAVE` | aviso | `<id>:beforeHistorySave` | `ypy:before:history:update` |
| `V2-EVENT-BEFORE-HISTORY-UPDATE` | aviso | `<id>:beforeHistoryUpdate` | `ypy:before:history:update` |
| `V2-EVENT-BEFORE-ON-LOAD` | aviso | `<id>:beforeOnLoad` | `ypy:before:response` |
| `V2-EVENT-BEFORE-PROCESS-NODE` | aviso | `<id>:beforeProcessNode` | `ypy:before:init` |
| `V2-EVENT-BEFORE-REQUEST` | aviso | `<id>:beforeRequest` | `ypy:before:request` |
| `V2-EVENT-BEFORE-SEND` | aviso | `<id>:beforeSend` | `ypy:before:request` |
| `V2-EVENT-BEFORE-SWAP` | aviso | `<id>:beforeSwap` | `ypy:before:swap`; the swap decision moved to `config.noSwap` and `ypy-status` |
| `V2-EVENT-BEFORE-TRANSITION` | aviso | `<id>:beforeTransition` | `ypy:before:viewTransition` |
| `V2-EVENT-CONFIG-REQUEST` | aviso | `<id>:configRequest` | `ypy:config:request`; `detail.ctx.request` replaces `detail.headers`, `parameters` and `path` |
| `V2-EVENT-HISTORY-CACHE-MISS` | aviso | `<id>:historyCacheMiss` | `ypy:before:history:restore` |
| `V2-EVENT-HISTORY-RESTORE` | aviso | `<id>:historyRestore` | `ypy:before:history:restore` |
| `V2-EVENT-LOAD` | aviso | `<id>:load` | `ypy:after:process` |
| `V2-EVENT-OOB-AFTER-SWAP` | aviso | `<id>:oobAfterSwap` | `ypy:after:swap` |
| `V2-EVENT-OOB-BEFORE-SWAP` | aviso | `<id>:oobBeforeSwap` | `ypy:before:swap` |
| `V2-EVENT-PUSHED-INTO-HISTORY` | aviso | `<id>:pushedIntoHistory` | `ypy:after:history:push` |
| `V2-EVENT-REPLACED-IN-HISTORY` | aviso | `<id>:replacedInHistory` | `ypy:after:history:replace` |
| `V2-EVENT-RESPONSE-ERROR` | aviso | `<id>:responseError` | `ypy:response:error` |
| `V2-EVENT-SEND-ABORT` | aviso | `<id>:sendAbort` | `ypy:error` |
| `V2-EVENT-SEND-ERROR` | aviso | `<id>:sendError` | `ypy:error` |
| `V2-EVENT-SWAP-ERROR` | aviso | `<id>:swapError` | `ypy:error` |
| `V2-EVENT-TARGET-ERROR` | aviso | `<id>:targetError` | `ypy:error` |
| `V2-EVENT-TIMEOUT` | aviso | `<id>:timeout` | `ypy:error` |
| `V2-EVENT-PROMPT` | aviso | `<id>:prompt` | `ypy:prompt` (the `ypy-prompt` extension) |
| `V2-EVENT-SSE-OPEN` | aviso | `<id>:sseOpen` | `ypy:sse:after:connection` |
| `V2-EVENT-SSE-ERROR` | aviso | `<id>:sseError` | `ypy:sse:error` |
| `V2-EVENT-SSE-BEFORE-MESSAGE` | aviso | `<id>:sseBeforeMessage` | `ypy:sse:before:message` |
| `V2-EVENT-SSE-MESSAGE` | aviso | `<id>:sseMessage` | `ypy:sse:after:message` |
| `V2-EVENT-SSE-CLOSE` | aviso | `<id>:sseClose` | `ypy:sse:close` |
| `V2-EVENT-WS-OPEN` | aviso | `<id>:wsOpen` | `ypy:ws:after:connection` |
| `V2-EVENT-WS-CLOSE` | aviso | `<id>:wsClose` | `ypy:ws:close` |
| `V2-EVENT-WS-ERROR` | aviso | `<id>:wsError` | `ypy:ws:error` |
| `V2-EVENT-WS-CONFIG-SEND` | aviso | `<id>:wsConfigSend` | `ypy:ws:before:message:outgoing` |
| `V2-EVENT-WS-BEFORE-SEND` | aviso | `<id>:wsBeforeSend` | `ypy:ws:before:message:outgoing` |
| `V2-EVENT-WS-AFTER-SEND` | aviso | `<id>:wsAfterSend` | `ypy:ws:after:message:outgoing` |
| `V2-EVENT-WS-BEFORE-MESSAGE` | aviso | `<id>:wsBeforeMessage` | `ypy:ws:before:message:incoming` |
| `V2-EVENT-WS-AFTER-MESSAGE` | aviso | `<id>:wsAfterMessage` | `ypy:ws:after:message:incoming` |
| `V2-EVENT-VALIDATION-VALIDATE` | quebra | `<id>:validation:validate` | removed (`quebra`): use native validation (`checkValidity`, the `invalid` event) or validate in a `ypy:config:request` listener |
| `V2-EVENT-VALIDATION-FAILED` | quebra | `<id>:validation:failed` | removed: use the native `invalid` event on the field |
| `V2-EVENT-VALIDATION-HALTED` | quebra | `<id>:validation:halted` | removed: use the native `invalid` event on the form |
| `V2-EVENT-XHR-LOADSTART` | quebra | `<id>:xhr:loadstart` | removed: use `ypy:before:request` |
| `V2-EVENT-XHR-LOADEND` | quebra | `<id>:xhr:loadend` | removed: use `ypy:finally:request` |
| `V2-EVENT-XHR-PROGRESS` | quebra | `<id>:xhr:progress` | removed, no equivalent for upload progress: use `ypy:before:request`, or a fetch-based upload |
| `V2-EVENT-XHR-ABORT` | quebra | `<id>:xhr:abort` | removed: use `ypy:abort` |
| `V2-EVENT-HISTORY-CACHE-MISS-LOAD` | muda | `<id>:historyCacheMissLoad` | load the `ypy-history-cache` extension, or remove the listener |
| `V2-EVENT-HISTORY-CACHE-MISS-LOAD-ERROR` | muda | `<id>:historyCacheMissLoadError` | load the `ypy-history-cache` extension, or remove the listener |
| `V2-EVENT-HISTORY-CACHE-ERROR` | muda | `<id>:historyCacheError` | load the `ypy-history-cache` extension, or remove the listener |
| `V2-EVENT-HISTORY-CACHE-HIT` | muda | `<id>:historyCacheHit` | load the `ypy-history-cache` extension, or remove the listener |
| `V2-EVENT-HISTORY-ITEM-CREATED` | muda | `<id>:historyItemCreated` | remove the listener |
| `V2-EVENT-VALIDATE-URL` | muda | `<id>:validateUrl` | set `config.mode`; veto a request in a `ypy:config:request` listener |
| `V2-EVENT-PROCESS-TEMPLATE` | muda | `<id>:processTemplate` | use `ypy-partial` and the out-of-band attributes instead of a template handler |
| `V2-EVENT-OOB-ERROR-NO-TARGET` | muda | `<id>:oobErrorNoTarget` | `ypy:error` |
| `V2-EVENT-ON-LOAD-ERROR` | muda | `<id>:onLoadError` | `ypy:error` |
| `V2-EVENT-EVAL-DISALLOWED-ERROR` | muda | `<id>:evalDisallowedError` | `ypy:error` |
| `V2-DETAIL-SWAP-DECISION` | muda | `detail.shouldSwap`, `detail.isError`, `detail.serverResponse`, ... | decide per status with `config.noSwap` or `ypy-status:<code>="swap:none"`; edit the response in `ypy:before:swap` (`detail.ctx.text`) |
| `V2-DETAIL-REQUEST` | muda | `detail.xhr`, `detail.requestConfig`, `detail.parameters`, ... | read and write `detail.ctx.request` (headers, body as `FormData`, action, method) and `detail.ctx.response` |
| `ANY-EVENT-NOT-EMITTED` | quebra | a `ypy:` event name that YPY does not emit | use a name from the event table of the [reference](reference.md#events) |
| `V2-DETAIL-ELT` | muda | `detail.elt` | `evt.target` |

#### Headers

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V2-HEADER-TRIGGER-AFTER-SWAP` | muda | `<P>-Trigger-After-Swap` | send `YPY-Trigger` and listen to `ypy:after:swap` |
| `V2-HEADER-TRIGGER-AFTER-SETTLE` | muda | `<P>-Trigger-After-Settle` | send `YPY-Trigger` and listen to `ypy:after:settle` |
| `V2-HEADER-REQ-TRIGGER` | muda | `<P>-Trigger` (read) | read `YPY-Source` (`button#save`); the old header carried only the element id |
| `V2-HEADER-REQ-TRIGGER-NAME` | muda | `<P>-Trigger-Name` | read `YPY-Source`, or send the value in the body |
| `V2-HEADER-REQ-TARGET` | muda | `<P>-Target` (read) | compare with the `tag#id` format |
| `V2-HEADER-REQ-PROMPT` | aviso | `<P>-Prompt` (read) | load the `ypy-prompt` extension |
| `ANY-HEADER-TARGET-PLAIN-ID` | muda | `<P>-Target` compared with a bare id | write the header as `tag#id`, for example `main#content` |

#### Configuration keys

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V1-CONFIG-TEMPLATE-FRAGMENTS` | aviso | `useTemplateFragments` | remove it |
| `V2-CONFIG-DEFAULT-SWAP-STYLE` | aviso | `defaultSwapStyle` | `defaultSwap` |
| `V2-CONFIG-GLOBAL-VIEW-TRANSITIONS` | aviso | `globalViewTransitions` | `transitions` |
| `V2-CONFIG-HISTORY-ENABLED` | aviso | `historyEnabled` | `history` |
| `V2-CONFIG-INCLUDE-INDICATOR-STYLES` | aviso | `includeIndicatorStyles` | `includeIndicatorCSS` |
| `V2-CONFIG-TIMEOUT` | muda | `timeout` | `defaultTimeout` (60000 ms by default; `0` is not "none") |
| `V2-CONFIG-SELF-REQUESTS-ONLY` | aviso | `selfRequestsOnly` | `mode`: `true` is `"same-origin"`, `false` is `"cors"` |
| `V2-CONFIG-DISABLE-INHERITANCE` | muda | `disableInheritance` | removed: inheritance is always explicit; add `:inherited` where it is needed |
| `V2-CONFIG-REQUEST-CLASS` | aviso | `requestClass` | `classes.request` |
| `V2-CONFIG-INDICATOR-CLASS` | aviso | `indicatorClass` | `classes.indicator` |
| `V2-CONFIG-ADDED-CLASS` | aviso | `addedClass` | `classes.added` |
| `V2-CONFIG-SETTLING-CLASS` | aviso | `settlingClass` | `classes.settling` |
| `V2-CONFIG-SWAPPING-CLASS` | aviso | `swappingClass` | `classes.swapping` |
| `V2-CONFIG-ALLOW-EVAL` | muda | `allowEval` | removed: the page's CSP decides, with `config.callbacks` |
| `V2-CONFIG-ALLOW-SCRIPT-TAGS` | muda | `allowScriptTags` | scripts in swapped content follow `config.scripts`; sanitize on the server |
| `V2-CONFIG-ALLOW-NESTED-OOB-SWAPS` | muda | `allowNestedOobSwaps` | nested out-of-band swaps are always processed |
| `V2-CONFIG-ATTRIBUTES-TO-SETTLE` | muda | `attributesToSettle` | settling is automatic |
| `V2-CONFIG-DEFAULT-SWAP-DELAY` | muda | `defaultSwapDelay` | use the `swap:` modifier of `ypy-swap` |
| `V2-CONFIG-DISABLE-SELECTOR` | muda | `disableSelector` | use `ypy-ignore` |
| `V2-CONFIG-GET-CACHE-BUSTER-PARAM` | muda | `getCacheBusterParam` | add the parameter in a `ypy:config:request` listener |
| `V2-CONFIG-HISTORY-CACHE-SIZE` | muda | `historyCacheSize` | load the `ypy-history-cache` extension for snapshots |
| `V2-CONFIG-METHODS-THAT-USE-URL-PARAMS` | muda | `methodsThatUseUrlParams` | parameters of GET requests go in the URL; other methods send a body |
| `V2-CONFIG-REFRESH-ON-HISTORY-MISS` | muda | `refreshOnHistoryMiss` | a history restore is always fetched; there is no cache-miss setting |
| `V2-CONFIG-RESPONSE-HANDLING` | muda | `responseHandling` | `config.noSwap` (a list of statuses) or `ypy-status:<code>="swap:none"` |
| `V2-CONFIG-SCROLL-BEHAVIOR` | muda | `scrollBehavior` | use `scroll-behavior` in the page CSS |
| `V2-CONFIG-SCROLL-INTO-VIEW-ON-BOOST` | muda | `scrollIntoViewOnBoost` | use the `show:` and `scroll:` modifiers of `ypy-swap` |
| `V2-CONFIG-TRIGGER-SPECS-CACHE` | muda | `triggerSpecsCache` | no replacement is needed |
| `V2-CONFIG-WITH-CREDENTIALS` | muda | `withCredentials` | `ypy-config="credentials:include"`, or `config.mode` |
| `V2-CONFIG-WS-BINARY-TYPE` | muda | `wsBinaryType` | set it on the connection in a `ypy:ws:before:connection` listener |
| `V2-CONFIG-WS-RECONNECT-DELAY` | muda | `wsReconnectDelay` | use the reconnect options of `config.ws` |
| `V2-CONFIG-REPORT-VALIDITY-OF-FORMS` | muda | `reportValidityOfForms` | forms are checked with `reportValidity`; `ypy-validate="false"` skips it |
| `V2-CONFIG-IGNORE-TITLE` | muda | `ignoreTitle` | use the `ignoreTitle:true` modifier of `ypy-swap` |

#### JavaScript API

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `V2-API-ADD-CLASS` | muda | `<id>.addClass` | `classList.add()` |
| `V2-API-REMOVE-CLASS` | muda | `<id>.removeClass` | `classList.remove()` |
| `V2-API-TOGGLE-CLASS` | muda | `<id>.toggleClass` | `classList.toggle()` |
| `V2-API-CLOSEST` | muda | `<id>.closest` | `element.closest()` |
| `V2-API-REMOVE` | muda | `<id>.remove` | `element.remove()` |
| `V2-API-OFF` | muda | `<id>.off` | `removeEventListener()` (`ypy.on()` returns the callback) |
| `V2-API-LOCATION` | muda | `<id>.location` | `ypy.ajax()` |
| `V2-API-TAKE-CLASS` | muda | `<id>.takeClass` | a few lines with `classList`, or `ypy.live.take()` from the live extension |
| `V2-API-LOG-ALL` | aviso | `<id>.logAll` | `ypy.config.logAll = true` |
| `V2-API-LOG-NONE` | aviso | `<id>.logNone` | `ypy.config.logAll = false` |
| `V2-API-LOGGER` | muda | `<id>.logger` | `console.*` |
| `V2-API-VALUES` | muda | `<id>.values` | `new FormData(form)` |
| `V2-API-DEFINE-EXTENSION` | muda | `<id>.defineExtension` | `ypy.registerExtension(name, hooks)`; the API is event based, so rewrite the extension |

#### Script tags

| Rule | Severity | Legacy | YPY and what to do |
|---|---|---|---|
| `ANY-SCRIPT-LIBRARY` | aviso | a script tag that loads the older library | `<script src="ypy.min.js">` |
| `V2-SCRIPT-EXTENSION` | aviso | a script tag of one of its extensions | remove it when YPY ships the equivalent (SSE, WebSocket, head, preload, morph); extensions written for the old extension API need a rewrite |

## After migrating

### Server side

After the conversion the server has four things to do: read `YPY-Request`, send `Vary: YPY-Request`, answer `422` for an
invalid form and `303` after a successful `POST`, and carry the CSRF token. [Server side](index.md#server-side) in the
Docs has the headers. A server that has to serve old and YPY pages during a transition can read both headers for a
while. The old name then appears in one place in your code, and you delete it with the last old page.

### Verify

```sh
ypy check .                      # the worklist: nothing "quebra" left
ypy migrate .                    # "0 of N file(s) would change": the conversion is complete and stable
```

Then run your own test suite in a browser, against the file you now serve.

### Roll back

The conversion is a normal commit. `git revert` it, put the previous script tag back, and you are where you started. YPY
does not touch your data or your server.

### Frequently asked

**Can I keep the old behavior for a while?** No. YPY has one behavior, so the `ypy check` review is part of the move.
You can do it page by page: convert, fix that page's worklist, ship.

**My `<p>-on` handlers stopped working.** Author code is off by default (`callbacks: "none"`). To turn it on, set it in
the configuration, knowing what that means ([Security](index.md#security) and
[Configuring](index.md#configuring)).

**Can the old and the new spelling live together?** No. The only precedence rule is between the two YPY spellings: on
one element, `ypy-*` wins over `data-ypy-*`.
