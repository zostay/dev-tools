package config

import "github.com/spf13/cobra"

// verbosity returns the setting on the verbose flag.
func verbosity(cmd *cobra.Command) int {
	v, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		panic(err)
	}
	return v
}
