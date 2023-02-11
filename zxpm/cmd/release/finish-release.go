package release

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/release"
)

var (
	finishReleaseCmd = &cobra.Command{
		Use:   "finish",
		Short: "complete the release process",
		Args:  cobra.NoArgs,
		RunE:  FinishRelease,
	}
)

func FinishRelease(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	process, err := release.NewProcessContinuation(ctx, MakeReleaseConfig())
	if err != nil {
		return err
	}

	process.CaptureChangesInfo()
	process.CheckReadyForMerge(ctx)
	process.MergePullRequest(ctx)
	process.TagRelease()
	process.CreateRelease(ctx)

	return nil
}
