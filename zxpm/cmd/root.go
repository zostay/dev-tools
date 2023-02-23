package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/cmd/config"
	"github.com/zostay/dev-tools/zxpm/cmd/release"
	config2 "github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
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
	cfg, err := config2.LocateAndLoad()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "zxpm failed to load: %v", err)
		os.Exit(1)
	}

	plugins := plugin.LoadPlugins(cfg)
	defer plugin.KillPlugins(plugins)

	configureTasks(cfg, plugins, rootCmd)

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
