package unit_test

import (
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
	if prev == nil || prev.Slug != "index" || next == nil || next.Slug != "a-early" {
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
