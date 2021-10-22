package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const ZxPrefix = "ZX"

type Config struct {
	App string

	Install   `mapstructure:"install"`
	SQLBoiler `mapstructure:"sqlboiler"`
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
	viper.SetConfigType("yaml")

	if verbosity > 0 {
		fmt.Fprintf(os.Stderr, "Reading configuration for defaults.yaml\n")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read defaults.yaml: %v\n", err)
	}

	viper.SetConfigName(".zx")
	viper.SetConfigType("yaml")

	if err := viper.MergeInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read .zx.yaml: %v\n", err)
	} else if verbosity > 0 {
		fmt.Fprintln(os.Stderr, "Read configuration for .zx.yaml")
	}

	envPrefix := viper.GetString("env_prefix")
	if envPrefix == "" {
		envPrefix = viper.GetString("app")
	}
	if envPrefix != "" {
		viper.SetEnvPrefix(ZxPrefix)
		viper.AutomaticEnv()
	} else {
		fmt.Fprintln(os.Stderr, "Not loading environment config. Please set the \"app\" or \"env_prefix\" key in .zx.yaml.")
	}

	// Expand some keys
	for _, k := range viper.AllKeys() {
		switch v := viper.Get(k).(type) {
		case []interface{}:
			if strings.HasSuffix(k, ".run") {
				vs := make([]string, len(v))
				for i, iv := range v {
					if siv, ok := iv.(string); ok {
						if strings.ContainsRune(siv, '$') {
							vs[i] = ExpandStdValue(siv)
						} else {
							vs[i] = siv
						}
					} else {
						vs[i] = fmt.Sprintf("%s", iv)
					}
				}
				viper.Set(k, vs)
			}
		}
	}
}

func Get() (*Config, error) {
	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}

func BasicEnv(n string) string {
	switch n {
	case "HOME":
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		} else {
			return ""
		}
	case "PWD":
		wd, err := os.Getwd()
		if err == nil {
			return wd
		} else {
			return ""
		}
	default:
		return ""
	}
}

func ExpandStdValue(v string) string {
	return os.Expand(v, BasicEnv)
}
