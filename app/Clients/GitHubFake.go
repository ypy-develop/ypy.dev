package clients

import "context"

// GitHubFake is GitHub for the tests: it answers what it was told to, records
// what it was asked, and never reaches the network. Code under test is handed
// one in place of GitHubClient.
type GitHubFake struct {
	// StargazersAnswer is what Stargazers answers.
	StargazersAnswer int64
	// Err, when set, is what every call answers instead.
	Err error
	// Calls are the methods called, in order.
	Calls []string
	// Repositories are the repositories Stargazers was asked about, in order.
	Repositories []string
}

// Compile-time proof that the fake stands in for the client.
var _ GitHub = (*GitHubFake)(nil)

// arandu:begin custom

// Stargazers records the call and the repository, and answers
// StargazersAnswer, or Err.
func (f *GitHubFake) Stargazers(_ context.Context, repository string) (int64, error) {
	f.Calls = append(f.Calls, "Stargazers")
	f.Repositories = append(f.Repositories, repository)
	if f.Err != nil {
		return 0, f.Err
	}
	return f.StargazersAnswer, nil
}

// arandu:end custom
