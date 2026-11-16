---
order: 40
title: "CSRF"
description: "Protect every unsafe request with a token bound to the session, and keep the page's token current."
---

# CSRF

`ypy.NewCSRF` protects every unsafe request (anything but GET, HEAD, OPTIONS and TRACE) with a token bound to the
session. The token is an HMAC-SHA256 of the session id, so the server stores nothing. The Go side and the
[`ypy-csrf`](../extensions/csrf.md) extension of the page are two halves of one contract.

1. The server writes the token of the session in the page head: `<meta name="ypy-csrf" content="TOKEN">`.
2. The extension sends it in `X-CSRF-Token` on every `POST`, `PUT`, `PATCH` and `DELETE` to the page's own origin, even
   when the element inherits no header. It never sends it to another origin. A plain form with no client sends the
   hidden field `csrf_token`.
3. `csrf.Middleware` reads the header, or else the field, compares it with the token of the session, and checks `Origin`.
4. When the token must change, the server answers with `YPY-CSRF: <new token>` and the extension replaces the content
   of the meta tag. The next request sends the new one.

## Wire it

```go
csrf, err := ypy.NewCSRF(key, // 32 random bytes or more
	ypy.CSRFSession(func(r *http.Request) string { return sessions.ID(r) }), // optional
	ypy.CSRFPreviousKeys(oldKey))                                              // optional, for rotation
if err != nil {
	log.Fatal(err)
}
handler := ypy.Middleware(csrf.Middleware(mux))
```

```html
<head>{{ .CSRFMeta }}</head>                 <!-- csrf.Meta(r) -->
<form method="post">{{ .CSRFField }}</form> <!-- csrf.Field(r), for a form that must work with no client -->
```

Put `csrf.Meta(r)` in the page data, and `csrf.Field(r)` in forms that must work without the client. A visitor with no
session gets a random one in a cookie (`HttpOnly`, `SameSite=Lax`, with the `__Host-` prefix over TLS or with
`ypy.CSRFSecure()`), so a login form is protected too. `CSRFSession` gives the application's own session id, which wins
when it is not empty. The middleware also adds `Vary: Cookie`.

| Option | Use |
|---|---|
| `CSRFSession(fn)` | take the session id from your own session store |
| `CSRFPreviousKeys(keys...)` | keep accepting tokens made with older keys while you rotate |
| `CSRFTrustedOrigins(origins...)` | list the public origin behind a proxy that rewrites `Host` |
| `CSRFHeaderName(name)` | match another header name set in the meta tag |
| `CSRFSecure()` | mark the cookie `Secure` with the `__Host-` prefix when a proxy ends TLS before the server |
| `CSRFOnFailure(fn)` | replace the 403 answer |

## What the middleware answers

| Error | Cause | Answer |
|---|---|---|
| `ErrCSRFOrigin` | `Origin` is another origin that is not trusted | 403 |
| `ErrCSRFMissing` | no token in the header or the field | 403; with the extension loaded, the client also fires `ypy:csrf:missing` |
| `ErrCSRFInvalid` | the token is from another session | 403 with `YPY-CSRF` holding the current token, so the client retries |
| valid, made with a previous key | key rotation | the request goes on and the response carries `YPY-CSRF` |

`csrf.Check(r)` runs the same check and returns the error, for one handler outside the middleware.

## When the session changes

After a sign-in the page still holds the token of the old session. Renew it in the handler that changes the session:

```go
mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
	id := sessions.Start(w, r.PostFormValue("email"))
	csrf.RenewFor(w, id) // YPY-CSRF: the client updates its meta tag
	ypy.Redirect(w, r, "/")
})
```

`csrf.Renew(w, r)` does the same for the session of the request. A stream that opens with GET needs no token and must not
change state. This protection is for cookie sessions, not for an API that authenticates with a header token.

## Test it

```go
csrf, _ := ypy.NewCSRF(key, ypy.CSRFSession(func(r *http.Request) string { return "user-7" }))
h := csrf.Middleware(mux)
form := ypytest.Form(url.Values{"title": {"Buy milk"}})

ypytest.Do(h, ypytest.NewRequest("POST", "/items", form)).AssertStatus(t, http.StatusForbidden)

token := csrf.TokenFor("user-7")
ypytest.Do(h, ypytest.NewRequest("POST", "/items", form, ypytest.CSRFToken(token))).
	AssertStatus(t, http.StatusOK)
```
