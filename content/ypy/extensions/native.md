---
title: "Native app bridge: the ypy-native extension"
description: "The web half of a native shell: a page in an iOS, Android or desktop web view talks to the app, so native navigation and titles can replace page parts."
order: 240
---
# Native app bridge: the `ypy-native` extension

A native shell is a small app (iOS, Android, or a desktop system web view) that owns the navigation chrome (the
stack, the tab bar, modals, the title) and draws each screen in a web view that loads the same server-rendered
page. Where it pays, the shell replaces a piece of the page with a native component: a menu, a date picker, a
share sheet. The server stays single, and the web page stays the reference implementation.

`ypy-native` is the web half of that design. The shell is native code in your own project: YPY ships neither a shell
nor a native UI kit. In a
plain browser the extension does nothing except hide what is marked `ypy-native-only`, and every element it
touches is valid HTML that works without it.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-native.min.js"></script>

<header ypy-native-hide>...the web header, replaced by the native bar...</header>
<button ypy-native-component="share" ypy-native-props='{"url": "/orders/7"}' ypy-native-click>Share</button>
```

The extension is not in `ypy.max`: load it in a page that has an app. It is 2,827 bytes min+gzip, and the core did not
grow for it.

Contents: [How the server learns it](#the-server-learns-it-is-an-app) · [The channel](#the-channel-and-the-handshake) ·
[Path configuration](#path-configuration) · [Visits](#visits) · [Components](#native-components) ·
[Messages](#messages) · [Security](#security) · [Go](#go-sdk) · [API](#javascript-api) · [Config](#configuration) · [Limits](#limits)

## The server learns it is an app

Three signals, all additive: an old client ignores them.

| Signal | Who sets it | Notes |
|---|---|---|
| user agent suffix `YpyNative/<n> (<platform>)` | the shell, on every request of the web view | always there, including the first load before any script runs; `<n>` is the bridge version |
| request header `YPY-Native: <platform>; v=<n>; caps=<a,b>` | this extension, on every same-origin request | a web view cannot add headers to `fetch`; `caps` lists the native components the installed build knows; never sent to another origin |
| `data-ypy-native="<platform>"` on `<html>` | this extension (the server may also render it) | for CSS |

A server that varies its HTML on these must send `Vary: YPY-Native, User-Agent`. It should decide on capabilities and
not on version numbers, so an old app build keeps working against a new server. The values are hints: they grant
nothing.

## The channel and the handshake

The shell gives the page one object, `window.ypyNative`, with `postMessage(text)` to talk to the shell and
`message` events (`event.data` is text) to hear it. Android's web message listener provides exactly that; an iOS
shell defines it with a user script that forwards to its message handler. Every message is JSON text, an object with
a string `type`. Nothing in a message is evaluated, on either side. The extension runs only in the top frame.

```text
extension  ->  {"type":"load","v":1}
shell      ->  {"type":"hello","v":1,"platform":"ios","caps":["menu","share"],"nonce":"…","announce":false}
extension  ->  {"type":"ready","v":1,"nonce":"…"}
```

Until `hello` is valid (a lower-case platform, a version from 1 to 9999, a non-empty nonce) the extension ignores
everything else. Another `hello` starts the handshake again. The shell may send `hello` on its own: it does not
have to wait for `load`.

## Path configuration

The shell does not hard-code routes. It reads one JSON document at launch and on resume, keeps it with its `ETag`,
and ships a copy in the app as the fallback, so a route change is a deploy and not a store review. The path is
`/.well-known/ypy-native.json` (`config.native.path`).

```json
{
  "version": 1,
  "settings": { "tabs": [{"title": "Orders", "path": "/orders"}], "bridge_origins": ["https://app.example.com"] },
  "rules": [
    { "patterns": ["^/orders/new$", "^/orders/\\d+/edit$"], "properties": { "context": "modal", "modal_style": "large" } },
    { "patterns": ["^/orders/\\d+$"],                      "properties": { "presentation": "push" } },
    { "patterns": ["^/signin$"],                           "properties": { "presentation": "replace_root", "pull_to_refresh": false } }
  ]
}
```

A rule has `patterns`, regular expressions over path plus query (keep to the subset Go, Swift, Kotlin and JavaScript
share), and `properties`. For one URL every matching rule applies in order and a later property overrides an
earlier one. An unknown property is ignored, so the document can grow.

| Property | Values | Meaning |
|---|---|---|
| `context` | `default`, `modal` | the screen goes on the main stack or in a presented sheet with a stack of its own |
| `modal_style` | `medium`, `large`, `full`, `drawer` | how a modal is presented |
| `presentation` | `push`, `replace`, `replace_root`, `pop`, `refresh`, `none` | what a visit to the URL does to the stack |
| `view` | `web`, `native:<name>` | a web view, or a native screen the app registered |
| `pull_to_refresh` | boolean | the native refresh gesture reloads the screen |
| `query_string` | `ignore`, `keep` | whether two URLs that differ only by query are the same screen |

The extension reads the same document, for two things: `presentation: none` leaves a route to YPY inside the
current page, and, outside the app, a route with `context: modal` can open as a layer ([below](#one-server-for-the-browser-and-the-app)).
`ypy.native.properties(url)` tells what the document says about a URL.

## Visits

A visit is a navigation of the main content (the `<body>` target): a boosted link or form, a `YPY-Location`
answer, or a post that ends in a redirect to a page. Most requests of a YPY page are not visits: they swap a fragment
into another target and keep doing so.

| What happens | `source` | When |
|---|---|---|
| a GET to the main target is not made | `boost`, `form`, `location` | a boosted link, a boosted GET form (its query included), an `ypy-push-url` or `YPY-Location` navigation |
| the response of a POST, PUT, PATCH or DELETE that was redirected is not swapped | `redirect` | the shell loads the page it ended on; a response that is a page with no redirect (a form with errors) stays in the page |
| an event | `event` | `ypy:native:navigate`, or `ypy.native.visit(url)` |

The extension sends `{"type":"visit","url":"https://…/orders/7","action":"advance","source":"boost"}` where
`action` is `advance`, `replace` (`ypy-replace-url`) or `restore`. The shell applies the path configuration of the
destination: `push` adds a screen, `replace` swaps the top one, `replace_root` resets the stack, `pop` closes the
current screen and `refresh` reloads the one below, and `none` is left to YPY. A modal is a stack of its own, so a
successful post inside it ends in a visit whose rule says `pop`, plus `refresh` of the screen below: what `303` to the
list means on the web.

Cancel `ypy:native:before:visit` (it fires on the source element with `{url, action, source}`) to keep one
navigation in the page. A click with Ctrl, Meta or Shift is left to the browser, as everywhere in YPY.

The server steers the app too: `YPY-Trigger: {"ypy:native:navigate": {"url": "/orders", "action": "replace"}}`, or
the same event from a stream. In the app it is a visit; in a browser it swaps the page.

The shell owns the history stack and each screen is its own document, so while the handshake is done the
extension cancels `ypy:before:history:update` and `ypy:before:history:restore`: the page writes no history entry and
restores none. These two events are the supported way for any extension to take history over.

On pull to refresh, the shell sends `{"type":"reload"}`. The extension requests the current URL into the main
target with a full response and the `YPY-Boosted` header, as a boosted link would, and that request is not a visit.

After the handshake, and whenever the title or the address changes, the page sends
`{"type":"page:update","url":"…","title":"…"}` so the native bar can show them. With `"announce": true` in `hello`, the
`ypy:announce` messages of the page are also sent as `{"type":"announce","message":"…","politeness":"polite"}`
so the platform can say them itself, and the shell should then silence the web live region.

### One server for the browser and the app

Outside the app the same rules map to web behavior, so the path configuration is the single description. With
`config.native.web: true` and `ypy-layers` loaded, a visit to a route with `context: modal` opens as a layer
(`modal_style: drawer` as a drawer) in a browser. Call `ypy.native.load()` to read the document, or let the extension
do it when it starts. Every other route is an ordinary swap.

## Native components

Markup declares what the native side may render in place of, or next to, an element. The element is the fallback.

| Attribute | Value | Effect in the app |
|---|---|---|
| `ypy-native-component` | a component name | asks the shell for that component; ignored unless the name is in `caps` |
| `ypy-native-props` | one JSON object, at most 2 KiB | the component's properties; data only; anything else is ignored |
| `ypy-native-hide` | present | hidden in the app, and `inert` (out of the accessibility tree and the tab order) |
| `ypy-native-only` | present | shown only in the app; in a browser it is hidden and `inert` |
| `ypy-native-click` | present, on the element itself | a native `action` message from the component clicks the element |

Hiding is done with an adopted style sheet keyed on `data-ypy-native`, so a strict CSP is fine. The `data-` spelling
of each attribute and `config.prefix` work as everywhere. Markup under `ypy-ignore`, and an element the nonce gate
of [`ypy-csp`](../reference/security.md) refuses, is not looked at.

Lifecycle, driven by the hooks every extension has: when an element is processed the extension sends
`component:connect`; when its props change, `component:update`; when it leaves the page, `component:disconnect`.
The shell answers with `component:event`, which becomes the cancelable DOM event `ypy:native:event` on the element
(`detail` is `{id, name, detail}`), or, for `ypy-native-click` and the name `action`, a click, unless a listener
cancels the event. Elements added to the page by code other than a swap are found after `ypy.process(root)`.

## Messages

| Direction | `type` | Fields |
|---|---|---|
| page to shell | `load` | `v` |
| shell to page | `hello` | `v`, `platform`, `caps`, `nonce`, optional `announce` |
| page to shell | `ready` | `v`, `nonce` |
| page to shell | `visit` | `url`, `action`, `source` |
| page to shell | `page:update` | `url`, `title` |
| page to shell | `announce` | `message`, `politeness` |
| page to shell | `component:connect` | `id`, `name`, `props` |
| page to shell | `component:update` | `id`, `props` |
| page to shell | `component:disconnect` | `id` |
| shell to page | `component:event` | `id`, `name`, `detail` |
| shell to page | `reload` | |

A message over 64 KiB, text that is not JSON, and an unknown `type` are ignored.

## Security

The bridge is a privilege boundary.

- The shell attaches the channel only to the origins in `bridge_origins` and drops any message from another frame. The extension speaks only after the handshake, and a custom URL scheme is not a channel: it has no origin check.
- The shell keeps an allowlist of component names and validates props against each component's schema; the extension also sends nothing for a name that is not in `caps`.
- A component that does something sensitive (share, buy, open another app) acts only after a user gesture on the native side.
- Anyone who can put markup in a page can write `ypy-native-*`. The same defense as for every `ypy-*` attribute applies: with [`ypy-csp`](../reference/security.md) the nonce gate removes the attributes the server did not stamp, and user-generated HTML must be sanitized with `ypy-native-*` and `data-ypy-native-*` in the deny list.
- Authentication stays in the web view's cookie jar, synchronized with the shell's networking. No token goes in a URL or a message. A native sign-in ends with the shell setting the cookie and sending a visit.
- The `YPY-Native` header is not sent to another origin.

## Go SDK

`ypy.Request.Native` (platform, version, `Caps()`, `Has`), `ypy.NativeMiddleware` and `ypy.VaryNative`,
`ypy.PathConfig` with `Handler()` (strong `ETag`, `304`, `Cache-Control: no-cache`) mounted at `ypy.NativePathConfigPath`,
and `ypy.NativeNavigate`. In tests: `ypytest.Native`, `ypytest.NativeUserAgent`, `AssertVaryNative` and
`AssertNativeNavigate`. The wire format is in the protocol specification of the Go SDK, section 3.5.

## JavaScript API

| Member | What it does |
|---|---|
| `ypy.native.connect(channel?)` | attach to `window.ypyNative` (or the object given) and send `load`; `false` when it is not a channel |
| `ypy.native.disconnect()` | leave the shell: components disconnect, the header and the visits stop |
| `ypy.native.receive(message)` | deliver one message from the shell; for hosts that call into the page instead of dispatching events |
| `ypy.native.send(type, fields)` | send a message of your own; `false` without a channel |
| `ypy.native.visit(url, action?)` | a visit by code, same origin only |
| `ypy.native.load()` | read the path configuration again |
| `ypy.native.properties(url)` | the properties the document gives a URL |
| `ypy.native.ready`, `.platform`, `.caps` | the state after the handshake |

## Events

| Event | On | `detail` |
|---|---|---|
| `ypy:native:event` | the component's element, cancelable | `{id, name, detail}` |
| `ypy:native:before:visit` | the source element, cancelable | `{url, action, source}` |
| `ypy:native:navigate` | listened on the document | `{url, action}` |

## Configuration

| Key | Default | Meaning |
|---|---|---|
| `native.path` | `/.well-known/ypy-native.json` | the path configuration document; `false` does not load it |
| `native.web` | `false` | in a browser, open `context: modal` routes as layers |

## Limits

- The shell is not part of YPY. Stack, presentation, the pool of web views, the path configuration cache and its bundled fallback, the native components, cookie synchronization, offline and error screens, file and camera pickers, push notifications, deep links, the app lifecycle and store packaging belong to the app.
- A thin wrapper around a website can be rejected by a store for minimum functionality: ship real native components and screens.
- The core did not change. The extension classifies a visit from the request context (the `YPY-Boosted` header, the target and `ctx.push`), not from a core field.
- The well-known path, the extension name and the header name are working names and may change. The extension speaks bridge version 1.
