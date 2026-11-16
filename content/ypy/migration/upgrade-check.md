---
title: "Upgrade check: ypy check"
description: "ypy check reads a project written for the older library and lists what breaks in YPY, what behaves differently and what only needs renaming, by file and line."
order: 30
---
# Upgrade check: `ypy check`

`ypy check` scans a project written for the older library and finds every use of a removed or changed behavior. For
each one it says whether it will break in YPY, whether it will behave differently, or whether it only needs
renaming. It reads files as text and never runs your code. It is a subcommand of the `ypy` binary, so it needs no
Python, no Node and no browser.

It is the worklist of the [migration guide](index.md). YPY has only its native behavior and no compatibility layer, so
`ypy migrate` renames and `ypy check` lists what else changes.

| Step | What you do | Where `ypy check` helps |
|---|---|---|
| 1. Before | Run `ypy check .` on the project. | The worklist: what breaks (`quebra`), what behaves differently (`muda`) and what is only spelled the legacy way (`aviso`). |
| 2. Convert | `ypy migrate --write` converts attributes, headers, events, classes and the script tags (see [`ypy migrate`](migrate.md)). | It does not: the conversion does not depend on the check. |
| 3. After | Fix the `quebra` and review the `muda` findings, page by page. | Run it again: it must be quiet. |

## Quick start

```sh
ypy check ./templates          # a directory, several paths or single files
ypy check .                    # the whole project (node_modules, vendor, dist and .git are skipped)
ypy check ./app --json > check.json
ypy check . --id acme              # name the legacy identifier when the code does not reveal it
```

Typical output:

```text
templates/items.html:1:6  muda   INHERIT-IMPLICIT  Attribute inherited implicitly by descendants; inheritance is explicit now. (acme-target is inherited implicitly (descendant on line 2 has acme-get))
    fix:   Add :inherited to the attribute on the container, for example ypy-target:inherited="#out". Do not add it blindly: only where a descendant relies on it.
    guide: https://ypy.dev/docs/guides/upgrade-check#inherit-implicit
templates/items.html:2:29  muda   ATTR-VARS  acme-vars is removed.
    fix:   Use ypy-vals with the js: prefix: ypy-vals="js:{a: 1}".
    guide: https://ypy.dev/docs/guides/upgrade-check#attr-vars

Legacy markup learned: attribute prefix acme, from acme-target (templates/items.html:1) and acme-vars (templates/items.html:2).
1 file(s) scanned, 1 with findings: 0 quebra, 2 muda, 2 aviso.
```

The listing is trimmed: the two `aviso` findings of that file are left out. Each finding gives `file:line:column`
(clickable in most terminals and editors), the severity, the rule id, what changes, the fix and a link to the rule in
the table at the end of this page. A rule that still awaits confirmation says so.

## Options

| Option | Meaning |
|---|---|
| `--json` | One JSON document on stdout (schema 1, below). |
| `--min-severity aviso\|muda\|quebra` | Hide the findings below that level. Default `aviso`. |
| `--fail-on quebra\|muda\|aviso\|never` | Lowest severity that makes the exit status 1. Default `quebra`. |
| `--prefix P` | The attribute prefix of the legacy markup when the code does not reveal it (also the header prefix unless `--header-prefix` is given). See [How the prefixes are recognized](#how-the-prefixes-are-recognized). |
| `--header-prefix P` | The header prefix, when it differs from the attribute prefix. |
| `--min-evidence N` | Distinct names that make a prefix learned (default 2). |
| `--id NAME[,NAME]` | The legacy identifier (its global, event prefix and class prefix), when the code does not reveal it. See [How legacy markup is recognized](#how-legacy-markup-is-recognized). |
| `--min-confidence medium\|high` | `high` hides the findings whose identifier rests on weak signs. Default `medium` (shows both). |
| `--ignore ID[,ID]` | Skip rules (repeatable). Use it for a rule you decided to live with. |
| `--only ID[,ID]` | Report only these rules. |
| `--ext .a,.b` | Scan more extensions as templates (for example `--ext .vue,.go`). `.md`, `.markdown` and `.json` are read by default. |
| `--list-rules` | Print the rule table (honors `--json`). |
| `--gen-guide [--check]` | Rewrite the rule tables of this guide from the rules; with `--check`, write nothing and exit 1 when the guide is out of date. `--guide FILE` picks another file. |
| `--no-color` | Plain output. Colors are also off when stdout is not a terminal or `NO_COLOR` is set. |

## Severities

| Severity | Meaning | Exit status |
|---|---|---|
| `quebra` | Stops working after the move to YPY. Fix before shipping. | 1 |
| `muda` | Keeps working, but behaves differently in YPY from what the project was written for. Review it. | 0 (1 with `--fail-on muda`) |
| `aviso` | Works. Only the spelling is legacy and a YPY spelling exists. `ypy migrate` converts it. | 0 (1 with `--fail-on aviso`) |

Exit status `2` means a usage error (unknown option, unreadable path, unknown rule id).

A rule marked "to confirm" was written from how the older markup behaves and has not been confirmed yet. Treat its
severity as provisional.

## How legacy markup is recognized

The rules never contain the name of the library the markup came from. What they know is its shape: the names of its events
(`afterSwap`, `responseError`), of its configuration keys (`historyCacheSize`), of its API methods (`defineExtension`),
and of its state classes (`request`, `indicator`). The identifier that goes in front of them (the global `<id>.ajax(`,
the event prefix `<id>:afterSwap`, the class prefix `<id>-request`, the script `<id>.min.js`) is **learned from your
project**, in a first pass over all the files:

- **Strong evidence** names it by itself: an event only the legacy API spells (`<id>:afterSwap`), a call of a method only it
  has (`<id>.defineExtension(`), a configuration key written as `<id>.config.historyCacheSize` or inside the
  `<id>-config` meta tag, two different state classes together (`<id>-request` and `<id>-indicator`), or the name given
  with `--id`.
- **Weak evidence** is ambiguous alone: a script called `<id>.min.js`, a single state class, an `<id>:load` event,
  an `<id>-config` meta tag. Two different kinds of weak evidence for the same identifier count as evidence, and the
  findings that rest on it get **medium** confidence.
- An event with a common one-word name (`<id>:load`, `<id>:timeout`) or an `<id>:error` is also reported with medium
  confidence when it is written in event-handling code (`addEventListener`, a trigger or `on` attribute) of a project
  that uses prefixed attributes, even if the identifier is not learned.

The tool never mistakes `ypy` for the older markup. Code that only looks like it (`Object.values(x)`, `el.remove()`,
`socket.on('chat:load')`, `$.ajax()`, a `loading-indicator` class) stays quiet while nothing in the project reveals the
identifier. When a project uses the legacy script without ever calling its API or listening to its events, its script tag
and classes are not reported until you name the identifier: the report says so, and `--id NAME` fixes it.
Attribute, header and structure rules do not depend on the identifier and always have high confidence. They
depend on the prefixes, which are learned too (next section).

## How the prefixes are recognized

The older markup spells its attributes with a short prefix (`<p>-get`, `<p>-target`) and its HTTP headers with the
same word in capitals (`<P>-Request`, `<P>-Trigger`). The tool does not contain those words either. It learns them
from the project, in the same first pass.

- **Attribute prefix**: the word in front of names of the request vocabulary (`get`, `post`, `put`, `patch`,
  `delete`, `target`, `swap`, `trigger`, `boost`, `select`, `vals`, `on`, `include`, `indicator`, `push-url`, ...), in
  the shape of an attribute (`name="..."`), of a selector (`[name]`) or of a quoted name, also with the `data-`
  spelling. The word `data` itself is never taken.
- **Header prefix**: the word in front of `Request`, `Trigger`, `Target`, `Redirect`, `Location`, `Push-Url`,
  `Reswap`, `Retarget`, `Current-URL`, `Boosted`, `Refresh`, ... inside a string, or in the CGI spelling
  (`HTTP_<P>_REQUEST`). It also counts when it is the attribute prefix.
- **Evidence**: a prefix is learned from at least two distinct names (`--min-evidence N` changes it); a request verb
  whose value is a URL counts double, and at least one name must be specific to the request vocabulary (`on` and
  `target` alone are not enough, a `x-on` of another framework stays quiet).
- **Forcing**: `--prefix P` (and `--header-prefix P` when the headers differ) names the prefix when the project does
  not reveal it, for example a back end repository that only reads one header. `P` is written without the hyphen
  (`acme`, not `acme-`).

The report says which prefixes and identifiers it learned (`legacy` in the JSON output).

## What is scanned

Default extensions: `.html .htm .xhtml` (HTML), `.gohtml .tmpl .tpl` (Go `html/template`), `.templ` (templ),
`.kyse.go` (Kyse templates), `.jinja .jinja2 .j2 .njk` (Jinja, Django), `.erb .rhtml` (ERB), `.php .phtml`
(PHP and Blade), `.twig` (Twig), `.hbs .handlebars .mustache .ejs .liquid .pug`, scripts (`.js .mjs .cjs .jsx .ts
.mts .cts .tsx .vue .svelte .astro`) and styles (`.css .scss .sass .less`). Server code (`.go .py .rb .java .kt .cs
.ex .exs .rs` and PHP) is read only for the HTTP header rules. Files over 2 MB, `*.min.js`, `*.min.css` and the
folders `node_modules`, `vendor`, `dist`, `build`, `coverage`, `.git` are skipped.

Templates are read **as text**: the template syntax (`{{ }}`, `{% %}`, `<% %>`, `<?php ?>`, Blade `@directive(...)`,
templ `{ expr }`) is blanked out keeping every line and column, then the markup is tokenized with a lenient parser.
Nothing is interpreted or executed.

## What it cannot see

- **Server responses.** The 4xx and 5xx swap change, or a header your server computes, only show where the code
  names them (`responseHandling`, a `Trigger-After-Swap` header).
- **Values built at run time**: an attribute name or an event name assembled from strings or template logic.
- **Inheritance across files.** A base template with a `headers` or `target` attribute on `<body>` is flagged because the
  requests live in other files, even when the scan of the child templates finds no request there.
- **Default values that changed** (for example the scroll behavior) when the project never sets them.
- **Extensions you wrote yourself**: the legacy extension API shows up as `defineExtension` (`muda`), but what the
  extension does is yours to review.

A hit can be a false positive when the same text means something else (an unrelated `preload` attribute on a custom
element). Rules that can do that skip the common cases (`<video preload>`); use `--ignore` for the rest. Findings
whose legacy identifier is only weakly evidenced say `medium confidence`; `--min-confidence high` hides
them.

## JSON output (schema 1)

`--json` prints one document, stable in field names and order:

```json
{
  "schema": 1,
  "tool": "ypy-check",
  "rulesVersion": "2026.11.1",
  "scanned": 1,
  "summary": {"quebra": 0, "muda": 2, "aviso": 2, "total": 4, "files": 1},
  "findings": [
    {"file": "templates/items.html", "line": 2, "column": 29, "severity": "muda", "rule": "ATTR-VARS",
     "message": "acme-vars is removed.", "match": "acme-vars",
     "fix": "Use ypy-vals with the js: prefix: ypy-vals=\"js:{a: 1}\".",
     "native": "ypy-vals=\"js:...\"", "guide": "https://ypy.dev/docs/guides/upgrade-check#attr-vars",
     "confirm": false, "count": 1, "confidence": "high"}
  ],
  "warnings": [],
  "legacy": {"attributePrefixes": ["acme"], "headerPrefixes": ["acme"], "identifiers": []}
}
```

| Field | Meaning |
|---|---|
| `schema` | Version of this format. It changes only when a field is removed or renamed. |
| `rulesVersion` | Version of the rule table (`internal/check/rules.json`). |
| `scanned` | Number of files read. |
| `summary` | Findings per severity, the total, and the number of files with findings. |
| `findings[].file` | Path relative to the working directory (absolute when outside it), with `/` separators. |
| `findings[].line`, `column` | 1-based position of the match. |
| `findings[].severity` | `quebra`, `muda` or `aviso`. |
| `findings[].rule` | Rule id, stable across versions. Never reused for another meaning. |
| `findings[].match` | The matched text (attribute name, event, key, header, URL). |
| `findings[].fix`, `native` | What to do, and the YPY equivalent when there is one (`null` otherwise). |
| `findings[].guide` | Link to the rule in this guide. |
| `findings[].confirm` | `true` when the rule awaits confirmation. |
| `findings[].count` | 1, or the number of occurrences for rules reported once per file (`ANY-ATTR-PREFIX`, `ANY-CLASS-PREFIX`, `ANY-API-GLOBAL`). |
| `findings[].confidence` | `high` or `medium` (see [How legacy markup is recognized](#how-legacy-markup-is-recognized)). |
| `legacy` | The attribute prefixes, header prefixes and identifiers of the legacy markup the scan learned or was given. |
| `warnings` | Files that could not be read. |

Findings are sorted by file, line and column. Rule ids are upper-case and start with the kind (`ATTR`, `EVENT`, `API`,
`CONFIG`, `HEADER`, ...), then the name.

## Use in CI and in an editor

```sh
# fail the build only on what breaks
ypy check ./app
# fail on anything that needs a human decision, ignore the pure renames
ypy check ./app --min-severity muda --fail-on muda
# a GitHub Actions annotation per finding
ypy check ./app --json | jq -r '.findings[] | "::\(if .severity == "quebra" then "error" else "warning" end) file=\(.file),line=\(.line),col=\(.column)::\(.rule) \(.message)"'
```

## Checking your stack

No setup is needed for any stack. Build the `ypy` binary (`go build -o ypy ./cmd/ypy` in the YPY repository) and point
it at the folder that holds your templates.

| Stack | Command | Notes |
|---|---|---|
| Django, Jinja, Flask | `ypy check templates/ static/js/` | Check the CSRF header on `<body>` (`INHERIT-IMPLICIT`) and the `HTTP_<P>_*` reads in views (`HEADER-REQ-*`). |
| Rails, ERB | `ypy check app/views app/javascript` | `rails-method` style extensions show up as `SCRIPT-EXTENSION`. |
| Laravel, Blade | `ypy check resources/views resources/js` | `@directive(...)` and `{{ }}` are blanked out; run it on the `public/` assets only when they are not minified. |
| Symfony, Twig | `ypy check templates/ assets/` | |
| Go `html/template`, templ | `ypy check views/ internal/web/` | Go handlers are read for the header rules (`w.Header().Set("<P>-Trigger-After-Swap", ...)`). |
| Arandu (Kyse) | `ypy check resources/views` | `.kyse.go` templates are read natively. |

## Reference: every rule

The tables below are generated from `internal/check/rules.json` (`ypy check --gen-guide`); a test and the CI fail when
they drift. `<id>` stands for the legacy identifier, `<p>-` for its attribute prefix and `<P>-` for its header prefix. The **Legacy** column is what the tool looks for and **YPY** is the equivalent. Anchors are the lowercase rule ids, so a link such as
`upgrade-check.md#attr-vars` always lands on the rule.

<!-- rules:begin -->

### Attributes

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v1-attr-on-legacy"></a>`V1-ATTR-ON-LEGACY` | v1 | quebra | `<p>-on` with a matching value | `ypy-on:click="code"` | The "event: code" form of the &lt;p&gt;-on attribute is not supported. Write one ypy-on:&lt;event&gt; attribute per handler, or the arrow form ypy-on="click -&gt; code". |
| <a id="v2-attr-vars"></a>`V2-ATTR-VARS` | v1, v2 | muda | `<p>-vars` | `ypy-vals="js:..."` | &lt;p&gt;-vars is removed. Use ypy-vals with the js: prefix: ypy-vals="js:{a: 1}". |
| <a id="v2-attr-params"></a>`V2-ATTR-PARAMS` | v1, v2 | muda | `<p>-params` | `ypy:config:request listener` | &lt;p&gt;-params is removed. Filter the parameters in a ypy:config:request listener (detail.ctx.request.body is a FormData). |
| <a id="v2-attr-prompt"></a>`V2-ATTR-PROMPT` | v1, v2 | aviso | `<p>-prompt` | `extension ypy-prompt` | &lt;p&gt;-prompt moved to an extension. Load the prompt extension (ypy-prompt.js); the attribute and the YPY-Prompt header stay the same. |
| <a id="v2-attr-ext"></a>`V2-ATTR-EXT` | v1, v2 | muda | `<p>-ext` | `(none)` | &lt;p&gt;-ext is removed: extensions are active when their script is included. Delete the attribute and keep the extension script; restrict extensions with config.extensions if needed. |
| <a id="v2-attr-disinherit"></a>`V2-ATTR-DISINHERIT` | v1, v2 | muda | `<p>-disinherit` | `(none)` | &lt;p&gt;-disinherit is removed: inheritance is explicit. Delete it, and add :inherited to the attributes that must reach descendants. |
| <a id="v2-attr-inherit"></a>`V2-ATTR-INHERIT` | v2 | muda | `<p>-inherit` | `(none)` | &lt;p&gt;-inherit is removed: inheritance is explicit. Delete it, and add :inherited to the attributes that must reach descendants. |
| <a id="v2-attr-request"></a>`V2-ATTR-REQUEST` | v1, v2 | muda | `<p>-request` | `ypy-config` | &lt;p&gt;-request is replaced by ypy-config. Rename to ypy-config; it accepts the same JSON (timeout, credentials, noHeaders). |
| <a id="v2-attr-history"></a>`V2-ATTR-HISTORY` | v1, v2 | aviso | `<p>-history` | `(none)` | &lt;p&gt;-history="false" is removed: history snapshots are no longer kept in localStorage. Delete it. |
| <a id="v2-attr-disable"></a>`V2-ATTR-DISABLE` | v1, v2 | muda | `<p>-disable` with a matching value | `ypy-ignore` | &lt;p&gt;-disable (without a value) means "skip processing" here, but a different thing in the native attribute set. Rename to ypy-ignore. Do this before renaming ypy-disabled-elt to ypy-disable. |
| <a id="v2-attr-disabled-elt"></a>`V2-ATTR-DISABLED-ELT` | v1, v2 | aviso | `<p>-disabled-elt` | `ypy-disable` | &lt;p&gt;-disabled-elt is renamed. Rename to ypy-disable (after ypy-disable became ypy-ignore). |
| <a id="v2-attr-on-dash"></a>`V2-ATTR-ON-DASH` | v1, v2 | muda | attribute names that match the rule | `ypy-on:click` | The dash spelling of &lt;p&gt;-on (&lt;p&gt;-on-click, &lt;p&gt;-on--after-swap) works only when config.metaCharacter is "-". Use the colon spelling (ypy-on:click, ypy-on::after:swap), or set config.metaCharacter to "-" and keep the dashes. |

### Extension attributes, swap styles and event sources

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v1-attr-sse"></a>`V1-ATTR-SSE` | v1 | quebra | `<p>-sse` | `ypy-sse:connect="/url"` | The &lt;p&gt;-sse attribute of the v1 markup is gone. Use the SSE attribute: ypy-sse:connect="/url" (swap targets and events follow the native model). |
| <a id="v1-attr-ws"></a>`V1-ATTR-WS` | v1 | quebra | `<p>-ws` | `ypy-ws:connect="/url"` | The &lt;p&gt;-ws attribute of the v1 markup is gone. Use ypy-ws:connect="/url" and ypy-ws:send. |
| <a id="v2-ext-sse-connect"></a>`V2-EXT-SSE-CONNECT` | v1, v2 | aviso | `sse-connect` | `ypy-sse:connect` | sse-connect is renamed. Rename to ypy-sse:connect. |
| <a id="v2-ext-sse-close"></a>`V2-EXT-SSE-CLOSE` | v1, v2 | aviso | `sse-close` | `ypy-sse:close` | sse-close is renamed. Rename to ypy-sse:close. |
| <a id="v2-ext-sse-swap"></a>`V2-EXT-SSE-SWAP` | v1, v2 | quebra | `sse-swap` | `(none)` | sse-swap is not supported: unnamed messages are swapped automatically and named events become DOM events. Drop sse-swap and listen to the named events (ypy-on) or send unnamed messages. |
| <a id="v2-ext-ws-connect"></a>`V2-EXT-WS-CONNECT` | v1, v2 | aviso | `ws-connect` | `ypy-ws:connect` | ws-connect is renamed. Rename to ypy-ws:connect. |
| <a id="v2-ext-ws-send"></a>`V2-EXT-WS-SEND` | v1, v2 | aviso | `ws-send` | `ypy-ws:send` | ws-send is renamed. Rename to ypy-ws:send. |
| <a id="v2-ext-preload"></a>`V2-EXT-PRELOAD` | v1, v2 | aviso | `preload` | `ypy-preload` | The preload attribute of the preload extension is renamed. Rename to ypy-preload (values such as mouseover are unchanged). |
| <a id="v2-ext-response-targets"></a>`V2-EXT-RESPONSE-TARGETS` | v1, v2 | aviso | attribute names that match the rule | `ypy-status:404="target:#x"` | The response-targets extension attributes are replaced by ypy-status. Use ypy-status:404="target:#x" (ypy-status:4xx, ypy-status:5xx for ranges). |
| <a id="v2-swap-morph"></a>`V2-SWAP-MORPH` | v1, v2 | aviso | `<p>-swap` with a matching value | `innerMorph, outerMorph` | The idiomorph swap style is built in and renamed. Use innerMorph or outerMorph (no extension needed). |

### Events

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v2-event-after-on-load"></a>`V2-EVENT-AFTER-ON-LOAD` | v1, v2 | aviso | `<id>:afterOnLoad` | `ypy:after:request` | Legacy event &lt;id&gt;:afterOnLoad is renamed. Listen to ypy:after:request. |
| <a id="v2-event-after-process-node"></a>`V2-EVENT-AFTER-PROCESS-NODE` | v1, v2 | aviso | `<id>:afterProcessNode` | `ypy:after:init` | Legacy event &lt;id&gt;:afterProcessNode is renamed. Listen to ypy:after:init. |
| <a id="v2-event-after-request"></a>`V2-EVENT-AFTER-REQUEST` | v1, v2 | aviso | `<id>:afterRequest` | `ypy:finally:request` | Legacy event &lt;id&gt;:afterRequest is renamed. Listen to ypy:finally:request (it fires on failures too, as the old event did; ypy:after:request fires before the swap). |
| <a id="v2-event-after-settle"></a>`V2-EVENT-AFTER-SETTLE` | v1, v2 | aviso | `<id>:afterSettle` | `ypy:after:settle` | Legacy event &lt;id&gt;:afterSettle is renamed. Listen to ypy:after:settle. |
| <a id="v2-event-after-swap"></a>`V2-EVENT-AFTER-SWAP` | v1, v2 | aviso | `<id>:afterSwap` | `ypy:after:swap` | Legacy event &lt;id&gt;:afterSwap is renamed. Listen to ypy:after:swap. |
| <a id="v2-event-before-cleanup-element"></a>`V2-EVENT-BEFORE-CLEANUP-ELEMENT` | v1, v2 | aviso | `<id>:beforeCleanupElement` | `ypy:before:cleanup` | Legacy event &lt;id&gt;:beforeCleanupElement is renamed. Listen to ypy:before:cleanup. |
| <a id="v2-event-before-history-save"></a>`V2-EVENT-BEFORE-HISTORY-SAVE` | v1, v2 | aviso | `<id>:beforeHistorySave` | `ypy:before:history:update` | Legacy event &lt;id&gt;:beforeHistorySave is renamed. Listen to ypy:before:history:update. |
| <a id="v2-event-before-history-update"></a>`V2-EVENT-BEFORE-HISTORY-UPDATE` | v1, v2 | aviso | `<id>:beforeHistoryUpdate` | `ypy:before:history:update` | Legacy event &lt;id&gt;:beforeHistoryUpdate is renamed. Listen to ypy:before:history:update. |
| <a id="v2-event-before-on-load"></a>`V2-EVENT-BEFORE-ON-LOAD` | v1, v2 | aviso | `<id>:beforeOnLoad` | `ypy:before:response` | Legacy event &lt;id&gt;:beforeOnLoad is renamed. Listen to ypy:before:response. |
| <a id="v2-event-before-process-node"></a>`V2-EVENT-BEFORE-PROCESS-NODE` | v1, v2 | aviso | `<id>:beforeProcessNode` | `ypy:before:init` | Legacy event &lt;id&gt;:beforeProcessNode is renamed. Listen to ypy:before:init. |
| <a id="v2-event-before-request"></a>`V2-EVENT-BEFORE-REQUEST` | v1, v2 | aviso | `<id>:beforeRequest` | `ypy:before:request` | Legacy event &lt;id&gt;:beforeRequest is renamed. Listen to ypy:before:request. |
| <a id="v2-event-before-send"></a>`V2-EVENT-BEFORE-SEND` | v1, v2 | aviso | `<id>:beforeSend` | `ypy:before:request` | Legacy event &lt;id&gt;:beforeSend is renamed. Listen to ypy:before:request. |
| <a id="v2-event-before-swap"></a>`V2-EVENT-BEFORE-SWAP` | v1, v2 | aviso | `<id>:beforeSwap` | `ypy:before:swap` | Legacy event &lt;id&gt;:beforeSwap is renamed. Listen to ypy:before:swap (the swap decision moved to config.noSwap and ypy-status). |
| <a id="v2-event-before-transition"></a>`V2-EVENT-BEFORE-TRANSITION` | v1, v2 | aviso | `<id>:beforeTransition` | `ypy:before:viewTransition` | Legacy event &lt;id&gt;:beforeTransition is renamed. Listen to ypy:before:viewTransition. |
| <a id="v2-event-config-request"></a>`V2-EVENT-CONFIG-REQUEST` | v1, v2 | aviso | `<id>:configRequest` | `ypy:config:request` | Legacy event &lt;id&gt;:configRequest is renamed. Listen to ypy:config:request (detail.ctx.request replaces detail.headers, parameters and path). |
| <a id="v2-event-history-cache-miss"></a>`V2-EVENT-HISTORY-CACHE-MISS` | v1, v2 | aviso | `<id>:historyCacheMiss` | `ypy:before:history:restore` | Legacy event &lt;id&gt;:historyCacheMiss is renamed. Listen to ypy:before:history:restore. |
| <a id="v2-event-history-restore"></a>`V2-EVENT-HISTORY-RESTORE` | v1, v2 | aviso | `<id>:historyRestore` | `ypy:before:history:restore` | Legacy event &lt;id&gt;:historyRestore is renamed. Listen to ypy:before:history:restore. |
| <a id="v2-event-load"></a>`V2-EVENT-LOAD` | v1, v2 | aviso | `<id>:load` | `ypy:after:process` | Legacy event &lt;id&gt;:load is renamed. Listen to ypy:after:process. |
| <a id="v2-event-oob-after-swap"></a>`V2-EVENT-OOB-AFTER-SWAP` | v1, v2 | aviso | `<id>:oobAfterSwap` | `ypy:after:swap` | Legacy event &lt;id&gt;:oobAfterSwap is renamed. Listen to ypy:after:swap. |
| <a id="v2-event-oob-before-swap"></a>`V2-EVENT-OOB-BEFORE-SWAP` | v1, v2 | aviso | `<id>:oobBeforeSwap` | `ypy:before:swap` | Legacy event &lt;id&gt;:oobBeforeSwap is renamed. Listen to ypy:before:swap. |
| <a id="v2-event-pushed-into-history"></a>`V2-EVENT-PUSHED-INTO-HISTORY` | v1, v2 | aviso | `<id>:pushedIntoHistory` | `ypy:after:history:push` | Legacy event &lt;id&gt;:pushedIntoHistory is renamed. Listen to ypy:after:history:push. |
| <a id="v2-event-replaced-in-history"></a>`V2-EVENT-REPLACED-IN-HISTORY` | v1, v2 | aviso | `<id>:replacedInHistory` | `ypy:after:history:replace` | Legacy event &lt;id&gt;:replacedInHistory is renamed. Listen to ypy:after:history:replace. |
| <a id="v2-event-response-error"></a>`V2-EVENT-RESPONSE-ERROR` | v1, v2 | aviso | `<id>:responseError` | `ypy:response:error` | Legacy event &lt;id&gt;:responseError is renamed. Listen to ypy:response:error. |
| <a id="v2-event-send-abort"></a>`V2-EVENT-SEND-ABORT` | v1, v2 | aviso | `<id>:sendAbort` | `ypy:error` | Legacy event &lt;id&gt;:sendAbort is renamed. Listen to ypy:error. |
| <a id="v2-event-send-error"></a>`V2-EVENT-SEND-ERROR` | v1, v2 | aviso | `<id>:sendError` | `ypy:error` | Legacy event &lt;id&gt;:sendError is renamed. Listen to ypy:error. |
| <a id="v2-event-swap-error"></a>`V2-EVENT-SWAP-ERROR` | v1, v2 | aviso | `<id>:swapError` | `ypy:error` | Legacy event &lt;id&gt;:swapError is renamed. Listen to ypy:error. |
| <a id="v2-event-target-error"></a>`V2-EVENT-TARGET-ERROR` | v1, v2 | aviso | `<id>:targetError` | `ypy:error` | Legacy event &lt;id&gt;:targetError is renamed. Listen to ypy:error. |
| <a id="v2-event-timeout"></a>`V2-EVENT-TIMEOUT` | v1, v2 | aviso | `<id>:timeout` | `ypy:error` | Legacy event &lt;id&gt;:timeout is renamed. Listen to ypy:error. |
| <a id="v2-event-prompt"></a>`V2-EVENT-PROMPT` | v1, v2 | aviso | `<id>:prompt` | `ypy:prompt (prompt extension)` | Legacy event &lt;id&gt;:prompt is renamed. Listen to ypy:prompt (prompt extension). |
| <a id="v2-event-sse-open"></a>`V2-EVENT-SSE-OPEN` | v1, v2 | aviso | `<id>:sseOpen` | `ypy:sse:after:connection` | Legacy event &lt;id&gt;:sseOpen is renamed. Listen to ypy:sse:after:connection. |
| <a id="v2-event-sse-error"></a>`V2-EVENT-SSE-ERROR` | v1, v2 | aviso | `<id>:sseError` | `ypy:sse:error` | Legacy event &lt;id&gt;:sseError is renamed. Listen to ypy:sse:error. |
| <a id="v2-event-sse-before-message"></a>`V2-EVENT-SSE-BEFORE-MESSAGE` | v1, v2 | aviso | `<id>:sseBeforeMessage` | `ypy:sse:before:message` | Legacy event &lt;id&gt;:sseBeforeMessage is renamed. Listen to ypy:sse:before:message. |
| <a id="v2-event-sse-message"></a>`V2-EVENT-SSE-MESSAGE` | v1, v2 | aviso | `<id>:sseMessage` | `ypy:sse:after:message` | Legacy event &lt;id&gt;:sseMessage is renamed. Listen to ypy:sse:after:message. |
| <a id="v2-event-sse-close"></a>`V2-EVENT-SSE-CLOSE` | v1, v2 | aviso | `<id>:sseClose` | `ypy:sse:close` | Legacy event &lt;id&gt;:sseClose is renamed. Listen to ypy:sse:close. |
| <a id="v2-event-ws-open"></a>`V2-EVENT-WS-OPEN` | v1, v2 | aviso | `<id>:wsOpen` | `ypy:ws:after:connection` | Legacy event &lt;id&gt;:wsOpen is renamed. Listen to ypy:ws:after:connection. |
| <a id="v2-event-ws-close"></a>`V2-EVENT-WS-CLOSE` | v1, v2 | aviso | `<id>:wsClose` | `ypy:ws:close` | Legacy event &lt;id&gt;:wsClose is renamed. Listen to ypy:ws:close. |
| <a id="v2-event-ws-error"></a>`V2-EVENT-WS-ERROR` | v1, v2 | aviso | `<id>:wsError` | `ypy:ws:error` | Legacy event &lt;id&gt;:wsError is renamed. Listen to ypy:ws:error. |
| <a id="v2-event-ws-config-send"></a>`V2-EVENT-WS-CONFIG-SEND` | v1, v2 | aviso | `<id>:wsConfigSend` | `ypy:ws:before:message:outgoing` | Legacy event &lt;id&gt;:wsConfigSend is renamed. Listen to ypy:ws:before:message:outgoing. |
| <a id="v2-event-ws-before-send"></a>`V2-EVENT-WS-BEFORE-SEND` | v1, v2 | aviso | `<id>:wsBeforeSend` | `ypy:ws:before:message:outgoing` | Legacy event &lt;id&gt;:wsBeforeSend is renamed. Listen to ypy:ws:before:message:outgoing. |
| <a id="v2-event-ws-after-send"></a>`V2-EVENT-WS-AFTER-SEND` | v1, v2 | aviso | `<id>:wsAfterSend` | `ypy:ws:after:message:outgoing` | Legacy event &lt;id&gt;:wsAfterSend is renamed. Listen to ypy:ws:after:message:outgoing. |
| <a id="v2-event-ws-before-message"></a>`V2-EVENT-WS-BEFORE-MESSAGE` | v1, v2 | aviso | `<id>:wsBeforeMessage` | `ypy:ws:before:message:incoming` | Legacy event &lt;id&gt;:wsBeforeMessage is renamed. Listen to ypy:ws:before:message:incoming. |
| <a id="v2-event-ws-after-message"></a>`V2-EVENT-WS-AFTER-MESSAGE` | v1, v2 | aviso | `<id>:wsAfterMessage` | `ypy:ws:after:message:incoming` | Legacy event &lt;id&gt;:wsAfterMessage is renamed. Listen to ypy:ws:after:message:incoming. |
| <a id="v2-event-validation-validate"></a>`V2-EVENT-VALIDATION-VALIDATE` | v1, v2 | quebra | `<id>:validation:validate` | `(none)` | &lt;id&gt;:validation:validate: Validation events are gone: validation uses the browser constraint API. Use native form validation (checkValidity, the invalid event) or validate in a ypy:config:request listener. |
| <a id="v2-event-validation-failed"></a>`V2-EVENT-VALIDATION-FAILED` | v1, v2 | quebra | `<id>:validation:failed` | `(none)` | &lt;id&gt;:validation:failed: Validation events are gone: validation uses the browser constraint API. Use the native invalid event on the field. |
| <a id="v2-event-validation-halted"></a>`V2-EVENT-VALIDATION-HALTED` | v1, v2 | quebra | `<id>:validation:halted` | `(none)` | &lt;id&gt;:validation:halted: Validation events are gone: validation uses the browser constraint API. Use the native invalid event on the form. |
| <a id="v2-event-xhr-loadstart"></a>`V2-EVENT-XHR-LOADSTART` | v1, v2 | quebra | `<id>:xhr:loadstart` | `(none)` | &lt;id&gt;:xhr:loadstart: XHR events are gone: requests use fetch. Use ypy:before:request. |
| <a id="v2-event-xhr-loadend"></a>`V2-EVENT-XHR-LOADEND` | v1, v2 | quebra | `<id>:xhr:loadend` | `(none)` | &lt;id&gt;:xhr:loadend: XHR events are gone: requests use fetch. Use ypy:finally:request. |
| <a id="v2-event-xhr-progress"></a>`V2-EVENT-XHR-PROGRESS` | v1, v2 | quebra | `<id>:xhr:progress` | `(none)` | &lt;id&gt;:xhr:progress: XHR progress events are gone: requests use fetch. No equivalent for upload progress; use ypy:before:request and a fetch based upload where needed. |
| <a id="v2-event-xhr-abort"></a>`V2-EVENT-XHR-ABORT` | v1, v2 | quebra | `<id>:xhr:abort` | `(none)` | &lt;id&gt;:xhr:abort: XHR events are gone: requests use fetch. Use ypy:abort. |
| <a id="v2-event-history-cache-miss-load"></a>`V2-EVENT-HISTORY-CACHE-MISS-LOAD` | v1, v2 | muda | `<id>:historyCacheMissLoad` | `(none)` | &lt;id&gt;:historyCacheMissLoad: History cache events no longer exist: the history snapshot cache is an optional extension. Load the ypy-history-cache extension or remove the listener. |
| <a id="v2-event-history-cache-miss-load-error"></a>`V2-EVENT-HISTORY-CACHE-MISS-LOAD-ERROR` | v1, v2 | muda | `<id>:historyCacheMissLoadError` | `(none)` | &lt;id&gt;:historyCacheMissLoadError: History cache events no longer exist: the history snapshot cache is an optional extension. Load the ypy-history-cache extension or remove the listener. |
| <a id="v2-event-history-cache-error"></a>`V2-EVENT-HISTORY-CACHE-ERROR` | v1, v2 | muda | `<id>:historyCacheError` | `(none)` | &lt;id&gt;:historyCacheError: History cache events no longer exist: the history snapshot cache is an optional extension. Load the ypy-history-cache extension or remove the listener. |
| <a id="v2-event-history-cache-hit"></a>`V2-EVENT-HISTORY-CACHE-HIT` | v1, v2 | muda | `<id>:historyCacheHit` | `(none)` | &lt;id&gt;:historyCacheHit: History cache events no longer exist: the history snapshot cache is an optional extension. Load the ypy-history-cache extension or remove the listener. |
| <a id="v2-event-history-item-created"></a>`V2-EVENT-HISTORY-ITEM-CREATED` | v1, v2 | muda | `<id>:historyItemCreated` | `(none)` | &lt;id&gt;:historyItemCreated: History cache events no longer exist: the history snapshot cache is an optional extension. Remove the listener. |
| <a id="v2-event-validate-url"></a>`V2-EVENT-VALIDATE-URL` | v1, v2 | muda | `<id>:validateUrl` | `(none)` | &lt;id&gt;:validateUrl: The egress check event no longer exists: cross-origin requests are controlled by config.mode. Set config.mode to "cors" or "same-origin"; veto a request in a ypy:config:request listener. |
| <a id="v2-event-process-template"></a>`V2-EVENT-PROCESS-TEMPLATE` | v1, v2 | muda | `<id>:processTemplate` | `(none)` | &lt;id&gt;:processTemplate: Partials work differently: &lt;ypy-partial&gt; is handled by the core. Use the partial attributes of the core instead of a custom template handler. |
| <a id="v2-event-oob-error-no-target"></a>`V2-EVENT-OOB-ERROR-NO-TARGET` | v1, v2 | muda | `<id>:oobErrorNoTarget` | `(none)` | &lt;id&gt;:oobErrorNoTarget: This diagnostic event no longer exists. Use ypy:error. |
| <a id="v2-event-on-load-error"></a>`V2-EVENT-ON-LOAD-ERROR` | v1, v2 | muda | `<id>:onLoadError` | `(none)` | &lt;id&gt;:onLoadError: This diagnostic event no longer exists. Use ypy:error. |
| <a id="v2-event-eval-disallowed-error"></a>`V2-EVENT-EVAL-DISALLOWED-ERROR` | v1, v2 | muda | `<id>:evalDisallowedError` | `(none)` | &lt;id&gt;:evalDisallowedError: This diagnostic event no longer exists. Use ypy:error. |
| <a id="v2-detail-swap-decision"></a>`V2-DETAIL-SWAP-DECISION` | v1, v2 | muda | see the description | `detail.ctx.text, ypy-status` | The swap decision properties of the old beforeSwap detail are gone. Decide per status with config.noSwap or ypy-status:&lt;code&gt;="swap:none"; edit the response in ypy:before:swap (detail.ctx.text). |
| <a id="v2-detail-request"></a>`V2-DETAIL-REQUEST` | v1, v2 | muda | see the description | `detail.ctx.request, detail.ctx.response` | The old event detail (xhr, requestConfig, parameters, headers, verb) is replaced by detail.ctx. Read and write detail.ctx.request (headers, body as FormData, action, method) and detail.ctx.response. |
| <a id="any-event-not-emitted"></a>`ANY-EVENT-NOT-EMITTED` | v1, v2, v4 | quebra | see the description | `ypy:<phase>:<verb>, for example ypy:after:settle` | The event is written in the ypy: namespace but YPY does not emit it, so the listener never runs (typical after a migration that only swapped the prefix). Use a name of the event table of YPY (ypy:after:swap, ypy:after:settle, ypy:before:request, ypy:finally:request, ypy:after:process...); `ypy migrate --write` repairs the legacy names. |
| <a id="v2-detail-elt"></a>`V2-DETAIL-ELT` | v1, v2, v4 | muda | see the description | `evt.target` | The element is not in the detail of a YPY event: detail.elt is undefined. Read evt.target (the element the event was fired on). |

### HTTP headers

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v2-header-trigger-after-swap"></a>`V2-HEADER-TRIGGER-AFTER-SWAP` | v1, v2 | muda | `<P>-Trigger-After-Swap` | `YPY-Trigger` | The &lt;P&gt;-Trigger-After-Swap response header is removed. Send YPY-Trigger (events fire after the swap) or dispatch the event from a ypy:after:swap listener. |
| <a id="v2-header-trigger-after-settle"></a>`V2-HEADER-TRIGGER-AFTER-SETTLE` | v1, v2 | muda | `<P>-Trigger-After-Settle` | `YPY-Trigger` | The &lt;P&gt;-Trigger-After-Settle response header is removed. Send YPY-Trigger or dispatch the event from a ypy:after:settle listener. |
| <a id="v2-header-req-trigger"></a>`V2-HEADER-REQ-TRIGGER` | v1, v2 | muda | `<P>-Trigger` (request, read by the server) | `YPY-Source` | The &lt;P&gt;-Trigger request header is replaced by YPY-Source, in the tag#id format. Read YPY-Source (for example button#save); the old header carried only the element id. |
| <a id="v2-header-req-trigger-name"></a>`V2-HEADER-REQ-TRIGGER-NAME` | v1, v2 | muda | `<P>-Trigger-Name` | `YPY-Source` | The &lt;P&gt;-Trigger-Name request header is removed. Read YPY-Source, or send the value in the request body. |
| <a id="v2-header-req-target"></a>`V2-HEADER-REQ-TARGET` | v1, v2 | muda | `<P>-Target` (request, read by the server) | `YPY-Target: tag#id` | The &lt;P&gt;-Target request header changes from the element id to tag#id. Compare against the tag#id format (for example main#content). |
| <a id="v2-header-req-prompt"></a>`V2-HEADER-REQ-PROMPT` | v1, v2 | aviso | `<P>-Prompt` (request, read by the server) | `extension ypy-prompt` | The &lt;P&gt;-Prompt request header needs the prompt extension. Load the prompt extension (ypy-prompt.js). |
| <a id="any-header-target-plain-id"></a>`ANY-HEADER-TARGET-PLAIN-ID` | v1, v2, v4 | muda | see the description | `YPY-Target: tag#id` | The target header carries the plain id of the element; YPY sends tag#id (a request in a test or a client written with the id only does not match). Write the header as tag#id, for example main#content. |

### Configuration keys

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v1-config-template-fragments"></a>`V1-CONFIG-TEMPLATE-FRAGMENTS` | v1 | aviso | `useTemplateFragments` |  | useTemplateFragments no longer exists; responses are always parsed through &lt;template&gt;. Remove the setting. |
| <a id="v2-config-default-swap-style"></a>`V2-CONFIG-DEFAULT-SWAP-STYLE` | v1, v2 | aviso | `defaultSwapStyle` | `defaultSwap` | Config key defaultSwapStyle is renamed to defaultSwap. Use defaultSwap. |
| <a id="v2-config-global-view-transitions"></a>`V2-CONFIG-GLOBAL-VIEW-TRANSITIONS` | v1, v2 | aviso | `globalViewTransitions` | `transitions` | Config key globalViewTransitions is renamed to transitions. Use transitions. |
| <a id="v2-config-history-enabled"></a>`V2-CONFIG-HISTORY-ENABLED` | v1, v2 | aviso | `historyEnabled` | `history` | Config key historyEnabled is renamed to history. Use history. |
| <a id="v2-config-include-indicator-styles"></a>`V2-CONFIG-INCLUDE-INDICATOR-STYLES` | v1, v2 | aviso | `includeIndicatorStyles` | `includeIndicatorCSS` | Config key includeIndicatorStyles is renamed to includeIndicatorCSS. Use includeIndicatorCSS. |
| <a id="v2-config-timeout"></a>`V2-CONFIG-TIMEOUT` | v1, v2 | muda | `timeout` | `defaultTimeout` | Config key timeout is renamed to defaultTimeout. Use defaultTimeout: the default is 60000 ms natively (0 meant "no timeout"). |
| <a id="v2-config-self-requests-only"></a>`V2-CONFIG-SELF-REQUESTS-ONLY` | v1, v2 | aviso | `selfRequestsOnly` | `mode` | Config key selfRequestsOnly is renamed to mode. Use mode: selfRequestsOnly: true is mode: "same-origin", false is mode: "cors". |
| <a id="v2-config-disable-inheritance"></a>`V2-CONFIG-DISABLE-INHERITANCE` | v1, v2 | muda | `disableInheritance` | `implicitInheritance` | Config key disableInheritance is renamed to implicitInheritance. Use implicitInheritance: the meaning is inverted: disableInheritance: false is implicitInheritance: true. |
| <a id="v2-config-request-class"></a>`V2-CONFIG-REQUEST-CLASS` | v1, v2 | aviso | `requestClass` | `classes.request` | Config key requestClass is renamed to classes.request. Use classes.request. |
| <a id="v2-config-indicator-class"></a>`V2-CONFIG-INDICATOR-CLASS` | v1, v2 | aviso | `indicatorClass` | `classes.indicator` | Config key indicatorClass is renamed to classes.indicator. Use classes.indicator. |
| <a id="v2-config-added-class"></a>`V2-CONFIG-ADDED-CLASS` | v1, v2 | aviso | `addedClass` | `classes.added` | Config key addedClass is renamed to classes.added. Use classes.added. |
| <a id="v2-config-settling-class"></a>`V2-CONFIG-SETTLING-CLASS` | v1, v2 | aviso | `settlingClass` | `classes.settling` | Config key settlingClass is renamed to classes.settling. Use classes.settling. |
| <a id="v2-config-swapping-class"></a>`V2-CONFIG-SWAPPING-CLASS` | v1, v2 | aviso | `swappingClass` | `classes.swapping` | Config key swappingClass is renamed to classes.swapping. Use classes.swapping. |
| <a id="v2-config-allow-eval"></a>`V2-CONFIG-ALLOW-EVAL` | v1, v2 | muda | `allowEval` | `(none)` | Config key allowEval is removed. Set the Content-Security-Policy instead; evaluation of ypy-vals js: and ypy-on is controlled by the page policy. |
| <a id="v2-config-allow-script-tags"></a>`V2-CONFIG-ALLOW-SCRIPT-TAGS` | v1, v2 | muda | `allowScriptTags` | `(none)` | Config key allowScriptTags is removed. Scripts in swapped content follow the browser rules; sanitize on the server. |
| <a id="v2-config-allow-nested-oob-swaps"></a>`V2-CONFIG-ALLOW-NESTED-OOB-SWAPS` | v1, v2 | muda | `allowNestedOobSwaps` | `(none)` | Config key allowNestedOobSwaps is removed. Nested OOB swaps are always processed. |
| <a id="v2-config-attributes-to-settle"></a>`V2-CONFIG-ATTRIBUTES-TO-SETTLE` | v1, v2 | muda | `attributesToSettle` | `(none)` | Config key attributesToSettle is removed. Settling is automatic. |
| <a id="v2-config-default-swap-delay"></a>`V2-CONFIG-DEFAULT-SWAP-DELAY` | v1, v2 | muda | `defaultSwapDelay` | `(none)` | Config key defaultSwapDelay is removed. Use the swap: modifier on ypy-swap. |
| <a id="v2-config-disable-selector"></a>`V2-CONFIG-DISABLE-SELECTOR` | v1, v2 | muda | `disableSelector` | `(none)` | Config key disableSelector is removed. Use ypy-ignore. |
| <a id="v2-config-get-cache-buster-param"></a>`V2-CONFIG-GET-CACHE-BUSTER-PARAM` | v1, v2 | muda | `getCacheBusterParam` | `(none)` | Config key getCacheBusterParam is removed. Add the parameter in a ypy:config:request listener. |
| <a id="v2-config-history-cache-size"></a>`V2-CONFIG-HISTORY-CACHE-SIZE` | v1, v2 | muda | `historyCacheSize` | `(none)` | Config key historyCacheSize is removed. Load the ypy-history-cache extension for snapshots. |
| <a id="v2-config-methods-that-use-url-params"></a>`V2-CONFIG-METHODS-THAT-USE-URL-PARAMS` | v1, v2 | muda | `methodsThatUseUrlParams` | `(none)` | Config key methodsThatUseUrlParams is removed. Parameters of GET requests go in the URL; other methods send a body. |
| <a id="v2-config-refresh-on-history-miss"></a>`V2-CONFIG-REFRESH-ON-HISTORY-MISS` | v1, v2 | muda | `refreshOnHistoryMiss` | `(none)` | Config key refreshOnHistoryMiss is removed. History restores are fetched; there is no cache miss setting. |
| <a id="v2-config-response-handling"></a>`V2-CONFIG-RESPONSE-HANDLING` | v1, v2 | muda | `responseHandling` | `(none)` | Config key responseHandling is removed. Use config.noSwap (a list of status codes) or ypy-status:&lt;code&gt;="swap:none". |
| <a id="v2-config-scroll-behavior"></a>`V2-CONFIG-SCROLL-BEHAVIOR` | v1, v2 | muda | `scrollBehavior` | `(none)` | Config key scrollBehavior is removed. Use the scroll behavior of the page CSS (scroll-behavior). |
| <a id="v2-config-scroll-into-view-on-boost"></a>`V2-CONFIG-SCROLL-INTO-VIEW-ON-BOOST` | v1, v2 | muda | `scrollIntoViewOnBoost` | `(none)` | Config key scrollIntoViewOnBoost is removed. Use the show:/scroll: modifiers on ypy-swap. |
| <a id="v2-config-trigger-specs-cache"></a>`V2-CONFIG-TRIGGER-SPECS-CACHE` | v1, v2 | muda | `triggerSpecsCache` | `(none)` | Config key triggerSpecsCache is removed. No replacement is needed. |
| <a id="v2-config-with-credentials"></a>`V2-CONFIG-WITH-CREDENTIALS` | v1, v2 | muda | `withCredentials` | `(none)` | Config key withCredentials is removed. Use ypy-config="credentials:include" or ypy.config.mode. |
| <a id="v2-config-ws-binary-type"></a>`V2-CONFIG-WS-BINARY-TYPE` | v1, v2 | muda | `wsBinaryType` | `(none)` | Config key wsBinaryType is removed. Set it on the connection in a ypy:ws:before:connection listener. |
| <a id="v2-config-ws-reconnect-delay"></a>`V2-CONFIG-WS-RECONNECT-DELAY` | v1, v2 | muda | `wsReconnectDelay` | `(none)` | Config key wsReconnectDelay is removed. Configure ypy.config.ws reconnect options. |
| <a id="v2-config-report-validity-of-forms"></a>`V2-CONFIG-REPORT-VALIDITY-OF-FORMS` | v1, v2 | muda | `reportValidityOfForms` | `(none)` | Config key reportValidityOfForms is removed. Forms are validated through reportValidity; use ypy-validate="false" to skip it. |
| <a id="v2-config-ignore-title"></a>`V2-CONFIG-IGNORE-TITLE` | v1, v2 | muda | `ignoreTitle` | `(none)` | Config key ignoreTitle is removed. Use the title handling of the swap (ypy-swap ignoreTitle:true). |

### JavaScript API

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v2-api-add-class"></a>`V2-API-ADD-CLASS` | v1, v2 | muda | `<id>.addClass()` | `element.classList.add()` | The addClass() method of the legacy API is removed. Use element.classList.add(). |
| <a id="v2-api-remove-class"></a>`V2-API-REMOVE-CLASS` | v1, v2 | muda | `<id>.removeClass()` | `element.classList.remove()` | The removeClass() method of the legacy API is removed. Use element.classList.remove(). |
| <a id="v2-api-toggle-class"></a>`V2-API-TOGGLE-CLASS` | v1, v2 | muda | `<id>.toggleClass()` | `element.classList.toggle()` | The toggleClass() method of the legacy API is removed. Use element.classList.toggle(). |
| <a id="v2-api-closest"></a>`V2-API-CLOSEST` | v1, v2 | muda | `<id>.closest()` | `element.closest()` | The closest() method of the legacy API is removed. Use element.closest(). |
| <a id="v2-api-remove"></a>`V2-API-REMOVE` | v1, v2 | muda | `<id>.remove()` | `element.remove()` | The remove() method of the legacy API is removed. Use element.remove(). |
| <a id="v2-api-off"></a>`V2-API-OFF` | v1, v2 | muda | `<id>.off()` | `removeEventListener() (ypy.on() returns the callback)` | The off() method of the legacy API is removed. Use removeEventListener() (ypy.on() returns the callback). |
| <a id="v2-api-location"></a>`V2-API-LOCATION` | v1, v2 | muda | `<id>.location()` | `ypy.ajax()` | The location() method of the legacy API is removed. Use ypy.ajax(). |
| <a id="v2-api-take-class"></a>`V2-API-TAKE-CLASS` | v1, v2 | muda | `<id>.takeClass()` | `a few lines with classList, or ypy.live.take() from the live extension` | The takeClass() method of the legacy API is removed. Use a few lines with classList, or ypy.live.take() from the live extension. |
| <a id="v2-api-log-all"></a>`V2-API-LOG-ALL` | v1, v2 | aviso | `<id>.logAll()` | `ypy.config.logAll = true` | The logAll() method of the legacy API is removed. Use ypy.config.logAll = true. |
| <a id="v2-api-log-none"></a>`V2-API-LOG-NONE` | v1, v2 | aviso | `<id>.logNone()` | `ypy.config.logAll = false` | The logNone() method of the legacy API is removed. Use ypy.config.logAll = false. |
| <a id="v2-api-logger"></a>`V2-API-LOGGER` | v1, v2 | muda | `<id>.logger()` | `console.* (the library logs there directly)` | The logger() method of the legacy API is removed. Use console.* (the library logs there directly). |
| <a id="v2-api-values"></a>`V2-API-VALUES` | v1, v2 | muda | `<id>.values()` | `new FormData(form)` | The values() method of the legacy API is removed. Use new FormData(form). |
| <a id="v2-api-define-extension"></a>`V2-API-DEFINE-EXTENSION` | v1, v2 | muda | `<id>.defineExtension()` | `ypy.registerExtension() (the extension API is event based; rewrite the extension)` | The defineExtension() method of the legacy API is removed. Use ypy.registerExtension() (the extension API is event based; rewrite the extension). |

### Behavior changes

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="v1-delete-body"></a>`V1-DELETE-BODY` | v1 | muda | `<p>-delete` |  | DELETE requests carried the form values in the body in v1 markup; later generations send them as URL parameters or not at all. Check the server handler: read DELETE data from the query string, or add ypy-include explicitly. |
| <a id="v1-cross-origin"></a>`V1-CROSS-ORIGIN` | v1 | muda | `<p>-get`, `<p>-post`, `<p>-put`, `<p>-patch`, `<p>-delete` with a matching value | `ypy.config.mode = "cors"` | Request to an absolute URL: cross-origin requests are blocked by default since v2 markup. If the URL is on another origin, set config.mode to "cors" (the old selfRequestsOnly=false). |
| <a id="v2-attr-oob"></a>`V2-ATTR-OOB` | v1, v2 | aviso | `<p>-swap-oob` |  | Out-of-band swaps now run after the main swap, and an OOB-only response no longer empties the main target. Review code that depends on the order; to restore the old empty swap use config.allowEmptySwapAfterOOB or swapEmpty:true. |
| <a id="v2-swap-show-syntax"></a>`V2-SWAP-SHOW-SYNTAX` | v1, v2 | muda | `<p>-swap` with a matching value | `show:top showTarget:#sel` | The show:&lt;selector&gt;:&lt;position&gt; and scroll:&lt;selector&gt;:&lt;position&gt; modifiers use separate keys now. Write show:top showTarget:#sel (or scroll:bottom scrollTarget:#sel). |
| <a id="v2-trigger-queue"></a>`V2-TRIGGER-QUEUE` | v1, v2 | muda | `<p>-trigger` with a matching value | `ypy-sync="this:queue last"` | The queue: trigger modifier is replaced by ypy-sync. Use ypy-sync="this:queue last" (or queue first, queue all, drop, abort). |
| <a id="v2-include-inherit"></a>`V2-INCLUDE-INHERIT` | v1, v2 | muda | `<p>-include`, `<p>-indicator` with a matching value | `ypy-include:inherited:append` | The inherit keyword of &lt;p&gt;-include and &lt;p&gt;-indicator is gone. Put the shared value on the parent with the :inherited:append modifier (ypy-include:inherited:append="#shared"). |
| <a id="v2-inherit-implicit"></a>`V2-INHERIT-IMPLICIT` | v1, v2 | muda | see the description | `ypy-target:inherited` | Attribute inherited implicitly by descendants; inheritance is explicit now. Add :inherited to the attribute on the container, for example ypy-target:inherited="#out". Do not add it blindly: only where a descendant relies on it. |
| <a id="v2-delete-in-form"></a>`V2-DELETE-IN-FORM` | v1, v2 | muda | see the description | `ypy-include="closest form"` | A DELETE request inside a &lt;form&gt; no longer includes the form values. Add ypy-include="closest form" to the element. |

### Script tags

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="any-script-library"></a>`ANY-SCRIPT-LIBRARY` | v1, v2, v4 | aviso | see the description | `<script src="ypy.min.js">` | This script tag loads the library the project used before YPY. Replace this tag with the YPY script tag (level 1 migration: nothing else has to change). |
| <a id="v2-script-extension"></a>`V2-SCRIPT-EXTENSION` | v1, v2 | aviso | see the description |  | This script tag loads an extension written for the previous markup. Remove it when YPY ships the equivalent (sse, ws, head, preload, morph); keep it only for extensions that use the legacy extension API, which YPY does not offer. |

### Spelling only (prefix)

| Rule | Markup | Severity | Legacy | YPY | What changes and what to do |
|---|---|---|---|---|---|
| <a id="any-attr-prefix"></a>`ANY-ATTR-PREFIX` | v1, v2, v4 | aviso | attribute names that match the rule | `ypy-*` | Attributes use the legacy &lt;p&gt;- prefix (ypy migrate converts it). Run `ypy migrate --write` to convert the legacy prefix to ypy- (the data- spelling becomes data-ypy-). |
| <a id="any-meta-config"></a>`ANY-META-CONFIG` | v1, v2, v4 | aviso | see the description | `<meta name="ypy-config">` | The configuration meta tag uses the legacy name (still read). Rename the meta tag; the legacy name is merged first, so ypy-config wins when both exist. |
| <a id="any-class-prefix"></a>`ANY-CLASS-PREFIX` | v1, v2, v4 | aviso | see the description | `ypy-request, ypy-indicator, ypy-swapping, ypy-settling, ypy-added` | Legacy state class (the default configuration applies both the legacy and the ypy- class); also inside nested CSS and CSS written in a Go string. Style the ypy- class (or the data-ypy-state attribute) and drop the legacy one. |
| <a id="any-api-global"></a>`ANY-API-GLOBAL` | v1, v2, v4 | aviso | see the description | `ypy` | Uses the legacy global name; `ypy` is the native global (the alias stays available). Use `ypy.&lt;method&gt;` (`ypy migrate --write` converts it). |
| <a id="any-event-prefix"></a>`ANY-EVENT-PREFIX` | v1, v2, v4 | aviso | `<id>:error`, `<id>:abort`, `<id>:confirm` | `ypy:error, ypy:abort, ypy:confirm` | Event with an unchanged name but the legacy prefix (still dispatched). Listen to the ypy: event with the same name. |
| <a id="v4-event-prefix"></a>`V4-EVENT-PREFIX` | v4 | aviso | see the description | `ypy:<same name>` | Event with the legacy prefix (generation v4 markup). Replace the &lt;id&gt;: prefix with ypy: (same event name); `ypy migrate --write` does it. |
| <a id="any-ident-prefix"></a>`ANY-IDENT-PREFIX` | v1, v2, v4 | aviso | see the description | `Ypy<Name>, ypy<Name>, IsYpy` | Identifier named after the legacy prefix (a prop, field or variable such as &lt;P&gt;Target, &lt;p&gt;_request, Is&lt;P&gt;); the templates now say ypy. Rename the whole identifier (`ypy migrate --write` renames it in every file; --no-identifiers keeps it). |

<!-- rules:end -->

## Changing the rules

1. Edit `internal/check/rules.json` (one object per rule: id, severity, kind, match, what, fix, native, confirm). Events are written without a prefix (`"name": "afterSwap"`); texts that show the old spelling write the
   identifier as `{P}` (`{P}:afterSwap`), which the report fills in and this guide shows as `<id>`.
2. Add the rule to `internal/check/testdata/cases.json` with one positive and one negative snippet (they use the
   made-up identifier `acme`); the test fails when a rule has none.
3. Run `go run ./cmd/ypy check --gen-guide` to refresh the tables above, then `go test ./internal/check`.

Never reuse a rule id for another meaning, and note a change of severity in the changelog.
