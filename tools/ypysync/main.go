// Command ypysync copies what the site shows from the YPY source tree into this
// repository, so the site never reads another repository at run time.
//
// It copies the name tables (src/attributes.json and src/attributes.d/*.json)
// into app/Reference/data, the version into app/Reference/data/VERSION, and the
// two extensions the site loads beside the core (dist/embed/ext) into assets/vendor.
// The core itself comes embedded in hesape.
//
// Run it with `go generate ./app/Reference` or `task sync`. The source tree is
// ../ypy next to this repository, or the directory named by YPY_DIR.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("ypy", os.Getenv("YPY_DIR"), "the YPY source tree (default: ../ypy next to this repository)")
	flag.Parse()

	root, err := moduleRoot()
	if err != nil {
		fail(err)
	}
	if *dir == "" {
		*dir = filepath.Join(root, "..", "ypy")
	}

	copies := []struct{ from, to string }{
		{filepath.Join(*dir, "src", "attributes.json"), filepath.Join(root, "app", "Reference", "data", "attributes.json")},
		{filepath.Join(*dir, "VERSION"), filepath.Join(root, "app", "Reference", "data", "VERSION")},
		{filepath.Join(*dir, "dist", "embed", "ext", "ypy-head.min.js"), filepath.Join(root, "assets", "vendor", "ypy-head.min.js")},
		{filepath.Join(*dir, "dist", "embed", "ext", "ypy-preload.min.js"), filepath.Join(root, "assets", "vendor", "ypy-preload.min.js")},
	}

	parts, err := filepath.Glob(filepath.Join(*dir, "src", "attributes.d", "*.json"))
	if err != nil {
		fail(err)
	}
	for _, p := range parts {
		copies = append(copies, struct{ from, to string }{p, filepath.Join(root, "app", "Reference", "data", "attributes.d", filepath.Base(p))})
	}

	// A file that left the source tree must leave the copy too, or the site would
	// keep documenting a name YPY no longer has.
	if err := os.RemoveAll(filepath.Join(root, "app", "Reference", "data", "attributes.d")); err != nil {
		fail(err)
	}

	for _, c := range copies {
		body, err := os.ReadFile(c.from)
		if err != nil {
			fail(fmt.Errorf("%w (is the YPY source tree at %s?)", err, *dir))
		}
		if strings.HasSuffix(c.to, "VERSION") {
			body = append([]byte(strings.TrimSpace(string(body))), '\n')
		}
		if err := os.MkdirAll(filepath.Dir(c.to), 0o755); err != nil {
			fail(err)
		}
		if err := os.WriteFile(c.to, body, 0o644); err != nil {
			fail(err)
		}
	}
	fmt.Printf("ypysync: %d files copied from %s\n", len(copies), *dir)
}

func moduleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := wd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d, nil
		}
		if filepath.Dir(d) == d {
			return "", fmt.Errorf("no go.mod above %s", wd)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "ypysync:", err)
	os.Exit(1)
}
