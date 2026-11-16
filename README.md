# ypy.dev

The website of [YPY](https://github.com/ypy-develop/ypy) at https://ypy.dev: the
documentation, a reference generated from the library's own name tables, and the
Go SDK guide. It is an [Arandu](https://arandu.io) application.

```sh
task serve          # http://localhost:8080
```

Without `task`: `aru view:build && go run . serve`.

## How it is made

- `content/` holds the pages as Markdown. A file is a page at the address its path
  gives; the sidebar, the page contents, "previous" and "next" and the search all
  come from the files. Nothing is listed by hand.
- `app/Reference/data/` is a copy of YPY's `src/attributes.json` and
  `src/attributes.d/*.json`, taken with `task sync`. The reference pages (attributes,
  events, headers, configuration, classes, elements, extensions) are generated
  from it, so the site documents exactly the names the library has.
- The site navigates with YPY itself: the script is the one hesape embeds, and two
  extensions are served from the project (`task sync` copies them from YPY's
  `dist/embed/ext`).
- Rendering is [`hyz-is/arandu-markdown`](https://github.com/hyz-is/arandu-markdown);
  the interface is Kyse and Basecoat, under a strict content security policy.

## Develop

```sh
task sync    # after YPY changes (needs the ypy repository next to this one)
task check   # the gates CI runs
```

Settings are in `.env` (copy `.env.example`). The site needs no database and no Redis.

## Licence

MIT. Copyright HYZIS - SERVICOS DIGITAIS LTDA - EPP. See `LICENSE.md`.
