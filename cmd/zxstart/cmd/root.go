package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zxstart",
	Short: "Start development servers.",
}

// init sets up the zxstart command.
func init() {
	rootCmd.AddCommand(serverCmd)
}

// Execute runst eh zxstart command.
func Execute() error {
	return rootCmd.Execute()
}
