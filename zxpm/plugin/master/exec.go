package master

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/zostay/dev-tools/zxpm/format"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Error []error

func (e Error) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

type InterfaceExecutor struct {
	m *Interface
}

func NewExecutor(m *Interface) *InterfaceExecutor {
	return &InterfaceExecutor{m}
}

func (e *InterfaceExecutor) SetTargetName(name string) {
	e.m.SetTargetName(name)
}

func (e *InterfaceExecutor) Define(values map[string]string) {
	e.m.Define(values)
}

func (e *InterfaceExecutor) tryCancel(
	ctx context.Context,
	taskName string,
	task plugin.Task,
	stage string,
) {
	logger := hclog.FromContext(ctx)
	cancelErr := e.m.Cancel(ctx, task)
	if cancelErr != nil {
		logger.Error("failed while canceling task due to error",
			"stage", stage,
			"task", taskName,
			"error", cancelErr)
	}
}

func (e *InterfaceExecutor) logFail(
	ctx context.Context,
	taskName string,
	stage string,
	err error,
) {
	logger := hclog.FromContext(ctx)
	logger.Error("task failed", "stage", stage, "task", taskName, "error", err)
}

func (e *InterfaceExecutor) prepare(
	ctx context.Context,
	taskName string,
) (plugin.Task, error) {
	task, err := e.m.Prepare(ctx, taskName)
	if err != nil {
		if task != nil {
			e.tryCancel(ctx, taskName, task, "Prepare")
		}
		e.logFail(ctx, taskName, "Prepare", err)
		return nil, err
	}
	return task, nil
}

func (e *InterfaceExecutor) taskOperation(
	ctx context.Context,
	taskName string,
	stage string,
	task plugin.Task,
	op func(context.Context) error,
) error {
	err := op(ctx)
	if err != nil {
		e.tryCancel(ctx, taskName, task, stage)
		e.logFail(ctx, taskName, stage, err)
		return err
	}
	return nil
}

func (e *InterfaceExecutor) setup(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Setup", task, task.Setup)
}

func (e *InterfaceExecutor) check(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Check", task, task.Check)
}

func (e *InterfaceExecutor) taskPriorityOperation(
	ctx context.Context,
	taskName string,
	stage string,
	task plugin.Task,
	prepare func(context.Context) (plugin.Operations, error),
) error {
	ops, err := prepare(ctx)
	if err != nil {
		e.tryCancel(ctx, taskName, task, stage)
		e.logFail(ctx, taskName, stage, err)
		return err
	}

	sort.Slice(ops, plugin.OperationLess(ops))
	for _, op := range ops {
		err := op.Action.Call(ctx)
		if err != nil {
			priStage := fmt.Sprintf("%s:%02d", stage, op.Order)
			e.tryCancel(ctx, taskName, task, priStage)
			e.logFail(ctx, taskName, priStage, err)
			return err
		}
	}

	return nil
}

func (e *InterfaceExecutor) begin(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "Begin", task, task.Begin)
}

func (e *InterfaceExecutor) run(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "Run", task, task.Run)
}

func (e *InterfaceExecutor) end(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "End", task, task.End)
}

func (e *InterfaceExecutor) finish(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Finish", task, task.Finish)
}

func (e *InterfaceExecutor) teardown(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Teardown", task, task.Teardown)
}

func (e *InterfaceExecutor) complete(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	err := e.m.Complete(ctx, task)
	if err != nil {
		logger := hclog.FromContext(ctx)
		logger.Error("failed while completing task due to error",
			"stage", "Complete",
			"task", taskName,
			"error", err)
	}
	return err
}

func (e *InterfaceExecutor) ExecuteAllStages(
	ctx context.Context,
	group *TaskGroup,
) error {
	stages, err := group.ExecutionGroups()
	if err != nil {
		return fmt.Errorf("failed to break down goal %q into stages: %v", group.Goal.Name(), err)
	}

	for _, stage := range stages {
		err := e.ExecuteStage(ctx, stage)
		if err != nil {
			stageNames := make([]string, len(stage))
			for i, task := range stage {
				stageNames[i] = task.Name()
			}
			return fmt.Errorf("failed to execute stage (%s): %v", format.And(stageNames...), err)
		}
	}

	return nil
}

func (e *InterfaceExecutor) ExecuteStage(
	ctx context.Context,
	stage []plugin.TaskDescription,
) error {
	return RunTasksAndAccumulateErrors[int, plugin.TaskDescription](ctx,
		NewSliceIterator[plugin.TaskDescription](stage),
		func(ctx context.Context, _ int, task plugin.TaskDescription) error {
			return e.Execute(ctx, task.Name())
		},
	)
}

func (e *InterfaceExecutor) Execute(
	ctx context.Context,
	taskName string,
) error {
	task, err := e.prepare(ctx, taskName)
	if err != nil {
		return fmt.Errorf("failed to prepare task %q: %v", taskName, err)
	}

	stdOps := []struct {
		name     string
		function func(context.Context, string, plugin.Task) error
	}{
		{"setup", e.setup},
		{"check", e.check},
		{"begin", e.begin},
		{"run", e.run},
		{"end", e.end},
		{"finish", e.finish},
		{"teardown", e.teardown},
		{"complete", e.complete},
	}

	for _, stdOp := range stdOps {
		err = stdOp.function(ctx, taskName, task)
		if err != nil {
			return fmt.Errorf("failed to execute operation %s: %v", stdOp.name, err)
		}
	}

	return nil
}

func (e *InterfaceExecutor) TaskGroups(
	ctx context.Context,
) ([]*TaskGroup, error) {
	tasks, err := e.m.Implements(ctx)
	if err != nil {
		return nil, err
	}

	taskGroupMap := make(map[string]*TaskGroup, len(tasks))

	for _, task := range tasks {
		goalName, err := plugin.GoalName(task)
		if err != nil {
			return nil, err
		}

		if group, groupExists := taskGroupMap[goalName]; groupExists {
			group.Tasks = append(group.Tasks, task)
			continue
		}

		goal, err := e.m.Goal(ctx, goalName)
		if err != nil {
			return nil, err
		}

		taskGroupMap[goalName] = &TaskGroup{
			Tree:  "/" + goal.Name(),
			Goal:  goal,
			Tasks: []plugin.TaskDescription{task},
		}
	}

	out := make([]*TaskGroup, 0, len(tasks))
	for _, group := range taskGroupMap {
		out = append(out, group)
	}

	for _, group := range out {
		sort.Slice(group.Tasks, func(i, j int) bool {
			return group.Tasks[i].Name() < group.Tasks[j].Name()
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Goal.Name() < out[j].Goal.Name()
	})

	return out, nil
}
