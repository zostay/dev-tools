package cmd

// TODO Convert this to use TOML. I have begun to loathe TOML files and want to
// always prefer YAML again instead.

import (
	"os"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zostay/dev-tools/pkg/config"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Output the contents of ZX configs as a TOML",
	RunE:  RunMerge,
}

// RunMerge loads the .zx.toml and whatever other configuration is present and
// detectable by the configuration tooling and outputs a fresh TOML file of all
// the gathered configuration information.
func RunMerge(cmd *cobra.Command, args []string) error {
	config.Init(verbosity)
	bs, err := toml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}

	os.Stdout.Write(bs)
	return nil
}
