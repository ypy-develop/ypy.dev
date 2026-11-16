// Package reference holds the name tables of YPY -- attributes, events,
// headers, configuration keys, classes and elements, of the core and of every
// official extension -- and the extensions themselves.
//
// The tables are not written here. They are a copy of YPY's own single source
// (src/attributes.json and src/attributes.d/*.json), taken by `go generate` and
// embedded, so the site documents exactly the names the library has and never
// reads another repository while it runs.
package reference

//go:generate go run ../../tools/ypysync

import (
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"sort"
	"strings"
)

//go:embed data
var data embed.FS

// Kind is one of the tables.
type Kind string

const (
	Attributes Kind = "attributes"
	Events     Kind = "events"
	Headers    Kind = "headers"
	Config     Kind = "config"
	Classes    Kind = "classes"
	Elements   Kind = "elements"
)

// Kinds is the order the tables are shown in.
var Kinds = []Kind{Attributes, Events, Headers, Config, Classes, Elements}

// Titles are the names of the pages.
var Titles = map[Kind]string{
	Attributes: "Attributes",
	Events:     "Events",
	Headers:    "Request and response headers",
	Config:     "Configuration",
	Classes:    "CSS classes",
	Elements:   "Elements",
}

// Value is one accepted value or modifier, with what it does.
type Value struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Item is one row of a table.
type Item struct {
	Kind        Kind
	Name        string
	Anchor      string
	Description string
	// Extension is the extension that adds the name, empty for the core.
	Extension string
	// Doc is the address of the page that explains it, on this site or not.
	Doc string

	// Attributes.
	ValueType  string
	Syntax     string
	Values     []Value
	Modifiers  []Value
	Inherited  bool
	Deprecated string

	// Events.
	Cancelable bool
	Target     string
	Detail     string

	// Headers.
	Direction string
	Value     string

	// Configuration.
	Type    string
	Default string

	// Elements.
	Attrs []string
}

// Extension is one official extension and what it adds.
type Extension struct {
	Name       string
	Script     string
	Attributes int
	Events     int
	Headers    int
	Config     int
	Classes    int
}

// Catalog is every table, merged from the core and the extensions.
type Catalog struct {
	// Version is YPY's version, from its VERSION file.
	Version    string
	Items      map[Kind][]Item
	Extensions []Extension
}

// Of returns the rows of one table, in the order the source lists them: the
// core first, then each extension by name.
func (c *Catalog) Of(k Kind) []Item { return c.Items[k] }

// Total is the number of rows of every table.
func (c *Catalog) Total() int {
	n := 0
	for _, items := range c.Items {
		n += len(items)
	}
	return n
}

type source struct {
	Extension string `json:"extension"`

	Elements []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Attributes  []string `json:"attributes"`
	} `json:"elements"`

	Attributes []struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ValueType   string  `json:"valueType"`
		Values      []Value `json:"values"`
		Syntax      string  `json:"syntax"`
		Modifiers   []Value `json:"modifiers"`
		Inherited   bool    `json:"inherited"`
		Doc         string  `json:"doc"`
		Deprecated  string  `json:"deprecated"`
	} `json:"attributes"`

	Events []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Detail      json.RawMessage `json:"detail"`
		Cancelable  bool            `json:"cancelable"`
		Target      string          `json:"target"`
		Doc         string          `json:"doc"`
	} `json:"events"`

	Headers []struct {
		Name        string `json:"name"`
		Direction   string `json:"direction"`
		Description string `json:"description"`
		Value       string `json:"value"`
	} `json:"headers"`

	Config []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Schema      json.RawMessage `json:"schema"`
		Default     json.RawMessage `json:"default"`
		Deprecated  string          `json:"deprecated"`
	} `json:"config"`

	Classes []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"classes"`
}

// Load reads the embedded tables.
func Load() (*Catalog, error) {
	return load(data, "data")
}

func load(fsys fs.FS, dir string) (*Catalog, error) {
	c := &Catalog{Items: map[Kind][]Item{}}

	version, err := fs.ReadFile(fsys, dir+"/VERSION")
	if err != nil {
		return nil, fmt.Errorf("reference: %w", err)
	}
	c.Version = strings.TrimSpace(string(version))

	files := []string{dir + "/attributes.json"}
	parts, err := fs.Glob(fsys, dir+"/attributes.d/*.json")
	if err != nil {
		return nil, err
	}
	sort.Strings(parts)
	files = append(files, parts...)

	for _, name := range files {
		raw, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("reference: %w", err)
		}
		var s source
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("reference: %s: %w", name, err)
		}
		c.add(s)
	}
	return c, nil
}

func (c *Catalog) add(s source) {
	ext := s.Extension
	add := func(it Item) {
		it.Extension = ext
		it.Anchor = Anchor(it.Name)
		it.Doc = siteAddress(it.Doc)
		c.Items[it.Kind] = append(c.Items[it.Kind], it)
	}

	for _, e := range s.Elements {
		add(Item{Kind: Elements, Name: e.Name, Description: e.Description, Attrs: e.Attributes})
	}
	for _, a := range s.Attributes {
		add(Item{
			Kind: Attributes, Name: a.Name, Description: a.Description, ValueType: a.ValueType,
			Values: a.Values, Syntax: a.Syntax, Modifiers: a.Modifiers, Inherited: a.Inherited,
			Doc: a.Doc, Deprecated: a.Deprecated,
		})
	}
	for _, e := range s.Events {
		add(Item{
			Kind: Events, Name: e.Name, Description: e.Description, Cancelable: e.Cancelable,
			Target: e.Target, Detail: detail(e.Detail), Doc: e.Doc,
		})
	}
	for _, h := range s.Headers {
		add(Item{Kind: Headers, Name: h.Name, Description: h.Description, Direction: h.Direction, Value: h.Value})
	}
	for _, k := range s.Config {
		add(Item{
			Kind: Config, Name: k.Name, Description: k.Description, Type: schemaType(k.Schema),
			Default: literal(k.Default), Deprecated: k.Deprecated,
		})
	}
	for _, k := range s.Classes {
		add(Item{Kind: Classes, Name: k.Name, Description: k.Description})
	}

	if ext != "" {
		c.Extensions = append(c.Extensions, Extension{
			Name:       ext,
			Script:     "ext/" + ext + ".min.js",
			Attributes: len(s.Attributes),
			Events:     len(s.Events),
			Headers:    len(s.Headers),
			Config:     len(s.Config),
			Classes:    len(s.Classes),
		})
	}
}

// siteAddress turns an address on ypy.dev into a path on this site.
func siteAddress(doc string) string {
	return strings.TrimPrefix(doc, "https://ypy.dev")
}

// detail writes the shape of an event's detail object on one line.
func detail(raw json.RawMessage) string {
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil || len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s: %v", k, m[k])
	}
	return strings.Join(parts, ", ")
}

// schemaType names the type a configuration key takes.
func schemaType(raw json.RawMessage) string {
	var s struct {
		Type  any   `json:"type"`
		Enum  []any `json:"enum"`
		OneOf []struct {
			Type any `json:"type"`
		} `json:"oneOf"`
	}
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	if len(s.Enum) > 0 {
		vals := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			vals[i] = fmt.Sprintf("%q", fmt.Sprint(v))
		}
		return strings.Join(vals, " | ")
	}
	switch t := s.Type.(type) {
	case string:
		return t
	case []any:
		parts := make([]string, len(t))
		for i, v := range t {
			parts[i] = fmt.Sprint(v)
		}
		return strings.Join(parts, " | ")
	}
	return ""
}

func literal(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return string(raw)
	}
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return strings.TrimSpace(string(raw))
}

// Anchor makes the fragment of a name: lower case, letters and digits, the rest
// a hyphen.
func Anchor(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// Inline writes a description for a page: escaped, with `code` spans as <code>.
// The descriptions are short sentences with backticks and nothing else, so this
// is all the markup they need, and it is the only place they become HTML.
func Inline(s string) string {
	var b strings.Builder
	parts := strings.Split(s, "`")
	for i, p := range parts {
		escaped := html.EscapeString(p)
		if i%2 == 1 && i < len(parts)-1 {
			b.WriteString("<code>" + escaped + "</code>")
			continue
		}
		if i%2 == 1 {
			// An unmatched backtick: keep it as it was written.
			b.WriteString("`" + escaped)
			continue
		}
		b.WriteString(escaped)
	}
	return b.String()
}
