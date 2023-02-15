package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/tools"
)

type ReleaseFinishTask struct {
	plugin.Boilerplate
	Git

	Version string

	TargetBranch        string
	TargetBranchRefName plumbing.ReferenceName

	Tag        string
	TagRefName plumbing.ReferenceName
	TagRefSpec config.RefSpec
}

func (f *ReleaseFinishTask) Setup(_ context.Context) error {
	return f.SetupGitRepo()
}

// TagRelease creates and pushes a tag for the newly merged release on master.
func (f *ReleaseFinishTask) TagRelease(ctx context.Context) error {
	err := f.wc.Checkout(&git.CheckoutOptions{
		Branch: f.TargetBranchRefName,
	})
	if err != nil {
		return fmt.Errorf("unable to switch to %s branch: %w", f.TargetBranch, err)
	}

	headRef, err := f.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get HEAD ref of %s branch: %w", f.TargetBranch, err)
	}

	head := headRef.Hash()
	_, err = f.repo.CreateTag(f.Tag, head, &git.CreateTagOptions{
		Message: fmt.Sprintf("Release tag for v%s", f.Version),
	})
	if err != nil {
		return fmt.Errorf("unable to tag release %s: %w", f.Tag, err)
	}

	tools.ForCleanup(ctx, func() { _ = f.repo.DeleteTag(f.Tag) })

	err = f.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{f.TagRefSpec},
	})
	if err != nil {
		return fmt.Errorf("unable to push tags to origin: %w", err)
	}

	tools.ForCleanup(ctx, func() {
		_ = f.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{f.TagRefSpec},
			Prune:      true,
		})
	})

	return nil
}

func (f *ReleaseFinishTask) End() plugin.Operations {
	return plugin.Operations{
		{
			Order:  75,
			Action: f.TagRelease,
		},
	}
}
