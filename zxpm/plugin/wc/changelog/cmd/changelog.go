package cmd

import "github.com/spf13/cobra"

var (
	Changelog = &cobra.Command{
		Use:   "changelog",
		Short: "Commands related to change logs",
	}
)
