---
version: alpha
name: YPY
description: The documentation site of YPY, in the house frame of arandu.io and yxy.dev, drawn in black, white, red and mint.
colors:
  background: "#ffffff"
  foreground: "#0a0a0a"
  card: "#fafafa"
  muted-foreground: "#5c5c5c"
  border: "color-mix(in srgb, #0a0a0a 12%, transparent)"
  primary: "#DC2C02"
  primary-foreground: "#ffffff"
  ring: "#DC2C02"
  accent-mint: "#6ECFB5"
colors-dark:
  background: "#000000"
  foreground: "#ffffff"
  card: "#0d0d0d"
  muted-foreground: "#a3a3a3"
  primary: "#DC2C02"
  primary-foreground: "#ffffff"
  ring: "#DC2C02"
typography:
  display: { fontFamily: "Montserrat, ui-sans-serif", fontWeight: 700, letterSpacing: "-0.03em" }
  body:    { fontFamily: "Montserrat, ui-sans-serif", fontSize: "15.5px", fontWeight: 450, lineHeight: 1.75 }
  label:   { fontFamily: "ui-monospace, SF Mono, JetBrains Mono, monospace", fontSize: "11px", letterSpacing: "0.14em" }
spacing: { frame-rule: "2px", header: "72px", gutter: "24px" }
---

# Design: YPY

What the stylesheets (`resources/css/app.css`, `home.css`, `docs.css`) and the views draw, and where each part
came from. Product facts and voice are in `PRODUCT.md`.

## Overview

**North star: a reference book bound in the logo's colours.** The reader comes to look a name up, so the page
is a long, plain column with an index on its left. The identity is the logo's: black, white, one red and one mint.
Nothing else is a colour.

**Signature detail.** On YPY, *the "+" that opens a section of the index* becomes *the red of the logo*: the only
mark on the sidebar that is not text. The competitor test passes: no other docs site has this one.

**Direction: a hybrid, by the author's decision (2026-11-15).** The house frame comes from two siblings; the
colour and the details are YPY's own. Each part below says its origin.

| Part | Origin | What we took |
|---|---|---|
| Frame: two hairline rails, registration crosses where a band meets a rail, 72 px header strip | arandu.io, yxy.dev | as they are |
| Footer: brand, one sentence, link columns, one centred line "Developed by [HYZIS] in [Arandu]" | arandu.io | structure and text; the columns are YPY's pages; no licence line |
| Header: logo, five links, search, theme, GitHub with the star count | arandu.io | the star count (`app/Stars`) and the strip |
| Type: Montserrat for words, monospace for labels, paths and code | arandu.io, yxy.dev | as they are; vendored |
| Numbered eyebrow on the home bands ("01 INSTALL") | yxy.dev | as it is |
| Code blocks: black panel, hairline, a bar with the language and Copy; syntax in red, mint and white | yxy.dev, arandu.io | the panel; the colours are YPY's |
| Palette: black, white, `#DC2C02`, `#6ECFB5` | YPY | not the brown and orange of arandu.io and not the warm tints of yxy.dev |
| Sidebar of a long page: the page's own headings, red "+" on the ones with subsections | YPY | the old portal's index, drawn in the identity |
| Logo plaque: the full logo always on a white plaque, in both themes | YPY | logos are never recoloured |

Not taken: the lighter red or orange for the dark theme that yxy.dev and arandu.io use. The red is `#DC2C02` in
both themes.

## Colors

- **Primary**: `#DC2C02`, used for links, the current section, the primary button, the focus ring and the "+".
- **Mint** `#6ECFB5`: the second voice. The status dot, a type name in code, the focus outline of a code panel,
  the third dot of a code window.
- **Neutral**: paper `#ffffff` and ink `#0a0a0a` in light; `#000000` and `#ffffff` in dark, with cards one step
  off the page (`#fafafa`, `#0d0d0d`).
- **The One Voice rule**: red marks what can be acted on or what is current; mint only ever decorates; they are
  never both on the same word.
- **Known limit**: `#DC2C02` on `#000000` is about 4.4:1, a little under 4.5:1 for small body text. Links are
  underlined, never colour alone. The author asked for this exact red in both themes.

## Typography

Montserrat 400-800, vendored. Page title 700 with tight tracking; body 15.5 px at weight 450 and line height 1.75,
measure about 65 characters (the article is 50 rem at most). Labels are monospace, 11 px, uppercase, +0.14em, for
section names, "min read" and "Updated". Sentence case everywhere else.

## Layout

- Frame width 1240 px; gutter 24 px (20 px under 640 px).
- A page of the manual has three columns from 1280 px: index (15 rem), article, contents (13 rem). A long page
  (`/docs`, `/reference`, `/migration`) has two: its headings are the index, so the contents column is gone and the
  article takes the room tables need.
- Under 1024 px the index is a disclosure above the article; the header links fold into a disclosure with the GitHub
  button and its stars at the end.

## Elevation and Depth

Borders, not shadows: hairlines at 12% ink and 2 px frame rules; shadows are rare and soft. Dark surfaces step `#000`, `#0d0d0d`, `#141414`, `#1a1a1a`.

## Shapes

Controls and panels take the Basecoat radius token; pills are for the version badge only. Registration crosses are the
Kyse `Plus` icon, never drawn by CSS.

## Components

Kyse components first (`ThemeToggle`, `icons.*`); the project's own classes for page structure: `docs-layout`,
`docs-sidebar`, `docs-group` (a `<details>` whose "+" is red), `docs-nav-link`, `ref-table` (the generated tables,
drawn by `partials.ReferenceTable`), `site-footer`, `site-footer-signature`, `brandmark-*`. The star count is a
`<span class="tabular-nums">` that is absent until fetched; an explicit zero is drawn.

## Do's and Don'ts

- Do keep one stroke weight (2 px) for rails and rules; do keep red for things to act on.
- Do not recolour, outline or make a dark version of the logo; do not draw an orange anywhere.
- Do not add a motion to the manual: the page changes with a 160 ms colour transition and a cross-fade between
  pages; the home may carry one authored moment.
- No `style=`, no inline script, no `on*=`; a class is a literal in a view or a rule in the stylesheets.
