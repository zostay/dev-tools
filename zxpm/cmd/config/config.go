package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Commands for working with ZX configuration.",
}

// init initializes the config subcommand.
func init() {
	Cmd.PersistentFlags().CountP("verbose", "v", "increase command verbosity")

	Cmd.AddCommand(EnvCmd)
	Cmd.AddCommand(MergeCmd)
	Cmd.AddCommand(ViperCmd)
}
