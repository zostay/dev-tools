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

func Init() {
	userdir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find HOME: %v\n", err)
	}

	viper.SetConfigName("zxdefaults")
	viper.AddConfigPath(filepath.Join(userdir, ".xz"))
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read .xz.toml: %v\n", err)
	}

	viper.SetConfigName(".zx")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	if err := viper.MergeInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read zxdefaults.toml: %v\n", err)
	}

	viper.AutomaticEnv()
}
