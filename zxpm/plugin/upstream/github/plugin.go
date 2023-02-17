package github

import (
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
	_ *config.Config,
	taskConfig any,
) (plugin.Task, error) {
	switch task {
	case release.StartTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseStartTask{
			Version:      releaseCfg.Version.String(),
			Owner:        releaseCfg.Owner,
			Project:      releaseCfg.Project,
			Branch:       releaseCfg.Branch,
			TargetBranch: releaseCfg.TargetBranch,
		}, nil
	case release.FinishTask:
		releaseCfg := taskConfig.(*release.Config)
		return &ReleaseFinishTask{
			Version:      releaseCfg.Version.String(),
			Owner:        releaseCfg.Owner,
			Project:      releaseCfg.Project,
			TargetBranch: releaseCfg.TargetBranch,
			Branch:       releaseCfg.Branch,
			Tag:          releaseCfg.Tag,
		}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}
