---
order: 220
title: Session expired
description: Send the user to the sign-in page with a YPY-Redirect header when a background request finds the session gone, or keep the page and show a 401 message.
---

# Session expired

When a request finds the session gone, the server answers with a redirect header and the browser loads the sign-in page in full, with a link back to where the user was.

## The page

A page stays open for an hour, and when the user finally clicks something the session is gone. A sign-in form swapped
into the middle of the page would be confusing, so the page needs nothing special. Every request on it is covered:

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<button ypy-get="/reports/42" ypy-target="#report" ypy-swap="innerHTML">Load report</button>
<div id="report"></div>
```

## What the server answers

For every protected route, a request with no valid session gets an answer with no body and one header:

```text
GET /reports/42
YPY-Request: true
YPY-Current-URL: https://app.example.com/reports?month=2026-10

HTTP/1.1 204 No Content
YPY-Redirect: /login?next=%2Freports%3Fmonth%3D2026-10
Vary: YPY-Request
```

| Request | Answer |
|---|---|
| YPY request, session valid | the fragment, `200` |
| YPY request, session gone | `204` with `YPY-Redirect: /login?next=<current page>` and no body |
| Any other request, session gone | `303 See Other` with `Location: /login?next=<requested page>` |

- Where to come back to is the page the user is on, not the fragment address. A YPY request sends it in the
  `YPY-Current-URL` header, so `next` carries that one.
- A body would be swapped in before the navigation, so there is none.
- The sign-in page accepts only a local path in `next` (it starts with `/` and not `//`) before it redirects there.

## How it works

- The client reads `YPY-Redirect` before it looks at the status code, so the header also works on a `401`. The browser
  navigates with a full page load, to the same origin only. A different origin fires `ypy:security:blocked` unless
  `config.mode` is `cors`.
- `YPY-Current-URL` is the address of the page that made the request.
- The `Vary: YPY-Request` header keeps a cache from serving the redirect to a plain request, or the reverse.

## Variation: keep the page and say so

Sometimes a redirect loses work, for instance a long form. Answer `401` with a short fragment and let the page put it
in a fixed place. Put the rule on a container, once:

```html
<body ypy-status:401:inherited="target:#session-notice swap:innerHTML">
  <div id="session-notice" role="alert"></div>
  ...
</body>
```

The server answers:

```text
HTTP/1.1 401 Unauthorized
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
```

```html
<p>Your session ended. <a href="/login?next=%2Freports%3Fmonth%3D2026-10">Sign in again</a>.</p>
```

`ypy-status:<code>` accepts `swap:`, `target:`, `select:`, `push:`, `replace:` and `transition:`. A range works too:
`ypy-status:4xx`.

## See also

- [Response headers](../reference/headers.md)
- [Security](../reference/security.md)
- [Dialogs with layers](dialogs-layers.md)
- [Error destinations: `ypy-error-*`](../reference/attributes/ypy-error.md)
