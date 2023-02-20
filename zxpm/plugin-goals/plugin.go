package plugin_goals

import (
	"context"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.TaskInterface = &Plugin{}

type Plugin struct{}

func (p *Plugin) Implements(context.Context) ([]plugin.TaskDescription, error) {
	return nil, nil
}

func (p *Plugin) Goal(
	_ context.Context,
	name string,
) (plugin.GoalDescription, error) {
	switch name {
	case goalBuild:
		return describeBuild(), nil
	case goalDeploy:
		return describeDeploy(), nil
	case goalGenerate:
		return describeGenerate(), nil
	case goalInstall:
		return describeInstall(), nil
	case goalLint:
		return describeLint(), nil
	case goalRequest:
		return describeRequest(), nil
	case goalTest:
		return describeTest(), nil
	default:
		return nil, plugin.ErrUnsupportedGoal
	}
}

func (p *Plugin) Prepare(
	context.Context,
	string,
	*config.Config,
) (plugin.Task, error) {
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Cancel(context.Context, plugin.Task) error {
	return plugin.ErrUnsupportedTask
}

func (p *Plugin) Complete(context.Context, plugin.Task) error {
	return plugin.ErrUnsupportedTask
}
