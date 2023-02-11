package release

import (
	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/release"
)

var (
	Cmd = &cobra.Command{
		Use:   "release",
		Short: "commands related to software releases",
	}

	targetBranch string
)

func init() {
	Cmd.AddCommand(startReleaseCmd)
	Cmd.AddCommand(finishReleaseCmd)

	Cmd.PersistentFlags().StringVar(&targetBranch, "target-branch", "master", "the branch to merge into during release")
}

func MakeReleaseConfig() *release.Config {
	cfg := release.GoEmailConfig
	cfg.TargetBranch = targetBranch
	return &cfg
}
