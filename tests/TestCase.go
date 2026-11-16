// Package tests holds what the feature tests share: a booted application over
// the real content/ folder, and a way to ask it for pages.
package tests

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arandu-io/framework/foundation"
	fwbootstrap "github.com/arandu-io/framework/foundation/bootstrap"
	"github.com/arandu-io/hesape/config"
	"github.com/arandu-io/hesape/database"
	"github.com/arandu-io/hesape/support"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
	providers "github.com/ypy-develop/ypy.dev/app/Providers"
	"github.com/ypy-develop/ypy.dev/bootstrap"
	appconfig "github.com/ypy-develop/ypy.dev/config"
)

// Root is the project root.
func Root(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("tests.Root is not the project root: %v", err)
	}
	return root
}

// Booted is an application and what it was built from.
type Booted struct {
	Handler  http.Handler
	Provider *providers.AppServiceProvider
}

// Boot composes the application the way bootstrap.App does, over the real
// content/ folder, with the environment given. The database is an in-memory
// SQLite that nothing writes to: the site stores nothing.
func Boot(t *testing.T, env config.Env, indexing bool) Booted {
	t.Helper()

	cfg := fwbootstrap.Configuration{
		App: config.App{
			Name:     "YPY",
			Env:      env,
			Debug:    env.Is(config.EnvDev),
			URL:      &url.URL{Scheme: "https", Host: "ypy.dev"},
			HTTPAddr: ":0",
			Timezone: time.UTC,
			Locale:   "en",
			Key:      []byte("0123456789abcdef0123456789abcdef"),
		},
		Database: database.Config{
			Connection: database.DialectSQLite,
			Database:   ":memory:",
		},
		Brand:         support.NewBrand("YPY"),
		Observability: fwbootstrap.Observability{LogLevel: slog.LevelError, Editor: "vscode"},
		Session:       fwbootstrap.Session{Lifetime: 2 * time.Hour},
	}
	if err := cfg.App.Validate(); err != nil {
		t.Fatalf("the test configuration is not valid: %v", err)
	}

	sqldb, err := sql.Open(cfg.Database.Connection.Driver(), cfg.Database.DSN())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })

	settings, err := fwbootstrap.SettingsFrom(cfg)
	if err != nil {
		t.Fatalf("reading the framework's settings: %v", err)
	}
	appCfg := appconfig.From(settings)
	appCfg.Site.Indexing = indexing
	appCfg.Site.ContentDir = filepath.Join(Root(t), "content")

	drivers, err := foundation.NewDriversOver(appCfg.Settings, database.Wrap(sqldb, cfg.Database.Connection))
	if err != nil {
		t.Fatalf("drivers: %v", err)
	}
	provider, err := bootstrap.NewProvider(appCfg, &clients.GitHubFake{})
	if err != nil {
		t.Fatalf("the site: %v", err)
	}
	k, err := bootstrap.Configure(drivers, provider).Create()
	if err != nil {
		t.Fatalf("composing the application: %v", err)
	}
	if err := k.Boot(context.Background()); err != nil {
		t.Fatalf("Boot: %v", err)
	}
	t.Cleanup(func() { _ = k.Shutdown() })
	return Booted{Handler: k.Handler(), Provider: provider}
}

// Get asks the application for an address, over https as a visitor would.
func (b Booted) Get(path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "https://ypy.dev"+path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	b.Handler.ServeHTTP(rec, req)
	return rec
}
