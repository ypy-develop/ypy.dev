package unit_test

import (
	"strings"
	"testing"
	"testing/fstest"

	markdown "github.com/hyz-is/arandu-markdown"

	site "github.com/ypy-develop/ypy.dev/app/Site"
)

// With content/ypy/reference.md present, /reference is that page, the tables of
// YPY's names sit where the markers are, and the generated pages are gone from
// the catalogue.
func TestTheLongReferenceCarriesTheGeneratedTables(t *testing.T) {
	md, err := markdown.New(markdown.Config{})
	if err != nil {
		t.Fatal(err)
	}
	body := "# Reference\n\n"
	for _, name := range site.Markers {
		body += "## " + strings.ToUpper(name[:1]) + name[1:] + "\n\n<!-- reference:" + name + " -->\n\n"
	}
	s, err := site.New(site.Config{
		Name: "YPY", URL: "https://ypy.dev", Indexing: true,
		Content: fstest.MapFS{
			"ypy/index.md":     {Data: []byte("# Docs\n\n## Install\n\nText.\n")},
			"ypy/reference.md": {Data: []byte(body)},
		},
	}, md)
	if err != nil {
		t.Fatal(err)
	}
	if !s.LongReference() {
		t.Fatal("content/ypy/reference.md is not the reference")
	}
	p, _ := s.Manual.Page("/reference")
	for _, want := range []string{`class="ref-table"`, `<code>ypy-get</code>`, `id="extensions"`} {
		if !strings.Contains(p.HTML, want) {
			t.Errorf("the reference page lacks %s", want)
		}
	}
	if strings.Contains(p.HTML, "@@reference") {
		t.Error("a marker was left on the page")
	}
	for _, e := range s.Entries() {
		if strings.HasPrefix(e.Path, "/reference/") {
			t.Errorf("the catalogue still lists the generated page %s", e.Path)
		}
	}
	if to, ok := s.Manual.Redirect("/reference/events"); !ok || to != "/reference#events" {
		t.Errorf("/reference/events -> %q, %v", to, ok)
	}
}
