---
order: 220
title: "Model answers: the ypy-model extension"
description: "The page side of a streamed language-model answer: an untrusted container, screen reader announcements, and controls to follow and stop it."
---

# Model answers: the `ypy-model` extension

A language model writes text nobody on your team has read. When that text streams into the page as HTML, the page needs a
second wall behind the server's renderer, a way to tell a screen reader what is happening, and a stop button. `ypy-model`
is the client side of that. The server side, with the Go helper that batches tokens and renders Markdown, is in
[Streaming a model answer](../server/model-streaming.md).

## Load it

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-stream.min.js"></script>
<script src="ext/ypy-model.min.js"></script>
```

The extension is in `ypy.max`, and it works with the [`ypy-stream`](stream.md) client that receives the answer. If you
set `ypy.config.extensions`, list `ypy-model` in it. It is listed with the other official extensions in
[Extensions](index.md).

## Attributes

| Attribute | On | Meaning |
|---|---|---|
| `ypy-untrusted` | the answer container | Marks content a model writes. Whatever enters it loses scripts, styles, frames, forms, `svg`, `ypy-*` and `data-ypy-*` attributes, `on*` handlers, `javascript:` and `data:` addresses and media that loads from the network. The value `images` keeps images. |
| `ypy-follow` | the answer container | The page scrolls down as the answer grows, only if the reader was already at the bottom and `prefers-reduced-motion` is not set. |
| `ypy-stop` | a button | A selector of the elements that get `ypy:abort` when the button is clicked. Their stream closes, and the open answer is closed in the page as stopped. |
| `ypy-announce-start`, `ypy-announce-stop`, `ypy-announce-error` | the answer container | The text said when the answer starts, is stopped or fails. |
| `ypy-answer` | the answer container | The server writes it (as `data-ypy-answer`) in the last patch: `done`, `stopped` or `failed`. |

Links that stay in an untrusted container get `rel="noopener noreferrer nofollow"`. A container that arrives in a response
with `ypy-untrusted` is cleaned the same way.

## Screen readers

A container with `aria-busy="true"` is an answer being written. The live region says one short message when it starts
and one when `aria-busy` goes away: the text of `ypy-announce` (or `answerReady`) when it finished, `ypy-announce-stop`
when `data-ypy-answer` is `stopped`, `ypy-announce-error` when it is `failed`. The last two are assertive. Nothing is
said while the text grows, unless `announce.stream` is `blocks`, which reads the text of each closed block, shortened,
at most once every 3 seconds.

## Settings

| Key | Meaning |
|---|---|
| `announce.messages.answerStart` | Said when an answer starts. |
| `announce.messages.answerReady` | Said when an answer is finished. |
| `announce.messages.answerStopped` | Said, assertive, when it was stopped. |
| `announce.messages.answerFailed` | Said, assertive, when it failed. |
| `announce.stream` | `blocks` reads each closed block. Unset: only the start and the end are said. |

## Example

```html
<div id="chat" ypy-stream="/chat/events"></div>

<form ypy-post="/chat" ypy-swap="none">
  <input name="prompt" required>
  <button>Ask</button>
</form>
<button ypy-post="/chat/stop" ypy-swap="none" ypy-stop="#chat">Stop</button>
```

The server patches an `<article id="a42" ypy-untrusted ypy-follow aria-busy="true">` into `#chat` and fills it. The Stop
button sends the stop to the server with its own request, and the extension closes the stream in the page. The Go side:

```go
ms, err := ypy.NewModelStream(s,
	ypy.WithModelTarget("#chat", ypy.ModeAppend),
	ypy.WithAnswerID("a42"),
	ypy.WithFollow())
```

## Notes

- The wall is a second defence. The server must still render Markdown with an allowlist (the `ypymd` package does).
- A stream that ends with an error leaves the answer closed as `failed`, and one that is aborted leaves it `stopped`.

## See also

[Streaming a model answer](../server/model-streaming.md), [`ypy-stream`](stream.md),
[Accessibility](../guides/accessibility.md), [`ypy-announce`](../reference/attributes/ypy-announce.md).
