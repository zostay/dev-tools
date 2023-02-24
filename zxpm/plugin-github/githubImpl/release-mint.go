package githubImpl

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
	github2 "github.com/zostay/dev-tools/zxpm/plugin-github/pkg/github"
)

type ReleaseMintTask struct {
	plugin.Boilerplate
	github2.Github
}

// CreateGithubPullRequest creates the PR on github for monitoring the test
// results for release testing. This will also be used to merge the release
// branch when testing passes.
func (s *ReleaseMintTask) CreateGithubPullRequest(ctx context.Context) error {
	owner, project, err := s.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	branch, err := github2.ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	_, _, err = s.Client().PullRequests.Create(ctx, owner, project, &github.NewPullRequest{
		Title: github.String("Release v" + github2.ReleaseVersion(ctx)),
		Head:  github.String(branch),
		Base:  github.String(github2.TargetBranch(ctx)),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", github2.ReleaseVersion(ctx))),
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