package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zxconfig",
	Short: "Work with .zx.toml files.",
}

var verbosity int

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase command verbosity")
	rootCmd.AddCommand(envCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
