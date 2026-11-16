---
order: 20
title: Reference
description: "Every YPY attribute, CSS class, header, event, JavaScript method, setting and extension in one place, generated from the same table of names the library uses."
---

# Reference

Look a name up here. The [Docs](index.md) page explains how the pieces fit together. The tables below come from the
table of names that YPY itself is built from, so they list exactly the names the library has. Extensions add rows to
the same tables and say which extension owns each one.

## Core attributes

The attributes of the core, in groups: request, trigger, target and swap, parameters, history, errors, accessibility
and morph. An attribute is written `ypy-<name>` or `data-ypy-<name>`. An attribute reaches the elements below it only
when it carries the `:inherited` modifier, and the table says which ones accept it. Families such as `ypy-status:<code>`
take a code or a name after the colon.

<!-- reference:attributes -->

## Additional attributes

Extensions bring their own attributes, such as `ypy-poll`, `ypy-validate-fields`, `ypy-sse:connect`, `ypy-signals` and
`ypy-layer`. They work only when the extension is loaded, and they appear in the table above with the extension that
adds them. A page for each extension explains its attributes with examples: see [Extensions](#extensions).

## CSS classes

The classes YPY writes on elements during a request, a swap and a connection loss, and the classes of the optional
`ypy.css`. The names can be changed with `config.classes`.

<!-- reference:classes -->

## Request headers

What the browser sends on every YPY request. The server reads `YPY-Request` to tell a fragment request from a page
load, and sends `Vary: YPY-Request` when one URL answers both.

<!-- reference:request-headers -->

## Response headers

What the server can send back to change the target, the swap, the address or the page. A header here wins over the
attribute of the element that sent the request.

<!-- reference:response-headers -->

## Events

Every step of a request, a swap and a history change fires a DOM event named `ypy:<phase>:<verb>`. The table says on
which element each one fires, what its `detail` holds and whether `preventDefault()` cancels it.

<!-- reference:events -->

## JavaScript API

Most pages never call the `ypy` object, because attributes do the work. Use it to start a request from a script, to
listen to events, to tell YPY about HTML you inserted, or to test the page.

```js
ypy.ajax("GET", "/items", "#list");
ypy.ajax("POST", "/save", { target: "#result", swap: "outerHTML", values: { id: 7 } });
```

<!-- reference:js-api -->

## Configuration

Settings live in `ypy.config`. Set them from a script or, before the core loads, from a `<meta name="ypy-config">` tag
that holds JSON. The `prefix` key only counts when it is set before the core loads.

```html
<meta name="ypy-config" content='{"history": false, "defaultTimeout": 10000}'>
```

<!-- reference:config -->

### Editor support

`ypy gen` writes the data editors need from the same table of names: HTML custom data for VS Code, web-types for the
JetBrains IDEs, a JSON Schema for the configuration, and `ypy.d.ts` with the `ypy` object and a typed `detail` for every
event. With VS Code, point `html.customData` at `ypy.html-data.json`, and typing `ypy-` lists the attributes.

## Extensions

The official extensions, with the attributes, events, headers, settings and classes each one adds. A page for every
extension has the details and an example; [how to load them](extensions/index.md) and how to write your own are on the
index page.

<!-- reference:extensions -->
