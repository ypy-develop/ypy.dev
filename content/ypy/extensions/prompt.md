---
order: 330
title: "Ask before sending: the ypy-prompt extension"
description: "Asks the visitor one question before the request, sends the answer in the YPY-Prompt header, and sends nothing if they cancel."
---

# Ask before sending: the `ypy-prompt` extension

`ypy-prompt` shows a question before the request goes out. The answer travels in the `YPY-Prompt` request header. If
the visitor cancels, no request is made. It is useful for a rename, a reason for a refusal or a number of items, where
one small question is enough and a form would be too much.

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-prompt.js"></script>
```

The extension is not part of `ypy.max`.

## Use it

```html
<button ypy-post="/projects/7/rename" ypy-target="#project-7" ypy-swap="outerHTML"
        ypy-prompt="New name for this project?">Rename</button>
```

The browser's own `prompt()` dialog asks the question. `ypy-prompt` is inherited, so you can set it on a container.
An empty answer is sent (an empty header); `Cancel` stops the request.

For your own dialog, define `window.ypyPrompt`. It receives the question and must return the answer as a string, or
`null` when the visitor cancels. It has to answer at once (a synchronous function). For a dialog that waits for the visitor, use
[layers](layers.md) and send a normal form.

## Attributes

| Attribute | Value | Meaning |
|---|---|---|
| `ypy-prompt` | text | the question; inherited |

## Events

| Event | Cancelable | `detail` | Fires |
|---|---|---|---|
| `ypy:prompt` | yes | `prompt` (the answer), `target` | on the element after the visitor answered; cancel it to stop the request |

## Headers

| Header | Direction | Value |
|---|---|---|
| `YPY-Prompt` | request | the answer, URI-encoded |

## Server

The Go SDK decodes the header for you: `Request.Prompt` holds the answer as text.

```go
import ypy "github.com/ypy-develop/go"

func (h *Projects) Rename(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(ypy.FromRequest(r).Prompt)
	if name == "" {
		ypy.ValidationError(w, `<p id="project-7" class="error">Name is required</p>`)
		return
	}
	// save the name, then answer with the new fragment
	ypy.Fragment(w, `<h2 id="project-7">`+html.EscapeString(name)+`</h2>`)
}
```

## Notes

- The answer comes from the visitor: validate it on the server as any other input, and escape it before it goes into
  HTML.
- To ask for a yes or no, use `ypy-confirm` from the core; see [the confirm example](../examples/confirm-and-prompt.md).

## See also

- [Confirm and prompt (example)](../examples/confirm-and-prompt.md)
- [Headers](../reference/headers.md)
