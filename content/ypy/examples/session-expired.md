---
order: 220
title: Session expired
description: Send the user to the sign-in page with a YPY-Redirect header when a background request finds the session gone, or keep the page and show a 401 message.
---

# Session expired

A page stays open for an hour, and when the user finally clicks something the session is gone. A sign-in form swapped
into the middle of the page would be confusing, so the server answers the YPY request with a redirect header and the
browser loads the sign-in page in full, with a link back to where the user was.

## The page

Nothing special. Every request on the page is covered:

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>

<button ypy-get="/reports/42" ypy-target="#report" ypy-swap="innerHTML">Load report</button>
<div id="report"></div>
```

## The server (Go)

One middleware in front of the protected routes:

```go
package auth

import (
	"net/http"
	"net/url"

	ypy "github.com/ypy-develop/go"
)

func RequireSession(sessions Sessions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sessions.Valid(r) {
				next.ServeHTTP(w, r)
				return
			}
			// Where to come back to: the page the user is on, not the fragment URL.
			back := r.URL.RequestURI()
			if q := ypy.FromRequest(r); q.IsYPY && q.CurrentURL != "" {
				if u, err := url.Parse(q.CurrentURL); err == nil {
					back = u.RequestURI()
				}
			}
			ypy.Redirect(w, r, "/login?next="+url.QueryEscape(back))
		})
	}
}
```

`ypy.Redirect` does two things. For a YPY request it sends `204` with the `YPY-Redirect` header and no body. For any
other request it sends `303 See Other`. A body would be swapped in before the navigation, so there is none.

The sign-in handler must accept only a local path in `next` (it starts with `/` and not `//`) before it redirects there.

## How it works

- The client reads `YPY-Redirect` before it looks at the status code, so the header also works on a `401`. The browser
  navigates with a full page load, to the same origin only. A different origin fires `ypy:security:blocked` unless
  `config.mode` is `cors`.
- `YPY-Current-URL` is the address of the page that made the request. `ypy.FromRequest(r).CurrentURL` reads it.

## Variation: keep the page and say so

Sometimes a redirect loses work, for instance a long form. Answer `401` with a short fragment and let the page put it
in a fixed place. Put the rule on a container, once:

```html
<body ypy-status:401:inherited="target:#session-notice swap:innerHTML">
  <div id="session-notice" role="alert"></div>
  ...
</body>
```

```go
w.WriteHeader(http.StatusUnauthorized)
io.WriteString(w, `<p>Your session ended. <a href="/login?next=`+url.QueryEscape(back)+`">Sign in again</a>.</p>`)
```

`ypy-status:<code>` accepts `swap:`, `target:`, `select:`, `push:`, `replace:` and `transition:`. A range works too:
`ypy-status:4xx`.

## See also

- [Response headers](../reference/headers.md)
- [Security](../reference/security.md)
- [Dialogs with layers](dialogs-layers.md)
- [Error destinations: `ypy-error-*`](../reference/attributes/ypy-error.md)
