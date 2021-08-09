package cmd

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

func RunMerge(cmd *cobra.Command, args []string) error {
	config.Init(verbosity)
	bs, err := toml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}

	os.Stdout.Write(bs)
	return nil
}
