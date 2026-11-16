---
order: 30
title: "Signals"
description: "Read the signals the page sends, patch them from the server, and sign the ones the page must not change."
---

# Signals

Signals are the small pieces of state the page keeps with the [`ypy-signals`](../extensions/signals.md) extension. The
page sends them with a request, and the server can change them in the answer. The SDK reads and writes both ways.

## Read

```go
type Search struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
}

s, err := ypy.ReadSignals[Search](r) // ypy.ErrNoSignals when the request has none
```

The client sends them in the `ypy-signals` query parameter on a GET, as a JSON body on POST, PUT and PATCH, or in the
`ypy-signals` form field. A JSON body above `ypy.MaxSignalsBytes` (1 MiB) is refused with `ypy.ErrSignalsTooLarge`.
`ypy.RawSignals(r)` returns the bytes if you want to decode them yourself.

## Patch

```go
ypy.PatchSignals(w, map[string]any{"page": 1, "draft": nil}) // JSON Merge Patch: null deletes a key
```

The answer is JSON, and the client merges it into its signals. On a stream use `Stream.PatchSignals` or
`Broadcaster.PatchSignals`, with `ypy.WithOnlyIfMissing()` to set a value only when the page has none.

## Signed read-only signals

A signal declared with `ypy-signals.server` cannot be written by the page. The browser still sends every signal back,
so a user can send any value. To trust one, sign it when you send it and verify it when it returns. The signature is an
HMAC-SHA256 over the name and the canonical JSON of the value, compared in constant time.

```go
signer, _ := ypy.NewSigner(newKey, oldKey) // sign with newKey; accept both while old pages live
signer = signer.Bind(sessionID)            // optional: a value signed for one session fails in another

// Send: the values and their signatures.
decl, _ := signer.Declaration(map[string]any{"plan": "free", "seats": 3}) // goes in ypy-signals.server="..."
ypy.PatchSignedSignals(w, signer, map[string]any{"plan": "pro"})

// Receive: the check runs before anything is decoded.
got, err := ypy.ReadSignedSignals[Account](r, signer, "plan", "seats")
if errors.Is(err, ypy.ErrSignatureMissing) || errors.Is(err, ypy.ErrSignatureInvalid) {
	http.Error(w, "forbidden", http.StatusForbidden)
	return
}
```

```html
<div ypy-signals.server="{{ .Declaration }}">...</div>
```

The signatures travel in one reserved signal, `ypySig`, so the client needs no change. Name the signals you rely on: a
signal that is absent or unsigned fails. `errors.As` with a `*ypy.SignatureError` gives the name of the signal that
failed. `NewSigner` takes the current key first and any previous keys after it, which is how you rotate a key without
breaking open pages.

In tests, `ypytest.Signals(v)` sends signals and `ypytest.SignedSignals(signer, v)` sends signed ones;
`res.AssertSignals(t, before, want)` applies a patch to a document and compares the result. See [`ypytest`](ypytest.md).
