package goalsImpl

import (
	"context"
	"os"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-goals/pkg/goals"
	"github.com/zostay/dev-tools/zxpm/storage"
)

var _ plugin.Interface = &Plugin{}

type Plugin struct{}

type InfoDisplayTask struct {
	plugin.TaskBoilerplate
}

func (p *Plugin) Implements(context.Context) ([]plugin.TaskDescription, error) {
	info := goals.DescribeInfo()
	return []plugin.TaskDescription{
		info.Task("display", "Display information."),
	}, nil
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
	_ context.Context,
	taskName string,
) (plugin.Task, error) {
	switch taskName {
	case "/info/display":
		return &InfoDisplayTask{}, nil
	}
	return nil, plugin.ErrUnsupportedTask
}

func (p *Plugin) Cancel(context.Context, plugin.Task) error {
	return nil
}

func (p *Plugin) Complete(ctx context.Context, task plugin.Task) error {
	values := plugin.KV(ctx)
	outputAll := InfoOutputAll(ctx)
	if !outputAll {
		values = storage.ExportsOnly(values)
	}
	formatter := InfoOutputFormatter(ctx)
	return formatter(os.Stdout, values)
}
