package master

import (
	"context"
	"errors"
	"fmt"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/storage"
)

var _ plugin.Interface = &Interface{}

type Interface struct {
	cfg        *config.Config
	is         map[string]plugin.Interface
	targetName string
	properties *storage.KVMem
}

func NewInterface(
	cfg *config.Config,
	is map[string]plugin.Interface,
) *Interface {
	return &Interface{cfg, is, "", storage.New()}
}

func (ti *Interface) GetInterface(name string) plugin.Interface {
	return ti.is[name]
}

func (ti *Interface) SetTargetName(name string) {
	ti.targetName = name
}

func (ti *Interface) Define(values map[string]string) {
	vals := make(map[string]any, len(values))
	for k, v := range values {
		vals[k] = v
	}
	ti.properties.Update(vals)
}

func (ti *Interface) ctxFor(
	ctx context.Context,
	taskName string,
	pluginName string,
) (context.Context, *plugin.Context) {
	pctx := plugin.NewConfigContext(ti.properties, taskName, ti.targetName, pluginName, ti.cfg)
	return plugin.InitializeContext(ctx, pctx), pctx
}

func (ti *Interface) Implements(ctx context.Context) ([]plugin.TaskDescription, error) {
	taskDescs := make([]plugin.TaskDescription, 0, 100)
	for pluginName, iface := range ti.is {
		ctx, _ := ti.ctxFor(ctx, "", pluginName)
		tds, err := iface.Implements(ctx)
		if err != nil {
			return nil, err
		}
		taskDescs = append(taskDescs, tds...)
	}

	return taskDescs, nil
}

func (ti *Interface) implements(
	ctx context.Context,
	iface plugin.Interface,
	taskName string,
) (bool, error) {
	taskDescs, err := iface.Implements(ctx)
	if err != nil {
		return false, err
	}

	for _, taskDesc := range taskDescs {
		if taskDesc.Name() == taskName {
			return true, nil
		}
	}
	return false, nil
}

func (ti *Interface) Goal(
	ctx context.Context,
	name string,
) (plugin.GoalDescription, error) {
	results, err := RunTasksAndAccumulate[string, plugin.Interface, plugin.GoalDescription](
		ctx,
		NewMapIterator[string, plugin.Interface](ti.is),
		func(ctx context.Context, _ string, p plugin.Interface) (plugin.GoalDescription, error) {
			goalDesc, err := p.Goal(ctx, name)
			if errors.Is(err, plugin.ErrUnsupportedGoal) {
				return nil, nil
			} else if err != nil {
				return nil, err
			}
			return goalDesc, nil
		},
	)

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		if result != nil {
			return result, nil
		}
	}

	return nil, plugin.ErrUnsupportedGoal
}

func (ti *Interface) Prepare(
	ctx context.Context,
	taskName string,
) (plugin.Task, error) {
	results, err := RunTasksAndAccumulate[string, plugin.Interface, *taskInfo](
		ctx,
		NewMapIterator[string, plugin.Interface](ti.is),
		func(ctx context.Context, pluginName string, iface plugin.Interface) (*taskInfo, error) {
			ctx, _ = ti.ctxFor(ctx, taskName, pluginName)
			mayPrepare, err := ti.implements(ctx, iface, taskName)
			if err != nil {
				return nil, fmt.Errorf("plugin %q failed implements check for task %q: %w", pluginName, taskName, err)
			}

			if mayPrepare {
				t, err := iface.Prepare(ctx, taskName)
				if err != nil {
					if t != nil {
						return newTaskInfo(pluginName, iface, t), err
					}
					return nil, fmt.Errorf("plugin %q failed to run task %q: %w", pluginName, taskName, err)
				}
				return newTaskInfo(pluginName, iface, t), nil
			}

			return nil, nil
		},
	)

	filteredResults := make([]*taskInfo, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		filteredResults = append(filteredResults, result)
	}
	results = filteredResults

	if len(results) > 0 {
		return newTask(taskName, ti, results), err
	}

	if err != nil {
		return nil, err
	}

	return nil, plugin.ErrUnsupportedTask
}

func (ti *Interface) Cancel(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[int, *taskInfo](
		ctx,
		NewSliceIterator[*taskInfo](task.taskInfo),
		func(ctx context.Context, _ int, p *taskInfo) error {
			return p.iface.Cancel(ctx, p.task)
		})
}

func (ti *Interface) Complete(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[int, *taskInfo](
		ctx,
		NewSliceIterator[*taskInfo](task.taskInfo),
		func(ctx context.Context, _ int, p *taskInfo) error {
			return p.iface.Complete(ctx, p.task)
		})
}
