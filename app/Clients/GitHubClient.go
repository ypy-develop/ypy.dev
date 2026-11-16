// Package clients holds this application's clients of other systems.
//
// Each one is a typed configuration, a small interface the callers depend on,
// the client that talks to the system over hesape/http/client, and a fake that
// a test hands over instead. A call that leaves the process is made here and
// nowhere else, so what the other system changes on its side changes one file.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/arandu-io/hesape/http/client"
)

// GitHubConfig is what the GitHub client needs, typed. bootstrap/app.go builds
// it once and hands it over; nothing here looks a setting up again.
type GitHubConfig struct {
	// BaseURL is where the REST API answers, with its scheme.
	BaseURL string
	// Token is the credential every request carries, as a bearer token. Empty
	// asks anonymously, which is what this site does.
	Token string
	// Timeout bounds each attempt. Zero leaves the client's own deadline.
	Timeout time.Duration
}

// LogValue implements slog.LogValuer, so a config passed to a log call, or
// dumped on the debug page, records where the client points and never the
// token.
func (c GitHubConfig) LogValue() slog.Value {
	return slog.GroupValue(slog.String("base_url", c.BaseURL), slog.Duration("timeout", c.Timeout))
}

// MarshalJSON writes the same fields LogValue does, for the same reason: the
// token stays out of anything encoded.
func (c GitHubConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"base_url": c.BaseURL, "timeout": c.Timeout.String()})
}

// GitHub is what the code that reads GitHub depends on: the calls this
// application makes to it, and nothing else. A test hands it GitHubFake.
type GitHub interface {
	// arandu:begin custom

	// Stargazers is how many stars a repository has, named as owner/name.
	Stargazers(ctx context.Context, repository string) (int64, error)
	// arandu:end custom
}

// GitHubClient talks to GitHub over hesape/http/client, which is the only way
// a request leaves this application: every attempt carries a deadline, reads a
// bounded body, and refuses an address inside the network.
//
// It holds no Grant, no model and no session, and answers with its own types,
// so the code that calls it never reads GitHub's wire format.
type GitHubClient struct {
	config GitHubConfig
	http   *client.Factory
}

// NewGitHubClient builds the client with its configuration and the factory its
// requests are made from: client.NewFactory(nil) in bootstrap/app.go, and the
// same factory faked in a test that wants to see the request.
func NewGitHubClient(config GitHubConfig, http *client.Factory) *GitHubClient {
	return &GitHubClient{config: config, http: http}
}

// Compile-time proof that the client is what the code calling it depends on.
var _ GitHub = (*GitHubClient)(nil)

// request is every call's starting point: the base address, the media type,
// the credential and the deadline, set once rather than at each call.
func (c *GitHubClient) request() *client.PendingRequest {
	// The documented media type and API version. Without them the API answers
	// whatever version it defaults to that year, and a field read here is one
	// rename away.
	r := c.http.CreatePendingRequest().BaseURL(c.config.BaseURL).
		Accept("application/vnd.github+json").
		WithHeader("X-GitHub-Api-Version", "2022-11-28")
	if c.config.Token != "" {
		r = r.WithToken(c.config.Token, "Bearer")
	}
	if c.config.Timeout > 0 {
		r = r.Timeout(c.config.Timeout)
	}
	return r
}

// arandu:begin custom

// repository is the one field of a repository this client reads.
//
// Named rather than a map, so the day GitHub renames it this stops instead of
// quietly treating an absent field as zero.
type repository struct {
	Stargazers *int64 `json:"stargazers_count"`
}

// Stargazers reads the repository and answers its stargazers_count.
//
// Anything but a 200 is an error, and so is an answer without the field: an
// explicit zero is a count, an absent field is not.
func (c *GitHubClient) Stargazers(ctx context.Context, name string) (int64, error) {
	res, err := c.request().Get(ctx, "/repos/"+name, nil)
	if err != nil {
		return 0, fmt.Errorf("github: %w", err)
	}
	if res.Status() != http.StatusOK {
		return 0, fmt.Errorf("github: the API answered %d", res.Status())
	}
	var repo repository
	if err := res.Object(&repo); err != nil {
		return 0, fmt.Errorf("github: %w", err)
	}
	if repo.Stargazers == nil {
		return 0, fmt.Errorf("github: the API response has no stargazers_count")
	}
	return *repo.Stargazers, nil
}

// arandu:end custom
