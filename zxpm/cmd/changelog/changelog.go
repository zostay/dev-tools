package changelog

import "github.com/spf13/cobra"

var (
	Cmd = &cobra.Command{
		Use:   "changelog",
		Short: "Commands related to change logs",
	}
)

func init() {
	Cmd.AddCommand(lintChangelogCmd)
	Cmd.AddCommand(extractChangelogCmd)
}
