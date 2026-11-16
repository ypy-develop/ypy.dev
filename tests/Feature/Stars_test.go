package feature_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arandu-io/hesape/config"
	"github.com/arandu-io/hesape/http/client"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
	stars "github.com/ypy-develop/ypy.dev/app/Stars"
	"github.com/ypy-develop/ypy.dev/tests"
)

// answer is what the faked GitHub sends back: a status and a body.
type answer struct {
	status int
	body   string
}

// clientFor is the real GitHub client, built the way bootstrap builds it, over a
// factory that answers every request in the process instead of sending it.
func clientFor(t *testing.T, a answer) clients.GitHub {
	t.Helper()
	factory := client.NewFactory(nil).Fake(func(r *http.Request) (*http.Response, error) {
		if got, want := r.Method+" "+r.URL.String(), "GET https://api.github.com/repos/ypy-develop/ypy"; got != want {
			t.Errorf("GitHub request = %q, want %q", got, want)
		}
		if r.Header.Get("X-GitHub-Api-Version") == "" {
			t.Error("the request names no API version")
		}
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("the request accepts %q, want GitHub's documented media type", got)
		}
		return client.NewResponseFromBytes(a.status, []byte(a.body), nil).HTTPResponse(), nil
	})
	return clients.NewGitHubClient(clients.GitHubConfig{BaseURL: "https://api.github.com"}, factory)
}

// The whole life of the value in one test, because the order is the behaviour:
// a good answer is the repository's count, a refusal afterwards leaves the last
// good number, and an explicit zero is a count.
func TestTheStarCountComesFromTheRepositoryAndSurvivesARefusal(t *testing.T) {
	good := clientFor(t, answer{http.StatusOK, `{"stargazers_count":42}`})
	if err := stars.Refresh(context.Background(), good); err != nil {
		t.Fatal(err)
	}
	if got := stars.Count(); got != "42" {
		t.Fatalf("Count() = %q, want 42", got)
	}

	app := tests.Boot(t, config.EnvDev, true)
	body := app.Get("/").Body.String()
	start, end := strings.Index(body, "<header "), strings.Index(body, "</header>")
	if start < 0 || end <= start {
		t.Fatal("the rendered page has no complete header")
	}
	if got := strings.Count(body[start:end], ">42</span>"); got != 2 {
		t.Fatalf("the star count appears %d times in the header, want the button and the mobile list", got)
	}

	limited := clientFor(t, answer{http.StatusForbidden, ``})
	if err := stars.Refresh(context.Background(), limited); err == nil {
		t.Fatal("a 403 was accepted")
	}
	if got := stars.Count(); got != "42" {
		t.Errorf("Count() = %q after a refused refresh, want the last fetched 42", got)
	}

	zero := clientFor(t, answer{http.StatusOK, `{"stargazers_count":0}`})
	if err := stars.Refresh(context.Background(), zero); err != nil {
		t.Fatal(err)
	}
	if got := stars.Count(); got != "0" {
		t.Fatalf("Count() = %q after an explicit zero, want 0", got)
	}
}

// A missing field must not be mistaken for an explicit zero, and a list is not
// the repository object.
func TestAnAnswerWithoutAStarCountIsRefused(t *testing.T) {
	for name, body := range map[string]string{"empty object": `{}`, "list": `[{"stargazers_count":42}]`} {
		if err := stars.Refresh(context.Background(), clientFor(t, answer{http.StatusOK, body})); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

// The repository is pinned, and a failed call or a missing client is an error.
func TestTheStarCountIsAskedOfTheYPYRepository(t *testing.T) {
	fake := &clients.GitHubFake{Err: context.DeadlineExceeded}
	if err := stars.Refresh(context.Background(), fake); err == nil {
		t.Fatal("a failed call was accepted")
	}
	if len(fake.Repositories) != 1 || fake.Repositories[0] != "ypy-develop/ypy" {
		t.Fatalf("asked of %v, want ypy-develop/ypy", fake.Repositories)
	}
	if err := stars.Refresh(context.Background(), nil); err == nil {
		t.Fatal("a refresh with no client was accepted")
	}
}

// Over the real transport: loopback is refused until the host is named, then the
// count is read back through the body limit and the decoding a fake skips.
func TestTheStarCountIsReadOverTheRealTransport(t *testing.T) {
	var asked []string
	github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = append(asked, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stargazers_count":7}`))
	}))
	defer github.Close()

	cfg := clients.GitHubConfig{BaseURL: github.URL}
	refused := clients.NewGitHubClient(cfg, client.NewFactory(nil))
	if err := stars.Refresh(context.Background(), refused); !errors.Is(err, client.ErrInternalAddress) {
		t.Fatalf("a loopback address = %v, want %v", err, client.ErrInternalAddress)
	}
	if len(asked) != 0 {
		t.Fatalf("the refused refresh reached the server: %v", asked)
	}

	host := strings.TrimPrefix(github.URL, "http://")
	host = host[:strings.LastIndex(host, ":")]
	allowed := clients.NewGitHubClient(cfg, client.NewFactory(nil).AllowInternalHosts(host))
	if err := stars.Refresh(context.Background(), allowed); err != nil {
		t.Fatal(err)
	}
	if got := stars.Count(); got != "7" {
		t.Fatalf("Count() = %q, want 7", got)
	}
	if len(asked) != 1 || asked[0] != "GET /repos/"+stars.Repository {
		t.Fatalf("asked %v, want one GET /repos/%s", asked, stars.Repository)
	}
}
