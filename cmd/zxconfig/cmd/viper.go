package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zostay/dev-tools/pkg/config"
)

var viperCmd = &cobra.Command{
	Use:   "viper",
	Short: "See things as viper sees things, for debugging config.",
	RunE:  RunViper,
}

// RunViper iterates through the keys viper knows about and displays them on
// standard output.
func RunViper(cmd *cobra.Command, args []string) error {
	config.Init(verbosity)
	for _, k := range viper.AllKeys() {
		v := viper.Get(k)
		fmt.Printf("%q = %q\n", k, v)
	}
	return nil
}
