---
order: 30
title: Reference
description: "Tables and pages that list every attribute, header, event, class, setting and JavaScript method of YPY, with a link to each full page."
---

# Reference

The reference is where you look a name up. The [Docs](../index.md) page explains how the pieces fit together;
the pages here list each name with its value and what it does. Extensions have their own pages under
[Extensions](../extensions/index.md).

## The lists

| Page | What it lists |
|---|---|
| [Attributes](attributes/index.md) | Every attribute of the core, by group: request, trigger, target and swap, parameters, history, errors, accessibility, morph. |
| [HTTP headers](headers.md) | Request and response headers, with Go examples, and the headers added by extensions. |
| [Events](events/index.md) | How events work, their order and what you can cancel. |
| [Event list](events/list.md) | Every event with the element it fires on and whether it can be cancelled. |
| [JavaScript API](javascript-api.md) | The methods of the `ypy` object. |
| [Configuration](config/index.md) | Every setting of `ypy.config`. |
| [CSS classes](css-classes.md) | The classes and state attributes YPY writes, and the optional `ypy.css`. |

## Attributes in detail

| Page | Subject |
|---|---|
| [`ypy-sync`](attributes/ypy-sync.md) | What happens when an element asks for a request while another is in flight. |
| [`ypy-error-target`, `ypy-error-swap`, `ypy-error-focus`](attributes/ypy-error.md) | Where error responses go and where the focus lands. |
| [`ypy-main` and `ypy-fallback`](attributes/ypy-main.md) | The main target, and what happens when a target is missing. |
| [`ypy-focus`](attributes/ypy-focus.md) | Where keyboard focus goes when a swap ends. |
| [`ypy-scroll`](attributes/ypy-scroll.md) | The scroll position after a swap. |
| [`ypy-announce`](attributes/ypy-announce.md) | What screen readers hear after a request. |
| [`ypy-busy`](attributes/ypy-busy.md) | The automatic `aria-busy` mark. |
| [`ypy-preserve` and `ypy-preserve-attrs`](attributes/ypy-preserve.md) | What a swap or morph leaves alone. |
| [`ypy-inert`](attributes/ypy-inert.md) | Making a part of the page inert during a request. |
| [`<ypy-partial>`](tags/ypy-partial.md) | Routing pieces of one response to different targets. |
| [Attribute inheritance](inheritance.md) | Which attributes pass down with `:inherited`, and how. |

## Events in detail

| Page | Subject |
|---|---|
| [Events](events/index.md) | Event names, order, cancelling and changing a request or a swap. |
| [Request queue events](events/request-queue.md) | `ypy:request:queued`, `ypy:request:dropped` and `ypy:request:aborted`. |
| [Missing targets](events/target.md) | `ypy:target:fallback` and `ypy:target:missing`. |
| [Offline state](events/ypy-offline.md) | `ypy:offline`, `ypy:online`, `ypy:request:retry` and `ctx.errorKind`. |

## Behavior

| Page | Subject |
|---|---|
| [Swap styles and morph](swap-and-morph.md) | The swap styles, their modifiers, and what a morph keeps. |
| [Atomic swaps](atomic.md) | All or nothing swaps with `YPY-Group: atomic`, and atomic navigation. |
| [History](history.md) | How YPY pushes, replaces and restores addresses. |
| [Request timing](timing.md) | `ctx.timing`, the clock of a request. |
| [Security](security.md) | How the core handles untrusted HTML, scripts, URLs and origins. |
| [`config.classes`](config/classes.md) | Renaming the state classes. |
| [Styling](styling.md) | Using YPY with Tailwind, Bootstrap, Bulma, Pico or plain CSS. |
| [Editor support](editors.md) | Completion and checking in VS Code and JetBrains, from the same data as these tables. |

## Where the names come from

Every name on these pages comes from one file in the repository, `src/attributes.json`, plus one file per
extension in `src/attributes.d/`. The command `ypy gen` turns them into the editor data and the JSON schema of
the configuration, so the editors, the checker and this reference agree.

## Extensions

The core handles requests, swaps, history, errors and accessibility. Everything else is an extension that you
load after the core: streaming, signals, layers, polling, lazy loading and the rest. Start from the
[list of extensions](../extensions/index.md), or from [Using and writing extensions](../extensions/using-and-writing.md).
