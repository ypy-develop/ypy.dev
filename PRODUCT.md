# PRODUCT: ypy.dev documentation

## What it is

The documentation of YPY, a hypermedia framework by HYZIS (MIT). Readers are developers who build server-rendered
applications and want one script tag instead of a front-end build. The site is in English; Portuguese pages mirror it
under `content/ypy/pt-BR/`.

## Voice

Plain and direct, written by someone who builds software and respects the reader. The reason comes before the
mechanism when it helps. One idea per sentence, concrete examples, the real name of every attribute, header and event.

## Evidence on hand (the only numbers the copy may use)

- Version 0.1.1.
- Size budgets enforced by `go run ./cmd/ypy build --sizes` in `ypy`: core 18,432 B min+gzip (measured 18,410 B),
  each extension 6,144 B (largest measured: `ypy-signals` 6,009 B).
- Everything else about size or speed must be measured and cited from `ypy` (for example `performance/baseline.md`).
- No customers, testimonials, benchmarks against other tools or guarantees exist. None may be written.

## Hard constraints

- Never name the library YPY came from, nor its old attribute or header prefix. Write "the older library".
- Technical facts (attribute, event and header names, example code, CDN addresses, version) change only after checking
  the source in `ypy/src` or the Go SDK.
- Logos are used as drawn (see `AGENTS.md`).

## Learned constraints

- The author asked for the Arandu interface skills (content, copywriting, humanizer) on the docs text (2026-11-15).
