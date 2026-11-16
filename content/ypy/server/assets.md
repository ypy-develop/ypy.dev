---
order: 50
title: "Assets and CSP"
description: "Write CSS and small scripts next to a template, ship them as hashed files, and keep a strict Content-Security-Policy."
---

# Assets and CSP

A strict Content-Security-Policy refuses inline `<style>`, inline `<script>` and `style=` or `on*=` attributes. The SDK
and the `ypy` tool let you keep a component's CSS and script in its own template and still serve a page with nothing
inline.

## Write next to the template

Mark the blocks with `ypy-asset`:

```html
<!-- ui/card.html -->
<style ypy-asset>
  .card { padding: 1rem; border: 1px solid var(--line); }
</style>

<article class="card">{{ .Title }}</article>

<script ypy-asset>
  document.addEventListener("click", function (e) { /* ... */ });
</script>
```

A file named `card.ypy.css` or `card.ypy.js` next to `card.html` does the same without the markup.

## Extract with `ypy gen assets`

```sh
ypy gen assets ./ui --out ./assets --strip-out ./ui-clean
```

| Option | Meaning |
|---|---|
| `--out DIR` | where the hashed files and `manifest.json` go (required) |
| `--strip-out DIR` | also write a copy of each template with the marked blocks removed; serve these copies |
| `--group package` or `template` | one stylesheet and script per directory (default), or one pair per template |
| `--check` | write nothing and exit 1 when a file is missing, differs or is stale; use it in CI |

The files are named by the hash of their content (`ui-cards.1a2b3c4d5e.css`), and the same input always gives the same
bytes. Each script is wrapped in its own function scope; assign to `window` to publish a global.

## Load in Go

```go
//go:embed assets
var assetsFS embed.FS

m, err := ypy.LoadManifest(assetsFS, "assets/manifest.json") // reads the files and computes SHA-384
if err != nil {
	log.Fatal(err) // a missing asset fails at start, not at the first page
}
mux.Handle("/assets/", m.Handler()) // immutable cache, only the files of the manifest

tmpl := template.Must(template.New("page").Funcs(m.FuncMap()).Parse(page))
```

```html
<head>
  {{ assets "ui" .Nonce }}
</head>
```

`m.Assets("ui", nonce)` returns the `<link rel="stylesheet" integrity=...>` and `<script src defer integrity=...>` tags
of the origin `ui`, with the nonce when the page has one. `m.WithBase("/static/")` changes the URL prefix (the default is
`ypy.DefaultAssetsBase`, `/assets/`), and `m.Origins()` lists what the manifest holds.

## The policy

With no inline code, the policy needs no `'unsafe-inline'`:

```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self'
```

If the page also carries an inline element of your own (a JSON data block is fine), give it a nonce and pass the same
nonce to `m.Assets`. The [`ypy-csp`](../extensions/csp.md) extension makes the page trust only YPY markup that carries that nonce (`ypy-nonce`),
and the [security checklist](../guides/security-checklist.md) lists the rest.

## Test it

```go
res := ypytest.Do(handler, ypytest.NewRequest("GET", "/", ypytest.NotYPY()))
res.AssertNoInline(t) // fails on <style>, <script> with code, style=, on*= and javascript: URLs
```

`AssertNoInline` also fails when an element is still marked `ypy-asset`, which means a template was served without
being stripped. Run it on the full page.
