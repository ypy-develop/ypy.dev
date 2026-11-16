---
order: 10
title: Why hypermedia
description: Why YPY keeps the rules and the links on the server and sends HTML fragments, what plain HTML controls cannot do, and where the approach costs you.
---

# Why hypermedia

A web page is a document with controls in it. A link says "you can go here". A form says "you can send this there". The
browser learns those options from the HTML it just received, so the server decides what the user can do next by what it
writes in the response. Nothing else has to know the routes.

That is hypermedia: the response carries the state of the page and the actions that are possible from it. Once an order
has shipped, the server leaves the "Cancel" button out of the HTML. The client never checks a rule, because there is no
button to click.

## What HTML already does

HTML has two controls that start a request, and both load a whole page:

- a link (`<a href>`) makes a `GET`;
- a form makes a `GET` or a `POST`.

Only a click on a link or a submit on a form starts a request, and the answer always replaces the whole window. Those
limits push many teams toward a JSON API and a client that draws the page from it. The server then describes data, and
the client has to know which actions each state allows. Two programs must agree on the rules, and over time they stop
agreeing.

## What YPY changes

YPY removes the limits of the controls and keeps the rest of the model:

| HTML limit | What YPY does |
|---|---|
| only links and forms start a request | any element can: `ypy-get`, `ypy-post`, `ypy-put`, `ypy-patch`, `ypy-delete` |
| only a click or a submit starts it | `ypy-trigger` names the event: `input changed delay:300ms`, `every 5s`, `revealed` |
| only `GET` and `POST` | the five verbs above |
| the answer replaces the window | `ypy-target` and `ypy-swap` put the answer where you say |

The server still answers with HTML. A search box looks like this:

```html
<input type="search" name="q" placeholder="Search"
       ypy-get="/people" ypy-trigger="input changed delay:300ms" ypy-target="#results">
<ul id="results"></ul>
```

The handler returns the `<li>` elements for the term, the same ones it would draw in a full page. The [Docs](../index.md#requests) walk through the request, the target and the swap.

## What stays on the server

The rules about who may do what stay where the data lives, and the HTML shows the result. One template draws a row,
whether it arrives inside a page or as a fragment ([Template fragments](template-fragments.md)). A link is a real URL: it
opens in a new tab, you can share it, and a search engine can find it. The session and the authorization stay on the
server too, so the client has no copy to keep in step.

## What the client keeps

A page still has state that belongs to the browser: what the user is typing, where the focus is, which panel is open
while they read. The default swap in YPY is a morph (`innerMorph`), which merges the new HTML into the existing nodes. A
field with the cursor in it keeps its value and its focus when the area around it changes. State that means nothing to
the server, such as a toggle that only changes how the page looks, can go in the [signals
extension](../extensions/signals.md) or in a few lines of script.

## What it costs

Every interaction is a round trip, and the server draws the HTML for each one. When a round trip is too slow for the
interaction, such as dragging an object under the pointer, the client has to act first. [When to use
YPY](when-to-use-ypy.md) lists those cases. For the ones where acting first and confirming later is enough, there is
[optimistic UI](../extensions/optimistic.md).

## See also

[Progressive enhancement](progressive-enhancement.md) · [Template fragments](template-fragments.md) ·
[Examples](../examples/index.md) · [Docs](../index.md)
