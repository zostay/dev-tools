package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleaseStartTask struct {
	plugin.Boilerplate
	Github

	Version string

	Owner   string
	Project string

	Branch       string
	TargetBranch string
}

// CreateGithubPullRequest creates the PR on github for monitoring the test
// results for release testing. This will also be used to merge the release
// branch when testing passes.
func (s *ReleaseStartTask) CreateGithubPullRequest(ctx context.Context) error {
	_, _, err := s.gh.PullRequests.Create(ctx, s.Owner, s.Project, &github.NewPullRequest{
		Title: github.String("Release v" + s.Version),
		Head:  github.String(s.Branch),
		Base:  github.String(s.TargetBranch),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", s.Version)),
	})

	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	return nil
}

func (s *ReleaseStartTask) End(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  80,
			Action: plugin.OperationFunc(s.CreateGithubPullRequest),
		},
	}, nil
}
