package plugin_github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleasePublishTask struct {
	plugin.Boilerplate
	Github
}

// CheckReadyForMerge ensures that all the required tests are passing.
func (f *ReleasePublishTask) CheckReadyForMerge(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	branch, err := ReleaseBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release branch name: %w", err)
	}

	bp, _, err := f.gh.Repositories.GetBranchProtection(ctx, owner, project, TargetBranch(ctx))
	if err != nil {
		return fmt.Errorf("unable to get branches %s: %w", branch, err)
	}

	checks := bp.GetRequiredStatusChecks().Checks
	passage := make(map[string]bool, len(checks))
	for _, check := range checks {
		passage[check.Context] = false
	}

	crs, _, err := f.gh.Checks.ListCheckRunsForRef(ctx, owner, project, branch, &github.ListCheckRunsOptions{})
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

	return nil
}

func (f *ReleasePublishTask) Check(ctx context.Context) error {
	return f.CheckReadyForMerge(ctx)
}

// MergePullRequest merges the PR into master.
func (f *ReleasePublishTask) MergePullRequest(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	prs, _, err := f.gh.PullRequests.List(ctx, owner, project, &github.PullRequestListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list pull requests: %w", err)
	}

	branch, err := ReleaseBranch(ctx)
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

	m, _, err := f.gh.PullRequests.Merge(ctx, owner, project, prId, "Merging release branch.", &github.PullRequestOptions{})
	if err != nil {
		return fmt.Errorf("unable to merge pull request %d: %w", prId, err)
	}

	if !m.GetMerged() {
		return fmt.Errorf("failed to merge pull request %d", prId)
	}

	return nil
}

// CreateRelease creates a release on github for the release.
func (f *ReleasePublishTask) CreateRelease(ctx context.Context) error {
	owner, project, err := f.OwnerProject(ctx)
	if err != nil {
		return fmt.Errorf("failed getting owner/project information: %w", err)
	}

	tag, err := ReleaseTag(ctx)
	if err != nil {
		return fmt.Errorf("failed to get release tag name: %w", err)
	}

	changesInfo := ReleaseDescription(ctx)
	releaseName := fmt.Sprintf("Release v%s", ReleaseVersion(ctx))
	_, _, err = f.gh.Repositories.CreateRelease(ctx, owner, project,
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
