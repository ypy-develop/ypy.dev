package feature_test

import (
	"encoding/json"
	"encoding/xml"
	"html"
	"image"
	_ "image/png"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/arandu-io/hesape/config"

	"github.com/ypy-develop/ypy.dev/public"
	"github.com/ypy-develop/ypy.dev/tests"
)

const origin = "https://ypy.dev"

var (
	graphPattern = regexp.MustCompile(`(?s)<script type="application/ld\+json" data-page-schema>(.*?)</script>`)
	metaPattern  = regexp.MustCompile(`<meta (?:name|property)="([^"]+)" content="([^"]*)"`)
	linkPattern  = regexp.MustCompile(`<link rel="([^"]+)"([^>]*)>`)
)

// metas is every meta tag of a page by its name or property; a name that
// repeats (og:locale:alternate) keeps its first value.
func metas(body string) map[string]string {
	out := map[string]string{}
	for _, m := range metaPattern.FindAllStringSubmatch(body, -1) {
		if _, ok := out[m[1]]; !ok {
			out[m[1]] = html.UnescapeString(m[2])
		}
	}
	return out
}

// attr reads an attribute out of the text of a tag.
func attr(tag, name string) string {
	m := regexp.MustCompile(name + `="([^"]*)"`).FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	return html.UnescapeString(m[1])
}

// links is the href of every <link rel=...> of a page with the given rel.
func links(body, rel string) []string {
	var out []string
	for _, m := range linkPattern.FindAllStringSubmatch(body, -1) {
		if m[1] == rel {
			out = append(out, attr(m[2], "href"))
		}
	}
	return out
}

// graph decodes the JSON-LD block of a page and returns its entities by type.
func graph(t *testing.T, path, body string) map[string]map[string]any {
	t.Helper()
	blocks := graphPattern.FindAllStringSubmatch(body, -1)
	if len(blocks) != 1 {
		t.Fatalf("%s carries %d JSON-LD blocks, want exactly one", path, len(blocks))
	}
	var doc struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(blocks[0][1]), &doc); err != nil {
		t.Fatalf("%s: the JSON-LD is not valid JSON: %v", path, err)
	}
	if doc.Context != "https://schema.org" {
		t.Errorf("%s: @context is %q", path, doc.Context)
	}
	out := map[string]map[string]any{}
	for _, n := range doc.Graph {
		switch typ := n["@type"].(type) {
		case string:
			out[typ] = n
		case []any:
			for _, v := range typ {
				out[v.(string)] = n
			}
		}
		if n["@id"] == nil {
			t.Errorf("%s: an entity of the graph has no @id: %v", path, n)
		}
	}
	return out
}

// The four surfaces answer in production with what each is for.
func TestDiscoverySurfacesInProduction(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	robots := app.Get("/robots.txt")
	if robots.Code != http.StatusOK || !strings.HasPrefix(robots.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("robots.txt: %d %s", robots.Code, robots.Header().Get("Content-Type"))
	}
	for _, want := range []string{"User-agent: *", "Allow: /\n", "Disallow: /search\n", "Disallow: /demo\n", "Sitemap: " + origin + "/sitemap.xml"} {
		if !strings.Contains(robots.Body.String(), want) {
			t.Errorf("robots.txt lacks %q:\n%s", want, robots.Body)
		}
	}
	if strings.Contains(robots.Body.String(), "Disallow: /\n") {
		t.Error("robots.txt in production disallows the whole site")
	}
	// The stylesheet and the script are what a crawler needs to draw the page.
	if strings.Contains(robots.Body.String(), "/_ypy") {
		t.Error("robots.txt blocks the assets a crawler needs to render the pages")
	}

	sitemap := app.Get("/sitemap.xml")
	if sitemap.Code != http.StatusOK || !strings.HasPrefix(sitemap.Header().Get("Content-Type"), "application/xml") {
		t.Fatalf("sitemap.xml: %d %s", sitemap.Code, sitemap.Header().Get("Content-Type"))
	}
	var set struct {
		URLs []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(sitemap.Body.Bytes(), &set); err != nil {
		t.Fatalf("sitemap.xml is not XML: %v", err)
	}
	dated := 0
	for _, u := range set.URLs {
		if !strings.HasPrefix(u.Loc, origin+"/") {
			t.Errorf("sitemap address %q is not under %s", u.Loc, origin)
		}
		if u.LastMod != "" {
			dated++
		}
	}
	if len(set.URLs) < 100 || dated == 0 {
		t.Errorf("sitemap lists %d addresses, %d with a lastmod", len(set.URLs), dated)
	}

	llms := app.Get("/llms.txt")
	if llms.Code != http.StatusOK || !strings.HasPrefix(llms.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("llms.txt: %d %s", llms.Code, llms.Header().Get("Content-Type"))
	}
	body := llms.Body.String()
	if !strings.HasPrefix(body, "# YPY\n\n> YPY is a hypermedia framework") {
		t.Errorf("llms.txt does not open with the name and the summary:\n%.300s", body)
	}
	for _, want := range []string{"## Reference", "## Docs", "## Guides", "## Server", "## Examples", "## Extensions", "## Migration", "## Go SDK", "## Optional",
		"](" + origin + "/reference/attributes)", "](" + origin + "/docs/guides/getting-started)", "](" + origin + "/llms-full.txt)"} {
		if !strings.Contains(body, want) {
			t.Errorf("llms.txt lacks %q", want)
		}
	}

	full := app.Get("/llms-full.txt")
	if full.Code != http.StatusOK || !strings.HasPrefix(full.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("llms-full.txt: %d %s", full.Code, full.Header().Get("Content-Type"))
	}
	for _, want := range []string{"# YPY documentation, in full", "Canonical URL: " + origin + "/docs/guides/getting-started", "Canonical URL: " + origin + "/reference/attributes", "### Public page content"} {
		if !strings.Contains(full.Body.String(), want) {
			t.Errorf("llms-full.txt lacks %q", want)
		}
	}
}

// Anything that is not production is closed: crawlers are turned away, the
// indexes list nothing and every page says noindex.
func TestDiscoveryIsClosedOutsideProduction(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, false)

	robots := app.Get("/robots.txt").Body.String()
	if !strings.Contains(robots, "Disallow: /\n") || strings.Contains(robots, "Allow:") || strings.Contains(robots, "Sitemap:") {
		t.Errorf("robots.txt of a closed deployment:\n%s", robots)
	}
	if strings.Contains(app.Get("/sitemap.xml").Body.String(), "<loc>") {
		t.Error("the sitemap of a closed deployment lists a page")
	}
	for _, path := range []string{"/llms.txt", "/llms-full.txt"} {
		body := app.Get(path).Body.String()
		if !strings.Contains(body, "Discovery disabled") || strings.Contains(body, origin+"/docs") {
			t.Errorf("%s of a closed deployment:\n%.300s", path, body)
		}
	}

	for _, path := range []string{"/", "/docs/guides/getting-started", "/reference/attributes"} {
		rec := app.Get(path)
		body := rec.Body.String()
		m := metas(body)
		if m["robots"] != "noindex, nofollow" {
			t.Errorf("%s: robots is %q on a closed deployment", path, m["robots"])
		}
		if rec.Header().Get("X-Robots-Tag") != "noindex, nofollow" {
			t.Errorf("%s: X-Robots-Tag is %q on a closed deployment", path, rec.Header().Get("X-Robots-Tag"))
		}
		if len(links(body, "canonical")) != 0 || m["og:image"] != "" || strings.Contains(body, "ld+json") {
			t.Errorf("%s: a closed deployment still names a canonical address, a card or a graph", path)
		}
	}
}

// The head of the front page, a page of the manual and a reference page has the
// title, the description, an absolute canonical address, a 1200 by 630 card that
// is a real image and one valid graph.
func TestHeadOfThreePages(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	for _, path := range []string{"/", "/docs/guides/getting-started", "/reference/attributes"} {
		rec := app.Get(path)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: %d", path, rec.Code)
		}
		body := rec.Body.String()
		m := metas(body)

		if title := regexp.MustCompile(`<title>([^<]*)</title>`).FindStringSubmatch(body); title == nil || (path != "/" && !strings.HasSuffix(html.UnescapeString(title[1]), " · YPY")) {
			t.Errorf("%s: title %v does not follow \"<title> · YPY\"", path, title)
		}
		if m["description"] == "" || m["og:description"] != m["description"] || m["twitter:description"] != m["description"] {
			t.Errorf("%s: description tags are missing or disagree", path)
		}
		want := origin + path
		if canon := links(body, "canonical"); len(canon) != 1 || canon[0] != want || m["og:url"] != want {
			t.Errorf("%s: canonical %v and og:url %q, want %s", path, canon, m["og:url"], want)
		}
		if m["robots"] != "index, follow, max-image-preview:large" {
			t.Errorf("%s: robots %q", path, m["robots"])
		}
		if m["og:site_name"] != "YPY" || m["og:type"] != "website" || m["og:title"] == "" || m["twitter:card"] != "summary_large_image" {
			t.Errorf("%s: Open Graph or Twitter tags incomplete: %v", path, m)
		}
		if !strings.HasPrefix(m["og:image"], origin+"/social-cover") || m["twitter:image"] != m["og:image"] {
			t.Errorf("%s: og:image %q is not an absolute address of a cover", path, m["og:image"])
		}
		if m["og:image:width"] != "1200" || m["og:image:height"] != "630" || m["og:image:type"] != "image/png" || m["og:image:alt"] == "" {
			t.Errorf("%s: the card's size, type or alt is missing: %v", path, m)
		}
		if got := links(body, "sitemap"); len(got) != 1 || got[0] != "/sitemap.xml" {
			t.Errorf("%s: rel=sitemap is %v", path, got)
		}
		if got := links(body, "describedby"); len(got) != 1 || got[0] != "/llms.txt" {
			t.Errorf("%s: rel=describedby is %v", path, got)
		}

		// The card is a real PNG of the size the tags state.
		cover := app.Get(strings.TrimPrefix(m["og:image"], origin))
		cfg, format, err := image.DecodeConfig(strings.NewReader(cover.Body.String()))
		if cover.Code != http.StatusOK || err != nil || format != "png" || cfg.Width != 1200 || cfg.Height != 630 {
			t.Errorf("%s: the card %s is %d %s %dx%d (%v)", path, m["og:image"], cover.Code, format, cfg.Width, cfg.Height, err)
		}

		g := graph(t, path, body)
		for _, typ := range []string{"WebSite", "WebPage", "Organization"} {
			if g[typ] == nil {
				t.Errorf("%s: the graph has no %s", path, typ)
			}
		}
		if g["WebPage"]["url"] != want {
			t.Errorf("%s: WebPage url is %v", path, g["WebPage"]["url"])
		}
		if _, ok := g["WebSite"]["potentialAction"]; !ok {
			t.Errorf("%s: the site's search is not declared", path)
		}
		sameAs, _ := g["Organization"]["sameAs"].([]any)
		if len(sameAs) != 2 || g["Organization"]["name"] != "HYZIS" {
			t.Errorf("%s: the publisher is %v", path, g["Organization"])
		}
		for typ := range g {
			if typ == "AggregateRating" || typ == "Review" || typ == "Offer" {
				t.Errorf("%s: the graph states %s, which the page does not show", path, typ)
			}
		}

		switch path {
		case "/":
			sw := g["SoftwareApplication"]
			if sw == nil || sw["license"] != "https://opensource.org/licenses/MIT" || sw["codeRepository"] != "https://github.com/ypy-develop/ypy" ||
				sw["programmingLanguage"] != "JavaScript" || sw["name"] != "YPY" || sw["version"] != app.Provider.Site.YPYVersion() {
				t.Errorf("/: the software entity is %v", sw)
			}
			if g["SoftwareSourceCode"] == nil {
				t.Error("/: the software is not also source code")
			}
		case "/docs/guides/getting-started":
			if g["TechArticle"] == nil || g["TechArticle"]["headline"] == "" {
				t.Errorf("%s: no TechArticle: %v", path, g)
			}
			crumbs := g["BreadcrumbList"]
			items, _ := crumbs["itemListElement"].([]any)
			if len(items) < 2 {
				t.Fatalf("%s: the trail is %v", path, crumbs)
			}
			last := items[len(items)-1].(map[string]any)
			if last["item"] != want || int(last["position"].(float64)) != len(items) {
				t.Errorf("%s: the trail does not end at the page: %v", path, last)
			}
			// The trail is the one the page shows.
			for _, it := range items {
				name := it.(map[string]any)["name"].(string)
				if !strings.Contains(body, ">"+html.EscapeString(name)+"</") {
					t.Errorf("%s: the trail names %q, which the page does not show", path, name)
				}
			}
		}
	}
}

// Every page of the catalogue is in the sitemap and answers 200 with itself as
// its canonical address; the sitemap lists nothing the catalogue does not.
func TestEveryCataloguePageIsInTheSitemap(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	sitemap := app.Get("/sitemap.xml").Body.String()

	entries := app.Provider.Site.Entries()
	if len(entries) < 150 {
		t.Fatalf("the catalogue has only %d pages", len(entries))
	}
	want := map[string]bool{}
	for _, e := range entries {
		loc := origin + e.Path
		want[loc] = true
		if !strings.Contains(sitemap, "<loc>"+loc+"</loc>") {
			t.Errorf("%s is in the catalogue and not in the sitemap", e.Path)
		}
		if e.Title == "" || e.Description == "" {
			t.Errorf("%s has no title or no description in the catalogue", e.Path)
		}
		rec := app.Get(e.Path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: %d", e.Path, rec.Code)
			continue
		}
		if canon := links(rec.Body.String(), "canonical"); len(canon) != 1 || canon[0] != loc {
			t.Errorf("%s: canonical %v", e.Path, canon)
		}
	}
	var set struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal([]byte(sitemap), &set); err != nil {
		t.Fatal(err)
	}
	if len(set.URLs) != len(entries) {
		t.Errorf("the sitemap lists %d addresses, the catalogue %d pages", len(set.URLs), len(entries))
	}
	for _, u := range set.URLs {
		if !want[u.Loc] {
			t.Errorf("the sitemap lists %s, which is not in the catalogue", u.Loc)
		}
	}
}

// A page that must not be indexed says so in the markup and in the header, and
// names no canonical address, no card and no graph.
func TestPagesThatAreNotIndexed(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	for path, status := range map[string]int{"/nope": http.StatusNotFound, "/docs/nope": http.StatusNotFound, "/search?q=swap": http.StatusOK} {
		rec := app.Get(path)
		body := rec.Body.String()
		m := metas(body)
		if rec.Code != status {
			t.Errorf("GET %s: %d, want %d", path, rec.Code, status)
		}
		if m["robots"] != "noindex, nofollow" || rec.Header().Get("X-Robots-Tag") != "noindex, nofollow" {
			t.Errorf("%s: robots %q, header %q", path, m["robots"], rec.Header().Get("X-Robots-Tag"))
		}
		if len(links(body, "canonical")) != 0 || m["og:url"] != "" || m["og:image"] != "" || strings.Contains(body, "ld+json") {
			t.Errorf("%s: a page that is not indexed names a canonical address, a card or a graph", path)
		}
	}
}

// Every language alternate of a page names the page back, itself included, and
// the cluster has an x-default: a cluster that is not reciprocal is discarded.
func TestLanguageAlternatesAreReciprocal(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	clusters := 0
	for _, e := range app.Provider.Site.Entries() {
		if e.Section == "" {
			continue
		}
		body := app.Get(e.Path).Body.String()
		alts := alternates(body)
		if len(alts) == 0 {
			continue
		}
		clusters++
		if alts["x-default"] == "" || alts[e.Lang] != origin+e.Path {
			t.Errorf("%s: the alternates %v lack x-default or the page itself", e.Path, alts)
		}
		for tag, href := range alts {
			if tag == "x-default" {
				continue
			}
			other := alternates(app.Get(strings.TrimPrefix(href, origin)).Body.String())
			if other[e.Lang] != origin+e.Path {
				t.Errorf("%s names %s as %s, which does not name it back: %v", e.Path, href, tag, other)
			}
		}
	}
	if clusters == 0 {
		t.Error("no page has a Portuguese twin, so nothing was checked")
	}
}

func alternates(body string) map[string]string {
	out := map[string]string{}
	for _, m := range linkPattern.FindAllStringSubmatch(body, -1) {
		if m[1] == "alternate" {
			out[attr(m[2], "hreflang")] = attr(m[2], "href")
		}
	}
	return out
}

// The covers and the icons are served with the type and the size they claim, and
// the manifest names the icons it declares.
func TestCoversIconsAndManifest(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)

	for _, c := range public.Covers() {
		rec := app.Get(c.Path)
		cfg, format, err := image.DecodeConfig(strings.NewReader(rec.Body.String()))
		if rec.Code != http.StatusOK || err != nil || format != "png" || cfg.Width != c.Width || cfg.Height != c.Height {
			t.Errorf("cover %s: %d %s %dx%d (%v)", c.Path, rec.Code, format, cfg.Width, cfg.Height, err)
		}
		if !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
			t.Errorf("cover %s is not cached for good: %q", c.Path, rec.Header().Get("Cache-Control"))
		}
	}

	for path, size := range map[string]int{
		"/favicon-16x16.png": 16, "/favicon-32x32.png": 32, "/apple-touch-icon.png": 180,
		"/android-chrome-192x192.png": 192, "/android-chrome-512x512.png": 512,
	} {
		rec := app.Get(path)
		cfg, _, err := image.DecodeConfig(strings.NewReader(rec.Body.String()))
		if rec.Code != http.StatusOK || err != nil || cfg.Width != size || cfg.Height != size {
			t.Errorf("%s: %d %dx%d (%v), want %d square", path, rec.Code, cfg.Width, cfg.Height, err, size)
		}
	}
	if rec := app.Get("/favicon.ico"); rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/x-icon" || len(rec.Body.Bytes()) < 100 {
		t.Errorf("/favicon.ico: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec := app.Get("/favicon.svg"); rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Errorf("/favicon.svg: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}

	rec := app.Get("/site.webmanifest")
	var manifest struct {
		Name  string `json:"name"`
		Icons []struct {
			Src   string `json:"src"`
			Sizes string `json:"sizes"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &manifest); err != nil || manifest.Name != "YPY" || len(manifest.Icons) != 2 {
		t.Fatalf("site.webmanifest: %v %+v", err, manifest)
	}
	for _, icon := range manifest.Icons {
		if got := app.Get(icon.Src); got.Code != http.StatusOK {
			t.Errorf("the manifest names %s, which answers %d", icon.Src, got.Code)
		}
	}

	// The head names them all.
	body := app.Get("/").Body.String()
	for _, want := range []string{`href="/favicon.ico"`, `href="/favicon-32x32.png"`, `href="/favicon-16x16.png"`, `rel="apple-touch-icon" href="/apple-touch-icon.png"`, `rel="manifest" href="/site.webmanifest"`} {
		if !strings.Contains(body, want) {
			t.Errorf("the head lacks %s", want)
		}
	}
}
