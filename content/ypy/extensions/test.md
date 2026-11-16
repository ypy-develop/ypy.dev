---
title: "Waiting for the page in browser tests: the ypy-test extension"
description: "Gives browser tests ypy.idle(), a promise that resolves when no request, swap or timer is pending, and data-ypy-busy on html while work runs."
order: 440
---
# Waiting for the page in browser tests: the `ypy-test` extension

A browser test clicks a button and then has to know when the page has finished. A fixed sleep is too short on a slow
machine and wastes time on a fast one. `ypy-test` gives the page a way to say "nothing is pending any more":
`ypy.idle()`, and a `data-ypy-busy` attribute on `<html>` while work is going on.

This page is about the hooks inside the page. For testing the server of a YPY app without a browser (`ypytest`), see
[Testing your app](../index.md#server-side).

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-test.js"></script>
```

It is not part of `ypy.max`. Load it in the pages your
browser tests open (a test build, or a staging page); a production page does not need it.

## `ypy.idle([options])`

Returns a promise. It resolves when nothing is left to wait for, and rejects with an `Error` after `timeout`.

```js
await ypy.idle()                    // defaults
await ypy.idle({ timeout: 3000 })   // reject after 3 s
await ypy.idle({ quiet: 0 })        // no grace period after the last piece of work
await ypy.idle({ timeout: '2s' })   // a time with a unit also works
```

| Option | Default | Meaning |
|---|---|---|
| `timeout` | `5000` | milliseconds before the promise rejects; `0` waits for ever |
| `quiet` | `50` | milliseconds the page must stay without pending work before it resolves; `0` resolves at once |

The defaults come from `ypy.config.test`, and the options of a call win:

```js
ypy.config.test = { timeout: '10s', quiet: 100 }
```

### What counts as work

The page is busy while any of these is pending:

| Kind | Pending from | until |
|---|---|---|
| `request` | `ypy:before:request` | `ypy:finally:request` |
| `queued` | a request waits its turn in an `ypy-sync` queue | it starts or is dropped |
| `swap` | `ypy:before:swap` | `ypy:finally:swap` |
| `settle` | `ypy:before:settle` | `ypy:after:settle`, also for content that a stream or a server-sent event inserts |

Not counted:

- A request that turned into a server-sent event stream: it stays open as long as the page lives.
- A WebSocket.
- Polling, except while a poll request is running.
- A `settle` that never ends is given up after 10 seconds, so a broken settle task cannot keep the page busy for ever.

### The quiet period

After the last piece of work ends, `ypy.idle()` waits `quiet` milliseconds and checks again, because the content that
was just swapped in may start the next request (a `load` trigger). If new work starts in that time, the wait goes on.
This is why a chain such as "click, swap, the swapped content loads another fragment" is one `ypy.idle()`.

When the promise resolves, `ypy:test:idle` fires on `document` with `{waited}`, the milliseconds the call waited. It is
not cancelable.

### The error on timeout

```
ypy.idle: still busy (request, swap) after 5000 ms
```

The message names the kinds of work still pending. A request that never answers shows as `request`; a request held
back by `ypy-sync` shows as `queued`.

## `data-ypy-busy` on `<html>`

While work is pending, `<html>` carries `data-ypy-busy`, a space-separated list of the kinds (`"request swap"`), and it
loses the attribute when the work ends. A test can wait for the DOM instead of calling the function:

```js
await page.waitForSelector('html:not([data-ypy-busy])')
```

It is output for tests and styles, not something to write by hand. It does not include the quiet period: the attribute
is gone as soon as the work ends, while `ypy.idle()` resolves `quiet` milliseconds later. For a chain of requests
prefer `ypy.idle()`.

## A complete test

The page:

```html
<button id="save" ypy-post="/notes" ypy-include="#note" ypy-target="#flash" class="rounded bg-sky-600 px-3 py-1 text-white">
  Save
</button>
<textarea id="note" name="note" class="block w-full rounded border"></textarea>
<p id="flash" class="text-emerald-700"></p>
```

The server answers `200` with `Saved`. The test with Playwright:

```js
await page.fill('#note', 'hello')
await page.click('#save')
await page.evaluate(() => ypy.idle())
await expect(page.locator('#flash')).toHaveText('Saved')
```

`ypy.idle()` is an ordinary promise: any driver that can await a script can call it. In a test of a failing case, pass a short timeout so a hang fails fast:

```js
await page.evaluate(() => ypy.idle({ timeout: 2000 }))
```

## Pitfalls

- Call it after the action, not before. A trigger with a `delay:` or a `throttle:` has not started a request yet when the click returns, and the page looks idle until it does. Wait for the visible result, or use a `quiet` longer than the delay.
- A page that polls or streams is never "done". Polls count only while they run, so `idle()` resolves between
  polls; a test that needs a state should assert on the DOM after it.
- `ypy.idle` is only there after the extension loaded. In a test that starts before the script, wait for
  `window.ypy?.idle`.
- A rejected `idle()` leaves the work running: it reports, it does not cancel.
- The extension uses the public API only. It does not change a request or a swap.

See also: [Testing your app](../index.md#server-side) for `ypytest` (the Go side), [Events](../index.md#events-and-logging),
[`ypy-telemetry`](telemetry.md) to collect failures during a test run.
