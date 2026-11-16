// Package public carries the files the site serves under fixed names, inside the
// binary: there is no document root.
//
// The two marks of the signature line are used by the stylesheet as CSS masks,
// and a stylesheet cannot name an asset by its content hash, so they have a
// fixed address of their own. The favicons, the web manifest and the social
// covers are here too: crawlers and browsers ask for them by name.
package public

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	fhttp "github.com/arandu-io/hesape/routing/controllers"
)

//go:embed logo-hyzis.svg aru-icon.svg
var marks embed.FS

//go:embed *.png favicon.ico site.webmanifest
var files embed.FS

// Routes serves the icon under the name browsers ask for, the two marks, the
// other icons, the web manifest and the social covers.
func Routes(r *fhttp.Router, icon []byte) {
	r.Get("/favicon.svg", svg(icon))
	for route, name := range map[string]string{"/marks/hyzis.svg": "logo-hyzis.svg", "/marks/arandu.svg": "aru-icon.svg"} {
		body, err := marks.ReadFile(name)
		if err != nil {
			panic("public: " + err.Error())
		}
		r.Get(route, svg(body))
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		panic("public: " + err.Error())
	}
	for _, e := range entries {
		name := e.Name()
		body, err := files.ReadFile(name)
		if err != nil {
			panic("public: " + err.Error())
		}
		r.Get("/"+name, file(name, body))
	}
}

// file answers an embedded file. A social cover is asked for with the start of
// its hash as a query (see Cover), so it can be cached for good; the icons and
// the manifest are cached for a day.
func file(name string, body []byte) http.HandlerFunc {
	contentType := "image/png"
	switch path.Ext(name) {
	case ".ico":
		contentType = "image/x-icon"
	case ".webmanifest":
		contentType = "application/manifest+json"
	}
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", contentType)
		if req.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		_, _ = w.Write(body)
	}
}

func svg(body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(body)
	}
}
