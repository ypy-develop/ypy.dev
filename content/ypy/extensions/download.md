---
order: 120
title: "File downloads: the ypy-download extension"
description: "Reads a response as a stream, reports progress events for a progress bar and hands the finished file to the browser as a download."
---

# File downloads: the `ypy-download` extension

A button that asks for a PDF should save the PDF, not swap bytes into the page. `ypy-download` reads the response as a
stream, tells the page how far it got, and hands the finished file to the browser as a download.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-download.min.js"></script>
```

The extension is in `ypy.max`. If you set `ypy.config.extensions`, list `ypy-download` in it. It is listed with the other
official extensions in [Extensions](index.md).

## When it acts

The extension takes over a response in any of these cases:

| Case | How |
|---|---|
| The element asks for it | `ypy-swap="download"` |
| The server marks the file | The response has `Content-Disposition: attachment` |
| The backend cannot send the file as the YPY response | The response has the header `YPY-Download` with a URL, and the extension fetches that URL as the file |

The file name comes from `Content-Disposition` (`filename*` first, then `filename`), or from the last part of the URL.

## Headers

| Header | Direction | Meaning |
|---|---|---|
| `YPY-Download` | response | A URL to fetch and save. The fetch uses `config.mode` and sends cookies only to the same origin. A failure fires `ypy:error`. |

## Events

All three fire on the element that made the request.

| Event | Detail | When |
|---|---|---|
| `ypy:download:start` | `total` (the `Content-Length` or `null`), `ctx` | The stream starts. |
| `ypy:download:progress` | `loaded`, `total`, `percent` (`null` when unknown), `ctx` | A chunk arrived. |
| `ypy:download:complete` | `filename`, `size`, `ctx` | The file was handed to the browser. |

## Example

```html
<button ypy-get="/reports/2026.pdf" ypy-swap="download">Download the report</button>
<progress id="report-progress" max="100" value="0" hidden></progress>

<script>
  document.addEventListener('ypy:download:start', () => {
    const bar = document.getElementById('report-progress'); bar.hidden = false; bar.value = 0;
  });
  document.addEventListener('ypy:download:progress', e => {
    if (e.detail.percent != null) document.getElementById('report-progress').value = e.detail.percent;
  });
  document.addEventListener('ypy:download:complete', () => {
    document.getElementById('report-progress').hidden = true;
  });
</script>
```

```go
func report(w http.ResponseWriter, r *http.Request) {
	data, _ := os.ReadFile("reports/2026.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Disposition", `attachment; filename="report-2026.pdf"`)
	w.Write(data)
}
```

With `Content-Disposition: attachment` set, `ypy-swap="download"` on the button is not required. Send `Content-Length`
so `percent` has a value.

## Notes

- The file is collected in memory before it is saved. For very large files use a plain link with the `download`
  attribute instead.
- A download has no fixed length, so the request timeout is dropped and the request stays open until the file is saved.
- An inline `<script>` as above needs a nonce on a page with a strict CSP; put it in a file instead.

## See also

[Upload progress](upload.md), [Events](../reference/events/list.md), [Security](../reference/security.md).
