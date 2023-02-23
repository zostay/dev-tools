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
	case goals.goalBuild:
		return goals.DescribeBuild(), nil
	case goals.goalDeploy:
		return goals.DescribeDeploy(), nil
	case goals.goalGenerate:
		return goals.DescribeGenerate(), nil
	case goals.goalInfo:
		return goals.DescribeInfo(), nil
	case goals.goalInit:
		return goals.DescribeInit(), nil
	case goals.goalInstall:
		return goals.DescribeInstall(), nil
	case goals.goalLint:
		return goals.DescribeLint(), nil
	case goals.goalRelease:
		return goals.DescribeRelease(), nil
	case goals.goalRequest:
		return goals.DescribeRequest(), nil
	case goals.goalTest:
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
