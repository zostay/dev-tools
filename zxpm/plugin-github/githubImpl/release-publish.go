package githubImpl

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
	zxGithub "github.com/zostay/dev-tools/zxpm/plugin-github/pkg/github"
)

type ReleasePublishTask struct {
	plugin.Boilerplate
	zxGithub.Github
}

// CheckReadyForMerge ensures that all the required tests are passing.
func (f *ReleasePublishTask) CheckReadyForMerge(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	branch, err := zxGithub.ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	bp, _, err := f.Client().Repositories.GetBranchProtection(ctx, owner, project, zxGithub.TargetBranch(ctx))
	if err != nil {
		return fmt.Errorf("unable to get branches %s: %w", branch, err)
	}

	checks := bp.GetRequiredStatusChecks().Checks
	passage := make(map[string]bool, len(checks))
	for _, check := range checks {
		passage[check.Context] = false
	}

	crs, _, err := f.Client().Checks.ListCheckRunsForRef(ctx, owner, project, branch, &github.ListCheckRunsOptions{})
	if err != nil {
		return fmt.Errorf("unable to list check runs for branch %s: %w", branch, err)
	}

	for _, run := range crs.CheckRuns {
		passage[run.GetName()] =
			run.GetStatus() == "completed" &&
				run.GetConclusion() == "success"
	}

	for k, v := range passage {
		if !v {
			return fmt.Errorf("cannot merge release branch because it has not passed check %q", k)
		}
	}

	plugin.Logger(ctx,
		"operation", "CheckReadyForMerge",
		"owner", owner,
		"project", project,
		"branch", branch,
	).Infof("All Github required checks appear to be passing")

	return nil
}

func (f *ReleasePublishTask) Check(ctx context.Context) error {
	const timeout = 15 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var err error
	for {
		if ctx.Err() != nil {
			break
		}

		err = f.CheckReadyForMerge(ctx)
		if err != nil {
			<-time.After(30 * time.Second)
		}
	}

	return err
}

// MergePullRequest merges the PR into master.
func (f *ReleasePublishTask) MergePullRequest(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	prs, _, err := f.Client().PullRequests.List(ctx, owner, project, &github.PullRequestListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list pull requests: %w", err)
	}

	branch, err := zxGithub.ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	prId := 0
	for _, pr := range prs {
		if pr.Head.GetRef() == branch {
			prId = pr.GetNumber()
			break
		}
	}

	if prId == 0 {
		return fmt.Errorf("cannot find pull request for branch %s", branch)
	}

	m, _, err := f.Client().PullRequests.Merge(ctx, owner, project, prId, "Merging release branch.", &github.PullRequestOptions{})
	if err != nil {
		return fmt.Errorf("unable to merge pull request %d: %w", prId, err)
	}

	if !m.GetMerged() {
		return fmt.Errorf("failed to merge pull request %d", prId)
	}

	plugin.Logger(ctx,
		"operation", "MergePullRequest",
		"owner", owner,
		"project", project,
		"branch", branch,
		"pullRequestID", prId,
	).Infof("Marged the pull request into the target branch.")

	return nil
}

// CreateRelease creates a release on github for the release.
func (f *ReleasePublishTask) CreateRelease(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	tag, err := zxGithub.ReleaseTag(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release tag name: %w", err)
	}

	changesInfo := zxGithub.ReleaseDescription(ctx)
	releaseName := fmt.Sprintf("Release v%s", zxGithub.ReleaseVersion(ctx))
	_, _, err = f.Client().Repositories.CreateRelease(ctx, owner, project,
		&github.RepositoryRelease{
			TagName:              github.String(tag),
			Name:                 github.String(releaseName),
			Body:                 github.String(changesInfo),
			Draft:                github.Bool(false),
			Prerelease:           github.Bool(false),
			GenerateReleaseNotes: github.Bool(false),
			MakeLatest:           github.String("true"),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create release %q: %w", releaseName, err)
	}

	plugin.Logger(ctx,
		"owner", owner,
		"project", project,
		"tag", tag,
		"releaseName", releaseName,
	).Infof("Created a release named %q", releaseName)

	return nil
}

func (f *ReleasePublishTask) Run(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  70,
			Action: plugin.OperationFunc(f.MergePullRequest),
		},
		{
			Order:  75,
			Action: plugin.OperationFunc(f.CreateRelease),
		},
	}, nil
}
