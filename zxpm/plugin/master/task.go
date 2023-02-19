package master

import (
	"context"
	"fmt"
	"sort"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.Task = &Task{}

type taskPair struct {
	iface plugin.TaskInterface
	task  plugin.Task
}

type Task struct {
	ts        plugin.Tasks
	taskPairs []*taskPair
}

func (t *Task) tasks() plugin.Tasks {
	if t.ts != nil {
		return t.ts
	}

	t.ts = make(plugin.Tasks, len(t.taskPairs))
	for i := range t.taskPairs {
		t.ts[i] = t.taskPairs[i].task
	}
	return t.ts
}

func (t *Task) Setup(ctx context.Context) error {
	return executeTaskOperation(ctx, t.tasks(), executeSetup)
}

func (t *Task) Check(ctx context.Context) error {
	return executeTaskOperation(ctx, t.tasks(), executeCheck)
}

func (t *Task) Begin(ctx context.Context) (plugin.Operations, error) {
	return prepareOperations(ctx, t.tasks(), plugin.Task.Begin)
}

func (t *Task) Run(ctx context.Context) (plugin.Operations, error) {
	return prepareOperations(ctx, t.tasks(), plugin.Task.Run)
}

func (t *Task) End(ctx context.Context) (plugin.Operations, error) {
	return prepareOperations(ctx, t.tasks(), plugin.Task.End)
}

func (t *Task) Finishing(ctx context.Context) error {
	return executeTaskOperation(ctx, t.tasks(), executeFinishing)
}

func (t *Task) Teardown(ctx context.Context) error {
	return executeTaskOperation(ctx, t.tasks(), executeTeardown)
}

type taskOperationFunc func(plugin.Task, context.Context) error

func executeBasicStage(
	opFunc taskOperationFunc,
	stage string,
) func(context.Context, plugin.Task) error {
	return func(ctx context.Context, t plugin.Task) error {
		err := opFunc(t, ctx)
		if err != nil {
			return fmt.Errorf("failed %s stage: %w", stage, err)
		}
		return nil
	}
}

var (
	executeSetup     = executeBasicStage(plugin.Task.Setup, "setup")
	executeCheck     = executeBasicStage(plugin.Task.Check, "check")
	executeFinishing = executeBasicStage(plugin.Task.Finishing, "finishing")
	executeTeardown  = executeBasicStage(plugin.Task.Teardown, "teardown")
)

func executeTaskOperation(
	ctx context.Context,
	ts plugin.Tasks,
	op func(context.Context, plugin.Task) error,
) error {
	opfs := make([]plugin.OperationFunc, len(ts))
	for i := range ts {
		t := ts[i]
		opfs = append(opfs, func(ctx context.Context) error {
			return op(ctx, t)
		})
	}

	return executeOperationFuncs(ctx, opfs)
}

func executeOperationFuncs(
	ctx context.Context,
	opfs []plugin.OperationFunc,
) error {
	return RunTasksAndAccumulateErrors[plugin.OperationFunc](ctx, opfs,
		func(ctx context.Context, opf plugin.OperationFunc) error {
			return opf.Call(ctx)
		})
}

func evaluateOperations(
	ctx context.Context,
	ts plugin.Tasks,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) (plugin.Operations, error) {
	ops := make(plugin.Operations, 0, len(ts))
	for _, t := range ts {
		thisOps, err := op(t, ctx)
		if err != nil {
			return nil, err
		}
		ops = append(ops, thisOps...)
	}

	sort.Slice(ops, plugin.OperationLess(ops))

	return ops, nil
}

func gatherOperations(
	ops plugin.Operations,
) []plugin.Operations {
	opss := make([]plugin.Operations, 0, len(ops))
	var lastOrder plugin.Ordering = -1
	for _, op := range ops {
		order := op.Order
		if order < 0 {
			order = 0
		} else if order > 100 {
			order = 100
		}

		if order > lastOrder {
			opss = append(opss, make(plugin.Operations, 0, len(ops)))
		}

		opss[len(opss)-1] = append(opss[len(opss)-1], op)

		lastOrder = order
	}

	return opss
}

func makeOperationHandlers(
	opss []plugin.Operations,
) plugin.Operations {
	handlers := make(plugin.Operations, len(opss))
	for i, ops := range opss {
		handlers[i] = plugin.Operation{
			Order:  ops[0].Order,
			Action: &OperationHandler{ops},
		}
	}
	return handlers
}

func prepareOperations(
	ctx context.Context,
	ts plugin.Tasks,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) (plugin.Operations, error) {
	ops, err := evaluateOperations(ctx, ts, op)
	if err != nil {
		return nil, err
	}

	opss := gatherOperations(ops)

	return makeOperationHandlers(opss), nil
}
