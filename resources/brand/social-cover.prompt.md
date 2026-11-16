# Social cover source

Outputs: `public/social-cover.png` and one cover per section
(`social-cover-docs.png`, `-reference`, `-examples`, `-extensions`, `-go`),
PNG, 1200 x 630.

Source: `resources/brand/social-cover.html`, a Go `html/template` that
`go run ./tools/brand` fills for each cover and rasterises in headless Chrome.
The page is the light one in every theme: paper ground, a dot grid, a frame
with four registration marks, the full logo `assets/ypy.svg` exactly as drawn
(never recoloured, never a dark version), the name, one line and `ypy.dev`.

Colours and the display face come from the site: paper `#f6f2e8`, ink `#2e2016`,
rust `#af3310`, Montserrat 600-800 (`assets/fonts/montserrat-600-800-latin.woff2`).

The text of each cover is the table `covers` in `tools/brand/main.go`; change it
there and run the tool again. Chrome is started headless with a profile in a
temporary folder that is removed at the end, and is stopped when the shot is taken.
