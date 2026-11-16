package feature_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/arandu-io/hesape/config"

	"github.com/ypy-develop/ypy.dev/tests"
)

// visitor is one browser: it keeps the cookies the site sets and sends them back,
// the way a real visitor does, so two visitors can be told apart.
type visitor struct {
	t       *testing.T
	app     tests.Booted
	cookies map[string]string
	token   string
	// header, when set, is sent as X-CSRF-Token, as ypy-headers does for a DELETE.
	header string
}

func newVisitor(t *testing.T, app tests.Booted) *visitor {
	return &visitor{t: t, app: app, cookies: map[string]string{}}
}

// do sends a request as a YPY request would be sent and keeps the cookies.
func (v *visitor) do(method, target string, form url.Values) *httptest.ResponseRecorder {
	v.t.Helper()
	var body *strings.Reader
	if form != nil && method != http.MethodGet && method != http.MethodDelete {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	if form != nil && (method == http.MethodGet || method == http.MethodDelete) {
		sep := "?"
		if strings.Contains(target, "?") {
			sep = "&"
		}
		target += sep + form.Encode()
	}
	req := httptest.NewRequest(method, "https://ypy.dev"+target, body)
	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("YPY-Request", "true")
	if v.header != "" {
		req.Header.Set("X-CSRF-Token", v.header)
	}
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	for k, val := range v.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: val})
	}
	rec := httptest.NewRecorder()
	v.app.Handler.ServeHTTP(rec, req)
	for _, c := range rec.Result().Cookies() {
		if c.MaxAge < 0 || c.Value == "" {
			delete(v.cookies, c.Name)
			continue
		}
		v.cookies[c.Name] = c.Value
	}
	if m := tokenField.FindStringSubmatch(rec.Body.String()); m != nil {
		v.token = m[1]
	}
	return rec
}

var tokenField = regexp.MustCompile(`name="_token" value="([^"]+)"`)

// open loads a page of the manual, as a browser navigating to it, and keeps the
// CSRF token its demo carries.
func (v *visitor) open(page string) string {
	v.t.Helper()
	req := httptest.NewRequest(http.MethodGet, "https://ypy.dev"+page, nil)
	for k, val := range v.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: val})
	}
	rec := httptest.NewRecorder()
	v.app.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		v.t.Fatalf("GET %s answered %d", page, rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		v.cookies[c.Name] = c.Value
	}
	body := rec.Body.String()
	if m := tokenField.FindStringSubmatch(body); m != nil {
		v.token = m[1]
	}
	return body
}

// form is a body with the visitor's CSRF token already in it.
func (v *visitor) form(pairs ...string) url.Values {
	f := url.Values{"_token": {v.token}}
	for i := 0; i+1 < len(pairs); i += 2 {
		f.Add(pairs[i], pairs[i+1])
	}
	return f
}

// wantFragment checks what every demo answer has in common: the status, a
// fragment and not a page, Vary: YPY-Request and noindex.
func wantFragment(t *testing.T, rec *httptest.ResponseRecorder, status int) string {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("answered %d, want %d: %s", rec.Code, status, rec.Body.String())
	}
	h := rec.Header()
	if !strings.Contains(h.Get("Vary"), "YPY-Request") {
		t.Errorf("Vary = %q, want it to name YPY-Request", h.Get("Vary"))
	}
	if !strings.Contains(h.Get("X-Robots-Tag"), "noindex") {
		t.Errorf("X-Robots-Tag = %q, want noindex", h.Get("X-Robots-Tag"))
	}
	if !strings.HasPrefix(h.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %q, want text/html", h.Get("Content-Type"))
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") || strings.Contains(body, "<body") {
		t.Errorf("a demo answers a fragment, not a page: %.80s", body)
	}
	return body
}

const clickToEditPage = "/examples/click-to-edit"

func TestClickToEdit(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open(clickToEditPage)
	if !strings.Contains(page, `id="contact"`) || !strings.Contains(page, "Marina Duarte") {
		t.Fatalf("the page does not carry the demo")
	}

	edit := wantFragment(t, a.do("GET", "/examples/demo/click-to-edit/edit", nil), 200)
	if !strings.Contains(edit, `ypy-put="/examples/demo/click-to-edit"`) || !strings.Contains(edit, `name="_token"`) {
		t.Errorf("the edit form is not wired: %s", edit)
	}

	bad := wantFragment(t, a.do("PUT", "/examples/demo/click-to-edit", a.form("name", "  ", "email", "nope@x")), 422)
	if !strings.Contains(bad, "Write the name") || !strings.Contains(bad, "name@company.com") {
		t.Errorf("422 does not carry both messages: %s", bad)
	}

	ok := wantFragment(t, a.do("PUT", "/examples/demo/click-to-edit", a.form("name", "Rafael Quintana", "email", "rafael@quintana.example")), 200)
	if !strings.Contains(ok, "Rafael Quintana") {
		t.Errorf("the saved contact is not shown: %s", ok)
	}

	// The same visitor sees the saved contact on the page; another one does not.
	if got := a.open(clickToEditPage); !strings.Contains(got, "Rafael Quintana") {
		t.Errorf("the visitor's own page lost the saved contact")
	}
	b := newVisitor(t, app)
	if got := b.open(clickToEditPage); strings.Contains(got, "Rafael Quintana") || !strings.Contains(got, "Marina Duarte") {
		t.Errorf("one visitor's edit leaked to another")
	}
	if rec := b.do("GET", "/examples/demo/click-to-edit", nil); strings.Contains(rec.Body.String(), "Rafael") {
		t.Errorf("one visitor's edit leaked through the route")
	}
}

func TestDemoWriteNeedsCSRFToken(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	v.open(clickToEditPage)
	rec := v.do("PUT", "/examples/demo/click-to-edit", url.Values{"name": {"X"}, "email": {"x@y.example"}})
	if rec.Code != 419 {
		t.Errorf("a PUT with no token answered %d, want 419", rec.Code)
	}
}

func TestBulkUpdate(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open("/examples/bulk-update")
	if !strings.Contains(page, `id="team"`) || !strings.Contains(page, "Helena Prado") {
		t.Fatalf("the page does not carry the demo")
	}

	none := wantFragment(t, a.do("PUT", "/examples/demo/bulk-update", a.form("status", "active")), 422)
	if !strings.Contains(none, "Tick at least one person") {
		t.Errorf("422 has no message: %s", none)
	}

	ok := wantFragment(t, a.do("PUT", "/examples/demo/bulk-update", a.form("ids", "3", "ids", "5", "status", "active")), 200)
	if !strings.Contains(ok, "2 people activated.") || strings.Contains(ok, `>Inactive</span>`) {
		t.Errorf("rows 3 and 5 were not activated: %s", ok)
	}
	if got := a.open("/examples/bulk-update"); strings.Contains(got, `>Inactive</span>`) {
		t.Errorf("the visitor's page does not keep the change")
	}

	b := newVisitor(t, app)
	if got := b.open("/examples/bulk-update"); !strings.Contains(got, `>Inactive</span>`) {
		t.Errorf("one visitor's bulk update leaked to another")
	}
}

func TestClickToLoad(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	page := v.open("/examples/click-to-load")
	if !strings.Contains(page, `id="more-orders"`) || !strings.Contains(page, "Harbor Coffee Roasters") {
		t.Fatalf("the page does not carry the demo")
	}
	p2 := wantFragment(t, v.do("GET", "/examples/demo/click-to-load?page=2", nil), 200)
	if !strings.Contains(p2, "#2036") || !strings.Contains(p2, "page=3") || !strings.Contains(p2, "10 of 14") {
		t.Errorf("page 2 is wrong: %s", p2)
	}
	p3 := wantFragment(t, v.do("GET", "/examples/demo/click-to-load?page=3", nil), 200)
	if !strings.Contains(p3, "#2028") || strings.Contains(p3, "<button") || !strings.Contains(p3, "All 14 orders are shown.") {
		t.Errorf("the last page must end the list: %s", p3)
	}
	beyond := wantFragment(t, v.do("GET", "/examples/demo/click-to-load?page=99", nil), 200)
	if strings.Contains(beyond, "<button") {
		t.Errorf("a page past the end offers another page: %s", beyond)
	}
}

func TestDeleteRow(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open("/examples/delete-row")
	if !strings.Contains(page, `ypy-delete="/examples/demo/delete-row/2"`) || !strings.Contains(page, `name="_token"`) {
		t.Fatalf("the page does not carry the demo")
	}
	headers := regexp.MustCompile(`ypy-headers='\{"X-CSRF-Token": "([^"]+)"\}'`).FindStringSubmatch(page)
	if headers == nil {
		t.Fatalf("a delete form must carry the token in ypy-headers")
	}

	if rec := a.do("DELETE", "/examples/demo/delete-row/2", nil); rec.Code != 419 {
		t.Errorf("a DELETE with no token answered %d, want 419", rec.Code)
	}
	a.header = headers[1]
	if body := wantFragment(t, a.do("DELETE", "/examples/demo/delete-row/2", nil), 200); body != "" {
		t.Errorf("a deleted row answers an empty body, got %q", body)
	}
	if rec := a.do("DELETE", "/examples/demo/delete-row/99", nil); rec.Code != 404 {
		t.Errorf("an unknown row answered %d, want 404", rec.Code)
	}
	if got := a.open("/examples/delete-row"); strings.Contains(got, `delete-row/2"`) || !strings.Contains(got, `delete-row/1"`) {
		t.Errorf("the visitor's page does not keep the deletion")
	}
	b := newVisitor(t, app)
	if got := b.open("/examples/delete-row"); !strings.Contains(got, `delete-row/2"`) {
		t.Errorf("one visitor's deletion leaked to another")
	}
	a.header = ""
	a.token = headers[1]
	back := wantFragment(t, a.do("POST", "/examples/demo/delete-row/restore", a.form()), 200)
	if !strings.Contains(back, `delete-row/2"`) {
		t.Errorf("restore did not bring the rows back: %s", back)
	}
}

func TestEditRow(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open("/examples/edit-row")
	if !strings.Contains(page, `id="member-3"`) {
		t.Fatalf("the page does not carry the demo")
	}
	form := wantFragment(t, a.do("GET", "/examples/demo/edit-row/3/edit", nil), 200)
	if !strings.HasPrefix(strings.TrimSpace(form), `<tr id="member-3">`) || !strings.Contains(form, `ypy-include="closest tr"`) {
		t.Errorf("the edit row is not a row: %s", form)
	}
	bad := wantFragment(t, a.do("PUT", "/examples/demo/edit-row/3", a.form("name", "", "role", "Lead")), 422)
	if !strings.Contains(bad, "cannot be empty") || !strings.HasPrefix(strings.TrimSpace(bad), "<tr") {
		t.Errorf("422 does not redraw the row with the message: %s", bad)
	}
	ok := wantFragment(t, a.do("PUT", "/examples/demo/edit-row/3", a.form("name", "Helena Prado-Silva", "role", "Head of product")), 200)
	if !strings.Contains(ok, "Head of product") {
		t.Errorf("the saved row is wrong: %s", ok)
	}
	if rec := a.do("GET", "/examples/demo/edit-row/9", nil); rec.Code != 404 {
		t.Errorf("an unknown row answered %d, want 404", rec.Code)
	}
	if got := a.open("/examples/edit-row"); !strings.Contains(got, "Head of product") {
		t.Errorf("the visitor's page does not keep the edit")
	}
	b := newVisitor(t, app)
	if got := b.open("/examples/edit-row"); strings.Contains(got, "Head of product") {
		t.Errorf("one visitor's edit leaked to another")
	}
}

func TestLazyLoading(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	page := v.open("/examples/lazy-loading")
	if !strings.Contains(page, `ypy-trigger="load"`) || !strings.Contains(page, "Loading the weekly report") {
		t.Fatalf("the page does not carry the placeholder")
	}
	rep := wantFragment(t, v.do("GET", "/examples/demo/lazy-loading/report", nil), 200)
	if !strings.Contains(rep, `id="report"`) || !strings.Contains(rep, "$14,320.50") {
		t.Errorf("the report is wrong: %s", rep)
	}
	again := wantFragment(t, v.do("GET", "/examples/demo/lazy-loading/placeholder", nil), 200)
	if !strings.Contains(again, `ypy-trigger="load"`) {
		t.Errorf("the placeholder does not start the request again: %s", again)
	}
}

func TestInlineValidation(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open("/examples/inline-validation")
	if !strings.Contains(page, `ypy-error-target="#email-msg"`) {
		t.Fatalf("the page does not carry the demo")
	}
	cases := []struct {
		email  string
		status int
		want   string
	}{
		{"", 422, "Enter your email"},
		{"nope@x", 422, "name@company.com"},
		{"Admin@Example.com", 422, "already has an account"},
		{"marina@company.example", 200, "is available"},
	}
	for _, c := range cases {
		body := wantFragment(t, a.do("POST", "/examples/demo/inline-validation/email", a.form("email", c.email)), c.status)
		if !strings.Contains(body, c.want) || !strings.HasPrefix(body, `<p`) {
			t.Errorf("%q: %s", c.email, body)
		}
	}
	bad := wantFragment(t, a.do("POST", "/examples/demo/inline-validation", a.form("email", "nope@x")), 422)
	if !strings.Contains(bad, `value="nope@x"`) || !strings.Contains(bad, `aria-invalid="true"`) {
		t.Errorf("the form does not keep the value or mark the field: %s", bad)
	}
	ok := wantFragment(t, a.do("POST", "/examples/demo/inline-validation", a.form("email", "rafael@quintana.example")), 200)
	if !strings.Contains(ok, "Account created for rafael@quintana.example") {
		t.Errorf("no confirmation: %s", ok)
	}
}

func TestActiveSearch(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	page := v.open("/examples/active-search")
	if !strings.Contains(page, `id="search-results"`) || !strings.Contains(page, "Rafael Quintana") {
		t.Fatalf("the page does not carry the demo")
	}
	porto := wantFragment(t, v.do("GET", "/examples/demo/active-search?q=PORTO", nil), 200)
	if !strings.Contains(porto, "Tomás Ribeiro") || !strings.Contains(porto, "Joaquim Teles") || strings.Contains(porto, "Marina") {
		t.Errorf("the filter is wrong: %s", porto)
	}
	none := wantFragment(t, v.do("GET", "/examples/demo/active-search?q=zzz", nil), 200)
	if !strings.Contains(none, "No one matches") || !strings.Contains(none, "zzz") {
		t.Errorf("the empty answer is wrong: %s", none)
	}
	if all := wantFragment(t, v.do("GET", "/examples/demo/active-search", nil), 200); strings.Count(all, "<tr>") != 12 {
		t.Errorf("an empty query must answer everyone: %d rows", strings.Count(all, "<tr>"))
	}
	if xss := wantFragment(t, v.do("GET", "/examples/demo/active-search?q=%3Cscript%3E", nil), 200); strings.Contains(xss, "<script>") {
		t.Errorf("the query is echoed unescaped: %s", xss)
	}
}

func TestInfiniteScroll(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	page := v.open("/examples/infinite-scroll")
	if !strings.Contains(page, `ypy-trigger="revealed"`) || !strings.Contains(page, `page=2`) {
		t.Fatalf("the page does not carry the demo")
	}
	p2 := wantFragment(t, v.do("GET", "/examples/demo/infinite-scroll?page=2", nil), 200)
	if strings.Count(p2, "<tr") != 9 || !strings.Contains(p2, "page=3") {
		t.Errorf("page 2 must have nine rows and ask for page 3: %s", p2)
	}
	p3 := wantFragment(t, v.do("GET", "/examples/demo/infinite-scroll?page=3", nil), 200)
	if strings.Contains(p3, "ypy-get") || strings.Count(p3, "<tr") != 9 {
		t.Errorf("the last page must stop asking: %s", p3)
	}
	if end := wantFragment(t, v.do("GET", "/examples/demo/infinite-scroll?page=4", nil), 200); strings.TrimSpace(end) != "" {
		t.Errorf("a page past the end is empty: %q", end)
	}
}

func TestValueSelect(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	v.open("/examples/value-select")
	toyota := wantFragment(t, v.do("GET", "/examples/demo/value-select?make=toyota", nil), 200)
	if !strings.Contains(toyota, `id="model"`) || !strings.Contains(toyota, "Hilux") || strings.Contains(toyota, "disabled") {
		t.Errorf("toyota: %s", toyota)
	}
	for _, q := range []string{"?make=", "?make=unknown", ""} {
		none := wantFragment(t, v.do("GET", "/examples/demo/value-select"+q, nil), 200)
		if !strings.Contains(none, "disabled") || !strings.Contains(none, "Choose a make first") {
			t.Errorf("%q: %s", q, none)
		}
	}
}

func TestTabs(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	v := newVisitor(t, app)
	v.open("/examples/tabs")
	two := wantFragment(t, v.do("GET", "/examples/demo/tabs/2", nil), 200)
	if !strings.Contains(two, `id="tab-panel"`) || !strings.Contains(two, "Next invoice") ||
		!strings.Contains(two, `ypy-swap-oob="true"`) || !strings.Contains(two, `id="tab-2" aria-selected="true"`) ||
		!strings.Contains(two, `id="tab-1" aria-selected="false"`) {
		t.Errorf("tab 2 is wrong: %s", two)
	}
	if rec := v.do("GET", "/examples/demo/tabs/9", nil); rec.Code != 404 {
		t.Errorf("an unknown tab answered %d, want 404", rec.Code)
	}
}

func TestProgressBar(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	a := newVisitor(t, app)
	page := a.open("/examples/progress-bar")
	headers := regexp.MustCompile(`ypy-headers='\{"X-CSRF-Token": "([^"]+)"\}'`).FindStringSubmatch(page)
	if headers == nil {
		t.Fatalf("the start button must carry the token in ypy-headers")
	}
	// Before a start there is nothing to follow: the button comes back.
	if idle := wantFragment(t, a.do("GET", "/examples/demo/progress-bar/progress", nil), 200); !strings.Contains(idle, "Start the export") {
		t.Errorf("a poll with no job must answer the start button: %s", idle)
	}
	if rec := a.do("POST", "/examples/demo/progress-bar/start", nil); rec.Code != 419 {
		t.Errorf("a start with no token answered %d, want 419", rec.Code)
	}
	a.header = headers[1]
	started := wantFragment(t, a.do("POST", "/examples/demo/progress-bar/start", nil), 200)
	if !strings.Contains(started, `ypy-trigger="load delay:600ms"`) || !strings.Contains(started, `value="0"`) {
		t.Errorf("the started job does not poll from zero: %s", started)
	}
	if poll := wantFragment(t, a.do("GET", "/examples/demo/progress-bar/progress", nil), 200); !strings.Contains(poll, "ypy-get") {
		t.Errorf("a running job must keep polling: %s", poll)
	}
	// Another visitor never sees this job.
	b := newVisitor(t, app)
	if idle := wantFragment(t, b.do("GET", "/examples/demo/progress-bar/progress", nil), 200); !strings.Contains(idle, "Start the export") {
		t.Errorf("one visitor's job leaked to another: %s", idle)
	}
}

// Every demo page draws its demo and draws no inline script, style or handler,
// and the demos stay out of the sitemap and the crawlers' way.
func TestDemosKeepTheStrictPolicy(t *testing.T) {
	app := tests.Boot(t, config.EnvDev, true)
	names := []string{"click-to-edit", "bulk-update", "click-to-load", "delete-row", "edit-row", "lazy-loading",
		"inline-validation", "active-search", "infinite-scroll", "progress-bar", "value-select", "tabs"}
	ldJSON := regexp.MustCompile(`(?s)<script type="application/ld\+json"[^>]*>.*?</script>`)
	inline := regexp.MustCompile(`(?i)<script[^>]*>[^<]|<style|\sstyle=|\son[a-z]+=`)
	for _, name := range names {
		rec := app.Get("/examples/" + name)
		if rec.Code != 200 {
			t.Errorf("/examples/%s answered %d", name, rec.Code)
			continue
		}
		// The JSON-LD block is data the browser never runs; everything else is checked.
		body := ldJSON.ReplaceAllString(rec.Body.String(), "")
		if !strings.Contains(body, `class="ex-demo"`) {
			t.Errorf("/examples/%s has no demo", name)
		}
		if strings.Contains(body, "YPYDEMOMARKER") {
			t.Errorf("/examples/%s leaks the demo marker", name)
		}
		if m := inline.FindString(body); m != "" {
			t.Errorf("/examples/%s carries inline code: %q", name, m)
		}
		if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "script-src 'self'") || strings.Contains(csp, "unsafe-inline") {
			t.Errorf("/examples/%s has a loose policy: %q", name, csp)
		}
	}
	if robots := app.Get("/robots.txt").Body.String(); !strings.Contains(robots, "/examples/demo") {
		t.Errorf("robots.txt does not close /examples/demo:\n%s", robots)
	}
	sitemap := app.Get("/sitemap.xml").Body.String()
	if strings.Contains(sitemap, "/examples/demo") || !strings.Contains(sitemap, "/examples/click-to-edit") {
		t.Errorf("the sitemap must list the example pages and not the demos")
	}
}
