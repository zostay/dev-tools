package master

import (
	"context"
	"sort"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.TaskInterface = &TaskInterface{}

type TaskInterface struct {
	is []plugin.TaskInterface
}

func New(is []plugin.TaskInterface) *TaskInterface {
	return &TaskInterface{is}
}

func (ti *TaskInterface) Implements(ctx context.Context) ([]string, error) {
	names := make(map[string]struct{}, 20)
	for _, iface := range ti.is {
		ins, err := iface.Implements(ctx)
		if err != nil {
			return nil, err
		}
		for _, name := range ins {
			names[name] = struct{}{}
		}
	}

	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}

	sort.Strings(out)

	return out, nil
}

func implements(
	ctx context.Context,
	iface plugin.TaskInterface,
	taskName string,
) (bool, error) {
	names, err := iface.Implements(ctx)
	if err != nil {
		return false, err
	}

	for _, name := range names {
		if name == taskName {
			return true, nil
		}
	}
	return false, nil
}

func (ti *TaskInterface) Prepare(
	ctx context.Context,
	taskName string,
	globalCfg *config.Config,
) (plugin.Task, error) {
	results, err := RunTasksAndAccumulate[plugin.TaskInterface, *taskPair](
		ctx,
		ti.is,
		func(ctx context.Context, iface plugin.TaskInterface) (*taskPair, error) {
			mayPrepare, err := implements(ctx, iface, taskName)
			if err != nil {
				return nil, err
			}

			if mayPrepare {
				t, err := iface.Prepare(ctx, taskName, globalCfg)
				if err != nil {
					if t != nil {
						return &taskPair{iface, t}, err
					}
					return nil, err
				}
				return &taskPair{iface, t}, nil
			}

			return nil, nil
		},
	)

	if len(results) > 0 {
		return &Task{taskPairs: results}, err
	}

	if err != nil {
		return nil, err
	}

	return nil, plugin.ErrUnsupportedTask
}

func (ti *TaskInterface) Cancel(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[*taskPair](ctx, task.taskPairs,
		func(ctx context.Context, p *taskPair) error {
			return p.iface.Cancel(ctx, p.task)
		})
}

func (ti *TaskInterface) Complete(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[*taskPair](ctx, task.taskPairs,
		func(ctx context.Context, p *taskPair) error {
			return p.iface.Complete(ctx, p.task)
		})
}
