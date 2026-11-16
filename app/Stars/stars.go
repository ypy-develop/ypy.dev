// Package stars is how many stars the YPY repository has, and the one number
// on this site that is not a fact about the tree.
//
// Every other figure the site states is read from something it holds: the
// names from the reference tables, the version from the VERSION file YPY ships.
// This one is held by GitHub and changes daily, so a number typed into markup
// would be stale by the end of the week.
//
// So it is fetched, and the honest consequence is written into the shape:
// nothing on a page waits for the network, and a count nobody has fetched yet
// is absent. An explicit zero is rendered as zero; an empty string is the mark
// on its own, which is what the header drew before this existed.
package stars

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	clients "github.com/ypy-develop/ypy.dev/app/Clients"
)

// Repository is the YPY repository the header's GitHub button counts, as GitHub
// names it.
//
// It is a constant and not an argument. One caller, one repository, and a
// configurable one would be a second place the answer comes from -- a test
// hands Refresh a fake or a client over a faked factory instead.
const Repository = "ypy-develop/ypy"

// count is the last total fetched, and available says whether one succeeded.
//
// Atomics and not a mutex: they are written by one task and read by every
// request, and the read has to be free. Availability cannot be inferred from
// count because zero is a valid answer from GitHub. Refresh stores the count
// before publishing availability, so a first reader never sees an available
// value that has not been written yet.
var (
	count     atomic.Int64
	available atomic.Bool
)

// Count is the total, as the digits a page draws, and the empty string when
// nothing has been fetched.
//
// A caller writes it beside the mark without asking whether it is there: the
// empty string draws nothing, which is the header as it was before this
// existed.
func Count() string {
	if !available.Load() {
		return ""
	}
	return strconv.FormatInt(count.Load(), 10)
}

// Refresh asks GitHub for the total and stores it.
//
// It replaces the stored value only on success. A failed fetch -- rate limit,
// outage, a machine with no network -- leaves the last good number in place,
// because a header that loses its count when GitHub has a bad minute is worse
// than one showing a number an hour old.
func Refresh(ctx context.Context, github clients.GitHub) error {
	if github == nil {
		return errors.New("stars: no GitHub client")
	}
	total, err := github.Stargazers(ctx, Repository)
	if err != nil {
		return fmt.Errorf("stars: %w", err)
	}
	count.Store(total)
	available.Store(true)
	return nil
}
