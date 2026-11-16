---
order: 130
title: File upload
description: Send a file with a multipart form and fill a progress bar while it uploads, using the ypy-upload extension and a server that reads an ordinary form.
---

# File upload

A form with a file field, a progress bar that fills while the file is sent, and a message from the server when it arrives.

## The page

The progress comes from the browser and needs the `ypy-upload` extension. Without it the upload still works and the bar
stays empty.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-upload.min.js"></script>

<form ypy-post="/files"
      ypy-encoding="multipart/form-data"
      ypy-upload-progress="#bar"
      ypy-target="#result"
      ypy-error-target="#result">
  <label>Document <input type="file" name="doc" required></label>
  <button>Send</button>
  <progress id="bar" max="100" value="0">0%</progress>
</form>
<div id="result" aria-live="polite"></div>
```

## What the server answers

The server sees an ordinary `multipart/form-data` request with one part named `doc`. It checks the size before it
parses anything, keeps the file under a name of its own (the name the client sent is only shown back) and answers with a
fragment:

```text
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
```

```html
<p>Received quarterly-report-2026-q3.pdf.</p>
```

When something is wrong with the file, it answers `422` with the message, and `ypy-error-target` puts it in `#result`:

```text
HTTP/1.1 422 Unprocessable Content
Content-Type: text/html; charset=utf-8
Vary: YPY-Request
```

```html
<p role="alert">The file is larger than 10 MB.</p>
```

| Case | Status | Body |
|---|---|---|
| File stored | `200` | `<p>Received ...</p>` |
| No file chosen | `422` | `<p role="alert">Choose a file.</p>` |
| File over 10 MB | `422` | `<p role="alert">The file is larger than 10 MB.</p>` |
| Storage failed | `500` | an error message |

## How it works

- `ypy-encoding="multipart/form-data"` sends the form as multipart, which a file field needs. A form with
  `enctype="multipart/form-data"` does the same.
- `ypy-upload-progress="#bar"` is the switch. A request whose body holds a file that is not empty and whose element has
  this attribute is sent with `XMLHttpRequest`, which reports progress. `fetch` cannot.
- The selector names the element that shows the percentage. A `<progress>` gets `value` and `max="100"`; any other
  element gets the CSS variable `--ypy-upload` and the attribute `data-ypy-upload`. The `ypy:upload:progress` event
  carries `loaded`, `total` and `percent` for your own code.
- The answer goes into `#result` like any other. Without `ypy-error-target`, a `422` goes into the form, as in
  [forms](../index.md#validation).
- The server should limit the size before it parses the body. The browser's progress is ahead of what the server has
  received.

## Variations

- Reset the bar between uploads: the value stays at 100 after a transfer. Listen to `ypy:finally:request` in a script
  file and set `value` back to 0.
- A bar that is not a `<progress>`: give an element `role="progressbar"` and a width of `var(--ypy-upload, 0%)`
  in your stylesheet. It also gets `aria-valuenow`. See [`ypy-upload`](../extensions/upload.md).
- Keep the file after an error in another field: see [preserve file inputs](preserve-file-inputs.md).

## See also

- [`ypy-upload`](../extensions/upload.md)
- [Forms](../index.md#validation)
