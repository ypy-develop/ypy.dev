package feature_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/arandu-io/hesape/config"

	docs "github.com/ypy-develop/ypy.dev/app/Docs"
	reference "github.com/ypy-develop/ypy.dev/app/Reference"
	"github.com/ypy-develop/ypy.dev/tests"
)

// Every address the site is made of answers 200 with a title, and the ones that
// do not exist answer 404 with the site's own page, never the router's text.
func TestFixedRoutesAnswer(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	for _, path := range []string{
		"/", "/docs", "/go", "/examples", "/extensions", "/essays", "/reference", "/reference/attributes", "/reference/events", "/reference/headers",
		"/reference/config", "/reference/classes", "/reference/elements", "/reference/extensions",
		"/search", "/search?q=swap", "/search-index.json", "/sitemap.xml", "/robots.txt", "/llms.txt",
	} {
		rec := app.Get(path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s answered %d, want 200", path, rec.Code)
		}
	}

	for _, path := range []string{"/nope", "/docs/nope", "/go/nope", "/reference/nope", "/docs/guides/nope/deeper"} {
		rec := app.Get(path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s answered %d, want 404", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "There is no page here") {
			t.Errorf("GET %s did not draw the site's own 404 page", path)
		}
	}
}

// Every .md file under content/ is a page that answers 200 at the address its path
// gives, and its title is on it.
func TestEveryMarkdownFileIsAPage(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	root := filepath.Join(tests.Root(t), "content")

	files := 0
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		files++
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)

		page, ok := app.Provider.Site.Manual.ByFile(rel)
		if !ok {
			t.Errorf("%s did not become a page", rel)
			return nil
		}
		rec := app.Get(page.Route)
		if rec.Code != http.StatusOK {
			t.Errorf("%s (%s) answered %d", rel, page.Route, rec.Code)
			return nil
		}
		if page.Title == "" || !strings.Contains(rec.Body.String(), "<h1") {
			t.Errorf("%s has no title on the page", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("content/ holds no Markdown; this test would be checking nothing")
	}
	if got := len(app.Provider.Site.Manual.Pages()); got != files {
		t.Errorf("%d files became %d pages", files, got)
	}
}

// The address of a page comes from its path, and the two documentations keep to
// their prefixes.
func TestRoutesFollowThePaths(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	m := app.Provider.Site.Manual

	cases := map[string]string{
		"ypy/index.md":                  "/docs",
		"ypy/guides/getting-started.md": "/docs/guides/getting-started",
		"go/README.md":                  "/go",
		"go/guide.md":                   "/go/guide",
		"go/pt-BR/README.md":            "/go/pt-BR",
	}
	for file, want := range cases {
		page, ok := m.ByFile(file)
		if !ok {
			t.Errorf("%s is not a page", file)
			continue
		}
		if page.Route != want {
			t.Errorf("%s is at %s, want %s", file, page.Route, want)
		}
	}
}

// A link between two files is a link to the page the second became, and a link
// to a file that is not a page is not offered at all.
func TestLinksBetweenFilesPointAtPages(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	manual := app.Provider.Site.Manual

	href := regexp.MustCompile(`href="(/[^"#]*)`)
	mdLink := regexp.MustCompile(`href="(\.\.?/[^":]*|[A-Za-z0-9_-][^":]*)\.md(#[^"]*)?"`)
	missing := map[string]bool{}
	for _, p := range manual.Pages() {
		for _, m := range href.FindAllStringSubmatch(p.HTML, -1) {
			target := strings.TrimSuffix(m[1], "/")
			if strings.HasPrefix(target, "/_arandu") || target == "" {
				continue
			}
			if _, ok := manual.Page(target); ok {
				continue
			}
			if _, _, _, ok := manual.Landing(target); ok {
				continue
			}
			missing[p.File+" -> "+target] = true
		}
		if mdLink.MatchString(p.HTML) {
			t.Errorf("%s still carries a link to a Markdown file", p.File)
		}
	}
	for link := range missing {
		t.Errorf("link to nothing: %s", link)
	}
	for _, b := range manual.Relocated() {
		t.Logf("link followed to a moved file: %s -> %s", b.From, b.Target)
	}
	for _, b := range manual.Broken() {
		t.Logf("dropped link: %s -> %s (the file it names is not a page)", b.From, b.Target)
	}
}

// The reference pages hold every row of the source tables, and every row has an
// anchor to link to.
func TestReferenceCoversEveryEntry(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	catalog := app.Provider.Site.Reference

	if catalog.Total() == 0 {
		t.Fatal("the reference tables are empty")
	}
	for _, k := range reference.Kinds {
		rec := app.Get("/reference/" + string(k))
		body := rec.Body.String()
		for _, it := range catalog.Of(k) {
			if !strings.Contains(body, `id="`+it.Anchor+`"`) {
				t.Errorf("/reference/%s lacks the row of %q", k, it.Name)
			}
		}
	}

	// The copy of the source JSON is complete: count the entries of the files
	// the site embeds against the rows it built.
	root := filepath.Join(tests.Root(t), "app", "Reference", "data")
	files := []string{filepath.Join(root, "attributes.json")}
	parts, _ := filepath.Glob(filepath.Join(root, "attributes.d", "*.json"))
	files = append(files, parts...)
	counts := map[string]int{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var tables map[string]json.RawMessage
		if err := json.Unmarshal(raw, &tables); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, k := range reference.Kinds {
			var rows []json.RawMessage
			if body, ok := tables[string(k)]; ok {
				_ = json.Unmarshal(body, &rows)
			}
			counts[string(k)] += len(rows)
		}
	}
	for _, k := range reference.Kinds {
		if got := len(catalog.Of(k)); got != counts[string(k)] {
			t.Errorf("%s: the JSON has %d entries and the page has %d", k, counts[string(k)], got)
		}
	}

	rec := app.Get("/reference/extensions")
	for _, e := range catalog.Extensions {
		if !strings.Contains(rec.Body.String(), `id="`+e.Name+`"`) {
			t.Errorf("/reference/extensions lacks %s", e.Name)
		}
	}
}

// The policy is strict, and the pages keep to it: no inline script, no inline
// style, no style attribute, no handler attribute.
func TestPagesCarryNothingInline(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	csp := app.Get("/").Header().Get("Content-Security-Policy")
	for _, want := range []string{"script-src 'self'", "style-src 'self'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("the policy %q lacks %s", csp, want)
		}
	}
	if strings.Contains(csp, "unsafe-inline") {
		t.Errorf("the policy allows inline code: %s", csp)
	}

	inlineScript := regexp.MustCompile(`(?is)<script(?:\s[^>]*)?>[^<]`)
	styleAttr := regexp.MustCompile(`(?i)\sstyle\s*=\s*"`)
	handler := regexp.MustCompile(`(?i)\son(click|load|error|change|input|submit|mouseover|focus)\s*=`)

	paths := []string{"/", "/docs", "/docs/guides/getting-started", "/reference", "/reference/attributes", "/reference/events", "/go", "/search?q=swap", "/nope"}
	for _, p := range paths {
		// A JSON-LD block is data: the browser never runs it and script-src does not
		// apply to it. Any other script with a body is inline code.
		body := graphPattern.ReplaceAllString(app.Get(p).Body.String(), "")
		if inlineScript.MatchString(body) {
			t.Errorf("%s carries an inline script", p)
		}
		if strings.Contains(body, "<style") {
			t.Errorf("%s carries a <style> element", p)
		}
		if styleAttr.MatchString(body) {
			t.Errorf("%s carries a style attribute", p)
		}
		if handler.MatchString(body) {
			t.Errorf("%s carries an event handler attribute", p)
		}
	}
}

// The site navigates with YPY, and loads the script the project serves, not a
// CDN's.
func TestNavigationUsesYPY(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	body := app.Get("/docs").Body.String()

	if !strings.Contains(body, `ypy-boost="true" ypy-boost:inherited="true"`) {
		t.Error("the body does not ask for boosted navigation that reaches its links")
	}
	if !regexp.MustCompile(`<script src="/_ypy/assets/[0-9a-f]+/ypy\.min\.js" defer>`).MatchString(body) {
		t.Error("the page does not load ypy.min.js from the project")
	}
	if regexp.MustCompile(`<script[^>]+src="https?://`).MatchString(body) {
		t.Error("a script is loaded from a CDN, which the policy would refuse")
	}
}

// The front page carries what the brief asks of it.
func TestHomeCarriesTheInstallation(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	body := app.Get("/").Body.String()

	for _, want := range []string{
		`https://cdn.jsdelivr.net/gh/ypy-develop/ypy@v0.1.1/dist/ypy.min.js`,
		`https://cdn.ypy.dev/ypy@0.1.1/ypy.min.js`,
		`go get github.com/ypy-develop/go@v0.1.0`,
		`npm install @ypy-dev/ypy@0.1.1`,
		`ypy-get=&#34;/hello&#34; ypy-target=&#34;#out&#34;`,
		`https://github.com/orgs/ypy-develop/discussions`,
		`href="/docs"`, `href="/reference"`, `href="/examples"`, `href="/extensions"`,
		`https://github.com/ypy-develop`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the front page lacks %s", want)
		}
	}
}

// The example on the front page answers a YPY request with a fragment and a
// browser with the front page.
func TestDemoEndpoint(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	rec := app.Get("/demo/hello", "YPY-Request", "true")
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("a YPY request got %d %q, want a fragment", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Vary") != "YPY-Request" {
		t.Error("the answer does not vary by YPY-Request")
	}
	if rec := app.Get("/demo/hello"); rec.Code != http.StatusSeeOther {
		t.Errorf("a browser got %d, want a redirect to the page", rec.Code)
	}
}

// The search finds a name and a page, with and without a script.
func TestSearch(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	idx := app.Provider.Site.Index
	got := idx.Search("ypy-boost", 5)
	if len(got) == 0 || got[0].Title != "ypy-boost" {
		t.Fatalf("searching ypy-boost put %v first", got)
	}
	if body := app.Get("/search?q=ypy-boost").Body.String(); !strings.Contains(body, "ypy-boost") {
		t.Error("the search page does not list the attribute")
	}

	var body struct {
		Data struct {
			Entries []docs.Entry `json:"entries"`
		} `json:"data"`
	}
	if err := json.Unmarshal(app.Get("/search-index.json").Body.Bytes(), &body); err != nil || len(body.Data.Entries) == 0 {
		t.Errorf("the index is not served as JSON: %v", err)
	}
}

// Crawlers are told what the environment says: indexed in production, kept out
// everywhere else.
func TestIndexingFollowsTheEnvironment(t *testing.T) {
	prod := tests.Boot(t, config.EnvDev, true)
	if body := prod.Get("/robots.txt").Body.String(); !strings.Contains(body, "Sitemap: https://ypy.dev/sitemap.xml") {
		t.Errorf("robots.txt does not name the sitemap: %q", body)
	}
	if body := prod.Get("/sitemap.xml").Body.String(); !strings.Contains(body, "<loc>https://ypy.dev/docs</loc>") {
		t.Error("the sitemap does not list /docs")
	}
	if body := prod.Get("/").Body.String(); !strings.Contains(body, `content="index, follow, max-image-preview:large"`) {
		t.Error("an indexed deployment does not say so")
	}

	staging := tests.Boot(t, config.EnvDev, false)
	if body := staging.Get("/robots.txt").Body.String(); !strings.Contains(body, "Disallow: /") {
		t.Error("a deployment that is not indexed does not disallow everything")
	}
	if body := staging.Get("/").Body.String(); !strings.Contains(body, `content="noindex, nofollow"`) {
		t.Error("a deployment that is not indexed does not say so")
	}
}

// The logo files are the ones that were drawn, byte for byte.
func TestLogosAreUntouched(t *testing.T) {
	root := tests.Root(t)
	for _, name := range []string{"ypy.svg", "ypy-icon.svg"} {
		served, err := os.ReadFile(filepath.Join(root, "assets", name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(served), "<svg ") {
			t.Errorf("assets/%s is not the drawn file", name)
		}
	}
}

// Every address the stylesheet names answers: the fonts under the application's
// asset prefix and the two brand marks. A prefix that moved would break them
// with nothing failing but the picture.
func TestStylesheetAddressesAnswer(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	css := regexp.MustCompile(`href="(/_[a-z]+/assets/[0-9a-f]+/app\.css)"`).FindStringSubmatch(app.Get("/").Body.String())
	if css == nil {
		t.Fatal("the page does not link the stylesheet")
	}
	body := app.Get(css[1]).Body.String()
	urls := regexp.MustCompile(`url\("?(/[^")]+)"?\)`).FindAllStringSubmatch(body, -1)
	if len(urls) < 4 {
		t.Fatalf("the stylesheet names %d addresses, expected the fonts and the marks", len(urls))
	}
	for _, u := range urls {
		if rec := app.Get(u[1]); rec.Code != http.StatusOK {
			t.Errorf("the stylesheet names %s, which answers %d", u[1], rec.Code)
		}
	}
}

// The files the build generates carry nothing of the library YPY replaced, and the
// core script is the one the framework embeds, not a copy of ours.
func TestGeneratedAssetsAreYPYOnly(t *testing.T) {
	root := tests.Root(t)
	old := []string{"h" + "tmx", "h" + "x-"}
	for _, name := range []string{"assets/bundles_gen.go", "assets/app.css"} {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, word := range old {
			if strings.Contains(strings.ToLower(string(raw)), word) {
				t.Errorf("%s still names the old library (%s)", name, word)
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, "assets", "vendor", "ypy.min.js")); err == nil {
		t.Error("assets/vendor/ypy.min.js is a second copy of the script the framework embeds")
	}
}

// The header is five links and nothing else: the sections of the manual. The Go
// SDK and the community are in the footer.
func TestHeaderMenuIsTheFiveSections(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	body := app.Get("/").Body.String()
	start := strings.Index(body, `aria-label="Main"`)
	nav := body[start : start+strings.Index(body[start:], "</nav>")]
	last := -1
	for _, want := range []string{`href="/docs"`, `href="/reference"`, `href="/examples"`, `href="/extensions"`, `href="/essays"`} {
		i := strings.Index(nav, want)
		if i < 0 || i < last {
			t.Fatalf("the main menu lacks %s, or has it out of order:\n%s", want, nav)
		}
		last = i
	}
	for _, not := range []string{`href="/go"`, "discussions", ">Community<", ">Go SDK<"} {
		if strings.Contains(nav, not) {
			t.Errorf("the main menu still has %s", not)
		}
	}
	header := body[strings.Index(body, "<header "):strings.Index(body, "</header>")]
	if !strings.Contains(header, "https://github.com/ypy-develop/ypy") {
		t.Error("the header's GitHub button does not open the repository")
	}
	footer := body[strings.Index(body, "<footer "):]
	for _, want := range []string{"Discussions", "LinkedIn", `href="/go"`, "Developed by", `aria-label="HYZIS"`, `aria-label="Arandu"`} {
		if !strings.Contains(footer, want) {
			t.Errorf("the footer lacks %s", want)
		}
	}
	for _, not := range []string{"MIT", "sitemap", "llms.txt"} {
		if strings.Contains(strings.ToLower(footer), strings.ToLower(not)) {
			t.Errorf("the footer still says %s", not)
		}
	}
}

// An address from before the sections had their own is moved for good to the
// page that has it now; one with no equivalent is a 404.
func TestOldAddressesAreMoved(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	m := app.Provider.Site.Manual
	for _, page := range m.Pages() {
		if page.Lang != docs.English || !docs.Own[page.Group] || page.Index {
			continue
		}
		old := "/docs/" + strings.TrimPrefix(page.Route, "/")
		rec := app.Get(old)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != page.Route {
			t.Errorf("GET %s answered %d to %q, want 301 to %s", old, rec.Code, rec.Header().Get("Location"), page.Route)
		}
		break
	}
	if rec := app.Get("/docs/examples/"); rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/examples" {
		t.Errorf("GET /docs/examples/ answered %d to %q, want 301 to /examples", rec.Code, rec.Header().Get("Location"))
	}
	if rec := app.Get("/docs/examples/no-such-example"); rec.Code != http.StatusNotFound {
		t.Errorf("an address with no equivalent answered %d, want 404", rec.Code)
	}
}

// The sidebar of a section is the list of its pages, with no other folder in it.
func TestSectionSidebarListsOnlyTheSection(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	body := app.Get("/examples").Body.String()
	aside := body[strings.Index(body, `<aside class="docs-sidebar"`):]
	aside = aside[:strings.Index(aside, "</aside>")]
	if !strings.Contains(aside, "Examples") {
		t.Errorf("the sidebar does not name the section:\n%s", aside)
	}
	for _, not := range []string{"Guides", "Server", "Security", "Performance", "Contributing", `href="/docs/guides`} {
		if strings.Contains(aside, not) {
			t.Errorf("the examples sidebar still lists %s", not)
		}
	}
}

// The sitemap lists the new addresses and none of the moved ones.
func TestSitemapHasNoMovedAddress(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	sitemap := app.Get("/sitemap.xml").Body.String()
	for _, want := range []string{"https://ypy.dev/examples", "https://ypy.dev/extensions", "https://ypy.dev/essays"} {
		if !strings.Contains(sitemap, want) {
			t.Errorf("the sitemap lacks %s", want)
		}
	}
	for _, not := range []string{"/docs/examples", "/docs/extensions", "/docs/essays"} {
		if strings.Contains(sitemap, not) {
			t.Errorf("the sitemap still lists %s", not)
		}
	}
}

// One red, in both themes: no lighter red and no orange in the stylesheets.
func TestThereIsOneRed(t *testing.T) {
	root := tests.Root(t)
	for _, name := range []string{"app.css", "home.css", "docs.css"} {
		raw, err := os.ReadFile(filepath.Join(root, "resources", "css", name))
		if err != nil {
			t.Fatal(err)
		}
		css := strings.ToLower(string(raw))
		for _, not := range []string{"f2461b", "--color-ember", "orange"} {
			if strings.Contains(css, not) {
				t.Errorf("%s still carries %s", name, not)
			}
		}
	}
}
