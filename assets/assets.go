// Package assets carries what `aru view:build` compiled, and the browser files
// the site runs on, into the binary.
package assets

import (
	_ "embed"

	"github.com/arandu-io/hesape/view"
)

// stylesheet is the compiled output of resources/css/app.css. It is committed,
// like generated Go, so `go build` works on a fresh clone before any tool has
// run; `aru view:build` overwrites it.
//
//go:embed app.css
var stylesheet []byte

// The core library is not here: hesape embeds YPY and serves it as
// view.Asset("ypy.min.js"). The two extensions the site uses are not part of
// that embed, so they are copied from YPY's dist/embed by `task sync` and served
// the same way, under their content hash: the page's policy allows scripts from
// 'self' only, so they cannot come from a CDN.
//
//go:embed vendor/ypy-head.min.js
var ypyHead []byte

//go:embed vendor/ypy-preload.min.js
var ypyPreload []byte

// The logos, exactly as they were drawn, with no ground behind them: never
// recoloured, redrawn or put on a plaque. ypy.svg is the full logo (front page,
// footer); ypy-logo.svg is the one the header uses.
//
//go:embed ypy.svg
var logo []byte

//go:embed ypy-logo.svg
var headerLogo []byte

//go:embed ypy-icon.svg
var icon []byte

// Icon is the icon file, which is also what /favicon.svg answers with.
func Icon() []byte { return icon }

func init() {
	view.RegisterAsset("ypy.svg", "image/svg+xml", logo)
	view.RegisterAsset("ypy-logo.svg", "image/svg+xml", headerLogo)
	view.RegisterAsset("ypy-icon.svg", "image/svg+xml", icon)
	view.RegisterStylesheet(stylesheet)
	view.RegisterAsset("ypy-head.min.js", "text/javascript; charset=utf-8", ypyHead)
	view.RegisterAsset("ypy-preload.min.js", "text/javascript; charset=utf-8", ypyPreload)
}
