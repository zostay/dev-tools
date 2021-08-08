package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "zxconfig",
	Short: "Work with .zx.toml files.",
}

func init() {
	cobra.OnInitialize(config.Init)

	rootCmd.AddCommand(envCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
