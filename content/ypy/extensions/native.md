---
title: "Native app bridge: the ypy-native extension"
description: "The web half of a native shell: a page in an iOS, Android or desktop web view talks to the app, so native navigation and titles can replace page parts."
order: 240
---
# Native app bridge: the `ypy-native` extension

A native shell is a small app (iOS, Android or a desktop web view) that owns the navigation chrome (the stack, the tab bar,
modals, the title) and draws each screen in a web view that loads the same server-rendered page. Where it pays, the shell
replaces a piece of the page with a native component: a menu, a date picker, a share sheet. The server stays single, and
the web page stays the reference implementation.

`ypy-native` is the web half of that design. The shell is native code in your own project, and YPY ships neither a shell
nor a native UI kit. In a plain browser the extension does nothing except hide what is marked `ypy-native-only`.

```html
<script src="ypy.min.js"></script>
<script src="ext/ypy-native.min.js"></script>

<header ypy-native-hide>...the web header, replaced by the native bar...</header>
<button ypy-native-component="share" ypy-native-props='{"url": "/orders/7"}' ypy-native-click>Share</button>
```

The extension is not in `ypy.max`. It is 2,827 bytes min+gzip.

## The server learns it is an app

| Signal | Who sets it | Notes |
|---|---|---|
| user agent suffix `YpyNative/<n> (<platform>)` | the shell, on every request of the web view | always there, including the first load; `<n>` is the bridge version |
| request header `YPY-Native: <platform>; v=<n>; caps=<a,b>` | this extension, on every same-origin request | `caps` lists the native components the installed build knows; never sent to another origin |
| `data-ypy-native="<platform>"` on `<html>` | this extension | for CSS |

A server that varies its HTML on these sends `Vary: YPY-Native, User-Agent`. Decide on capabilities and not on version
numbers, so an old app build keeps working against a new server. The values are hints and grant nothing.

## The channel and the handshake

The shell gives the page one object, `window.ypyNative`, with `postMessage(text)` to talk to the shell and `message` events
to hear it. Every message is JSON text, an object with a string `type`, and nothing in a message is evaluated. The extension
runs only in the top frame.

```text
extension  ->  {"type":"load","v":1}
shell      ->  {"type":"hello","v":1,"platform":"ios","caps":["menu","share"],"nonce":"...","announce":false}
extension  ->  {"type":"ready","v":1,"nonce":"..."}
```

Until `hello` is valid (a lower-case platform, a version from 1 to 9999, a non-empty nonce) the extension ignores
everything else. The shell may send `hello` without waiting for `load`.

## Path configuration

The shell reads one JSON document at launch and on resume, keeps it with its `ETag` and ships a copy as the fallback, so a
route change is a deploy and not a store review. The path is `/.well-known/ypy-native.json` (`config.native.path`).

```json
{
  "version": 1,
  "settings": { "tabs": [{"title": "Orders", "path": "/orders"}], "bridge_origins": ["https://app.example.com"] },
  "rules": [
    { "patterns": ["^/orders/new$", "^/orders/\\d+/edit$"], "properties": { "context": "modal", "modal_style": "large" } },
    { "patterns": ["^/orders/\\d+$"], "properties": { "presentation": "push" } },
    { "patterns": ["^/signin$"], "properties": { "presentation": "replace_root", "pull_to_refresh": false } }
  ]
}
```

A rule has `patterns`, regular expressions over path plus query, and `properties`. Every matching rule applies in order
and a later property overrides an earlier one. An unknown property is ignored, so the document can grow.

| Property | Values | Meaning |
|---|---|---|
| `context` | `default`, `modal` | the main stack, or a presented sheet with a stack of its own |
| `modal_style` | `medium`, `large`, `full`, `drawer` | how a modal is presented |
| `presentation` | `push`, `replace`, `replace_root`, `pop`, `refresh`, `none` | what a visit to the URL does to the stack |
| `view` | `web`, `native:<name>` | a web view, or a native screen the app registered |
| `pull_to_refresh` | boolean | the native refresh gesture reloads the screen |
| `query_string` | `ignore`, `keep` | whether URLs that differ only by query are the same screen |

The extension reads the same document: `presentation: none` leaves a route to YPY inside the current page, and in a browser
with `config.native.web: true` and [`ypy-layers`](layers.md) loaded, a route with `context: modal` opens as a layer (a
drawer for `modal_style: drawer`). `ypy.native.properties(url)` tells what the document says about a URL.

## Visits

A visit is a navigation of the main content (the `<body>` target): a boosted link or form, a `YPY-Location` answer, or a
write that ends in a redirect to a page. Most requests of a YPY page swap a fragment and are not visits.

| `source` | When |
|---|---|
| `boost`, `form`, `location` | a boosted link, a boosted GET form, an `ypy-push-url` or `YPY-Location` navigation; the GET is not made |
| `redirect` | the response of a POST, PUT, PATCH or DELETE that was redirected is not swapped; the shell loads the page it ended on |
| `event` | `ypy:native:navigate`, or `ypy.native.visit(url)` |

The extension sends `{"type":"visit","url":"https://.../orders/7","action":"advance","source":"boost"}`, where `action` is
`advance`, `replace` (`ypy-replace-url`) or `restore`. The shell applies the path configuration of the destination. A modal
is a stack of its own, so a successful post inside it ends in a visit whose rule says `pop`, plus `refresh` of the screen
below. That is what `303` to the list means on the web.

Cancel `ypy:native:before:visit`, fired on the source element with `{url, action, source}`, to keep one navigation in the
page. A click with Ctrl, Meta or Shift is left to the browser. The server steers the app too, with
`YPY-Trigger: {"ypy:native:navigate": {"url": "/orders", "action": "replace"}}`.

The shell owns the history stack, so after the handshake the extension cancels `ypy:before:history:update` and
`ypy:before:history:restore`. On pull to refresh the shell sends `{"type":"reload"}`, and the extension requests the current
URL into the main target as a boosted link would, without a visit. After the handshake, and whenever the title or address
changes, the page sends `{"type":"page:update","url":"...","title":"..."}`. With `"announce": true` in `hello`, the
`ypy:announce` messages go to the shell too, which should then silence the web live region.

## Native components

Markup declares what the native side may render in place of, or next to, an element. The element is the fallback.

| Attribute | Value | Effect in the app |
|---|---|---|
| `ypy-native-component` | a component name | asks the shell for that component; ignored unless the name is in `caps` |
| `ypy-native-props` | one JSON object, at most 2 KiB | the component's properties; data only |
| `ypy-native-hide` | present | hidden and `inert` in the app |
| `ypy-native-only` | present | shown only in the app; in a browser it is hidden and `inert` |
| `ypy-native-click` | present, on the element itself | a native `action` message from the component clicks the element |

Hiding uses an adopted style sheet keyed on `data-ypy-native`, so a strict CSP is fine. When an element is processed the
extension sends `component:connect`, when its props change `component:update`, and when it leaves `component:disconnect`.
The shell answers with `component:event`, which becomes the cancelable DOM event `ypy:native:event` on the element, with
`detail` `{id, name, detail}`. Elements added by code other than a swap are found after `ypy.process(root)`.

## Messages

| Direction | `type` | Fields |
|---|---|---|
| page to shell | `load` | `v` |
| shell to page | `hello` | `v`, `platform`, `caps`, `nonce`, optional `announce` |
| page to shell | `ready` | `v`, `nonce` |
| page to shell | `visit` | `url`, `action`, `source` |
| page to shell | `page:update` | `url`, `title` |
| page to shell | `announce` | `message`, `politeness` |
| page to shell | `component:connect`, `component:update`, `component:disconnect` | `id`, `name`, `props` |
| shell to page | `component:event` | `id`, `name`, `detail` |
| shell to page | `reload` | |

A message over 64 KiB, text that is not JSON and an unknown `type` are ignored.

## Security

The bridge is a privilege boundary.

- The shell attaches the channel only to the origins in `bridge_origins` and drops any message from another frame. A custom
  URL scheme is not a channel, because it has no origin check.
- The shell keeps an allowlist of component names and validates props against each component's schema. A component that
  does something sensitive (share, buy, open another app) acts only after a user gesture on the native side.
- Anyone who can put markup in a page can write `ypy-native-*`. With [`ypy-csp`](csp.md) the nonce gate removes attributes
  the server did not stamp, and user-generated HTML must be sanitized with `ypy-native-*` and `data-ypy-native-*` in the deny
  list.
- Authentication stays in the web view's cookie jar. No token goes in a URL or a message.

## JavaScript API, events and configuration

| Member | What it does |
|---|---|
| `ypy.native.connect(channel?)` | attach to `window.ypyNative` (or the object given) and send `load` |
| `ypy.native.disconnect()` | leave the shell; components disconnect and visits stop |
| `ypy.native.receive(message)` | deliver one message from the shell |
| `ypy.native.send(type, fields)` | send a message of your own |
| `ypy.native.visit(url, action?)` | a visit by code, same origin only |
| `ypy.native.load()`, `properties(url)` | read the path configuration again; what it says about a URL |
| `ypy.native.ready`, `.platform`, `.caps` | the state after the handshake |

| Event | On | `detail` |
|---|---|---|
| `ypy:native:event` | the component's element, cancelable | `{id, name, detail}` |
| `ypy:native:before:visit` | the source element, cancelable | `{url, action, source}` |
| `ypy:native:navigate` | listened on the document | `{url, action}` |

| Key | Default | Meaning |
|---|---|---|
| `native.path` | `/.well-known/ypy-native.json` | the path configuration document; `false` does not load it |
| `native.web` | `false` | in a browser, open `context: modal` routes as layers |

## Limits

The shell is not part of YPY: stack, presentation, the pool of web views, the native components, cookie synchronization,
offline screens, pickers, push notifications, deep links and store packaging belong to the app. A thin wrapper around a
website can be rejected by a store for minimum functionality, so ship real native components and screens. The extension
speaks bridge version 1, and the well-known path, the extension name and the header name may still change.
