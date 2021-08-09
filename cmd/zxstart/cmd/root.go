package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zxstart",
	Short: "Start development servers.",
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
