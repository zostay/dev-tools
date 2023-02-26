package githubImpl

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
	zxGithub "github.com/zostay/dev-tools/zxpm/plugin-github/pkg/github"
)

type ReleaseMintTask struct {
	plugin.Boilerplate
	zxGithub.Github
}

// CreateGithubPullRequest creates the PR on github for monitoring the test
// results for release testing. This will also be used to merge the release
// branch when testing passes.
func (s *ReleaseMintTask) CreateGithubPullRequest(ctx context.Context) error {
	owner, project, err := s.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	branch, err := zxGithub.ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	prName := "Release v" + zxGithub.ReleaseVersion(ctx)
	_, _, err = s.Client().PullRequests.Create(ctx, owner, project, &github.NewPullRequest{
		Title: github.String(prName),
		Head:  github.String(branch),
		Base:  github.String(zxGithub.TargetBranch(ctx)),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", zxGithub.ReleaseVersion(ctx))),
	})

	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	plugin.Logger(ctx,
		"operation", "CreateGithubPullRequest",
		"owner", owner,
		"project", project,
		"branch", branch,
		"pullRequestName", prName,
	).Infof("Created Github pull request %q", prName)

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
