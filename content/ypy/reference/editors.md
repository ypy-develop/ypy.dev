---
title: "Editor support"
description: "YPY ships data for editors, so ypy- attributes autocomplete in HTML, ypy-on:: offers the ypy: events, and the ypy-config meta tag has a schema."
order: 150
---
# Editor support

YPY ships data for editors, so `ypy-*` attributes autocomplete in HTML, `ypy-on::` offers the `ypy:*` events, and
the `ypy-config` meta tag has a schema. The data is generated from one table of names, never written by hand.

| File | For | Contents |
|---|---|---|
| `editors/vscode/ypy.html-data.json` | VS Code | HTML Custom Data: attributes with their values, the `ypy-partial` tag, `ypy-on::<event>` for every event |
| `src/editors/jetbrains/ypy.web-types.json` (`dist/editors/jetbrains/`) | WebStorm and the other JetBrains IDEs | web-types: attributes, `ypy-partial`, events for `ypy-on`, CSS classes |
| `src/editors/schema/ypy.config.schema.json` (`dist/editors/schema/`) | any editor with JSON Schema support | JSON Schema (draft 2020-12) of the `ypy-config` object and of `ypy.config` |
| `src/ypy.d.ts` (`dist/ypy.d.ts`) | TypeScript, and editors that read types | the `ypy` object, `ypy.config`, the typed event map |

The editors offer the names YPY reads. Names from markup written before YPY (`ypy-vars`, `ypy:afterRequest`, ...)
are not offered, so a new page does not start with them. Hooks that only extensions receive (`ypy:scope`, `ypy:before:morph:node`, ...) are not
events of the DOM and are left out too.

## VS Code

Point `html.customData` at the file. In `.vscode/settings.json`:

```json
{
  "html.customData": ["./tools/ypy.html-data.json"]
}
```

The path is relative to the workspace folder and points at a copy of `editors/vscode/ypy.html-data.json` (here
in `tools/`), or at the file inside a checkout of the repository. Reload the window. In an `.html` file (and in templates VS Code treats as HTML), typing
`ypy-` lists the attributes with a description, the values they take (`ypy-swap="innerMorph"`) and the modifiers.
Typing `ypy-on::` lists the events.

VS Code cannot complete a family whose name is open (`ypy-computed:<name>`), so the data lists the common
spellings of each family (`ypy-live:text`, `ypy-status:422`, `ypy-on:click`) and no more.

The schema of the configuration applies to JSON, not to the `content` of a `<meta>` tag. Use it for a JSON file
that holds the configuration, for example one your server reads to render the tag:

```json
{
  "json.schemas": [
    { "fileMatch": ["ypy.config.json"], "url": "./tools/ypy.config.schema.json" }
  ]
}
```

Autocomplete and diagnostics for the keys then work in that file. The schema is closed: a key that is not in the
table is reported, which catches a typo (`defaultSwapp`). A key of a third-party extension is reported too; add
it to your own schema with `allOf` if you use one.

## JetBrains

The web-types file describes the attributes and events to the HTML support of WebStorm, IntelliJ IDEA Ultimate,
PhpStorm and the others. JetBrains IDEs load web-types from the `web-types` field of a `package.json`:

```json
{
  "web-types": "./path/to/ypy.web-types.json"
}
```

Put the field in the `package.json` of your project (or of the package that ships the file), and make sure the
IDE has indexed it (File | Invalidate Caches, or reopen the project). Then `ypy-` completes attributes with their
documentation, and `ypy-on:` and `ypy-on::` complete DOM events and YPY events.

For the schema, add a mapping under Settings | Languages & Frameworks | Schemas and DTDs | JSON Schema Mappings:
choose `ypy.config.schema.json` and the files that hold your configuration.

The data is generated and checked in tests; it has not been opened in a JetBrains IDE for this release, so report
anything an IDE does not show as expected.

## TypeScript

`ypy.d.ts` declares the global `ypy`, an ES module default export, and the types of everything the object takes:

```ts
import ypy, { type YpyConfig } from './vendor/ypy'; // wherever ypy.d.ts and the script sit

ypy.on('ypy:before:request', (evt) => {
  evt.detail.ctx.request.headers['X-Token'] = token; // evt.detail is typed for this event
});

document.addEventListener('ypy:response:error', (evt) => evt.detail.ctx.response?.status);

ypy.config.defaultSwap = 'innerHTML';
const config: Partial<YpyConfig> = { sse: { reconnect: true }, retry: { attempts: 3 } };
```

- `YpyEventMap` maps each event name to the type of its `detail`. `ypy.on()`, `ypy.trigger()` and
  `addEventListener()` on `document`, `window` and elements read it.
- `YpyHookMap` types the hooks an extension receives (`ypy_before_inherit` and the like). They are not events of
  the DOM.
- `YpyConfig` is `ypy.config`. Extensions add their keys by merging into the interface, so a key appears when its
  extension is part of the declaration file.
- Members that exist only when an extension is loaded are optional: `ypy.signals?`, `ypy.live?`,
  `ypy.config.lock?()`.

## The table of names

Everything above comes from `src/attributes.json` (the core) and one fragment per extension in
`src/attributes.d/<extension>.json`, where `<extension>` is the file name in `src/ext/` without `.js`
(`src/attributes.d/ypy-sse.json` for `src/ext/ypy-sse.js`). A fragment has the same sections as the core table:

| Section | One entry per | Main fields |
|---|---|---|
| `attributes` | attribute, or family such as `ypy-live:*` | `name`, `description`, `valueType`, `values`, `modifiers`, `inherited`, `variants`, `pattern`, `doc` |
| `elements` | custom tag | `name`, `description`, `attributes` |
| `events` | `ypy:*` event or extension hook | `name`, `description`, `detail` (field to TypeScript type), `cancelable`, `target`, `listen`, `extensionOnly` |
| `headers` | `YPY-*` header | `name`, `direction` (`request` or `response`), `description`, `value` |
| `config` | `ypy.config` key, nested keys with a dotted path | `name`, `description`, `schema` (JSON Schema), `default`, `deprecated` |
| `classes` | CSS class | `name`, `description` |
| `internalNames` | name in the sources that is not public | the names |

The core table also holds `version`, which must be the version `src/ypy.js` reports. An unknown field or a name that
is written twice is an error.

### Adding the names of a new extension

1. Write `src/attributes.d/<extension>.json` with `"extension": "<extension>"` and the names the extension adds:
   attributes, the events it fires with the fields of `detail`, headers, config keys with their defaults, classes.
   Names that appear in its source but are not public (a private data attribute, the name it registers under) go
   in `internalNames`.
2. Declare the same events and config keys in `src/ypy.d.ts`: events in `YpyNativeEventMap`, config keys in
   `YpyConfig` (or in an interface that a key points to).
3. Run `ypy gen`. It rewrites the three generated files.
4. Run `ypy gen --check`. It must pass.

### What `ypy gen --check` verifies

- The table is valid: names follow the contract (`ypy-`, `ypy:`, `YPY-`), nothing is defined twice, every entry
  has a description, defaults fit their schema, nested config keys have their parent.
- The three generated files are exactly what the table produces.
- `src/ypy.d.ts` and the table agree: every DOM event of the table is in the event map with the same `detail`
  fields, every event of the map is in the table, and the same holds for the keys of `YpyConfig`, nested ones
  included.
- The sources use no `ypy-`, `ypy:` or `YPY-` name that the table does not know. A name the table lists that no
  source mentions is reported as a warning. An extension file without a fragment gets one warning, not an
  error, until its fragment exists.

`ypy gen --check` is the `gen` gate of `ypy ci`. When the TypeScript compiler is installed, `go test ./internal/gen`
also compiles `src/ypy.d.ts` together with a file that uses it (`internal/gen/testdata/types/consume.ts`);
the project itself needs no Node.
