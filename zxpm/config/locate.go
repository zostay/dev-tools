package config

import (
	"os"
	"path/filepath"

	"github.com/zostay/dev-tools/zxpm/storage"
)

func LocateAndLoadHome() (*Config, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}

	homeConf := filepath.Join(userDir, ".zxpm.conf")
	r, err := os.Open(homeConf)
	if err != nil {
		return nil, nil
	}

	return Load(homeConf, r)
}

func LocateAndLoadProject() (*Config, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, nil
	}

	for i := 0; i < 3; i++ {
		curConf := filepath.Join(curDir, "zxpm.conf")
		r, err := os.Open(curConf)
		if err == nil {
			return Load(curConf, r)
		}

		// if we encounter a go.mod, assume this is the project dir
		goMod := filepath.Join(curDir, "go.mod")
		goModStat, err := os.Stat(goMod)
		if err == nil && !goModStat.IsDir() {
			return nil, nil
		}

		// if we encounter a .git, assume this is the project dir
		gitDir := filepath.Join(curDir, ".git")
		gitDirStat, err := os.Stat(gitDir)
		if err == nil && !gitDirStat.IsDir() {
			return nil, nil
		}

		curConf = filepath.Dir(curConf)
	}

	return nil, nil
}

func DefaultConfig() *Config {
	return &Config{
		Properties: storage.New().RO(),
		Plugins: []PluginConfig{
			{
				Name:       "goals",
				Command:    "zxpm-plugin-goals",
				Properties: storage.New().RO(),
			},
			{
				Name:       "changelog",
				Command:    "zxpm-plugin-changelog",
				Properties: storage.New().RO(),
			},
			{
				Name:       "git",
				Command:    "zxpm-plugin-git",
				Properties: storage.New().RO(),
			},
			{
				Name:       "github",
				Command:    "zxpm-plugin-github",
				Properties: storage.New().RO(),
			},
		},
	}
}

func LocateAndLoad() (*Config, error) {
	cfg, err := LocateAndLoadProject()
	if err != nil || cfg != nil {
		return cfg, err
	}

	cfg, err = LocateAndLoadHome()
	if err != nil || cfg != nil {
		return cfg, err
	}

	return DefaultConfig(), nil
}
