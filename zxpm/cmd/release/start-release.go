package release

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/release"
)

var (
	startReleaseCmd = &cobra.Command{
		Use:   "start <version>",
		Short: "Start a release",
		Args:  cobra.ExactArgs(1),
		RunE:  StartRelease,
	}
)

func StartRelease(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	process, err := release.NewProcess(ctx, args[0], MakeReleaseConfig())
	if err != nil {
		return err
	}

	process.SetupGitRepo()
	process.CheckGitCleanliness()
	process.LintChangelog(changes.CheckPreRelease)
	process.MakeReleaseBranch()
	process.FixupChangelog()
	process.LintChangelog(changes.CheckRelease)
	process.AddAndCommit()
	process.PushReleaseBranch()
	process.CreateGithubPullRequest(ctx)

	return nil
}
