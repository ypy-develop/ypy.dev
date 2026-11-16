---
order: 130
title: File upload
description: Send a file with a multipart form and fill a progress bar while it uploads, using the ypy-upload extension and an ordinary Go handler.
---

# File upload

A form with a file field, a progress bar that fills while the file is sent, and a message from the server when it
arrives. The progress comes from the browser; the server only sees an ordinary multipart request.

## The page

The progress needs the `ypy-upload` extension. Without it the upload still works and the bar stays empty.

```html
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ext/ypy-upload.min.js"></script>

<form ypy-post="/files"
      ypy-encoding="multipart/form-data"
      ypy-upload-progress="#bar"
      ypy-target="#result">
  <label>Document <input type="file" name="doc" required></label>
  <button>Send</button>
  <progress id="bar" max="100" value="0">0%</progress>
</form>
<div id="result" aria-live="polite"></div>
```

## The server (Go)

```go
package files

import (
	"crypto/rand"
	"encoding/hex"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"

	ypy "github.com/ypy-develop/go"
)

const maxUpload = 10 << 20 // 10 MiB

type Handler struct{ Dir string }

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		ypy.ValidationError(w, `<p role="alert">The file is larger than 10 MB.</p>`)
		return
	}
	file, head, err := r.FormFile("doc")
	if err != nil {
		ypy.ValidationError(w, `<p role="alert">Choose a file.</p>`)
		return
	}
	defer file.Close()

	// The name on disk is ours. The name the client sent is only shown back.
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(filepath.Join(h.Dir, hex.EncodeToString(raw[:])))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ypy.Fragment(w, `<p>Received `+html.EscapeString(head.Filename)+`.</p>`)
}

func Routes(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /files", h.Upload)
	return ypy.Middleware(mux)
}
```

## How it works

- `ypy-encoding="multipart/form-data"` sends the form as multipart, which a file field needs. A form with
  `enctype="multipart/form-data"` does the same.
- `ypy-upload-progress="#bar"` is the switch. A request whose body holds a file that is not empty and whose element has
  this attribute is sent with `XMLHttpRequest`, which reports progress. `fetch` cannot.
- The selector names the element that shows the percentage. A `<progress>` gets `value` and `max="100"`; any other
  element gets the CSS variable `--ypy-upload`. The `ypy:upload:progress` event carries `loaded`, `total` and
  `percent` for your own code.
- The answer goes into `#result` like any other. A `422` from `ypy.ValidationError` goes into the form, as in
  [forms](../guides/forms.md).
- The server limits the size with `http.MaxBytesReader` before it parses anything. The browser's progress is ahead
  of what the server has received.

## Variations

- Reset the bar between uploads: the values stay at 100 after a transfer. Listen to `ypy:finally:request` and set
  `value` back to 0 in a script file.
- A bar that is not a `<progress>`: give an element `role="progressbar"` and a width of `var(--ypy-upload, 0%)`
  in your CSS. See the Tailwind example in [`ypy-upload`](../extensions/upload.md).
- Keep the file after an error in another field: see [preserve file inputs](preserve-file-inputs.md).

## See also

- [`ypy-upload`](../extensions/upload.md)
- [Forms](../guides/forms.md)
