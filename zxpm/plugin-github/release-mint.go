package plugin_github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleaseMintTask struct {
	plugin.Boilerplate
	Github
}

// CreateGithubPullRequest creates the PR on github for monitoring the test
// results for release testing. This will also be used to merge the release
// branch when testing passes.
func (s *ReleaseMintTask) CreateGithubPullRequest(ctx context.Context) error {
	owner, project, err := s.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	branch, err := ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	_, _, err = s.gh.PullRequests.Create(ctx, owner, project, &github.NewPullRequest{
		Title: github.String("Release v" + ReleaseVersion(ctx)),
		Head:  github.String(branch),
		Base:  github.String(TargetBranch(ctx)),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", ReleaseVersion(ctx))),
	})

	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	return nil
}

func (s *ReleaseMintTask) End(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  80,
			Action: plugin.OperationFunc(s.CreateGithubPullRequest),
		},
	}, nil
}
