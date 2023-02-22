package config

import (
	"io"

	"github.com/hashicorp/hcl/v2/hclsimple"

	"github.com/zostay/dev-tools/zxpm/storage"
)

type Config struct {
	Properties storage.KV

	Goals   []GoalConfig
	Plugins []PluginConfig
}

type GoalConfig struct {
	Name string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV

	Tasks   []TaskConfig
	Targets []TargetConfig
}

type PluginConfig struct {
	Name    string
	Package string

	Properties storage.KV

	Goals []GoalConfig
}

type TaskConfig struct {
	Name    string
	SubTask string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV

	Targets []TargetConfig
	Tasks   []TaskConfig
}

type TargetConfig struct {
	Name string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV
}

func Load(filename string, in io.Reader) (*Config, error) {
	var raw RawConfig
	fileBytes, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	err = hclsimple.Decode(filename, fileBytes, nil, &raw)
	if err != nil {
		return nil, err
	}

	return decodeRawConfig(&raw)
}
