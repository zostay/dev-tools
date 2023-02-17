package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v49/github"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/release"
)

type ReleaseFinishTask struct {
	plugin.Boilerplate
	Github

	Version string

	Owner   string
	Project string

	TargetBranch string
	Branch       string
	Tag          string
}

// CheckReadyForMerge ensures that all the required tests are passing.
func (f *ReleaseFinishTask) CheckReadyForMerge(ctx context.Context) error {
	bp, _, err := f.gh.Repositories.GetBranchProtection(ctx, f.Owner, f.Project, f.TargetBranch)
	if err != nil {
		return fmt.Errorf("unable to get branches %s: %w", f.Branch, err)
	}

	checks := bp.GetRequiredStatusChecks().Checks
	passage := make(map[string]bool, len(checks))
	for _, check := range checks {
		passage[check.Context] = false
	}

	crs, _, err := f.gh.Checks.ListCheckRunsForRef(ctx, f.Owner, f.Project, f.Branch, &github.ListCheckRunsOptions{})
	if err != nil {
		return fmt.Errorf("unable to list check runs for branch %s: %w", f.Branch, err)
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

func (f *ReleaseFinishTask) Check(ctx context.Context) error {
	return f.CheckReadyForMerge(ctx)
}

// MergePullRequest merges the PR into master.
func (f *ReleaseFinishTask) MergePullRequest(ctx context.Context) error {
	prs, _, err := f.gh.PullRequests.List(ctx, f.Owner, f.Project, &github.PullRequestListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list pull requests: %w", err)
	}

	prId := 0
	for _, pr := range prs {
		if pr.Head.GetRef() == f.Branch {
			prId = pr.GetNumber()
			break
		}
	}

	if prId == 0 {
		return fmt.Errorf("cannot find pull request for branch %s", f.Branch)
	}

	m, _, err := f.gh.PullRequests.Merge(ctx, f.Owner, f.Project, prId, "Merging release branch.", &github.PullRequestOptions{})
	if err != nil {
		return fmt.Errorf("unable to merge pull request %d: %w", prId, err)
	}

	if !m.GetMerged() {
		return fmt.Errorf("failed to merge pull request %d", prId)
	}

	return nil
}

// CreateRelease creates a release on github for the release.
func (f *ReleaseFinishTask) CreateRelease(ctx context.Context) error {
	changesInfo := plugin.Get(ctx, release.ValueDescription)
	releaseName := fmt.Sprintf("Release v%s", f.Version)
	_, _, err := f.gh.Repositories.CreateRelease(ctx, f.Owner, f.Project, &github.RepositoryRelease{
		TagName:              github.String(f.Tag),
		Name:                 github.String(releaseName),
		Body:                 github.String(changesInfo),
		Draft:                github.Bool(false),
		Prerelease:           github.Bool(false),
		GenerateReleaseNotes: github.Bool(false),
		MakeLatest:           github.String("true"),
	})

	if err != nil {
		return fmt.Errorf("failed to create release %q: %w", releaseName, err)
	}

	return nil
}

func (f *ReleaseFinishTask) Run(context.Context) (plugin.Operations, error) {
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
