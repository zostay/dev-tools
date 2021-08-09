package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	App       string
	EnvPrefix string `toml:"env_prefix"`

	Install   `toml:"install"`
	SQLBoiler `toml:"sqlboiler"`
	Web
}

func Init(verbosity int) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find PWD: %v\n", err)
	} else {
		viper.AddConfigPath(wd)
	}

	userdir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find HOME: %v\n", err)
	} else {
		viper.AddConfigPath(filepath.Join(userdir, ".zx"))
	}

	viper.SetConfigName("defaults")
	viper.SetConfigType("toml")

	if verbosity > 0 {
		fmt.Fprintf(os.Stderr, "Reading configuration for defaults.toml\n")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read defaults.toml: %v\n")
	}

	viper.SetConfigName(".zx")
	viper.SetConfigType("toml")

	if err := viper.MergeInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read .zx.toml: %v\n", err)
	} else if verbosity > 0 {
		fmt.Fprintln(os.Stderr, "Read configuration for .zx.toml")
	}

	envPrefix := viper.GetString("env_prefix")
	if envPrefix == "" {
		envPrefix = viper.GetString("app")
	}
	if envPrefix != "" {
		viper.SetEnvPrefix(envPrefix)
		viper.AutomaticEnv()
	} else {
		fmt.Fprintln(os.Stderr, "Not loading environment config. Please set the \"app\" or \"env_prefix\" key in .zx.toml.")
	}
}

func Get() (*Config, error) {
	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}
