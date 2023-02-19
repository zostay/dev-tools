package plugin_changelog

import (
	"github.com/zostay/dev-tools/zxpm/plugin-changelog/cmd"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/release"
)

type Plugin struct{}

func (p *Plugin) Implements() ([]string, error) {
	return []string{release.StartTask, release.FinishTask}, nil
}

func (p *Plugin) Prepare(
	task string,
	cfg *config.Config,
) (plugin.Task, error) {
	pluginCfg := cfg.Section("github.com/zostay/zxpm/plugin-changelog")

	switch task {
	case release.StartTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseStartTask{
			Version:   releaseCfg.Version.String(),
			Today:     releaseCfg.Today,
			Changelog: pluginCfg.Get("changelog"),
		}, nil
	case release.FinishTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseFinishTask{
			Version:   releaseCfg.Version,
			Changelog: pluginCfg.Get("changelog"),
		}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Commands() []plugin.CommandDescriptor {
	return []plugin.CommandDescriptor{
		{
			Command: cmd.Changelog,
		},
		{
			Parents: []string{"changelog"},
			Command: cmd.ExtractChangelog,
		},
		{
			Parents: []string{"changelog"},
			Command: cmd.LintChangelog,
		},
	}
}
