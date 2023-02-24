package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/cmd/config"
	config2 "github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
)

var (
	rootCmd = &cobra.Command{
		Use:   "zxpm",
		Short: "Golang project management tools by zostay",
	}
)

func init() {
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(templateFileCmd)
	rootCmd.AddCommand(runCmd)
}

func Execute() {
	cfg, err := config2.LocateAndLoad()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "zxpm failed to load: %v\n", err)
		os.Exit(1)
	}

	plugins, err := metal.LoadPlugins(cfg)
	if err != nil {
		panic(err) // TODO Fix this panic, it's temporary
	}
	defer metal.KillPlugins(plugins)

	err = configureTasks(cfg, plugins, rootCmd)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "zxpm failed to configure goals: %v\n", err)
		os.Exit(1)
	}

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
