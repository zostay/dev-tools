package plugin_goals

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.Interface = &Plugin{}

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
		return DescribeBuild(), nil
	case goalDeploy:
		return DescribeDeploy(), nil
	case goalGenerate:
		return DescribeGenerate(), nil
	case goalInfo:
		return DescribeInfo(), nil
	case goalInit:
		return DescribeInit(), nil
	case goalInstall:
		return DescribeInstall(), nil
	case goalLint:
		return DescribeLint(), nil
	case goalRelease:
		return DescribeRelease(), nil
	case goalRequest:
		return DescribeRequest(), nil
	case goalTest:
		return DescribeTest(), nil
	default:
		return nil, plugin.ErrUnsupportedGoal
	}
}

func (p *Plugin) Prepare(
	context.Context,
	string,
) (plugin.Task, error) {
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Cancel(context.Context, plugin.Task) error {
	return plugin.ErrUnsupportedTask
}

func (p *Plugin) Complete(context.Context, plugin.Task) error {
	return plugin.ErrUnsupportedTask
}
