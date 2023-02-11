package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/cmd/changelog"
	"github.com/zostay/dev-tools/zxpm/cmd/config"
	"github.com/zostay/dev-tools/zxpm/cmd/release"
)

var (
	rootCmd = &cobra.Command{
		Use:   "zxpm",
		Short: "Golang project management tools by zostay",
	}
)

func init() {
	rootCmd.AddCommand(changelog.Cmd)
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(release.Cmd)
	rootCmd.AddCommand(templateFileCmd)
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
