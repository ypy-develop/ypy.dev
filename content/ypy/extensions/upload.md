---
title: "Upload progress: the ypy-upload extension"
description: "Progress for file uploads: sends the request with XMLHttpRequest and fires ypy:upload:progress on the element while the file goes out."
order: 460
---
# Upload progress: the `ypy-upload` extension

`fetch()` cannot report how much of a request body has been sent, so it cannot drive a progress bar for a file
upload. The `ypy-upload` extension sends the requests that ask for progress with `XMLHttpRequest`, which can, and
fires `ypy:upload:progress` on the element while the file goes out. The core sees the same result as with `fetch`
(status, headers, text), so swaps, events, queues, timeouts and `ypy-retry` work as usual.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-upload.js"></script>   <!-- provided by the official extension -->

<form ypy-post="/files" ypy-encoding="multipart/form-data" ypy-target="#result" ypy-upload-progress="#bar">
  <input type="file" name="doc">
  <button>Send</button>
</form>
<progress id="bar" max="100" value="0"></progress>
<div id="result"></div>
```

If you set `ypy.config.extensions`, list `ypy-upload` in it, or the extension is refused. It is listed with the other
official extensions in [Extensions](index.md); it is part of `ypy.max`. Without it, `ypy-upload-progress` does
nothing and breaks nothing: the upload goes through `fetch`.

## When the transport changes

The rule is one and simple. A request is sent with `XMLHttpRequest` when both are true:

1. the body is a `FormData` that holds at least one `File` that is not empty (use `ypy-encoding="multipart/form-data"`,
   or a form with that `enctype`), and
2. the element has the `ypy-upload-progress` attribute (not `false`). The attribute is not inherited unless you write
   `ypy-upload-progress:inherited` on a container.

Everything else, including a file upload without the attribute, a form without a file and a `GET`, keeps using `fetch`.
Registering a listener for `ypy:upload:progress` does not turn it on: the attribute is the switch.

## The attribute

| Value | Meaning |
|---|---|
| none | send with `XMLHttpRequest` and fire the event |
| a CSS selector | the same, and show the percentage in the first element that matches |
| `false` | off for this element; the request uses `fetch` |

The element named by the selector is updated through the DOM API only (no `style` attribute, no inline script), so a
strict CSP is fine:

- `--ypy-upload` is set to the percentage, for example `42%`, and `data-ypy-upload` to the number (`42`);
- a `<progress>` also gets `max="100"` and `value`;
- an element with `role="progressbar"` also gets `aria-valuenow`.

The values stay at the end of the transfer (`100`); reset or hide the bar yourself on `ypy:finally:request`. A selector
that matches nothing or is not valid is ignored; the event still fires.

## The event

`ypy:upload:progress` fires on the element that made the request and bubbles. It cannot be cancelled. The detail is:

| Field | Type | Meaning |
|---|---|---|
| `loaded` | number | bytes sent so far |
| `total` | number | bytes in the body, or `0` when the browser does not know |
| `percent` | number | a whole number from `0` to `100`; `0` while `total` is unknown |
| `ctx` | `YpyRequestCtx` | the request, so a listener can call `ctx.request.abort()` |

The first event reports `0`, and the last one reports `100` when the body was sent completely. A progress event does
not mean the server answered: the request ends with the usual `ypy:after:request` and `ypy:finally:request`.

## Abort, timeout, errors

- `ctx.request.abort()` (and the request timeout, which aborts with the reason `timeout`) aborts the transfer. An abort
  you ask for is silent: no `ypy:error` and nothing is swapped. A request aborted before it starts is never sent. A
  timeout is reported with `ctx.errorKind` set to `"timeout"`.
- A transfer that fails (no connection, server gone) rejects with a `TypeError` like `fetch`, so `ctx.errorKind` is
  `network` or `offline` and `ypy-retry` can repeat it.
- An error status (`4xx`, `5xx`) is a normal response and goes through `ypy:response:error`.

## Differences from `fetch`

These come from `XMLHttpRequest` and are the reason the extension changes the transport only when you ask:

- `credentials: "include"` sets `withCredentials`; `"omit"` cannot hide the cookies of the same origin.
- `mode`, `cache` and `redirect` of the request are not used; redirects are followed.
- The body of the answer is read as a whole before the core gets it (no streaming of the answer).
- The event is a progress of the bytes handed to the network, which can be ahead of what the server has received.

## A progress bar with Tailwind CSS

The bar is a `div` that takes its width from the variable the extension sets. The initial `0%` is the fallback of
`var()`, so nothing else is needed before the first event.

```html
<form ypy-post="/files" ypy-encoding="multipart/form-data" ypy-target="#result" ypy-upload-progress="#bar"
      class="space-y-3">
  <input type="file" name="doc" class="block w-full text-sm">
  <button class="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50">Send</button>

  <div class="h-2 w-full overflow-hidden rounded bg-gray-200" aria-hidden="true">
    <div id="bar" role="progressbar" aria-label="Upload" aria-valuemin="0" aria-valuemax="100" aria-valuenow="0"
         class="h-full w-[var(--ypy-upload,0%)] bg-blue-600 transition-[width] duration-150"></div>
  </div>
</form>
<div id="result"></div>
```

To show the number as text, listen to the event:

```js
document.addEventListener('ypy:upload:progress', e => {
  document.querySelector('#percent').textContent = e.detail.percent + '%';
});
```

## The server

The server receives an ordinary `multipart/form-data` request: the progress is entirely in the browser. Limit the size
of what you accept, then answer a fragment. With the Go SDK (net/http):

```go
package files

import (
	"net/http"

	ypy "github.com/ypy-develop/go"
)

const maxUpload = 10 << 20 // 10 MiB

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		_ = ypy.ValidationError(w, `<p class="text-red-700">File too large (10 MB at most).</p>`)
		return
	}
	file, _, err := r.FormFile("doc")
	if err != nil {
		_ = ypy.ValidationError(w, `<p class="text-red-700">Choose a file.</p>`)
		return
	}
	defer file.Close()
	// store the file under a name you generate, never under the name the client sent
	_ = ypy.Fragment(w, `<p class="text-emerald-700">Uploaded.</p>`)
}
```

The fragment is swapped into `ypy-target` as for any request. `ypy.ValidationError` answers `422`, the status a form
uses for a rejected input (see [forms](../guides/forms.md)).

## Pitfalls

- The attribute is the switch. A file upload without `ypy-upload-progress` works, with no progress.
- An empty file does not count. A body whose only file has size `0` goes through `fetch`.
- `100` is not "done". The last event means the bytes were handed to the network; wait for `ypy:after:request` or
  `ypy:finally:request` for the answer.
- Reset the bar between uploads: the values stay at `100` after a transfer.
- A proxy or the server may refuse a large body with `413` before the whole file went out; the bar can then stop short.
