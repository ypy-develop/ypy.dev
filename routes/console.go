package routes

import (
	"context"
	"time"

	"github.com/arandu-io/hesape/auth"
	"github.com/arandu-io/hesape/console"
	"github.com/arandu-io/hesape/foundation"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
	stars "github.com/ypy-develop/ypy.dev/app/Stars"
)

// StarsRefreshTaskID identifies the one scheduled path that refreshes the
// project repository's star count.
const StarsRefreshTaskID = "stars.refresh"

// ConsoleDeps is what this application's commands and scheduled tasks reach,
// built by app/Providers.
type ConsoleDeps struct {
	// GitHub is the client the star count is read through.
	GitHub clients.GitHub
}

// Commands are this application's own console commands. The framework's --
// serve, routes, migrate and the rest -- are composed by the kernel and are not
// listed here.
func Commands() []console.Command {
	return []console.Command{
		// arandu:begin custom
		// arandu:end custom
	}
}

// Schedule is this application's scheduled work. The manual is read once at
// start, so the star count is the only thing that changes while the site runs.
func Schedule(d ConsoleDeps) []foundation.Task {
	return []foundation.Task{
		{
			ID: StarsRefreshTaskID,
			// The seventeenth minute rather than the zeroth: an hourly task on
			// the hour shares the hour with every other hourly task on the
			// machine, and this one is not urgent enough to be in that queue.
			Spec: "17 * * * *",
			// Global, because the number is about the project repository and
			// not about a customer. A global task gets the zero Grant and can
			// reach nothing else, which is right: it reads an API and writes a
			// variable.
			Scope: foundation.Global,
			// Singleton is false. The work is one GET and an assignment, so two
			// replicas doing it is two requests rather than a conflict.
			Singleton: false,
			Timeout:   time.Minute,
			Action:    "site.stars.refresh",
			Run: func(ctx context.Context, _ auth.Grant) error {
				return stars.Refresh(ctx, d.GitHub)
			},
		},
	}
}
