---
title: "CSRF token on every request: the ypy-csrf extension"
description: "Sends the CSRF token from one meta tag with every request except GET and HEAD to your own origin, so no element can fail with a silent 403."
order: 80
---
# CSRF token on every request: the `ypy-csrf` extension

A server that checks a CSRF token answers `403` to a request that arrives without it. When the token travels in a
header that an element *inherits*, one element that forgets the inheritance sends nothing, and the page fails with no
message on the client. `ypy-csrf` removes that failure: the token lives in one meta tag, and every request that is
not a `GET` or `HEAD` and goes to the page's own origin carries it, whatever the element did.

```html
<head>
  <meta name="ypy-csrf" content="{{ token }}">
  <script src="ypy.min.js"></script>
  <script src="ext/ypy-csrf.js"></script>
</head>

<button ypy-delete="/items/7">Delete</button>   <!-- sends X-CSRF-Token: <token>, nothing else to write -->
```

It is part of `ypy.max`; with
the plain core, load it with a `<script>` after the core.

## The meta tag

| Part | Meaning |
|---|---|
| `<meta name="ypy-csrf" content="TOKEN">` | the token. The tag is read at each request, so a swapped head or a script that sets `content` is always followed |
| `data-header="X-XSRF-Token"` | the request header; default `X-CSRF-Token` |

## What is sent, and where

- Sent: every request that is not a `GET` or `HEAD` (`POST`, `PUT`, `PATCH`, `DELETE`, `QUERY`...) to the page's own
  origin (a relative URL, or an absolute one with the same scheme, host and port). The header is set from the meta tag even when the element or an ancestor
  set a header of the same name (`ypy-headers`): a stale inherited token never wins over the current one.
- Not sent: `GET` and `HEAD` (they must not change anything, so they need no token).
- Never to another origin. A request to another origin carries no token, and a header of that name that an
  element inherited is removed from it, so the token cannot leak to a third party by inheritance.

## Renewal

When a same-origin response carries the header `YPY-CSRF`, its value replaces the `content` of the meta tag (the tag is
created in `<head>` when the page has none), and the next request sends the new token. It works on any status, a `403`
included, so a server can answer "your token is old" and hand over the new one at the same time. A `YPY-CSRF` header
on a response from another origin is ignored. The extension does not repeat the request that was refused: the user
(or your code) sends it again, now with the new token.

The server decides when to renew: after a sign-in (the session changed), or when it rotates its signing key. The
Go package does it with `csrf.Renew(w, r)` ([server side](../index.md#server-side)).

## A request without a token

An unsafe request to the page's own origin that leaves with no token (no meta tag, or an empty `content`) is the
silent failure this extension exists to end. It is still sent, because the server decides, but:

- the cancelable event `ypy:csrf:missing` fires on the source element, with `{ctx, header}`;
- unless the event is canceled, a `console.warn` names the method and the URL. It is shown once per URL.

```js
document.addEventListener('ypy:csrf:missing', evt => {
  report(evt.detail.ctx.request.method, evt.detail.ctx.request.action);
  evt.preventDefault(); // you handled it: no console warning
});
```

## The server

The server checks the header (or, for a plain form post, the field `csrf_token`) against the token of the session and
answers `403` when it is missing or wrong. Any server can follow this contract:

| | |
|---|---|
| request header | `X-CSRF-Token` (or the name in `data-header`) |
| response header | `YPY-CSRF`: the new token |

The Go package ships it as `ypy.NewCSRF`. The token is an HMAC of the session identifier, so it needs no storage and a token
of one session fails in another. `Middleware` lets `GET`, `HEAD`, `OPTIONS` and `TRACE` through and, for the other
methods, refuses with `403` a request from another origin (the `Origin` header, or `Sec-Fetch-Site: cross-site`) or
without a valid token. It also adds `Vary: Cookie`. When a token was made with a previous key (`CSRFPreviousKeys`), or is
refused as invalid, the response carries `YPY-CSRF` with the current token.

A visitor with no session yet (the login form) gets a cookie of the middleware (`HttpOnly`, `SameSite=Lax`, with the
`__Host-` prefix over TLS), and the token is bound to it. `csrf.Field(r)` writes a hidden `csrf_token` input for a form
that must work without the client. 

## Limits

- The token protects against cross-site requests made by a browser. It does not replace authentication, and it does
  not stop a script that already runs in your page (that is XSS; see [security](../index.md#security)).
- A request made with a `fetch` of your own, outside YPY, is not covered: send the header yourself, from the meta tag.
- A redirect to another origin that the browser follows by itself is outside the extension's reach, as for any custom
  header. Do not redirect an unsafe request off-site.
- Size: under 0.7 kB min+gzip (`ypy build --sizes`).

## Pitfalls

- No meta tag, no token. A layout that forgot the tag sends unsafe requests with no token. The extension fires
  `ypy:csrf:missing` and warns in the console, and the server answers `403`.
- A new session needs a new token. After a sign-in the old token is for the old session. Send `YPY-CSRF` in the
  response of the sign-in (`RenewFor`), or the next `POST` gets a `403`.
- A `403` is not retried by the extension. The `YPY-CSRF` header on the `403` updates the tag; the request has to be
  sent again by the user or your code.
- The token is for requests YPY sends. A `fetch` of your own or a plain `<form>` post has to add it (`csrf.Field(r)`).
