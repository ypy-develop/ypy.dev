package unit_test

import (
	"errors"
	"testing"
	"testing/fstest"

	markdown "github.com/hyz-is/arandu-markdown"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
)

func manual(t *testing.T, files map[string]string) *docs.Manual {
	t.Helper()
	fsys := fstest.MapFS{}
	for name, body := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(body)}
	}
	md, err := markdown.New(markdown.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m, err := docs.Load(fsys, md)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRoutesTitlesAndOrder(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/index.md":                  "---\ntitle: \"Home\"\n---\n\n# Ignored\n\nIntro.\n",
		"ypy/guides/b-later.md":         "# Later\n\nText.\n",
		"ypy/guides/getting-started.md": "# Getting started\n\nFirst words of the guide.\n\n## Part\n\nMore.\n",
		"ypy/guides/a-early.md":         "---\norder: 5\n---\n# Early\n\nText.\n",
		"go/README.md":                  "# Go SDK\n\nThe SDK.\n",
	})

	if p, ok := m.Page("/docs"); !ok || p.Title != "Home" {
		t.Errorf("/docs = %v, want the page titled Home", p)
	}
	gs, ok := m.Page("/docs/guides/getting-started")
	if !ok || gs.Title != "Getting started" || gs.Description != "First words of the guide." {
		t.Fatalf("getting started = %+v", gs)
	}
	if len(gs.Headings) != 1 || gs.Headings[0].Text != "Part" {
		t.Errorf("headings = %+v", gs.Headings)
	}
	if _, ok := m.Page("/go"); !ok {
		t.Error("go/README.md is not /go")
	}

	var order []string
	for _, g := range m.Tree("docs", docs.English) {
		for _, p := range g.Pages {
			order = append(order, p.Slug)
		}
	}
	want := []string{"index", "getting-started", "a-early", "b-later"}
	if len(order) != len(want) {
		t.Fatalf("order = %v", order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order = %v, want %v", order, want)
			break
		}
	}

	prev, next := m.Neighbours(gs)
	if prev != nil || next == nil || next.Slug != "a-early" {
		t.Errorf("neighbours = %v, %v", prev, next)
	}
}

func TestLinksAreRewrittenAndCodeIsLeftAlone(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/guides/one.md": "# One\n\nSee [two](two.md#part), [home](../index.md) and [gone](missing.md) and [site](https://example.com/x.md).\n\n```md\n[kept](two.md)\n```\n",
		"ypy/guides/two.md": "# Two\n\n## Part\n\nText.\n",
		"ypy/index.md":      "# Home\n\nText.\n",
	})
	p, _ := m.Page("/docs/guides/one")

	for _, want := range []string{`href="/docs/guides/two#part"`, `href="/docs"`, `href="https://example.com/x.md"`, `[kept](two.md)`} {
		if !contains(p.HTML, want) {
			t.Errorf("the page lacks %s:\n%s", want, p.HTML)
		}
	}
	if contains(p.HTML, `missing.md`) && contains(p.HTML, `href="missing.md"`) {
		t.Error("a link to a file that is not a page was kept")
	}
	if len(m.Broken()) != 1 || m.Broken()[0].Target != "missing.md" {
		t.Errorf("broken = %+v", m.Broken())
	}
}

func TestTwoFilesForOneAddressIsAnError(t *testing.T) {
	md, _ := markdown.New(markdown.Config{})
	fsys := fstest.MapFS{
		"ypy/index.md":  {Data: []byte("# A\n")},
		"ypy/README.md": {Data: []byte("# B\n")},
	}
	if _, err := docs.Load(fsys, md); err == nil {
		t.Fatal("two files for /docs were accepted")
	}
}

func TestCodeBlocksAreColouredWithClasses(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/index.md": "# Home\n\n```html\n<button ypy-get=\"/x\">Go</button>\n```\n",
	})
	p, _ := m.Page("/docs")
	for _, want := range []string{`class="chroma"`, `data-copy-text=`, `<span class="nt">button</span>`} {
		if !contains(p.HTML, want) {
			t.Errorf("the code block lacks %s:\n%s", want, p.HTML)
		}
	}
	if contains(p.HTML, ` style=`) {
		t.Error("the highlighted code carries a style attribute")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestALinkToAFileThatMovedIsFollowed(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/guides/one.md":      "# One\n\nSee [two](two.md) and [three](../old/three.md) and [home](index.md).\n",
		"ypy/server/two.md":      "# Two\n",
		"ypy/reference/three.md": "# Three\n",
		"ypy/reference/dup.md":   "# Dup\n",
		"ypy/server/dup.md":      "# Dup too\n",
		"ypy/index.md":           "# Home\n",
		"ypy/guides/lost.md":     "# Lost\n\n[dup](dup.md) and [nothing](nowhere.md)\n",
	})
	p, _ := m.Page("/docs/guides/one")
	for _, want := range []string{`href="/docs/server/two"`, `href="/docs/reference/three"`, `href="/docs"`} {
		if !contains(p.HTML, want) {
			t.Errorf("the page lacks %s:\n%s", want, p.HTML)
		}
	}
	if got := len(m.Relocated()); got != 3 {
		t.Errorf("relocated = %+v, want the three links that moved", m.Relocated())
	}
	// Two files of that name leave no way to know which was meant, so the link goes.
	if got := len(m.Broken()); got != 2 {
		t.Errorf("broken = %+v, want the ambiguous and the missing", m.Broken())
	}
}

// The long pages and the sections of the header have addresses of their own, and
// a file of the old layout is still found by the old address.
func TestLongPagesAndSectionsHaveTheirOwnAddresses(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/index.md":             "# Manual\n\n## Install\n\nText.\n\n### With npm\n\nText.\n",
		"ypy/reference.md":         "# Reference\n\n## Attributes\n\nText.\n",
		"ypy/migration.md":         "# Migration\n\nText.\n",
		"ypy/examples/index.md":    "# Examples\n\nText.\n",
		"ypy/examples/click.md":    "# Click\n\nText.\n",
		"ypy/extensions/ypy-ws.md": "# WebSocket\n\nText.\n",
		"ypy/essays/why.md":        "# Why\n\nText.\n",
		"ypy/guides/forms.md":      "# Forms\n\nText.\n",
	})
	want := map[string]string{
		"ypy/index.md": "/docs", "ypy/reference.md": "/reference", "ypy/migration.md": "/migration",
		"ypy/examples/index.md": "/examples", "ypy/examples/click.md": "/examples/click",
		"ypy/extensions/ypy-ws.md": "/extensions/ypy-ws", "ypy/essays/why.md": "/essays/why",
	}
	for file, route := range want {
		p, ok := m.ByFile(file)
		if !ok || p.Route != route {
			t.Errorf("%s is at %v, want %s", file, p, route)
		}
	}

	// The sidebar of a long page is its headings; of a section, its pages.
	for _, route := range []string{"/docs", "/reference", "/migration"} {
		if p, _ := m.Page(route); !p.Outline || len(m.SectionOf(p)) != 0 {
			t.Errorf("%s should have its own headings for a sidebar", route)
		}
	}
	click, _ := m.Page("/examples/click")
	if click.Outline || len(m.SectionOf(click)) != 1 || m.SectionOf(click)[0].Key != "examples" {
		t.Errorf("an example's sidebar should be the examples, got %+v", m.SectionOf(click))
	}
	if prev, next := m.Neighbours(click); prev == nil || prev.Route != "/examples" || next != nil {
		t.Errorf("neighbours of an example = %v, %v", prev, next)
	}

	redirects := map[string]string{
		"/docs/examples/click":    "/examples/click",
		"/docs/examples":          "/examples",
		"/docs/extensions/ypy-ws": "/extensions/ypy-ws",
		"/docs/essays/why":        "/essays/why",
		"/docs/migration/migrate": "/migration",
		"/docs/guides/install":    "/docs#install",
		"/reference/attributes":   "/reference#attributes",
	}
	for from, to := range redirects {
		if got, ok := m.Redirect(from); !ok || got != to {
			t.Errorf("%s -> %q (%v), want %s", from, got, ok, to)
		}
	}
	for _, from := range []string{"/docs/guides/nope", "/docs/examples/nope-at-all", "/docs/reference/nope", "/reference/nope"} {
		if got, ok := m.Redirect(from); ok {
			t.Errorf("%s redirected to %s, want no equivalent", from, got)
		}
	}
}

// A marker line becomes the table generated from the name tables, and a marker
// no table answers stops the start.
func TestReferenceMarkersAreFilled(t *testing.T) {
	m := manual(t, map[string]string{
		"ypy/reference.md": "# Reference\n\n## Attributes\n\nBefore.\n\n<!-- reference:attributes -->\n\n## Events\n\n<!-- reference:events -->\n",
	})
	if err := m.Fill(func(name string) (string, error) { return "<table data-t=\"" + name + "\"></table>", nil }); err != nil {
		t.Fatal(err)
	}
	p, _ := m.Page("/reference")
	for _, want := range []string{`<table data-t="attributes">`, `<table data-t="events">`} {
		if !contains(p.HTML, want) {
			t.Errorf("the page lacks %s:\n%s", want, p.HTML)
		}
	}
	if contains(p.HTML, "@@reference") || contains(p.Text, "@@reference") {
		t.Error("a marker token was left on the page")
	}
	if len(p.Headings) != 2 {
		t.Errorf("headings = %+v, want Attributes and Events", p.Headings)
	}

	m = manual(t, map[string]string{"ypy/reference.md": "# R\n\n<!-- reference:nothing -->\n"})
	if err := m.Fill(func(name string) (string, error) { return "", errors.New("no table") }); err == nil {
		t.Error("a marker with no table was accepted")
	}
}
