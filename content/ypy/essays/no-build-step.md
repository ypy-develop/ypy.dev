---
order: 50
title: No build step
description: YPY loads with one script tag and runs the file as shipped. What that saves you, what the core weighs against its budget, and where a bundler still fits.
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
browser runs is what you wrote. There is no compiled copy of the interface to keep in step with it.

The dependency is one file with a version in its address. To upgrade you change the number and read the
[changelog](https://github.com/ypy-develop/ypy/blob/main/CHANGELOG.md). Whatever language writes your server can serve
the file as a static asset, and the build machine does not need Node.

You can also copy `dist/ypy.min.js` into your project and serve it yourself. The page then keeps working with no
registry or CDN online, and a strict Content Security Policy is easier to write.

## What it weighs

The build enforces a size budget on every run: `ypy build --sizes` fails when the core goes over 18,432 B min+gzip or
an extension goes over 6,144 B. In version 0.1.1 the core measures 18,410 B, so it uses nearly all of its budget, and
the largest extension, `ypy-signals`, measures 6,009 B. The budget is what keeps the core from growing a feature at a
time.

Load only the extensions you use. `ypy.max.min.js` carries the core and the official extensions that are safe to ship,
which is handy for trying things out and larger than any page needs.

## Extensions are separate files

The core covers requests, swaps, morphing, history, errors and accessibility. Everything else is an
[extension](../extensions/index.md), loaded with another `<script>` tag after the core:

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-sse.min.js"></script>
<script src="ext/ypy-validate.min.js"></script>
```

A page pays only for what it loads. [Extensions](../index.md#extensions) in the docs explains how to choose and
restrict them.

## When a bundler is fine

YPY does not forbid one. `dist/ypy.esm.min.js` is an ES module for a bundler, and `dist/ypy.d.ts` has the types. If your
project already has a bundler, import YPY there. It only means you are not required to.

## What you give up

You get all of the core, because it cannot be tree-shaken. There is also no JSX or template compilation. The markup is
plain HTML, so the mistakes a compiler would catch, such as a misspelled attribute or a target that does not exist,
show up in the browser. Load the [`ypy-dev`](../extensions/dev.md) extension in development and it warns about them.
Editors get completion from the data files that `ypy gen` writes from the table of names.

## See also

[Extensions](../extensions/index.md) · [Performance](performance.md) · [Installing](../index.md#installing)
