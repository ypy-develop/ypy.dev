---
order: 60
title: "Streaming a model answer"
description: "Write the answer of a language model into the page as it is generated, with safe Markdown."
---

# Streaming a model answer into the page

A language model writes a few bytes per token, tens of tokens per second, and the answer is Markdown that
should look right while it grows. YPY does this with the streaming protocol you already have: no new event, a
recipe of `ypy-patch-elements`, a helper in the Go SDK and the `ypy-model` extension.

The server holds the provider connection (the key never reaches the browser), batches the tokens, renders and
sanitizes the Markdown, and sends HTML. The page needs no Markdown parser.

## The shape of one answer

The server patches an answer container into the page once:

```html
<article id="a42" ypy-untrusted ypy-follow aria-live="off" aria-busy="true">
  <div id="a42-done"></div>   <!-- closed blocks: append only -->
  <div id="a42-tail"></div>   <!-- the one open block: replaced at each flush -->
</article>
```

While the model writes, each flush sends, in one write, an `append` to `#a42-done` with the blocks that just
closed and an `inner` on `#a42-tail` with the open block. The cost of a flush follows the open block, not the
answer. At the end one `outer` patch of `#a42` carries the whole final HTML, without `aria-busy` and with
`data-ypy-answer="done"` (`stopped` or `failed` otherwise). That last event repeats the state in full, so a client that
missed events or reconnected ends correct.

Markdown does not jump: a block is stable once it is closed. An unfinished `**`, a link without its `)` or a
table without its separator row shows as plain text until it completes, and an open code fence is already a
`pre`. For every way to cut the stream into chunks, the final HTML equals the rendering of the whole text.

## The server, in Go

```go
func answer(w http.ResponseWriter, r *http.Request) {
	s, err := ypy.NewStream(w, r)
	if err != nil { http.Error(w, "streaming is not supported", 500); return }
	defer s.Close()

	ms, err := ypy.NewModelStream(s,
		ypy.WithModelTarget("#chat", ypy.ModeAppend),
		ypy.WithAnswerID("a42"),
		ypy.WithRenderer(ypymd.Renderer{}),
		ypy.WithAnnounce("Answer started", "Answer ready", "Answer stopped", "The answer failed"),
		ypy.WithFollow())
	if err != nil { return }
	for tok := range provider.Stream(ms.Context(), prompt) { // the provider call carries the context
		ms.WriteString(tok)
	}
	ms.Close() // Cancel() after a Stop, Fail() when the provider failed
}
```

- Tokens go into a buffer flushed every 50 ms (`WithFlushInterval`), when a block closes or when the buffer
  passes a size. A UTF-8 sequence is never split.
- A slow client gets only the newest tail (last wins); closed blocks are never skipped.
- `ypymd` is a small Markdown renderer without dependencies. It escapes raw HTML, drops images, keeps `http`,
  `https`, `mailto` and relative links, and emits no `ypy-*` attribute. Plug your own renderer through
  `WithRenderer` if you need more; it must follow the same allowlist.
- `WithPlainText()` is the low-cost mode for logs and code completion: one `append` of escaped text per flush.
- `WithMaxBytes` caps the answer. Past it, the answer ends as stopped with a text your application wrote.
- **Bound** (default): the end of the connection cancels the provider request, so an abandoned stream burns no
  tokens. **Detached**: your application keeps the generation and `ms.Text()` in its own registry, and a
  reconnect gets the current state as its snapshot.

## The page

```html
<div id="gen" ypy-stream.post="/ask" ypy-config='{"stream":{"frame":true}}'>
  <button ypy-post="/generations/42/stop" ypy-swap="none" ypy-stop="#gen">Stop</button>
  <section id="chat"></section>
</div>
```

Load `ypy.max.js` or `ext/ypy-stream.js` and `ext/ypy-model.js`.

- `stream.merge` (on by default): patches that wait to be applied and have the same mode (`inner`, `outer`,
  `append`) and selector are merged. The last `inner` or `outer` wins and `append` payloads join. A client that
  does not merge stays correct, only slower.
- `stream.frame` (off by default): apply events one animation frame apart, so the browser does one round of DOM
  work per frame however fast the tokens arrive.
- `ypy-stop="#gen"` sends `ypy:abort` to the stream element, which closes it. The open answer is closed in the
  page as stopped. The button's own request (`ypy-post`) tells the server, with your CSRF protection, that the stop
  was intended.

## The answer is untrusted content

What a model writes is shaped by its prompt, by documents it retrieved and by users: assume an attacker wrote it.

- The server renderer emits only `p br strong em del code pre ul ol li blockquote h1`..`h6` `hr a table thead tbody
  tr th td`, `href` on `a` and `class="language-x"` on `code`. Images are off: an image is a request that can
  carry data out in its address.
- `ypy-untrusted` is the second wall on the client. Whatever enters the container loses scripts, styles,
  frames, forms, `svg` and `math`, every `ypy-*` and `data-ypy-*` attribute, `on*` handlers, Alpine-style
  attributes, `target`, `javascript:` and `data:` addresses and media (`ypy-untrusted="images"` keeps images). An
  injected `ypy-post` is a complete attack without a single script, which is why this matters as much as
  removing `script`.
- Selectors, modes, URLs, events and signals are written only from application values, never from model text.
  Tool calls of the model produce events only after your application validated their arguments.
- Set caps for bytes and events per answer, limit concurrent generations per user, and do not log tokens by
  default. See the [security checklist](../guides/security-checklist.md).

## Accessibility

A live region that announces every change would read the answer token by token, so it does not.

- The container is `aria-live="off"` with `aria-busy="true"` while it is written. The managed live region says one
  short message at the start and one at the end (`ypy-announce-start`, `ypy-announce`, or
  `config.announce.messages.answerStart` and `answerReady`); a stop or a failure uses the assertive form
  (`ypy-announce-stop`, `ypy-announce-error`, `answerStopped`, `answerFailed`).
- `config.announce.stream = "blocks"` is opt-in: it says the text of each closed block, shortened, at most once
  every three seconds.
- Focus never moves. `ypy-follow` scrolls with the answer only when the reader was already at the bottom, and
  never when `prefers-reduced-motion` is set. No view transition is used for token patches.
- The typing cursor is decorative CSS on `[aria-busy="true"]`. The Stop button comes before the answer in the
  DOM and is a real `button`. Test with a screen reader on each platform you support.

The protocol side is in section 5.8 of the SDK's `spec/PROTOCOL.md`, and the extension options in the
[streaming reference](../extensions/stream.md).
