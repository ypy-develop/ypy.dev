package unit_test

import (
	"context"
	"testing"
	"testing/fstest"

	fwbootstrap "github.com/arandu-io/framework/foundation/bootstrap"
	"github.com/arandu-io/hesape/config"
	"github.com/arandu-io/hesape/geo"
	markdown "github.com/hyz-is/arandu-markdown"

	site "github.com/ypy-develop/ypy.dev/app/Site"
	appconfig "github.com/ypy-develop/ypy.dev/config"
)

// Indexing is a switch of the deployment: on in production, off everywhere
// else, and SITE_INDEXING can say otherwise either way.
func TestIndexingFollowsTheEnvironmentSetting(t *testing.T) {
	cases := []struct {
		env      config.Env
		override string
		want     bool
	}{
		{config.EnvProd, "", true},
		{config.EnvDev, "", false},
		{config.EnvStaging, "", false},
		{config.EnvProd, "false", false},
		{config.EnvDev, "true", true},
		{config.EnvDev, "1", true},
	}
	for _, c := range cases {
		t.Setenv("SITE_INDEXING", c.override)
		settings := fwbootstrap.Settings{}
		settings.App.Env = c.env
		if got := appconfig.From(settings).Site.Indexing; got != c.want {
			t.Errorf("env %s with SITE_INDEXING=%q: indexing is %v, want %v", c.env, c.override, got, c.want)
		}
	}
}

// A deployment that is indexed with an origin that is not https does not
// start: it would publish addresses nobody can fetch.
func TestIndexedDeploymentNeedsAnHTTPSOrigin(t *testing.T) {
	md, err := markdown.New(markdown.Config{})
	if err != nil {
		t.Fatal(err)
	}
	content := fstest.MapFS{"ypy/index.md": &fstest.MapFile{Data: []byte("# YPY\n\nA page.\n")}}

	for url, indexing := range map[string]bool{"http://ypy.dev": true, "https://ypy.dev": true, "http://localhost:8080": false} {
		s, err := site.New(site.Config{Name: "YPY", URL: url, Indexing: indexing, Content: content}, md)
		if err != nil {
			t.Fatal(err)
		}
		err = geo.NewModule(s.GeoConfig(), s.Catalog()).Boot(context.Background())
		refused := err != nil
		if want := indexing && url[:5] == "http:"; refused != want {
			t.Errorf("origin %s with indexing %v: boot error %v", url, indexing, err)
		}
	}
}
