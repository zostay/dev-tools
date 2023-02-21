package plugin_git

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

const (
	defaultReleaseBranchPrefix = "release-v"
	defaultReleaseTagPrefix    = "v"
)

var ignoreStatus = map[string]struct{}{
	".session.vim": {},
}

type Git struct {
	repo   *git.Repository
	remote *git.Remote
	wc     *git.Worktree
}

func ref(t, n string) plumbing.ReferenceName {
	return plumbing.ReferenceName(path.Join("refs", t, n))
}

func refSpec(r plumbing.ReferenceName) gitConfig.RefSpec {
	sr := string(r)
	return gitConfig.RefSpec(strings.Join([]string{sr, sr}, ":"))
}

func ReleaseVersion(ctx context.Context) string {
	return plugin.GetString(ctx, "release.version")
}

func TargetBranch(ctx context.Context) string {
	cfg := plugin.ConfigFor(ctx, gitPlugin)
	if plugin.IsConfigSet(ctx, gitPlugin+".target_branch") {
		return cfg.GetString("target_branch")
	}
	return "master"
}

func TargetBranchRefName(ctx context.Context) plumbing.ReferenceName {
	return ref("heads", TargetBranch(ctx))
}

func ReleaseBranch(ctx context.Context) (string, error) {
	if plugin.IsSet(ctx, "release.branch") {
		return plugin.GetString(ctx, "release.branch"), nil
	}
	return "", fmt.Errorf("missing required \"release.branch\" setting")
}

func ReleaseBranchRefName(ctx context.Context) (plumbing.ReferenceName, error) {
	branch, err := ReleaseBranch(ctx)
	if err != nil {
		return "", err
	}
	return ref("heads", branch), nil
}

func ReleaseBranchRefSpec(ctx context.Context) (gitConfig.RefSpec, error) {
	branchRefName, err := ReleaseBranchRefName(ctx)
	if err != nil {
		return "", err
	}
	return refSpec(branchRefName), nil
}

func ReleaseTag(ctx context.Context) (string, error) {
	if plugin.IsSet(ctx, "release.tag") {
		return plugin.GetString(ctx, "release.tag"), nil
	}
	return "", fmt.Errorf("missing required \"release.tag\" setting")
}

func ReleaseTagRefSpec(ctx context.Context) (gitConfig.RefSpec, error) {
	tag, err := ReleaseTag(ctx)
	if err != nil {
		return "", err
	}
	tagRefName := ref("tags", tag)
	return refSpec(tagRefName), nil
}

func SetDefaultReleaseBranch(ctx context.Context) error {
	if !plugin.IsSet(ctx, "release.branch") && ReleaseVersion(ctx) != "" {
		branchName := defaultReleaseBranchPrefix + ReleaseVersion(ctx)
		plugin.Set(ctx, "release.branch", branchName)
	}
	return nil
}

func SetDefaultReleaseTag(ctx context.Context) error {
	if !plugin.IsSet(ctx, "release.tag") && ReleaseVersion(ctx) != "" {
		tagName := defaultReleaseTagPrefix + ReleaseVersion(ctx)
		plugin.Set(ctx, "release.tag", tagName)
	}
	return nil
}

func (g *Git) SetupGitRepo(ctx context.Context) error {
	l, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("unable to open git repository at .: %w", err)
	}

	g.repo = l

	r, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("unable to connect to remote origin: %w", err)
	}

	g.remote = r

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to examine the working copy: %w", err)
	}

	g.wc = w

	return nil
}
