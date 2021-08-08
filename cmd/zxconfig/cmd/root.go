package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zxconfig",
	Short: "Work with .zx.toml files.",
}

func init() {
	rootCmd.AddCommand(envCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
