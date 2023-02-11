package changelog

import (
	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/release"
)

const todayFormat = "2006-01-02"

type Plugin struct{}

func (p *Plugin) Implements() []string {
	return []string{"start-release", "finish-release"}
}

func (p *Plugin) Prepare(task string, cfg *config.Config, taskConfig any) plugin.Task {
	switch task {
	case "start-release":
		releaseCfg := taskConfig.(*release.TaskConfig)
		return &ReleaseStartTask{
			Version:   releaseCfg.Version,
			Today:     releaseCfg.Now.Format(todayFormat),
			Changelog: cfg.Paths["changelog"],
		}
	case "finish-release":
		releaseCfg := taskConfig.(*release.TaskConfig)
		return &ReleaseFinishTask{
			Version:   releaseCfg.SemVer(),
			Changelog: cfg.Paths["changelog"],
		}
	}
	return nil
}
