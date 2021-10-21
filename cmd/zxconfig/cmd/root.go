package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zxconfig",
	Short: "Work with .zx.yaml files.",
}

var verbosity int // the verbosity level to use when working with zxconfig

// init initializes the zxconfig command.
func init() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase command verbosity")
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(viperCmd)
}

// Execute runs the zxconfig command.
func Execute() error {
	return rootCmd.Execute()
}
