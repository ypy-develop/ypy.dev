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

const clickToEditPage = "/docs/examples/click-to-edit"

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
