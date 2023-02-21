package plugin_changelog

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin-changelog/cmd"
	plugin_goals "github.com/zostay/dev-tools/zxpm/plugin-goals"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Plugin struct{}

var _ plugin.TaskInterface = &Plugin{}

func (p *Plugin) Goal(context.Context, string) (plugin.GoalDescription, error) {
	return nil, plugin.ErrUnsupportedGoal
}

func (p *Plugin) Implements() ([]plugin.TaskDescription, error) {
	release := plugin_goals.DescribeRelease()
	return []plugin.TaskDescription{
		release.Task("mint", "Check and prepare changelog for release."),
		release.Task("publish", "Capture changelog data to prepare for release.",
			release.TaskName("mint")),
	}, nil
}

func (p *Plugin) Prepare(
	ctx context.Context,
	task string,
	cfg *config.Config,
) (plugin.Task, error) {
	pluginCfg := cfg.Section("github.com/zostay/zxpm/plugin-changelog")

	switch task {
	case "/release/mint":
		return &ReleaseStartTask{
			Version:   plugin.GetString(ctx, "release.version"),
			Today:     plugin.GetString(ctx, "release.date"),
			Changelog: pluginCfg.Get("changelog"),
		}, nil
	case "/release/publish":
		return &ReleaseFinishTask{
			Version:   plugin.GetString(ctx, "release.version"),
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
