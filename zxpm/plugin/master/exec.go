package master

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/zostay/dev-tools/pkg/config"
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

type TaskInterfaceExecutor struct {
	iface plugin.Interface
}

func NewExecutor(iface plugin.Interface) *TaskInterfaceExecutor {
	return &TaskInterfaceExecutor{iface}
}

func (e *TaskInterfaceExecutor) tryCancel(
	ctx context.Context,
	taskName string,
	task plugin.Task,
	stage string,
) {
	logger := hclog.FromContext(ctx)
	cancelErr := e.iface.Cancel(ctx, task)
	if cancelErr != nil {
		logger.Error("failed while canceling task due to error",
			"stage", stage,
			"task", taskName,
			"error", cancelErr)
	}
}

func (e *TaskInterfaceExecutor) logFail(
	ctx context.Context,
	taskName string,
	stage string,
	err error,
) {
	logger := hclog.FromContext(ctx)
	logger.Error("task failed", "stage", stage, "task", taskName, "error", err)
}

func (e *TaskInterfaceExecutor) prepare(
	ctx context.Context,
	cfg *config.Config,
	taskName string,
) (plugin.Task, error) {
	task, err := e.iface.Prepare(ctx, taskName, cfg)
	if err != nil {
		if task != nil {
			e.tryCancel(ctx, taskName, task, "Prepare")
		}
		e.logFail(ctx, taskName, "Prepare", err)
		return nil, err
	}
	return task, nil
}

func (e *TaskInterfaceExecutor) taskOperation(
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

func (e *TaskInterfaceExecutor) setup(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Setup", task, task.Setup)
}

func (e *TaskInterfaceExecutor) check(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Check", task, task.Check)
}

func (e *TaskInterfaceExecutor) taskPriorityOperation(
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

func (e *TaskInterfaceExecutor) begin(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "Begin", task, task.Begin)
}

func (e *TaskInterfaceExecutor) run(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "Run", task, task.Run)
}

func (e *TaskInterfaceExecutor) end(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskPriorityOperation(ctx, taskName, "End", task, task.End)
}

func (e *TaskInterfaceExecutor) finish(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Finish", task, task.Finish)
}

func (e *TaskInterfaceExecutor) teardown(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	return e.taskOperation(ctx, taskName, "Teardown", task, task.Teardown)
}

func (e *TaskInterfaceExecutor) complete(
	ctx context.Context,
	taskName string,
	task plugin.Task,
) error {
	err := e.iface.Complete(ctx, task)
	if err != nil {
		logger := hclog.FromContext(ctx)
		logger.Error("failed while completing task due to error",
			"stage", "Complete",
			"task", taskName,
			"error", err)
	}
	return err
}

func (e *TaskInterfaceExecutor) Execute(
	ctx context.Context,
	cfg *config.Config,
	taskName string,
) error {
	pctx := plugin.NewContext(cfg)
	ctx = plugin.InitializeContext(ctx, pctx)

	task, err := e.prepare(ctx, cfg, taskName)
	if err != nil {
		return err
	}

	stdOps := []func(context.Context, string, plugin.Task) error{
		e.setup,
		e.check,
		e.begin,
		e.run,
		e.end,
		e.finish,
		e.teardown,
		e.complete,
	}

	for _, stdOp := range stdOps {
		err = stdOp(ctx, taskName, task)
		if err != nil {
			return err
		}
	}

	return nil
}
