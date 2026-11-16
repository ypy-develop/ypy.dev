---
order: 50
title: No build step
description: YPY loads with one script tag and runs the file as shipped. What that saves you, what it weighs in bytes, and where a bundler still fits.
---

# No build step

You use YPY by putting one tag in the page:

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
```

or, from the project's own CDN:

```html
<script src="https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js"></script>
```

That is the whole setup. There is no bundler, no transpiler and no package to resolve, and no generated folder to add
to `.gitignore`. The file the browser runs is the file in `dist/`.

## What you save

The markup is the source. A page that uses `ypy-get` and `ypy-target` can be read with "view source", and what the
browser runs is what you wrote. There is no compiled copy of the interface that has to be kept in step with it.

The dependency is one file with a version in its address. To upgrade, you change the number and read the
[changelog](https://github.com/ypy-develop/ypy/blob/main/CHANGELOG.md). A team that writes Go, Python, Ruby or PHP on
the server can serve the file as a static asset and does not need Node on the build machine.

You can also copy `dist/ypy.min.js` into your project and serve it yourself. The page then keeps working without
any registry or CDN being online.

## What it weighs

These are the files of version 0.1.1, measured on `dist/`:

| File | Raw | gzip -9 | Brotli |
|---|---|---|---|
| `ypy.min.js` (the core) | 50,407 B | 18,316 B | 16,512 B |
| `ext/ypy-sse.min.js` | 7,801 B | 3,072 B | 2,783 B |
| `ext/ypy-nav.min.js` | 6,942 B | 2,936 B | 2,603 B |
| `ext/ypy-validate.min.js` | 3,530 B | 1,586 B | 1,368 B |
| `ext/ypy-signals.min.js` | 13,088 B | 5,992 B | 5,429 B |
| `ypy.max.min.js` (the core and the official extensions that are safe to ship) | 186,525 B | 65,716 B | 57,420 B |

The build checks a size budget on every run: `ypy build --sizes` fails when the core goes over 18,432 B or an
extension over 6,144 B. That check uses Go's gzip, which gives 18,410 B for the core, so the core uses almost all of
its budget. `ypy-signals` and `ypy-live` come within about 200 B of the extension limit. Each `.min.js` has a
Brotli copy (`.br`) next to it.

Load only the extensions you use. `ypy.max.min.js` is handy for trying things out, but it is more than three times
the size of the core.

## Extensions are separate files

The core covers requests, swaps, morphing, history, errors and accessibility. Everything else is an
[extension](../extensions/index.md), loaded with another `<script>` tag after the core:

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-sse.min.js"></script>
<script src="ext/ypy-validate.min.js"></script>
```

A page pays only for what it loads. [Using and writing extensions](../extensions/using-and-writing.md) explains how to
choose, restrict and write one.

## When a bundler is fine

YPY does not forbid one. `dist/ypy.esm.min.js` is an ES module for a bundler, and `dist/ypy.d.ts` has the types. If
your project already has a bundler, import YPY there. It only means you are not required to.

## What you give up

You get all of the core, because it cannot be tree-shaken, and the size budget above is what keeps it from growing.
There is also no JSX or template compilation. The markup is plain HTML, so the mistakes a compiler would catch, such
as a misspelled attribute or a target that does not exist, show up in the browser instead. Load the
[`ypy-dev`](../extensions/dev.md) extension in development and it warns about them. Editors get completion from the
data files described in [Editor support](../reference/editors.md).

## See also

[Extensions](../extensions/index.md) · [Performance](performance.md) · [Core concepts](../guides/core-concepts.md)
