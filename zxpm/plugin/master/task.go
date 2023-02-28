package master

import (
	"context"
	"fmt"
	"sort"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.Task = &Task{}

type taskInfo struct {
	pluginName string
	iface      plugin.Interface
	task       plugin.Task
}

func newTaskInfo(
	pluginName string,
	pluginIface plugin.Interface,
	pluginTask plugin.Task,
) *taskInfo {
	return &taskInfo{pluginName, pluginIface, pluginTask}
}

type Task struct {
	taskName string
	ti       *Interface
	ts       plugin.Tasks
	taskInfo []*taskInfo
}

func newTask(taskName string, ti *Interface, taskInfo []*taskInfo) *Task {
	return &Task{
		taskName: taskName,
		ti:       ti,
		taskInfo: taskInfo,
	}
}

func (t *Task) tasks() plugin.Tasks {
	if t.ts != nil {
		return t.ts
	}

	t.ts = make(plugin.Tasks, len(t.taskInfo))
	for i := range t.taskInfo {
		t.ts[i] = t.taskInfo[i].task
	}
	return t.ts
}

func (t *Task) Setup(ctx context.Context) error {
	return t.executeTaskOperation(ctx, t.taskName, executeSetup)
}

func (t *Task) Check(ctx context.Context) error {
	return t.executeTaskOperation(ctx, t.taskName, executeCheck)
}

func (t *Task) Begin(ctx context.Context) (plugin.Operations, error) {
	return t.prepareOperations(ctx, plugin.Task.Begin)
}

func (t *Task) Run(ctx context.Context) (plugin.Operations, error) {
	return t.prepareOperations(ctx, plugin.Task.Run)
}

func (t *Task) End(ctx context.Context) (plugin.Operations, error) {
	return t.prepareOperations(ctx, plugin.Task.End)
}

func (t *Task) Finish(ctx context.Context) error {
	return t.executeTaskOperation(ctx, t.taskName, executeFinish)
}

func (t *Task) Teardown(ctx context.Context) error {
	return t.executeTaskOperation(ctx, t.taskName, executeTeardown)
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
	executeSetup    = executeBasicStage(plugin.Task.Setup, "setup")
	executeCheck    = executeBasicStage(plugin.Task.Check, "check")
	executeFinish   = executeBasicStage(plugin.Task.Finish, "finish")
	executeTeardown = executeBasicStage(plugin.Task.Teardown, "teardown")
)

func (t *Task) executeTaskOperation(
	ctx context.Context,
	taskName string,
	op func(context.Context, plugin.Task) error,
) error {
	opfs := make([]plugin.OperationFunc, 0, len(t.taskInfo))
	for i := range t.taskInfo {
		info := t.taskInfo[i]
		opfs = append(opfs, func(ctx context.Context) error {
			ctx = t.ti.ctxFor(ctx, taskName, info.pluginName)
			return op(ctx, info.task)
		})
	}

	return executeOperationFuncs(ctx, opfs)
}

func executeOperationFuncs(
	ctx context.Context,
	opfs []plugin.OperationFunc,
) error {
	return RunTasksAndAccumulateErrors[int, plugin.OperationFunc](
		ctx,
		NewSliceIterator[plugin.OperationFunc](opfs),
		func(ctx context.Context, _ int, opf plugin.OperationFunc) error {
			return opf.Call(ctx)
		})
}

func (t *Task) evaluateOperations(
	ctx context.Context,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) ([]*operationInfo, error) {
	opInfo := make([]*operationInfo, 0, len(t.taskInfo))
	for _, tInfo := range t.taskInfo {
		ctx = t.ti.ctxFor(ctx, t.taskName, tInfo.pluginName)
		theseOps, err := op(tInfo.task, ctx)
		if err != nil {
			return nil, err
		}
		for _, thisOp := range theseOps {
			info := newOperationInfo(tInfo.pluginName, thisOp)
			opInfo = append(opInfo, info)
		}
	}

	sort.Slice(opInfo, operationInfoLess(opInfo))

	return opInfo, nil
}

func (t *Task) gatherOperations(
	opInfo []*operationInfo,
) plugin.Operations {
	ophs := make(plugin.Operations, 0, len(opInfo))
	var lastOrder plugin.Ordering = -1
	var curOp *OperationHandler
	for _, info := range opInfo {
		order := info.op.Order
		if order < 0 {
			order = 0
		} else if order > 100 {
			order = 100
		}

		if order > lastOrder {
			curOp = newOperationHandler(
				t.taskName,
				t.ti,
				make([]*operationInfo, 0, len(opInfo)),
			)
			ophs = append(ophs, plugin.Operation{
				Order:  order,
				Action: curOp,
			})
		}

		curOp.opInfo = append(curOp.opInfo, info)

		lastOrder = order
	}

	return ophs
}

func (t *Task) prepareOperations(
	ctx context.Context,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) (plugin.Operations, error) {
	opInfo, err := t.evaluateOperations(ctx, op)
	if err != nil {
		return nil, err
	}

	return t.gatherOperations(opInfo), nil
}
