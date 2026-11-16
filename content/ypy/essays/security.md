---
order: 90
title: Security
description: Why attributes that start requests change what an injection can do, and the defaults YPY ships to answer it: escaping, no author code, same origin, CSP.
---

# Security

YPY turns attributes into behavior. An element with `ypy-post` sends a request, and an element with `ypy-on:click` can
run code. So an attacker who can put markup into your page does not need a script, because the attributes are enough.
The defenses start on the server and end in the browser.

## Escape what you write

A fragment is HTML like any page, so everything that goes into it goes through the escaping of your template engine.
Two rules are specific to YPY. User text never lands inside a tag YPY reads or in a value that starts with `ypy-`, and
user content is sanitized on the server with every `ypy-*` and `data-ypy-*` attribute removed from it.

## Author code is off

Three kinds of attribute text can turn into code: `ypy-on` handlers, `js:` values and the `[filter]` of `ypy-trigger`.
`config.callbacks` decides what happens to them, and its default is `"none"`, so nothing is compiled. You turn it on
with `"all"` or `"nonce"`. Any other value, a typo included, is read as `"none"`.

## Same origin and a strict policy

Every request uses the fetch mode `same-origin` (`config.mode`), so an injected URL cannot make the page send data to
another origin. Change `mode` only for a page that really talks to another origin, and set CORS on that server.

YPY runs under a policy with no `unsafe-eval` and no `unsafe-inline`. [`ypy-expr`](../extensions/expr.md) evaluates
trigger filters without `eval`, and [`ypy-csp`](../extensions/csp.md) checks a nonce on each element before it acts.

## Tokens and caches

A site that uses cookies needs a token on every request that changes state. [`ypy-csrf`](../extensions/csrf.md) reads
it from one meta tag and adds it to every request except `GET` that goes to the page's own origin. A route that answers
a fragment to YPY and a page to a browser sends `Vary: YPY-Request`, or a cache can serve the wrong one.

## What YPY does not do

YPY cannot protect you from a script that already runs on your page. That script can call `fetch` or replace YPY as
easily as it can use it, and only a strict script policy stops it. [`ypy-lock`](../extensions/lock.md) guards against
careless code and markup injected into the head, and goes no further.

YPY does not decide who may do what. A button missing from the HTML hides an action from a reader and not from someone
who sends the request by hand, so every handler checks the permission itself.

The settings are in [Security](../index.md#security) and [Configuration](../reference.md#configuration). To report a
vulnerability, follow the `SECURITY.md` file of the repository.
