package routes

import (
	"github.com/arandu-io/hesape/console"
	"github.com/arandu-io/hesape/foundation"
)

// Commands are this application's own console commands. The framework's --
// serve, routes, migrate and the rest -- are composed by the kernel and are not
// listed here.
func Commands() []console.Command {
	return []console.Command{
		// arandu:begin custom
		// arandu:end custom
	}
}

// Schedule is this application's scheduled work. The site has none: the manual
// is read once at start, so there is nothing to refresh.
func Schedule() []foundation.Task { return nil }
