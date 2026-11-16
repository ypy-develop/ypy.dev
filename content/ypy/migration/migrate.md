---
title: "Migrate a project: ypy migrate"
description: "ypy migrate rewrites a project written for the older library to use YPY names: attributes, headers, events, classes and script tags, as a diff you can review."
order: 20
---
# Migrate a project: `ypy migrate`

`ypy migrate` rewrites a project written for the older library so that it uses YPY names. The result loads
`ypy.min.js`, and there is no compatibility layer behind it. The tool reads files as text and edits only where it
recognizes the construct. It is a subcommand of the `ypy` binary, with no Node and no other dependency. Every Arandu
project (Kyse, hesape, the framework and the products built on them) moves to YPY with it.

```sh
ypy migrate .            # report and unified diff, nothing is written
ypy migrate --write .    # apply; run it twice and the second run changes nothing
```

## What it converts

Below, `<p>-` is the old attribute prefix, `<P>-` the old header prefix, and `<id>` the old identifier (its global,
event prefix and state classes). The tool does not contain any of them. It learns them from your project (see [How the prefixes are recognized](upgrade-check.md#how-the-prefixes-are-recognized)), or you give
them with `--prefix`, `--header-prefix` and `--id`.

| Before | After | Where |
|---|---|---|
| `<p>-get`, `data-<p>-target`, `<p>-on::after-swap` | `ypy-get`, `data-ypy-target`, `ypy-on::after-swap` | attributes in templates (HTML, Go `html/template`, templ, `.kyse.go`, Jinja, ERB, Blade, Twig, PHP, JSX, Vue), strings and selectors in JavaScript, TypeScript and CSS |
| `<<p>-partial>` | `<ypy-partial>` | markup |
| `<P>-Request`, `Vary: <P>-Request`, `HTTP_<P>_REQUEST` | `YPY-Request`, `Vary: YPY-Request`, `HTTP_YPY_REQUEST` | strings in Go, PHP, Python, Ruby, JavaScript, Java, Kotlin, C#, templates |
| `<id>:afterSwap`, `<id>:afterSettle`, `<id>:beforeRequest`, `<id>:afterRequest`, `<id>:load`, `<id>:before:request` | `ypy:after:swap`, `ypy:after:settle`, `ypy:before:request`, `ypy:finally:request`, `ypy:after:process`, `ypy:before:request` | strings and attribute values, `ypy-on::` and `ypy-on:` attributes: the name YPY emits is written, not only the prefix (the table is in [events](../reference/events/index.md)). An event with no YPY equivalent (`<id>:xhr:loadstart`) and a read of `detail.elt` (use `evt.target`) are left as they are and listed. A tree migrated by an earlier version (`ypy:afterSwap`) is repaired |
| `<id>.process(el)`, `window.<id>` | `ypy.process(el)`, `window.ypy` | JavaScript, inline scripts, inline handlers |
| `<id>-request`, `<id>-indicator`, `<id>-swapping`, ... | `ypy-request`, `ypy-indicator`, `ypy-swapping`, ... | class attributes, selectors, CSS |
| `<p>-sse.js`, `sse-connect`, `sse-swap`, `ws-connect`, `ws-send`, `preload="mouseover"` | `ypy-sse.js`, `ypy-sse:connect`, `ypy-sse-swap`, `ypy-ws:connect`, `ypy-ws:send`, `ypy-preload` | extension files and attributes |
| `<meta name="<id>-config">` | `<meta name="ypy-config">` | markup |
| `<script src=".../<id>.min.js">` | `<script src=".../ypy.min.js">` | markup, and the file name in strings such as `view.Asset("<id>.min.js")` |
| `<P>Target`, `<p>On`, `<p>_request`, `HTTP_<P>_REQUEST` (the prefix spelled as a word of an identifier, in front of a name of the vocabulary) | `YpyTarget`, `ypyOn`, `ypy_request`, `HTTP_YPY_REQUEST` | all code |
| `Is<P>`, `Has<P>`, `With<P>` (a predicate that ends in the prefix, in capitals or capitalized) | `IsYPY`, `HasYPY`, `WithYPY` | Go files only (`.go`, `.kyse.go`) |
| `extensions:"ypy-lock,sse,ws"`, `config.extensions = 'preload'` (the short names of the official extensions) | `extensions:"ypy-lock,ypy-sse,ypy-ws"`, `config.extensions = 'ypy-preload'` | the `extensions` setting in markup, scripts and server code; the names `sse`, `ws`, `preload`, `ptag`, `upsert`, `download`, `history-cache`, `browser-indicator`, `alpine-compat`, and no other |

What it does not do:

- It does not edit running text, and it leaves comments alone (`--comments` converts the recognized names in them too;
  identifiers such as `YpyTarget` are renamed everywhere, comments included).
- It skips `node_modules`, `vendor`, `dist`, `build`, `.git`, minified files, binary files and files over 2 MB. A
  copy of the old script or of one of its extensions (`<id>.js`, `<p>-sse.js`) is skipped and listed; replace it
  with the YPY file.
- A script tag with an external URL is left as it is and listed, because the tool cannot know where you serve the YPY
  build. `--script-src URL` rewrites it. A file that imports the legacy script as a module keeps its `<id>`
  global (the import binding is yours to change).
- Identifiers are renamed as a whole project: another module that uses your exported field `<P>Target` has to be
  migrated too (see the order below). `--no-identifiers` skips them.

It ends with a list of what it left alone: the places that still carry an old name (comments, prose, external tags)
and the number of lines that still name it in words. Edit those by hand.

## Step by step for an Arandu or Kyse project

1. **Start from a clean tree.** Commit what you have. The conversion is a normal diff you can review and revert.
2. **Read the worklist.** `ypy check .` lists what breaks (`quebra`) or behaves differently (`muda`) besides the
   spelling. It changes nothing.
3. **Dry run.** `ypy migrate .` prints the prefixes it learned and the diff. Check that the prefixes and the
   identifier are the right ones; if the project does not reveal them (a back end repository that reads one
   header, say), pass `--prefix`, `--header-prefix` and `--id`.
4. **Write.** `ypy migrate --write .`, then `ypy migrate .` again: it must report `0 of N file(s) would change`.
   In a Kyse project both the `.kyse.go` sources and the compiled `.go` files beside them are converted the same way,
   so they stay in step; `aru view:build` regenerates the compiled files from the sources with the same result.
5. **Swap the script.** Put `dist/ypy.min.js` where the old script was (the asset the layout loads,
   for example the one `view.Asset("ypy.min.js")` names after the conversion) and delete the old file and its
   extension scripts. YPY has one behavior and no compatibility layer: what behaves differently is the worklist of
   `ypy check` ([the migration guide](index.md#what-changes-in-behavior)).
6. **Run the gates of the project**: `gofmt`, `go build ./...`, `go vet ./...`, `go test ./...`, and your browser
   tests. A check that asserts the old spelling of an attribute or header is converted with the rest.
7. **Migrate in dependency order.** The code that writes and sanitizes attributes lives in the libraries (`hesape`,
   the framework, `kyse`). A product converted before them would write attributes its dependencies still refuse. Convert
   and release the libraries first (hesape, framework, kyse), then each product, and bump the dependency versions
   together.
8. **Fix the behavior, page by page.** Work through the worklist of `ypy check` ([upgrade check](upgrade-check.md)),
   page by page, until it is quiet.

## Options

| Option | Meaning |
|---|---|
| `--write` | Apply the changes. Without it nothing is written. |
| `--prefix P` | The attribute prefix, when the project does not reveal it (`acme`, as in `acme-get`); also the header prefix. |
| `--header-prefix P` | The header prefix, when it differs from the attribute prefix. |
| `--id NAME` | The legacy identifier (global, events, classes), when the code does not reveal it. Repeatable. |
| `--min-evidence N` | Distinct names that make a prefix learned (default 2). |
| `--comments` | Also convert the recognized names inside comments. |
| `--no-identifiers` | Leave identifiers (`<P>Target`, `<p>_request`) alone. |
| `--script-src URL` | Replace the URL of an external script tag of a legacy script. |
| `--ext .a,.b` | Convert more extensions as templates. |
| `--json` | One JSON document on stdout (schema 1). |
| `--no-diff` | Text output without the diff. |

Exit status: `0` when it ran, `2` on a usage error or an unreadable path.

## JSON output (schema 1)

`ypy migrate --json` prints one document, stable in field names and order: `schema`, `tool` (`ypy-migrate`),
`rulesVersion`, `mode` (`dry-run` or `write`), `legacy` (the learned
`attributePrefixes`, `headerPrefixes` and `identifiers`), `scanned`, `summary` (`files`, `replacements`, `byKind`,
`unconverted`, `mentions`), `files` (`file`, `replacements`, `byKind`, `written`, and the unified `diff` in a dry run),
`unconverted` (`file`, `line`, `kind`, `text`, `reason`), `skipped` and `warnings`. The kinds are `attribute`,
`header`, `event`, `class`, `global`, `identifier`, `script`, `extension`, `partial` and `config`.
