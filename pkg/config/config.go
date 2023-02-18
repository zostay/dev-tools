package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// ZxPrefix is the environment prefix to add before dev-tools environment
// variables.
const ZxPrefix = "ZX"

type Config struct {
	Values      map[string]string
	SubSections map[string]*Config
}

func New() *Config {
	return NewCap(10, 10)
}

func NewCap(vCap, ssCap int) *Config {
	return &Config{
		Values:      make(map[string]string, vCap),
		SubSections: make(map[string]*Config, ssCap),
	}
}

func (c *Config) navigate(key string) (*Config, string) {
	parts := strings.SplitN(key, ".", 2)
	thisKey := parts[0]
	if len(parts) == 1 {
		return c, thisKey
	}
	return c.SubSections[thisKey].navigate(parts[1])
}

func (c *Config) Get(key string) string {
	cfg, key := c.navigate(key)
	return cfg.Values[key]
}

func (c *Config) Section(key string) *Config {
	cfg, key := c.navigate(key)
	return cfg.SubSections[key]
}

func FromAnyMap(in map[string]any) *Config {
	out := NewCap(len(in), len(in))
	for k, v := range in {
		if nestedV, isNested := v.(map[string]any); isNested {
			out.SubSections[k] = FromAnyMap(nestedV)
		} else {
			out.Values[k] = cast.ToString(v)
		}
	}
	return out
}

func FromViper() *Config {
	return FromAnyMap(viper.AllSettings())
}

// Init initializes the ZX configuration. If verbosity is set to a non-zero
// value it reports which configuration files it is reading from. Use the Get()
// function to return the configuration as a pointer to a Config object.
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

	// TODO The value interpolation for .run keys is pretty half-assed and needs
	// fixing.

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

// BasicEnv is a tool for mapping HOME and PWD environment variables in some
// values in the configuration file which can be interpolated.
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

// ExpandStdValue performs environment value interpolation with variables
// returned by BasicEnv.
func ExpandStdValue(v string) string {
	return os.Expand(v, BasicEnv)
}
