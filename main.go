// Command ypy.dev is the entry point of the YPY website: the server and the
// command line in one binary. No arguments is serve.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ypy-develop/ypy.dev/bootstrap"
)

// version and commit are stamped by the build. See the Dockerfile.
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	app, err := bootstrap.App(version, commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(app.HandleCommand(context.Background(), os.Args[1:]))
}
