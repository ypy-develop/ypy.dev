---
title: "YPY Testing Guide"
description: "Everything runs through the ypy tool, written in Go. No Node.js is needed."
order: 30
---
# YPY Testing Guide

Everything runs through the `ypy` tool, written in Go. No Node.js is needed.

## Prerequisites

- Go (the version in `go.mod`).
- Google Chrome installed (the runner drives it through the DevTools protocol).
- Optional: Safari with "Allow Remote Automation" turned on and `safaridriver --enable` run once,
  and Firefox with `geckodriver`, to run the suites in those browsers too.

## Running the tests

```bash
go run ./cmd/ypy test                         # main suite, Chrome
go run ./cmd/ypy test --suite migration --strict   # `ypy migrate` matrix: converted example apps pass the same scenarios
go run ./cmd/ypy test --browsers chrome,safari,firefox
go run ./cmd/ypy test test/tests/attributes/ypy-error.js   # one file
go run ./cmd/ypy test --grep "morph"          # tests whose title matches
go run ./cmd/ypy test --headed                # watch the browser
```

`go run ./cmd/ypy test --serve` serves the repository so the pages in `test/manual/` can be opened
by hand.

## Before pushing

```bash
go run ./cmd/ypy ci
```

It runs every gate the GitHub workflow runs: names, licenses, SPDX headers, gofmt, go vet, the Go
tests, bundle sizes, the generated upgrade guide, the browser suites and the migration matrix.

## Writing tests

- Test files live under `test/tests/` and use Mocha with Chai (both embedded in the runner).
- `test/lib/helpers.js` has `setupTest(test)`, `cleanupTest`, `mockResponse` and the other helpers.
- A new behavior needs a test that shows the behavior and a test that shows what stays the same.
