package site

import (
	"fmt"
	"strings"

	reference "github.com/ypy-develop/ypy.dev/app/Reference"
	partials "github.com/ypy-develop/ypy.dev/storage/framework/views/partials"
)

// Markers are the names a page of content/ can ask a table for, written as a
// line of its own: <!-- reference:attributes -->.
//
//	attributes  events  classes  config  elements
//	request-headers  response-headers  extensions  js-api
//
// js-api has no name table behind it yet, so it draws nothing; the prose around
// the marker is the page's own.
var Markers = []string{"attributes", "events", "classes", "config", "elements", "request-headers", "response-headers", "extensions", "js-api"}

// LongReference says whether the reference is one long page of content/ (the
// file content/ypy/reference.md) and not the generated pages under /reference.
func (s *Site) LongReference() bool { return s.longReference }

// Table is the HTML a marker is replaced with.
func (s *Site) Table(name string) (string, error) {
	var kind reference.Kind
	direction := ""
	switch name {
	case "attributes":
		kind = reference.Attributes
	case "events":
		kind = reference.Events
	case "classes":
		kind = reference.Classes
	case "config":
		kind = reference.Config
	case "elements":
		kind = reference.Elements
	case "request-headers":
		kind, direction = reference.Headers, "request"
	case "response-headers":
		kind, direction = reference.Headers, "response"
	case "js-api":
		return "", nil
	case "extensions":
		return s.extensionsTable()
	default:
		return "", fmt.Errorf("no table is called %q (known: %s)", name, strings.Join(Markers, ", "))
	}

	var rows []partials.ReferenceRow
	for _, it := range s.Reference.Of(kind) {
		if direction != "" && it.Direction != direction {
			continue
		}
		r := partials.ReferenceRow{
			Name: it.Name, Anchor: it.Anchor, Extension: it.Extension, Guide: s.DocFor(it),
			Text: reference.Inline(it.Description), ValueText: reference.Inline(it.Value),
			Syntax: it.Syntax, ValueType: it.ValueType, Type: it.Type, Default: it.Default,
			Detail: it.Detail, Target: it.Target, Value: it.Value, Deprecated: it.Deprecated,
			Attrs: strings.Join(it.Attrs, ", "), Cancelable: it.Cancelable, Inherited: it.Inherited,
		}
		if direction == "" {
			r.Direction = it.Direction
		}
		for _, v := range it.Values {
			r.Values = append(r.Values, partials.ReferenceValue{Name: v.Name, Text: reference.Inline(v.Description)})
		}
		for _, v := range it.Modifiers {
			r.Modifiers = append(r.Modifiers, partials.ReferenceValue{Name: v.Name, Text: reference.Inline(v.Description)})
		}
		rows = append(rows, r)
	}
	return string(partials.ReferenceTable(partials.ReferenceTableProps{Rows: rows})), nil
}

func (s *Site) extensionsTable() (string, error) {
	var rows []partials.ReferenceExtension
	for _, e := range s.Reference.Extensions {
		var parts []string
		for _, n := range []struct {
			count int
			noun  string
		}{{e.Attributes, "attributes"}, {e.Events, "events"}, {e.Headers, "headers"}, {e.Config, "config keys"}, {e.Classes, "classes"}} {
			if n.count > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", n.count, n.noun))
			}
		}
		adds := strings.Join(parts, ", ")
		if adds == "" {
			adds = "behaviour only"
		}
		rows = append(rows, partials.ReferenceExtension{Name: e.Name, Script: e.Script, Guide: s.ExtensionGuide(reference.Item{Name: e.Name}), Adds: adds})
	}
	return string(partials.ReferenceExtensions(partials.ReferenceExtensionsProps{Rows: rows})), nil
}
