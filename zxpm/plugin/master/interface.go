package master

import (
	"context"
	"errors"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.Interface = &Interface{}

type Interface struct {
	is map[string]plugin.Interface
}

func New(is map[string]plugin.Interface) *Interface {
	return &Interface{is}
}

func (ti *Interface) GetInterface(name string) plugin.Interface {
	return ti.is[name]
}

func (ti *Interface) Implements(ctx context.Context) ([]plugin.TaskDescription, error) {
	taskDescs := make([]plugin.TaskDescription, 0, 100)
	for _, iface := range ti.is {
		tds, err := iface.Implements(ctx)
		if err != nil {
			return nil, err
		}
		taskDescs = append(taskDescs, tds...)
	}

	return taskDescs, nil
}

func implements(
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
		func(ctx context.Context, p plugin.Interface) (plugin.GoalDescription, error) {
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
	results, err := RunTasksAndAccumulate[string, plugin.Interface, *taskPair](
		ctx,
		NewMapIterator[string, plugin.Interface](ti.is),
		func(ctx context.Context, iface plugin.Interface) (*taskPair, error) {
			mayPrepare, err := implements(ctx, iface, taskName)
			if err != nil {
				return nil, err
			}

			if mayPrepare {
				t, err := iface.Prepare(ctx, taskName)
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

	filteredResults := make([]*taskPair, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		filteredResults = append(filteredResults, result)
	}
	results = filteredResults

	if len(results) > 0 {
		return &Task{taskPairs: results}, err
	}

	if err != nil {
		return nil, err
	}

	return nil, plugin.ErrUnsupportedTask
}

func (ti *Interface) Cancel(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[int, *taskPair](
		ctx,
		NewSliceIterator[*taskPair](task.taskPairs),
		func(ctx context.Context, p *taskPair) error {
			return p.iface.Cancel(ctx, p.task)
		})
}

func (ti *Interface) Complete(ctx context.Context, pluginTask plugin.Task) error {
	task := pluginTask.(*Task)
	return RunTasksAndAccumulateErrors[int, *taskPair](
		ctx,
		NewSliceIterator[*taskPair](task.taskPairs),
		func(ctx context.Context, p *taskPair) error {
			return p.iface.Complete(ctx, p.task)
		})
}
