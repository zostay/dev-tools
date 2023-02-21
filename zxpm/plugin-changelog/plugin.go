package plugin_changelog

import (
	"context"

	plugin_goals "github.com/zostay/dev-tools/zxpm/plugin-goals"
	"github.com/zostay/dev-tools/zxpm/storage"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Plugin struct{}

var _ plugin.TaskInterface = &Plugin{}

func (p *Plugin) Goal(context.Context, string) (plugin.GoalDescription, error) {
	return nil, plugin.ErrUnsupportedGoal
}

func (p *Plugin) Implements(context.Context) ([]plugin.TaskDescription, error) {
	lint := plugin_goals.DescribeLint()
	release := plugin_goals.DescribeRelease()
	return []plugin.TaskDescription{
		lint.Task("changelog", "Check changelog for correctness."),
		lint.Task("changelog", "Extract the changes for a version."),
		release.Task("mint/changelog", "Check and prepare changelog for release."),
		release.Task("publish/changelog", "Capture changelog data to prepare for release.",
			release.TaskName("mint")),
	}, nil
}

func (p *Plugin) Prepare(
	ctx context.Context,
	task string,
	cfg storage.KV,
) (plugin.Task, error) {
	pluginCfg := cfg.Sub("github.com/zostay/zxpm/plugin-changelog")

	switch task {
	case "/lint/changelog":
		return &LintChangelogTask{
			Changelog: pluginCfg.GetString("changelog"),
		}, nil
	case "/info/extract-changelog":
		return &InfoChangelogTask{
			Changelog: pluginCfg.GetString("changelog"),
		}, nil
	case "/release/mint/changelog":
		return &ReleaseStartTask{
			Changelog: pluginCfg.GetString("changelog"),
		}, nil
	case "/release/publish/changelog":
		return &ReleaseFinishTask{
			Changelog: pluginCfg.GetString("changelog"),
		}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Cancel(ctx context.Context, task plugin.Task) error {
	return nil
}

func (p *Plugin) Complete(ctx context.Context, task plugin.Task) error {
	return nil
}
