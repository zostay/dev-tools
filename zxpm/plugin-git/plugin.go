package plugin_git

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
	cfg *config.Config,
	taskConfig any,
) (plugin.Task, error) {
	switch task {
	case release.StartTask:
		releaseCfg := taskConfig.(*release.Config)
		branchRefName := ref("heads", releaseCfg.Branch)
		return &ReleaseStartTask{
			Version:             releaseCfg.Version.String(),
			BranchRefSpec:       refSpec(branchRefName),
			Branch:              releaseCfg.Branch,
			BranchRefName:       branchRefName,
			TargetBranch:        releaseCfg.TargetBranch,
			TargetBranchRefName: ref("heads", releaseCfg.TargetBranch),
		}, nil
	case release.FinishTask:
		releaseCfg := taskConfig.(*release.Config)
		tagRefName := ref("tags", releaseCfg.Tag)
		return &ReleaseFinishTask{
			Version:             releaseCfg.Version.String(),
			TargetBranch:        releaseCfg.TargetBranch,
			TargetBranchRefName: ref("heads", releaseCfg.TargetBranch),
			Tag:                 releaseCfg.Tag,
			TagRefName:          tagRefName,
			TagRefSpec:          refSpec(tagRefName),
		}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}