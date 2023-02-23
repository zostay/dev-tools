package githubImpl

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
	plugin_goals "github.com/zostay/dev-tools/zxpm/plugin-goals"
	"github.com/zostay/dev-tools/zxpm/release"
)

var githubPlugin = plugin.ConfigName(Plugin{})

var _ plugin.Interface = &Plugin{}

type Plugin struct{}

func (p *Plugin) Implements(context.Context) ([]plugin.TaskDescription, error) {
	release := plugin_goals.DescribeRelease()
	return []plugin.TaskDescription{
		release.Task("mint/github", "Create a Github pull request."),
		release.Task("publish/github", "Publish a release.",
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
	case release.StartTask:
		return &ReleaseMintTask{}, nil
	case release.FinishTask:
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
