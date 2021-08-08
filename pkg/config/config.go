package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	App string

	Install   `toml:"install"`
	SQLBoiler `toml:"sqlboiler"`
}

func Init(verbosity int) {
	userdir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find HOME: %v\n", err)
	} else {

		viper.SetConfigName("defaults")
		viper.AddConfigPath(filepath.Join(userdir, ".zx"))
		viper.SetConfigType("toml")

		if verbosity > 0 {
			fmt.Fprintf(os.Stderr, "Reading configuration from %q\n", filepath.Join(userdir, ".zx/defaults.toml"))
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read ~/.zx/defaults.toml: %v\n", err)
	}

	viper.SetConfigName(".zx")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	if err := viper.MergeInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read zxdefaults.toml: %v\n", err)
	} else if verbosity > 0 {
		fmt.Fprintf(os.Stderr, "Reading configuration from %q\n", "./.zx.toml")
	}

	viper.AutomaticEnv()
}
