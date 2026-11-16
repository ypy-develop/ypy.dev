---
order: 90
title: Security
description: How YPY handles escaping, author code, same-origin requests, CSP, CSRF and caches, in one page, with links to the checklist and the full settings.
---

# Security

YPY turns attributes into behavior. An element with `ypy-post` sends a request, and an element with `ypy-on:click` can
run code. So an attacker who can put markup into your page does not need a script, because the attributes are enough.
The model below is the answer to that, and it starts on the server.

## 1. Escape what you write

Every value that goes into HTML goes through the escaping of your template engine. A fragment is HTML like any page,
and `html/template` in Go, Jinja, ERB, Blade and Twig all escape by default. Two rules are specific to YPY:

- Never let user text land inside a tag that YPY reads, or in an attribute value that starts with `ypy-`.
- Sanitize user content on the server and remove `ypy-*` and `data-ypy-*` attributes from it.

The reasons and the code are in the [security checklist](../guides/security-checklist.md#1-the-server-escapes-everything).

## 2. Author code is off

Three kinds of attribute text can turn into code: `ypy-on` handlers, `js:` values and the `[filter]` of `ypy-trigger`.
`config.callbacks` decides what happens to them, and its default is `"none"`, so no author code is compiled. To turn it
on you set `"all"` or `"nonce"` yourself. Any other value, a typo included, is read as `"none"`
([details](../reference/security.md#author-code-callbacks)).

## 3. Same origin by default

Every request is sent with the fetch mode `same-origin` (`config.mode`), so an injected URL cannot make the page send
data to another origin. `selfRequestsOnly` is an alias of the same switch. Change `mode` only for a page that really
talks to another origin, and set CORS on that server.

## 4. A strict Content Security Policy

YPY runs under a policy with no `unsafe-eval` and no `unsafe-inline`. The [`ypy-expr`](../extensions/expr.md)
extension evaluates trigger filters and `ypy-on` code without `eval`. The [`ypy-csp`](../extensions/csp.md) extension
checks a nonce on each element before it acts. The Go SDK can also serve the scripts and styles it extracts from your
templates as hashed files, so the page has nothing inline ([Assets](../server/assets.md)).

## 5. CSRF

A site that uses cookies needs a token on every request that changes state. The
[`ypy-csrf`](../extensions/csrf.md) extension reads the token from one meta tag and adds it to every request except
`GET` that goes to the page's own origin. It removes the token from a request to another origin. The Go SDK has the
server half ([CSRF](../server/csrf.md)).

## 6. Caches

A route that answers a fragment to YPY and a page to a browser must send `Vary: YPY-Request`, or a cache can serve the
wrong one. The Go middleware sets it for you. A fragment is not something a search engine should index, so the SDK
also adds `X-Robots-Tag: noindex` to fragments.

## What YPY does not do

YPY cannot protect you from a script that already runs on your page. That script can call `fetch` or replace YPY as
easily as it can use it, and only a strict script policy stops it. The settings lock
([`ypy-lock`](../extensions/lock.md)) guards against careless code and markup injected into the head, and goes no
further.

YPY does not decide who is allowed to do what. A button missing from the HTML hides an action from a reader but not
from someone who sends the request by hand, so every handler checks the permission itself.

## Where to go next

| You want | Read |
|---|---|
| a list to go through before you ship | [Security checklist](../guides/security-checklist.md) |
| every setting and the threat model | [Security reference](../reference/security.md) |
| what the last review found and fixed | [Security audit](../security/audit-2026-11-15.md) |
| the extensions | [`ypy-csp`](../extensions/csp.md), [`ypy-csrf`](../extensions/csrf.md), [`ypy-lock`](../extensions/lock.md), [`ypy-expr`](../extensions/expr.md) |

To report a vulnerability, follow the `SECURITY.md` file of the repository.
