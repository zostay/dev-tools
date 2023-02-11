package git

import (
	"path"
	"strings"

	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/release"
)

type Plugin struct{}

func ref(t, n string) plumbing.ReferenceName {
	return plumbing.ReferenceName(path.Join("refs", t, n))
}

func refSpec(r plumbing.ReferenceName) gitConfig.RefSpec {
	return gitConfig.RefSpec(strings.Join([]string{r, r}, ":"))
}

func (p *Plugin) Implements() []string {
	return []string{release.StartTask}
}

func (p *Plugin) Prepare(
	task string,
	cfg *config.Config,
	taskConfig any,
) plugin.Task {
	switch task {
	case release.StartTask:
		releaseCfg := taskConfig.(*release.Config)
		branchRefName := ref("heads", releaseCfg.Branch)
		return &ReleaseStartTask{
			Version:             releaseCfg.Version.String(),
			BranchRefSpec:       refSpec(branchRefName),
			Branch:              releaseCfg.Branch,
			BranchRefName:       ref("heads", releaseCfg.Branch),
			TargetBranch:        releaseCfg.TargetBranch,
			TargetBranchRefName: ref("heads", releaseCfg.TargetBranch),
		}
	}
	return nil
}
