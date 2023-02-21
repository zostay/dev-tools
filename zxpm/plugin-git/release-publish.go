package plugin_git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleasePublishTask struct {
	plugin.Boilerplate
	Git
}

func (f *ReleasePublishTask) Setup(ctx context.Context) error {
	return f.SetupGitRepo(ctx)
}

func (s *ReleasePublishTask) Begin(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  20,
			Action: plugin.OperationFunc(SetDefaultReleaseTag),
		},
	}, nil
}

// TagRelease creates and pushes a tag for the newly merged release on master.
func (f *ReleasePublishTask) TagRelease(ctx context.Context) error {
	err := f.wc.Checkout(&git.CheckoutOptions{
		Branch: TargetBranchRefName(ctx),
	})
	if err != nil {
		return fmt.Errorf("unable to switch to %s branch: %w", TargetBranch(ctx), err)
	}

	headRef, err := f.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get HEAD ref of %s branch: %w", TargetBranch(ctx), err)
	}

	tag, err := ReleaseTag(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine release tag: %w", err)
	}

	head := headRef.Hash()
	_, err = f.repo.CreateTag(tag, head, &git.CreateTagOptions{
		Message: fmt.Sprintf("Release tag for v%s", ReleaseVersion(ctx)),
	})
	if err != nil {
		return fmt.Errorf("unable to tag release %s: %w", tag, err)
	}

	plugin.ForCleanup(ctx, func() { _ = f.repo.DeleteTag(tag) })

	tagRefSpec, err := ReleaseTagRefSpec(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine release tag ref spec: %w", err)
	}

	err = f.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{tagRefSpec},
	})
	if err != nil {
		return fmt.Errorf("unable to push tags to origin: %w", err)
	}

	plugin.ForCleanup(ctx, func() {
		_ = f.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{tagRefSpec},
			Prune:      true,
		})
	})

	return nil
}

func (f *ReleasePublishTask) End(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  75,
			Action: plugin.OperationFunc(f.TagRelease),
		},
	}, nil
}
