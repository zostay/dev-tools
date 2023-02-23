package gitImpl

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleaseMintTask struct {
	plugin.Boilerplate
	Git
}

func (s *ReleaseMintTask) Setup(ctx context.Context) error {
	return s.SetupGitRepo(ctx)
}

// IsDirty returns true if we consider the tree dirty. We do not consider
// Untracked to dirty the directory and we also ignore some filenames that are
// in the global .gitignore and not in the local .gitignore.
func IsDirty(status git.Status) bool {
	for fn, fstat := range status {
		if _, ignorable := ignoreStatus[fn]; ignorable {
			continue
		}

		if fstat.Worktree != git.Unmodified && fstat.Worktree != git.Untracked {
			return true
		}

		if fstat.Staging != git.Unmodified && fstat.Staging != git.Untracked {
			return true
		}
	}
	return false
}

// CheckGitCleanliness ensures that the current git repository is clean and that
// we are on the correct branch from which to trigger a release.
func (s *ReleaseMintTask) CheckGitCleanliness(ctx context.Context) error {
	headRef, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to find HEAD: %w", err)
	}

	if headRef.Name() != TargetBranchRefName(ctx) {
		return fmt.Errorf("you must checkout %s to release", TargetBranch(ctx))
	}

	remoteRefs, err := s.remote.List(&git.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list remote git references: %w", err)
	}

	var masterRef *plumbing.Reference
	for _, ref := range remoteRefs {
		if ref.Name() == TargetBranchRefName(ctx) {
			masterRef = ref
			break
		}
	}

	if headRef.Hash() != masterRef.Hash() {
		return fmt.Errorf("local copy differs from remote, you need to push or pull")
	}

	stat, err := s.wc.Status()
	if err != nil {
		return fmt.Errorf("unable to check working copy status: %w", err)
	}

	if IsDirty(stat) {
		return fmt.Errorf("your working copy is dirty")
	}

	return nil
}

func (s *ReleaseMintTask) Check(ctx context.Context) error {
	return s.CheckGitCleanliness(ctx)
}

// MakeReleaseBranch creates the branch that will be used to manage the release.
func (s *ReleaseMintTask) MakeReleaseBranch(ctx context.Context) error {
	headRef, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to retrieve the HEAD ref: %w", err)
	}

	branchRefName, err := ReleaseBranchRefName(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine release branch references: %w", err)
	}

	branch, _ := ReleaseBranch(ctx)
	err = s.wc.Checkout(&git.CheckoutOptions{
		Hash:   headRef.Hash(),
		Branch: branchRefName,
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("unable to checkout branch %s: %v", branch, err)
	}

	plugin.ForCleanup(ctx, func() {
		_ = s.repo.Storer.RemoveReference(branchRefName)
	})
	plugin.ForCleanup(ctx, func() {
		_ = s.wc.Checkout(&git.CheckoutOptions{
			Branch: TargetBranchRefName(ctx),
		})
	})

	return nil
}

func (s *ReleaseMintTask) Begin(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  20,
			Action: plugin.OperationFunc(SetDefaultReleaseBranch),
		},
	}, nil
}

func (s *ReleaseMintTask) Run(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  30,
			Action: plugin.OperationFunc(s.MakeReleaseBranch),
		},
	}, nil
}

// AddAndCommit adds changes made as part of the release process to the release
// branch.
func (s *ReleaseMintTask) AddAndCommit(ctx context.Context) error {
	addedFiles := plugin.ListAdded(ctx)
	for _, fn := range addedFiles {
		_, err := s.wc.Add(fn)
		if err != nil {
			return fmt.Errorf("error adding file %s to git: %w", fn, err)
		}
	}

	version := plugin.GetString(ctx, "release.version")
	_, err := s.wc.Commit("releng: v"+version, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("error committing changes to git: %w", err)
	}

	return nil
}

// PushReleaseBranch pushes the release branch to github for release testing.
func (s *ReleaseMintTask) PushReleaseBranch(ctx context.Context) error {
	branchRefSpec, err := ReleaseBranchRefSpec(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine the ref spec: %w", err)
	}

	err = s.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{branchRefSpec},
	})
	if err != nil {
		return fmt.Errorf("error pushing changes to github: %w", err)
	}

	plugin.ForCleanup(ctx, func() {
		_ = s.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{branchRefSpec},
			Prune:      true,
		})
	})

	return nil
}

func (s *ReleaseMintTask) End(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  70,
			Action: plugin.OperationFunc(s.AddAndCommit),
		},
		{
			Order:  75,
			Action: plugin.OperationFunc(s.PushReleaseBranch),
		},
	}, nil
}
