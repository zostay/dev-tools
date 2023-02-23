package goalsImpl

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-goals/pkg/goals"
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
	case goals.NameBuild:
		return goals.DescribeBuild(), nil
	case goals.NameDeploy:
		return goals.DescribeDeploy(), nil
	case goals.NameGenerate:
		return goals.DescribeGenerate(), nil
	case goals.NameInfo:
		return goals.DescribeInfo(), nil
	case goals.NameInit:
		return goals.DescribeInit(), nil
	case goals.NameInstall:
		return goals.DescribeInstall(), nil
	case goals.NameLint:
		return goals.DescribeLint(), nil
	case goals.NameRelease:
		return goals.DescribeRelease(), nil
	case goals.NameRequest:
		return goals.DescribeRequest(), nil
	case goals.NameTest:
		return goals.DescribeTest(), nil
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
