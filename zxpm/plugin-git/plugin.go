package plugin_git

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
	plugin_goals "github.com/zostay/dev-tools/zxpm/plugin-goals"
)

var gitPlugin = plugin.ConfigName(Plugin{})

var _ plugin.Interface = &Plugin{}

type Plugin struct{}

func (p *Plugin) Implements(context.Context) ([]plugin.TaskDescription, error) {
	release := plugin_goals.DescribeRelease()
	return []plugin.TaskDescription{
		release.Task("mint/git", "Verify work directory is clean and push a release branch."),
		release.Task("publish/git", "Push a release tag.",
			release.TaskName("mint")),
	}, nil
}

func (p *Plugin) Goal(context.Context, string) (plugin.GoalDescription, error) {
	return nil, plugin.ErrUnsupportedGoal
}

func (p *Plugin) Prepare(
	ctx context.Context,
	task string,
) (plugin.Task, error) {
	switch task {
	case "/release/mint/git":
		return &ReleaseMintTask{}, nil
	case "/release/publish/git":
		return &ReleasePublishTask{}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Cancel(ctx context.Context, task plugin.Task) error {
	return nil
}

func (p *Plugin) Complete(ctx context.Context, task plugin.Task) error {
	return nil
}
