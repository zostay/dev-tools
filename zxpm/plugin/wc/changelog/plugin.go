package changelog

import (
	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/wc/changelog/cmd"
	"github.com/zostay/dev-tools/zxpm/release"
)

type Plugin struct{}

func (p *Plugin) Implements() []string {
	return []string{release.StartTask, release.FinishTask}
}

func (p *Plugin) Prepare(
	task string,
	cfg *config.Config,
	taskConfig any,
) plugin.Task {
	switch task {
	case release.StartTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseStartTask{
			Version:   releaseCfg.Version.String(),
			Today:     releaseCfg.Today,
			Changelog: cfg.Paths["changelog"],
		}
	case release.FinishTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseFinishTask{
			Version:   releaseCfg.Version,
			Changelog: cfg.Paths["changelog"],
		}
	}
	return nil
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
