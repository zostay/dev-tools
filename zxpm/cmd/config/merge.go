package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/zostay/dev-tools/pkg/config"
)

var MergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Output the contents of a completely merged ZX config as YAML",
	RunE:  RunMerge,
}

// RunMerge loads the .zx.yaml and whatever other configuration is present and
// detectable by the configuration tooling and outputs a fresh YAML file of all
// the gathered configuration information.
func RunMerge(cmd *cobra.Command, _ []string) error {
	config.Init(verbosity(cmd))
	bs, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}

	os.Stdout.Write(bs)
	return nil
}
