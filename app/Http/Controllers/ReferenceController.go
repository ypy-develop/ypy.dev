package controllers

import (
	"fmt"
	"net/http"
	"strings"

	hhttp "github.com/arandu-io/hesape/http"

	reference "github.com/ypy-develop/ypy.dev/app/Reference"
	site "github.com/ypy-develop/ypy.dev/app/Site"
	"github.com/ypy-develop/ypy.dev/storage/framework/views"
)

// ReferenceController answers the pages generated from YPY's name tables.
type ReferenceController struct {
	Controller
}

// NewReferenceController builds it.
func NewReferenceController(s *site.Site) *ReferenceController {
	return &ReferenceController{Controller{site: s}}
}

var tableBlurbs = map[reference.Kind]string{
	reference.Attributes: "Every ypy-* attribute: what it takes, what it inherits, what it does.",
	reference.Events:     "The events YPY fires on the page, with what they carry and whether they can be cancelled.",
	reference.Headers:    "The request headers YPY sends and the response headers a server can answer with.",
	reference.Config:     "The configuration keys, with their types and defaults.",
	reference.Classes:    "The CSS classes YPY puts on elements while it works.",
	reference.Elements:   "The custom elements, such as the partial that routes part of a response.",
}

// Index is /reference.
func (c *ReferenceController) Index(ctx *hhttp.Context) error {
	s := c.site
	chrome := s.Chrome(ctx, "/reference", "reference", "Reference",
		"Every attribute, event, header and configuration key of YPY, with the official extensions.")

	tables := make([]views.Table, 0, len(reference.Kinds)+1)
	for _, k := range reference.Kinds {
		tables = append(tables, views.Table{Title: reference.Titles[k], Href: "/reference/" + string(k), Count: len(s.Reference.Of(k)), Body: tableBlurbs[k]})
	}
	tables = append(tables, views.Table{Title: "Extensions", Href: "/reference/extensions", Count: len(s.Reference.Extensions), Body: "The official extensions, the script of each and what it adds."})

	return ctx.View("reference", views.ReferenceData{
		Chrome:  chrome,
		Version: s.YPYVersion(),
		Title:   "Reference",
		Intro:   reference.Inline("The tables below are generated from the name tables that ship with YPY, so they list exactly the names the library has. Each row links to the guide that explains it."),
		Tables:  tables,
	})
}

// Table is /reference/{kind}.
func (c *ReferenceController) Table(ctx *hhttp.Context) error {
	s := c.site
	kind := ctx.Request.PathValue("kind")

	if kind == "extensions" {
		return c.extensions(ctx)
	}
	var k reference.Kind
	for _, candidate := range reference.Kinds {
		if string(candidate) == kind {
			k = candidate
		}
	}
	if k == "" {
		return c.NotFound(ctx)
	}

	items := s.Reference.Of(k)
	rows := make([]views.Row, len(items))
	for i, it := range items {
		rows[i] = views.Row{
			Item:      it,
			Text:      reference.Inline(it.Description),
			Guide:     s.DocFor(it),
			ValueText: reference.Inline(it.Value),
			AttrList:  strings.Join(it.Attrs, ", "),
			ValueList: valueRows(it.Values),
			ModList:   valueRows(it.Modifiers),
		}
	}

	title := reference.Titles[k]
	chrome := s.Chrome(ctx, "/reference/"+kind, "reference", title+" reference", "Every YPY "+title+", generated from the library's own name tables.")
	return ctx.View("reference", views.ReferenceData{
		Chrome:  chrome,
		Version: s.YPYVersion(),
		Kind:    kind,
		Title:   title,
		Intro:   reference.Inline(tableBlurbs[k]),
		Rows:    rows,
	})
}

func (c *ReferenceController) extensions(ctx *hhttp.Context) error {
	s := c.site
	rows := make([]views.ExtensionRow, len(s.Reference.Extensions))
	for i, e := range s.Reference.Extensions {
		item := reference.Item{Kind: reference.Attributes, Name: e.Name}
		rows[i] = views.ExtensionRow{
			Extension: e,
			Guide:     s.ExtensionGuide(item),
			Names:     e.Attributes + e.Events + e.Headers + e.Config + e.Classes,
			Adds:      adds(e),
		}
	}
	chrome := s.Chrome(ctx, "/reference/extensions", "reference", "Extensions reference", "The official YPY extensions, the script of each and the names it adds.")
	return ctx.View("reference", views.ReferenceData{
		Chrome:     chrome,
		Version:    s.YPYVersion(),
		Kind:       "extensions",
		Title:      "Extensions",
		Intro:      reference.Inline("Load an extension with a script tag after the core. Each is one file; an extension that is not on the page costs nothing."),
		Extensions: rows,
	})
}

func valueRows(in []reference.Value) []views.ValueRow {
	out := make([]views.ValueRow, len(in))
	for i, v := range in {
		out[i] = views.ValueRow{Name: v.Name, Text: reference.Inline(v.Description)}
	}
	return out
}

var _ = http.StatusOK

// adds says in words what an extension adds to the tables.
func adds(e reference.Extension) string {
	var parts []string
	for _, n := range []struct {
		count int
		noun  string
	}{{e.Attributes, "attributes"}, {e.Events, "events"}, {e.Headers, "headers"}, {e.Config, "config keys"}, {e.Classes, "classes"}} {
		if n.count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n.count, n.noun))
		}
	}
	if len(parts) == 0 {
		return "behaviour only"
	}
	return strings.Join(parts, ", ")
}
