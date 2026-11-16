package public

import (
	"crypto/sha256"
	"encoding/hex"
)

// Image is a link-preview picture: where it is, the size it was drawn at, its
// type and what it shows, for a screen reader.
type Image struct {
	Path   string
	Width  int
	Height int
	Type   string
	Alt    string
}

// The covers are 1200 by 630 PNG files drawn by tools/brand, one for the front
// page and one for each section.
const (
	coverWidth  = 1200
	coverHeight = 630
)

// cover is the file and the alternative text of each cover, by key.
var coverFiles = map[string]struct{ file, alt string }{
	"":           {"social-cover.png", "YPY, a hypermedia framework, with its logo and the address ypy.dev"},
	"docs":       {"social-cover-docs.png", "YPY Docs: guides, core concepts and the server side, with the YPY logo"},
	"reference":  {"social-cover-reference.png", "YPY Reference: every attribute, event, header and configuration key, with the YPY logo"},
	"examples":   {"social-cover-examples.png", "YPY Examples: click to edit, active search, bulk update and delete a row, with the YPY logo"},
	"extensions": {"social-cover-extensions.png", "YPY Extensions: streaming, WebSocket, history cache and more, with the YPY logo"},
	"go":         {"social-cover-go.png", "YPY Go SDK: the server side in Go, with the YPY logo"},
}

// Cover is the link-preview image of a section: "docs", "reference",
// "examples", "extensions" or "go". Any other key, the empty one included, is
// the front page's cover.
func Cover(key string) Image {
	c, ok := coverFiles[key]
	if !ok {
		c = coverFiles[""]
	}
	return Image{Path: versioned(c.file), Width: coverWidth, Height: coverHeight, Type: "image/png", Alt: c.alt}
}

// Covers lists every cover's address, so a test can ask for each.
func Covers() []Image {
	out := make([]Image, 0, len(coverFiles))
	for key := range coverFiles {
		out = append(out, Cover(key))
	}
	return out
}

// versioned is the address of an embedded image with the start of its SHA-256
// as a query. A link preview (WhatsApp, LinkedIn, Slack) keeps the image it
// fetched under an address for days; a new cover gets a new address, so the
// preview fetches it.
func versioned(name string) string {
	body, err := files.ReadFile(name)
	if err != nil {
		panic("public: missing " + name + ": " + err.Error())
	}
	sum := sha256.Sum256(body)
	return "/" + name + "?v=" + hex.EncodeToString(sum[:])[:12]
}
